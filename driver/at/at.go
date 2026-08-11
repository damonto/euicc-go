package at

import (
	"errors"
	"fmt"
	"slices"

	"github.com/damonto/euicc-go/driver"
	"github.com/damonto/euicc-go/driver/iso7816"
	wwanat "github.com/damonto/wwan-go/at"
)

const defaultBaudRate = 115200

// AT is an AT smart card channel. It is not safe for concurrent use.
type AT struct {
	device  string
	options []iso7816.Option
	channel *iso7816.Channel
	closed  bool
}

var _ driver.SmartCardChannel = (*AT)(nil)

// New creates an unconnected AT smart card channel. Options configure its ISO
// 7816 operations.
func New(device string, options ...iso7816.Option) (*AT, error) {
	if device == "" {
		return nil, errors.New("AT device is required")
	}
	return &AT{device: device, options: slices.Clone(options)}, nil
}

// Connect opens the serial port and initializes the eUICC transport.
func (a *AT) Connect() error {
	if a.closed {
		return errors.New("AT channel is closed")
	}
	if a.channel != nil {
		return nil
	}
	reader, err := wwanat.Open(a.device, defaultBaudRate)
	if err != nil {
		return fmt.Errorf("open serial port %s: %w", a.device, err)
	}
	channel := iso7816.NewChannel(reader, a.options...)
	if err := channel.Connect(); err != nil {
		return errors.Join(err, channel.Disconnect())
	}
	a.channel = channel
	return nil
}

// Disconnect closes the serial port and releases its resources.
func (a *AT) Disconnect() error {
	if a.closed {
		return nil
	}
	a.closed = true
	if a.channel == nil {
		return nil
	}
	return a.channel.Disconnect()
}

// Transmit exchanges one raw APDU with the modem.
func (a *AT) Transmit(command []byte) ([]byte, error) {
	channel, err := a.smartCardChannel()
	if err != nil {
		return nil, err
	}
	return channel.Transmit(command)
}

// OpenLogicalChannel opens a logical channel and selects aid.
func (a *AT) OpenLogicalChannel(aid []byte) (byte, error) {
	channel, err := a.smartCardChannel()
	if err != nil {
		return 0, err
	}
	return channel.OpenLogicalChannel(aid)
}

// CloseLogicalChannel closes a logical channel.
func (a *AT) CloseLogicalChannel(logicalChannel byte) error {
	channel, err := a.smartCardChannel()
	if err != nil {
		return err
	}
	return channel.CloseLogicalChannel(logicalChannel)
}

func (a *AT) smartCardChannel() (*iso7816.Channel, error) {
	if a.closed {
		return nil, errors.New("AT channel is closed")
	}
	if a.channel == nil {
		return nil, errors.New("AT channel is not connected")
	}
	return a.channel, nil
}
