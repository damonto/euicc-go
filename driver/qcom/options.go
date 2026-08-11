package qcom

import (
	"errors"
	"time"

	wwanqcom "github.com/damonto/wwan-go/qcom"
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
	client    *wwanqcom.Client
	slot      uint8
	timeout   time.Duration
	clientSet bool
	slotSet   bool
}

// Option configures a Qualcomm smart-card channel.
type Option func(*config)

// WithAutoDetect uses qmi-proxy when available and otherwise accesses device
// directly.
func WithAutoDetect(device string) Option {
	return func(config *config) {
		config.access = accessAutoDetect
		config.device = device
	}
}

// WithDirect accesses the QMI device directly.
func WithDirect(device string) Option {
	return func(config *config) {
		config.access = accessDirect
		config.device = device
	}
}

// WithProxy accesses the QMI device through qmi-proxy.
func WithProxy(device string) Option {
	return func(config *config) {
		config.access = accessProxy
		config.device = device
	}
}

// WithClient configures NewQMI to use an already connected Qualcomm UIM client.
// The channel takes ownership of client and closes it on Disconnect.
func WithClient(client *wwanqcom.Client) Option {
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

// WithTimeout sets the timeout for each QMI or QRTR operation. The default is
// 30 seconds. A non-positive timeout makes operations expire immediately.
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

func (c config) validateQMI() error {
	if c.clientSet {
		return c.validateClient()
	}
	if err := validateSlot(c.slot); err != nil {
		return err
	}
	if c.access == accessUnset {
		return errors.New("QMI access method is required")
	}
	if c.device == "" {
		return errors.New("QMI device is required")
	}
	return nil
}

func (c config) validateQRTR() error {
	if c.clientSet {
		return errors.New("QCOM client option cannot be used with QRTR")
	}
	if err := validateSlot(c.slot); err != nil {
		return err
	}
	if c.access != accessUnset {
		return errors.New("QMI access option cannot be used with QRTR")
	}
	return nil
}

func (c config) validateClient() error {
	if c.client == nil {
		return errors.New("QCOM client is nil")
	}
	if c.access != accessUnset {
		return errors.New("QMI access option cannot be used with an existing client")
	}
	if c.slotSet {
		return errors.New("QCOM slot option cannot be used with an existing client")
	}
	return nil
}
