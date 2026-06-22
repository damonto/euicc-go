package mbim

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/damonto/euicc-go/driver"
	uiccmbim "github.com/damonto/uicc-go/mbim"
)

type reader interface {
	OpenChannel(ctx context.Context, aid []byte) (uint32, error)
	TransmitAPDU(ctx context.Context, channel uint32, command []byte) ([]byte, uint32, error)
	CloseChannel(ctx context.Context, channel uint32) error
	Close() error
}

const defaultTimeout = 30 * time.Second

type mbimOpener func(context.Context, ...uiccmbim.Option) (reader, error)

var openReader mbimOpener = func(ctx context.Context, opts ...uiccmbim.Option) (reader, error) {
	return uiccmbim.Open(ctx, opts...)
}

// MBIM implements driver.SmartCardChannel over an MBIM proxy connection.
type MBIM struct {
	mu      sync.Mutex
	device  string
	slot    uint8
	reader  reader
	channel uint32
	closed  bool
}

// New creates a new MBIM proxy channel to the specified device.
func New(device string, slot uint8) (driver.SmartCardChannel, error) {
	if slot == 0 {
		return nil, fmt.Errorf("slot must be >= 1")
	}
	return &MBIM{device: device, slot: slot}, nil
}

// Connect establishes the MBIM session and opens the device.
func (m *MBIM) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("mbim reader is closed")
	}
	if m.reader != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	reader, err := openReader(ctx, uiccmbim.WithProxy(m.device), uiccmbim.WithSlot(int(m.slot)))
	if err != nil {
		return err
	}
	m.reader = reader
	return nil
}

// OpenLogicalChannel opens a logical channel for the specified Application ID.
func (m *MBIM) OpenLogicalChannel(AID []byte) (byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.ensureOpen(); err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	channel, err := m.reader.OpenChannel(ctx, AID)
	if err != nil {
		return 0, err
	}
	m.channel = channel
	return byte(channel), nil
}

// Transmit implements driver.SmartCardChannel.
func (m *MBIM) Transmit(command []byte) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.ensureOpen(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	response, status, err := m.reader.TransmitAPDU(ctx, m.channel, command)
	if err != nil {
		return nil, err
	}
	return append(response, uiccStatusWord(status)...), nil
}

// CloseLogicalChannel closes the specified logical channel.
func (m *MBIM) CloseLogicalChannel(channel byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.ensureOpen(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := m.reader.CloseChannel(ctx, uint32(channel)); err != nil {
		return err
	}
	if m.channel == uint32(channel) {
		m.channel = 0
	}
	return nil
}

// Disconnect closes the MBIM connection and releases resources.
func (m *MBIM) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}
	m.closed = true
	if m.reader == nil {
		return nil
	}
	return m.reader.Close()
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
