package qcom

import (
	"context"

	"github.com/damonto/euicc-go/driver"
	"github.com/damonto/wwan-go/qcom"
	wwanqmi "github.com/damonto/wwan-go/qcom/qmi"
)

// QMI implements driver.SmartCardChannel over a QMI proxy connection.
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
	transport, err := wwanqmi.Open(ctx, wwanqmi.WithProxy(device))
	if err != nil {
		return nil, err
	}
	reader, err := qcom.NewClient(transport, qcom.WithSlot(slot))
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	return &QMI{channel: newChannel(reader)}, nil
}
