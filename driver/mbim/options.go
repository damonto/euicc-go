package mbim

import (
	"errors"
	"time"

	wwanmbim "github.com/damonto/wwan-go/mbim"
)

type accessMode uint8

const (
	accessUnset accessMode = iota
	accessAutoDetect
	accessDirect
	accessProxy
)

type config struct {
	access    accessMode
	device    string
	client    *wwanmbim.Client
	slot      uint8
	timeout   time.Duration
	clientSet bool
	slotSet   bool
}

// Option configures an MBIM channel.
type Option func(*config)

// WithAutoDetect uses mbim-proxy when available and otherwise accesses device
// directly.
func WithAutoDetect(device string) Option {
	return func(config *config) {
		config.access = accessAutoDetect
		config.device = device
	}
}

// WithDirect accesses the MBIM device directly.
func WithDirect(device string) Option {
	return func(config *config) {
		config.access = accessDirect
		config.device = device
	}
}

// WithProxy accesses the MBIM device through mbim-proxy.
func WithProxy(device string) Option {
	return func(config *config) {
		config.access = accessProxy
		config.device = device
	}
}

// WithClient uses an already connected MBIM client. The channel takes
// ownership of client and closes it on Disconnect.
func WithClient(client *wwanmbim.Client) Option {
	return func(config *config) {
		config.client = client
		config.clientSet = true
	}
}

// WithSlot selects the physical UICC slot. The default is slot 1.
func WithSlot(slot uint8) Option {
	return func(config *config) {
		config.slot = slot
		config.slotSet = true
	}
}

// WithTimeout sets the timeout for each MBIM operation. The default is 30
// seconds. A non-positive timeout makes operations expire immediately.
func WithTimeout(timeout time.Duration) Option {
	return func(config *config) {
		config.timeout = timeout
	}
}

func applyOptions(options []Option) config {
	config := config{slot: 1, timeout: defaultTimeout}
	for _, option := range options {
		option(&config)
	}
	return config
}

func (c config) validate() error {
	if c.clientSet {
		if c.client == nil {
			return errors.New("MBIM client is nil")
		}
		if c.access != accessUnset {
			return errors.New("MBIM client and access options are mutually exclusive")
		}
		if c.slotSet {
			return errors.New("MBIM slot option cannot be used with an existing client")
		}
		return nil
	}

	if c.slot == 0 {
		return errors.New("MBIM slot must be >= 1")
	}
	if c.access == accessUnset {
		return errors.New("MBIM access method is required")
	}
	if c.device == "" {
		return errors.New("MBIM device is required")
	}
	return nil
}
