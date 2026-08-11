package qcom

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	maxSlot           = 5
	maxLogicalChannel = 19
	defaultTimeout    = 30 * time.Second
)

type uimReader interface {
	ActivateSlot(ctx context.Context) error
	OpenLogicalChannel(ctx context.Context, aid []byte) (uint8, error)
	SendAPDU(ctx context.Context, channel uint8, command []byte) ([]byte, error)
	CloseLogicalChannel(ctx context.Context, channel uint8) error
	Close() error
}

type channel struct {
	reader  uimReader
	channel uint8
	closed  bool
}

func newChannel(reader uimReader) *channel {
	return &channel{reader: reader}
}

func validateSlot(slot uint8) error {
	if slot < 1 || slot > maxSlot {
		return fmt.Errorf("slot must be between 1 and %d", maxSlot)
	}
	return nil
}

func (c *channel) Connect() error {
	if c.closed {
		return errors.New("smart card channel is closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := c.reader.ActivateSlot(ctx); err != nil {
		return fmt.Errorf("activate QCOM slot: %w", err)
	}
	return nil
}

func (c *channel) Disconnect() error {
	if c.closed {
		return nil
	}
	var channelErr error
	if c.channel != 0 {
		channelErr = c.closeLogicalChannel(c.channel)
	}
	c.closed = true
	closeErr := c.reader.Close()
	if closeErr != nil {
		closeErr = fmt.Errorf("close QCOM reader: %w", closeErr)
	}
	return errors.Join(channelErr, closeErr)
}

func (c *channel) OpenLogicalChannel(aid []byte) (byte, error) {
	if c.closed {
		return 0, errors.New("smart card channel is closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	channel, err := c.reader.OpenLogicalChannel(ctx, aid)
	if err != nil {
		return 0, fmt.Errorf("open QCOM logical channel: %w", err)
	}
	if channel == 0 || channel > maxLogicalChannel {
		var cleanupErr error
		if channel != 0 {
			cleanupErr = c.closeLogicalChannel(channel)
		}
		return 0, errors.Join(
			fmt.Errorf("QCOM returned invalid logical channel %d", channel),
			cleanupErr,
		)
	}
	c.channel = channel
	return channel, nil
}

func (c *channel) Transmit(command []byte) ([]byte, error) {
	if c.closed {
		return nil, errors.New("smart card channel is closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	response, err := c.reader.SendAPDU(ctx, c.channel, command)
	if err != nil {
		return nil, fmt.Errorf("transmit QCOM APDU: %w", err)
	}
	return response, nil
}

func (c *channel) CloseLogicalChannel(channel byte) error {
	if c.closed {
		return errors.New("smart card channel is closed")
	}
	if channel == 0 || channel > maxLogicalChannel {
		return fmt.Errorf("invalid logical channel %d", channel)
	}
	return c.closeLogicalChannel(channel)
}

func (c *channel) closeLogicalChannel(channel byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := c.reader.CloseLogicalChannel(ctx, channel); err != nil {
		return fmt.Errorf("close QCOM logical channel %d: %w", channel, err)
	}
	if c.channel == channel {
		c.channel = 0
	}
	return nil
}
