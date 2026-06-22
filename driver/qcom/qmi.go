package qcom

import (
	"context"

	"github.com/damonto/euicc-go/driver"
	uiccqmi "github.com/damonto/uicc-go/qcom/qmi"
	"github.com/damonto/uicc-go/qcom/uim"
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
	transport, err := uiccqmi.Open(ctx, uiccqmi.WithProxy(device))
	if err != nil {
		return nil, err
	}
	reader, err := uim.New(ctx, transport, uim.WithSlot(slot))
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	return &QMI{channel: newChannel(reader)}, nil
}
