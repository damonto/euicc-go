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
	device  string
	slot    uint8
	reader  reader
	channel uint32
	closed  bool
}

// New creates an MBIM channel whose access method is resolved by Connect.
func New(device string, slot uint8) (driver.SmartCardChannel, error) {
	if slot == 0 {
		return nil, fmt.Errorf("slot must be >= 1")
	}
	return &MBIM{device: device, slot: slot}, nil
}

// NewWithClient creates a channel backed by an already connected MBIM client.
// The channel takes ownership of client and closes it on Disconnect.
func NewWithClient(client *wwanmbim.Client) (driver.SmartCardChannel, error) {
	if client == nil {
		return nil, errors.New("mbim client is nil")
	}
	return &MBIM{reader: client}, nil
}

// Connect establishes the MBIM session and opens the device.
func (m *MBIM) Connect() error {
	if m.closed {
		return errors.New("mbim reader is closed")
	}
	if m.reader != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	reader, err := wwanmbim.Open(ctx, wwanmbim.WithAutoDetect(m.device), wwanmbim.WithSlot(int(m.slot)))
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
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
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
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
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
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
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
