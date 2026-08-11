package qcom

import (
	"context"
	"errors"
	"fmt"

	"github.com/damonto/euicc-go/driver"
	"github.com/damonto/wwan-go/qcom"
	wwanqrtr "github.com/damonto/wwan-go/qcom/qrtr"
)

// QRTR implements driver.SmartCardChannel over QRTR. It is not safe for
// concurrent use.
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
		return nil, fmt.Errorf("open QRTR transport: %w", err)
	}
	reader, err := qcom.NewClient(transport, qcom.WithSlot(slot))
	if err != nil {
		closeErr := transport.Close()
		if closeErr != nil {
			closeErr = fmt.Errorf("close QRTR transport: %w", closeErr)
		}
		return nil, errors.Join(fmt.Errorf("create QRTR client: %w", err), closeErr)
	}
	return &QRTR{channel: newChannel(reader)}, nil
}
