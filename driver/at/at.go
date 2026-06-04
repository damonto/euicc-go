package at

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/damonto/euicc-go/apdu"
	uiccat "github.com/damonto/uicc-go/at"
)

const (
	maxLogicalChannel      = 19
	maxShortAPDUDataLength = 255
	defaultBaudRate        = 115200
)

var connectAPDU = []byte{0x80, 0xAA, 0x00, 0x00, 0x0A, 0xA9, 0x08, 0x81, 0x00, 0x82, 0x01, 0x01, 0x83, 0x01, 0x07}

type transmitter interface {
	Transmit(ctx context.Context, command []byte) ([]byte, error)
	Close() error
}

// AT is an AT smart card channel.
type AT struct {
	*channel
}

// New creates an AT smart card channel.
func New(device string) (apdu.SmartCardChannel, error) {
	reader, err := uiccat.Open(context.Background(), device, defaultBaudRate, uiccat.WithoutInit())
	if err != nil {
		return nil, fmt.Errorf("open serial port %s: %w", device, err)
	}
	return &AT{channel: newChannel(reader)}, nil
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
	response, err := c.tx.Transmit(context.Background(), connectAPDU)
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
	return c.tx.Transmit(context.Background(), command)
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
	channel, err := c.openChannel()
	if err != nil {
		return 0, err
	}
	if err := c.selectAID(channel, AID); err != nil {
		return 0, errors.Join(err, c.closeLogicalChannel(channel))
	}
	c.channel = channel
	return channel, nil
}

func (c *channel) CloseLogicalChannel(channel byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closeLogicalChannel(channel)
}

func (c *channel) openChannel() (byte, error) {
	response, err := c.tx.Transmit(context.Background(), []byte{0x00, 0x70, 0x00, 0x00, 0x01})
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

func (c *channel) selectAID(channel byte, AID []byte) error {
	command, err := selectAIDCommand(channel, AID)
	if err != nil {
		return err
	}
	response, err := c.tx.Transmit(context.Background(), command)
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

func (c *channel) closeLogicalChannel(channel byte) error {
	if channel == 0 || channel > maxLogicalChannel {
		return fmt.Errorf("invalid logical channel %d", channel)
	}
	response, err := c.tx.Transmit(context.Background(), []byte{0x00, 0x70, 0x80, channel, 0x00})
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
