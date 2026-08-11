// Package iso7816 implements logical-channel APDUs for smart card transports
// that expose raw command transmission.
package iso7816

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	maxLogicalChannel      = 19
	maxShortAPDUDataLength = 255
	defaultTimeout         = 30 * time.Second
)

const connectAPDU = "\x80\xAA\x00\x00\x0A\xA9\x08\x81\x00\x82\x01\x01\x83\x01\x07"

// Transmitter exchanges raw APDUs with a smart card transport.
type Transmitter interface {
	Transmit(ctx context.Context, command []byte) ([]byte, error)
	Close() error
}

// Option configures a Channel.
type Option func(*Channel)

// WithTimeout sets the timeout for each transport operation. A non-positive
// timeout makes operations expire immediately.
func WithTimeout(timeout time.Duration) Option {
	return func(channel *Channel) {
		channel.timeout = timeout
	}
}

// Channel manages ISO 7816 logical channels over a Transmitter.
// Channel is not safe for concurrent use.
type Channel struct {
	tx      Transmitter
	timeout time.Duration
	channel byte
	closed  bool
}

// NewChannel creates a logical-channel adapter over tx. Tx must not be nil.
func NewChannel(tx Transmitter, options ...Option) *Channel {
	channel := &Channel{tx: tx, timeout: defaultTimeout}
	for _, option := range options {
		option(channel)
	}
	return channel
}

// Connect initializes the eUICC transport.
func (c *Channel) Connect() error {
	if c.closed {
		return errors.New("smart card channel is closed")
	}
	ctx, cancel := c.newContext()
	defer cancel()
	response, err := c.tx.Transmit(ctx, []byte(connectAPDU))
	if err != nil {
		return fmt.Errorf("initialize eUICC transport: %w", err)
	}
	if !statusOK(response) && !statusHasMore(response) {
		return fmt.Errorf("connect APDU: %X", response)
	}
	return nil
}

// Disconnect closes the underlying transport.
func (c *Channel) Disconnect() error {
	if c.closed {
		return nil
	}
	var channelErr error
	if c.channel != 0 {
		ctx, cancel := c.newContext()
		channelErr = c.closeLogicalChannel(ctx, c.channel)
		cancel()
	}
	c.closed = true
	closeErr := c.tx.Close()
	if closeErr != nil {
		closeErr = fmt.Errorf("close smart card transport: %w", closeErr)
	}
	return errors.Join(channelErr, closeErr)
}

// Transmit exchanges one raw APDU with the underlying transport.
func (c *Channel) Transmit(command []byte) ([]byte, error) {
	if c.closed {
		return nil, errors.New("smart card channel is closed")
	}
	ctx, cancel := c.newContext()
	defer cancel()
	response, err := c.tx.Transmit(ctx, command)
	if err != nil {
		return nil, fmt.Errorf("transmit APDU: %w", err)
	}
	return response, nil
}

// OpenLogicalChannel opens a channel and selects AID on it.
func (c *Channel) OpenLogicalChannel(aid []byte) (byte, error) {
	if c.closed {
		return 0, errors.New("smart card channel is closed")
	}
	if len(aid) > maxShortAPDUDataLength {
		return 0, fmt.Errorf("AID length %d exceeds short APDU limit", len(aid))
	}
	if c.channel != 0 {
		return 0, fmt.Errorf("logical channel %d is already open", c.channel)
	}
	ctx, cancel := c.newContext()
	channel, err := c.openChannel(ctx)
	cancel()
	if err != nil {
		return 0, err
	}
	c.channel = channel
	ctx, cancel = c.newContext()
	err = c.selectAID(ctx, channel, aid)
	cancel()
	if err != nil {
		cleanupCtx, cleanupCancel := c.newContext()
		cleanupErr := c.closeLogicalChannel(cleanupCtx, channel)
		cleanupCancel()
		return 0, errors.Join(err, cleanupErr)
	}
	return channel, nil
}

// CloseLogicalChannel closes channel.
func (c *Channel) CloseLogicalChannel(channel byte) error {
	if c.closed {
		return errors.New("smart card channel is closed")
	}
	ctx, cancel := c.newContext()
	defer cancel()
	return c.closeLogicalChannel(ctx, channel)
}

func (c *Channel) newContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}

func (c *Channel) openChannel(ctx context.Context) (byte, error) {
	response, err := c.tx.Transmit(ctx, []byte{0x00, 0x70, 0x00, 0x00, 0x01})
	if err != nil {
		return 0, fmt.Errorf("open logical channel APDU: %w", err)
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

func (c *Channel) selectAID(ctx context.Context, channel byte, aid []byte) error {
	command, err := selectAIDCommand(channel, aid)
	if err != nil {
		return err
	}
	response, err := c.tx.Transmit(ctx, command)
	if err != nil {
		return fmt.Errorf("select AID: %w", err)
	}
	if len(response) < 2 {
		return fmt.Errorf("select AID returned short response: %X", response)
	}
	if !statusOK(response) && !statusHasMore(response) {
		return fmt.Errorf("select AID: %X", response)
	}
	return nil
}

func (c *Channel) closeLogicalChannel(ctx context.Context, channel byte) error {
	if channel == 0 || channel > maxLogicalChannel {
		return fmt.Errorf("invalid logical channel %d", channel)
	}
	response, err := c.tx.Transmit(ctx, []byte{0x00, 0x70, 0x80, channel, 0x00})
	if err != nil {
		return fmt.Errorf("close logical channel APDU: %w", err)
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

func selectAIDCommand(channel byte, aid []byte) ([]byte, error) {
	cla, err := classByteForChannel(0x00, channel)
	if err != nil {
		return nil, err
	}
	if len(aid) > maxShortAPDUDataLength {
		return nil, fmt.Errorf("AID length %d exceeds short APDU limit", len(aid))
	}
	command := make([]byte, 0, 5+len(aid))
	command = append(command, cla, 0xA4, 0x04, 0x00, byte(len(aid)))
	command = append(command, aid...)
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
