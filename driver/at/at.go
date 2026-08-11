package at

import (
	"fmt"

	"github.com/damonto/euicc-go/driver"
	"github.com/damonto/euicc-go/driver/iso7816"
	wwanat "github.com/damonto/wwan-go/at"
)

const defaultBaudRate = 115200

// AT is an AT smart card channel. It is not safe for concurrent use.
type AT struct {
	*iso7816.Channel
}

// New creates an AT smart card channel. Options configure its ISO 7816
// operations.
func New(device string, options ...iso7816.Option) (driver.SmartCardChannel, error) {
	reader, err := wwanat.Open(device, defaultBaudRate)
	if err != nil {
		return nil, fmt.Errorf("open serial port %s: %w", device, err)
	}
	return &AT{Channel: iso7816.NewChannel(reader, options...)}, nil
}
