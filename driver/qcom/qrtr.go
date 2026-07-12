package qcom

import (
	"context"

	"github.com/damonto/euicc-go/driver"
	"github.com/damonto/wwan-go/qcom"
	wwanqrtr "github.com/damonto/wwan-go/qcom/qrtr"
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
	transport, err := wwanqrtr.Open(ctx)
	if err != nil {
		return nil, err
	}
	reader, err := qcom.NewClient(transport, qcom.WithSlot(slot))
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	return &QRTR{channel: newChannel(reader)}, nil
}
