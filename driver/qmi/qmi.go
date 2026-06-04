package qmi

import (
	"context"

	"github.com/damonto/euicc-go/apdu"
	uiccqmi "github.com/damonto/uicc-go/qualcomm/qmi"
	"github.com/damonto/uicc-go/qualcomm/uim"
)

// QMI implements apdu.SmartCardChannel over a QMI proxy connection.
type QMI struct {
	*channel
}

// New creates a new QMI connection to the specified device.
func New(device string, slot uint8) (apdu.SmartCardChannel, error) {
	if err := validateSlot(slot); err != nil {
		return nil, err
	}

	ctx := context.Background()
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
