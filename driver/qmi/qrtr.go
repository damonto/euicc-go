package qmi

import (
	"context"

	"github.com/damonto/euicc-go/apdu"
	uiccqrtr "github.com/damonto/uicc-go/qualcomm/qrtr"
	"github.com/damonto/uicc-go/qualcomm/uim"
)

// QRTR implements apdu.SmartCardChannel over QRTR.
type QRTR struct {
	*channel
}

// NewQRTR creates a new QRTR connection to the UIM service.
func NewQRTR(slot uint8) (apdu.SmartCardChannel, error) {
	if err := validateSlot(slot); err != nil {
		return nil, err
	}

	ctx := context.Background()
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
