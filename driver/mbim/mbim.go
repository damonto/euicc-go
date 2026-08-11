package mbim

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/damonto/euicc-go/driver"
	wwanmbim "github.com/damonto/wwan-go/mbim"
)

type reader interface {
	OpenChannel(ctx context.Context, aid []byte) (uint32, error)
	TransmitAPDU(ctx context.Context, channel uint32, command []byte) ([]byte, uint32, error)
	CloseChannel(ctx context.Context, channel uint32) error
	Close() error
}

const (
	defaultTimeout    = 30 * time.Second
	maxLogicalChannel = 19
)

// MBIM implements driver.SmartCardChannel over an MBIM connection. It is not
// safe for concurrent use.
type MBIM struct {
	access  accessMode
	device  string
	slot    uint8
	timeout time.Duration
	reader  reader
	channel uint32
	closed  bool
}

var _ driver.SmartCardChannel = (*MBIM)(nil)

// New creates an MBIM channel. Configure its connection using WithAutoDetect,
// WithDirect, WithProxy, or WithClient.
func New(options ...Option) (*MBIM, error) {
	config := applyOptions(options)
	if err := config.validate(); err != nil {
		return nil, err
	}
	if config.client != nil {
		return &MBIM{reader: config.client, timeout: config.timeout}, nil
	}
	return &MBIM{
		access:  config.access,
		device:  config.device,
		slot:    config.slot,
		timeout: config.timeout,
	}, nil
}

// Connect establishes the MBIM session and opens the device.
func (m *MBIM) Connect() error {
	if m.closed {
		return errors.New("mbim reader is closed")
	}
	if m.reader != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()
	options := []wwanmbim.Option{wwanmbim.WithSlot(int(m.slot))}
	switch m.access {
	case accessAutoDetect:
		options = append(options, wwanmbim.WithAutoDetect(m.device))
	case accessDirect:
		options = append(options, wwanmbim.WithDirect(m.device))
	case accessProxy:
		options = append(options, wwanmbim.WithProxy(m.device))
	}
	reader, err := wwanmbim.Open(ctx, options...)
	if err != nil {
		return fmt.Errorf("open MBIM reader: %w", err)
	}
	m.reader = reader
	return nil
}

// OpenLogicalChannel opens a logical channel for the specified Application ID.
func (m *MBIM) OpenLogicalChannel(aid []byte) (byte, error) {
	if err := m.ensureOpen(); err != nil {
		return 0, err
	}
	if m.channel != 0 {
		return 0, fmt.Errorf("MBIM logical channel %d is already open", m.channel)
	}
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()
	channel, err := m.reader.OpenChannel(ctx, aid)
	if err != nil {
		return 0, fmt.Errorf("open MBIM logical channel: %w", err)
	}
	if channel == 0 || channel > maxLogicalChannel {
		var cleanupErr error
		if channel != 0 {
			cleanupErr = m.closeLogicalChannel(channel)
		}
		return 0, errors.Join(
			fmt.Errorf("MBIM returned invalid logical channel %d", channel),
			cleanupErr,
		)
	}
	m.channel = channel
	return byte(channel), nil
}

// Transmit implements driver.SmartCardChannel.
func (m *MBIM) Transmit(command []byte) ([]byte, error) {
	if err := m.ensureOpen(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()
	response, status, err := m.reader.TransmitAPDU(ctx, m.channel, command)
	if err != nil {
		return nil, fmt.Errorf("transmit MBIM APDU: %w", err)
	}
	return append(response, uiccStatusWord(status)...), nil
}

// CloseLogicalChannel closes the specified logical channel.
func (m *MBIM) CloseLogicalChannel(channel byte) error {
	if err := m.ensureOpen(); err != nil {
		return err
	}
	if channel == 0 || channel > maxLogicalChannel {
		return fmt.Errorf("invalid logical channel %d", channel)
	}
	return m.closeLogicalChannel(uint32(channel))
}

func (m *MBIM) closeLogicalChannel(channel uint32) error {
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()
	if err := m.reader.CloseChannel(ctx, channel); err != nil {
		return fmt.Errorf("close MBIM logical channel %d: %w", channel, err)
	}
	if m.channel == channel {
		m.channel = 0
	}
	return nil
}

// Disconnect closes the MBIM connection and releases resources.
func (m *MBIM) Disconnect() error {
	if m.closed {
		return nil
	}
	var channelErr error
	if m.reader != nil && m.channel != 0 {
		channelErr = m.closeLogicalChannel(m.channel)
	}
	m.closed = true
	if m.reader == nil {
		return channelErr
	}
	closeErr := m.reader.Close()
	if closeErr != nil {
		closeErr = fmt.Errorf("close MBIM reader: %w", closeErr)
	}
	return errors.Join(channelErr, closeErr)
}

func (m *MBIM) ensureOpen() error {
	if m.closed {
		return errors.New("mbim reader is closed")
	}
	if m.reader == nil {
		return errors.New("mbim reader is not connected")
	}
	return nil
}

func uiccStatusWord(status uint32) []byte {
	sw := make([]byte, 2)
	binary.LittleEndian.PutUint16(sw, uint16(status&0xffff))
	return sw
}
