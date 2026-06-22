package qcom

import (
	"context"

	"github.com/damonto/euicc-go/driver"
	uiccqrtr "github.com/damonto/uicc-go/qcom/qrtr"
	"github.com/damonto/uicc-go/qcom/uim"
)

// QRTR implements driver.SmartCardChannel over QRTR.
type QRTR struct {
	*channel
}

// NewQRTR creates a new QRTR connection to the UIM service.
func NewQRTR(slot uint8) (driver.SmartCardChannel, error) {
	if err := validateSlot(slot); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	transport, err := uiccqrtr.Open(ctx)
	if err != nil {
		return nil, err
	}
	reader, err := uim.New(ctx, transport, uim.WithSlot(slot))
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	return &QRTR{channel: newChannel(reader)}, nil
}
