package qcom

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

const maxSlot = 5
const defaultTimeout = 30 * time.Second

type uimReader interface {
	ActivateSlot(ctx context.Context) error
	OpenLogicalChannel(ctx context.Context, aid []byte) (uint8, error)
	SendAPDU(ctx context.Context, channel uint8, command []byte) ([]byte, error)
	CloseLogicalChannel(ctx context.Context, channel uint8) error
	Close() error
}

type channel struct {
	mu      sync.Mutex
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
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errors.New("smart card channel is closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return c.reader.ActivateSlot(ctx)
}

func (c *channel) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true
	return c.reader.Close()
}

func (c *channel) OpenLogicalChannel(AID []byte) (byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return 0, errors.New("smart card channel is closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	channel, err := c.reader.OpenLogicalChannel(ctx, AID)
	if err != nil {
		return 0, err
	}
	c.channel = channel
	return channel, nil
}

func (c *channel) Transmit(command []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, errors.New("smart card channel is closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return c.reader.SendAPDU(ctx, c.channel, command)
}

func (c *channel) CloseLogicalChannel(channel byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errors.New("smart card channel is closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := c.reader.CloseLogicalChannel(ctx, channel); err != nil {
		return err
	}
	if c.channel == channel {
		c.channel = 0
	}
	return nil
}
