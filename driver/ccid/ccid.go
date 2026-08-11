package ccid

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/damonto/euicc-go/driver"
	"github.com/damonto/euicc-go/driver/iso7816"
	wwanccid "github.com/damonto/wwan-go/ccid"
)

const defaultTimeout = 30 * time.Second

var _ driver.SmartCardChannel = (*Reader)(nil)

// Reader is a PC/SC smart card channel. Create one with New or
// NewWithReader. It is not safe for concurrent use.
type Reader struct {
	reader    string
	channel   *iso7816.Channel
	options   []iso7816.Option
	connected bool
	closed    bool
}

// New creates a CCID channel. Options configure its ISO 7816 operations.
func New(options ...iso7816.Option) *Reader {
	return NewWithReader("", options...)
}

// NewWithReader creates a CCID channel with reader preselected. Options
// configure its ISO 7816 operations.
func NewWithReader(reader string, options ...iso7816.Option) *Reader {
	return &Reader{
		reader:  reader,
		options: slices.Clone(options),
	}
}

func (c *Reader) ListReaders() ([]string, error) {
	if c.closed {
		return nil, errors.New("ccid reader is closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	readers, err := wwanccid.ListReaders(ctx)
	if err != nil {
		return nil, fmt.Errorf("list CCID readers: %w", err)
	}
	return readers, nil
}

// SetReader selects the reader used by Connect.
func (c *Reader) SetReader(reader string) error {
	if c.closed {
		return errors.New("ccid reader is closed")
	}
	if c.connected {
		return errors.New("ccid reader is connected")
	}
	c.reader = reader
	return nil
}

func (c *Reader) Connect() error {
	if c.closed {
		return errors.New("ccid reader is closed")
	}
	if c.connected {
		return nil
	}
	if c.reader == "" {
		return errors.New("ccid reader is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	reader, err := wwanccid.Open(ctx, c.reader)
	if err != nil {
		return fmt.Errorf("open CCID reader %q: %w", c.reader, err)
	}
	channel := iso7816.NewChannel(reader, c.options...)
	if err := channel.Connect(); err != nil {
		return errors.Join(err, channel.Disconnect())
	}
	c.channel = channel
	c.connected = true
	return nil
}

func (c *Reader) Disconnect() error {
	if c.closed {
		return nil
	}
	c.closed = true
	c.connected = false
	if c.channel == nil {
		return nil
	}
	return c.channel.Disconnect()
}

func (c *Reader) Transmit(command []byte) ([]byte, error) {
	channel, err := c.smartCardChannel()
	if err != nil {
		return nil, err
	}
	return channel.Transmit(command)
}

func (c *Reader) OpenLogicalChannel(aid []byte) (byte, error) {
	channel, err := c.smartCardChannel()
	if err != nil {
		return 0, err
	}
	return channel.OpenLogicalChannel(aid)
}

func (c *Reader) CloseLogicalChannel(logicalChannel byte) error {
	channel, err := c.smartCardChannel()
	if err != nil {
		return err
	}
	return channel.CloseLogicalChannel(logicalChannel)
}

func (c *Reader) smartCardChannel() (*iso7816.Channel, error) {
	if c.closed {
		return nil, errors.New("ccid reader is closed")
	}
	if c.channel == nil {
		return nil, errors.New("ccid reader is not connected")
	}
	return c.channel, nil
}
