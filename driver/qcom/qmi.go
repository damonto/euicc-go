package qcom

import (
	"context"
	"errors"

	"github.com/damonto/euicc-go/driver"
	"github.com/damonto/wwan-go/qcom"
	wwanqmi "github.com/damonto/wwan-go/qcom/qmi"
)

// QMI implements driver.SmartCardChannel over a QMI connection.
type QMI struct {
	*channel
}

// NewQMI creates a new QMI connection to the specified device.
func NewQMI(device string, slot uint8) (driver.SmartCardChannel, error) {
	if err := validateSlot(slot); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	transport, err := wwanqmi.Open(ctx, wwanqmi.WithAutoDetect(device))
	if err != nil {
		return nil, err
	}
	reader, err := qcom.NewClient(transport, qcom.WithSlot(slot))
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	return NewQMIWithClient(reader)
}

// NewQMIWithClient creates a channel backed by an already connected QMI
// client. The channel takes ownership of client and closes it on Disconnect.
func NewQMIWithClient(client *qcom.Client) (driver.SmartCardChannel, error) {
	if client == nil {
		return nil, errors.New("qmi client is nil")
	}
	return &QMI{channel: newChannel(client)}, nil
}
