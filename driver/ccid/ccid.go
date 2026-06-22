package ccid

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/damonto/euicc-go/driver"
	uiccccid "github.com/damonto/uicc-go/ccid"
)

const (
	maxLogicalChannel      = 19
	maxShortAPDUDataLength = 255
	defaultTimeout         = 30 * time.Second
)

var connectAPDU = []byte{0x80, 0xAA, 0x00, 0x00, 0x0A, 0xA9, 0x08, 0x81, 0x00, 0x82, 0x01, 0x01, 0x83, 0x01, 0x07}

type transmitter interface {
	Transmit(ctx context.Context, command []byte) ([]byte, error)
	Close() error
}

type readerOpener func(context.Context, string) (transmitter, error)
type readerLister func(context.Context) ([]string, error)

var (
	openReader  readerOpener = openUICCReader
	listReaders readerLister = uiccccid.ListReaders
)

// CCID is a PC/SC smart card channel.
type CCID interface {
	driver.SmartCardChannel
	ListReaders() ([]string, error)
	SetReader(reader string) error
}

// CCIDReader is a PC/SC smart card channel.
type CCIDReader struct {
	mu     sync.Mutex
	reader string
	tx     *channel
	open   bool
	closed bool
}

// New creates a CCID channel.
func New() (*CCIDReader, error) {
	return NewWithReader("")
}

// NewWithReader creates a CCID channel with reader preselected.
func NewWithReader(reader string) (*CCIDReader, error) {
	return &CCIDReader{reader: reader}, nil
}

func openUICCReader(ctx context.Context, reader string) (transmitter, error) {
	return uiccccid.Open(ctx, reader)
}

func (c *CCIDReader) ListReaders() ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, errors.New("ccid reader is closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return listReaders(ctx)
}

// SetReader selects the reader used by Connect.
func (c *CCIDReader) SetReader(reader string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errors.New("ccid reader is closed")
	}
	if c.open {
		return errors.New("ccid reader is connected")
	}
	c.reader = reader
	return nil
}

func (c *CCIDReader) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errors.New("ccid reader is closed")
	}
	if c.open {
		return nil
	}
	if c.reader == "" {
		return errors.New("ccid reader is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	reader, err := openReader(ctx, c.reader)
	if err != nil {
		return err
	}
	tx := newChannel(reader)
	if err := tx.Connect(); err != nil {
		return errors.Join(err, tx.Disconnect())
	}
	c.tx = tx
	c.open = true
	return nil
}

func (c *CCIDReader) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true
	c.open = false
	if c.tx == nil {
		return nil
	}
	return c.tx.Disconnect()
}

func (c *CCIDReader) Transmit(command []byte) ([]byte, error) {
	c.mu.Lock()
	tx := c.tx
	closed := c.closed
	c.mu.Unlock()

	if tx == nil {
		if closed {
			return nil, errors.New("ccid reader is closed")
		}
		return nil, errors.New("ccid reader is not connected")
	}
	return tx.Transmit(command)
}

func (c *CCIDReader) OpenLogicalChannel(AID []byte) (byte, error) {
	c.mu.Lock()
	tx := c.tx
	closed := c.closed
	c.mu.Unlock()

	if tx == nil {
		if closed {
			return 0, errors.New("ccid reader is closed")
		}
		return 0, errors.New("ccid reader is not connected")
	}
	return tx.OpenLogicalChannel(AID)
}

func (c *CCIDReader) CloseLogicalChannel(channel byte) error {
	c.mu.Lock()
	tx := c.tx
	closed := c.closed
	c.mu.Unlock()

	if tx == nil {
		if closed {
			return errors.New("ccid reader is closed")
		}
		return errors.New("ccid reader is not connected")
	}
	return tx.CloseLogicalChannel(channel)
}

type channel struct {
	mu      sync.Mutex
	tx      transmitter
	channel byte
	closed  bool
}

func newChannel(tx transmitter) *channel {
	return &channel{tx: tx}
}

func (c *channel) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errors.New("smart card channel is closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	response, err := c.tx.Transmit(ctx, connectAPDU)
	if err != nil {
		return err
	}
	if !statusOK(response) && !statusHasMore(response) {
		return fmt.Errorf("connect APDU: %X", response)
	}
	return nil
}

func (c *channel) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true
	return c.tx.Close()
}

func (c *channel) Transmit(command []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, errors.New("smart card channel is closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return c.tx.Transmit(ctx, command)
}

func (c *channel) OpenLogicalChannel(AID []byte) (byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return 0, errors.New("smart card channel is closed")
	}
	if len(AID) > maxShortAPDUDataLength {
		return 0, fmt.Errorf("AID length %d exceeds short APDU limit", len(AID))
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	channel, err := c.openChannel(ctx)
	if err != nil {
		return 0, err
	}
	if err := c.selectAID(ctx, channel, AID); err != nil {
		return 0, errors.Join(err, c.closeLogicalChannel(ctx, channel))
	}
	c.channel = channel
	return channel, nil
}

func (c *channel) CloseLogicalChannel(channel byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return c.closeLogicalChannel(ctx, channel)
}

func (c *channel) openChannel(ctx context.Context) (byte, error) {
	response, err := c.tx.Transmit(ctx, []byte{0x00, 0x70, 0x00, 0x00, 0x01})
	if err != nil {
		return 0, err
	}
	if len(response) < 3 {
		return 0, fmt.Errorf("open logical channel returned short response: %X", response)
	}
	if !statusOK(response) {
		return 0, fmt.Errorf("open logical channel: %X", response)
	}
	channel := response[0]
	if channel == 0 || channel > maxLogicalChannel {
		return 0, fmt.Errorf("open logical channel returned invalid logical channel %d", channel)
	}
	return channel, nil
}

func (c *channel) selectAID(ctx context.Context, channel byte, AID []byte) error {
	command, err := selectAIDCommand(channel, AID)
	if err != nil {
		return err
	}
	response, err := c.tx.Transmit(ctx, command)
	if err != nil {
		return err
	}
	if len(response) < 2 {
		return fmt.Errorf("select AID returned short response: %X", response)
	}
	if !statusOK(response) && !statusHasMore(response) {
		return fmt.Errorf("select AID: %X", response)
	}
	return nil
}

func (c *channel) closeLogicalChannel(ctx context.Context, channel byte) error {
	if channel == 0 || channel > maxLogicalChannel {
		return fmt.Errorf("invalid logical channel %d", channel)
	}
	response, err := c.tx.Transmit(ctx, []byte{0x00, 0x70, 0x80, channel, 0x00})
	if err != nil {
		return err
	}
	if len(response) < 2 {
		return fmt.Errorf("close logical channel returned short response: %X", response)
	}
	if !statusOK(response) {
		return fmt.Errorf("close logical channel: %X", response)
	}
	if c.channel == channel {
		c.channel = 0
	}
	return nil
}

func selectAIDCommand(channel byte, AID []byte) ([]byte, error) {
	cla, err := classByteForChannel(0x00, channel)
	if err != nil {
		return nil, err
	}
	if len(AID) > maxShortAPDUDataLength {
		return nil, fmt.Errorf("AID length %d exceeds short APDU limit", len(AID))
	}
	command := make([]byte, 0, 5+len(AID))
	command = append(command, cla, 0xA4, 0x04, 0x00, byte(len(AID)))
	command = append(command, AID...)
	return command, nil
}

func classByteForChannel(cla, channel byte) (byte, error) {
	if channel < 4 {
		return (cla & 0x9C) | channel, nil
	}
	if channel <= maxLogicalChannel {
		return (cla & 0xB0) | 0x40 | (channel - 4), nil
	}
	return 0, fmt.Errorf("logical channel %d exceeds maximum %d", channel, maxLogicalChannel)
}

func statusOK(response []byte) bool {
	return len(response) >= 2 && response[len(response)-2] == 0x90 && response[len(response)-1] == 0x00
}

func statusHasMore(response []byte) bool {
	return len(response) >= 2 && response[len(response)-2] == 0x61
}
