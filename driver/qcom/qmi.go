package qcom

import (
	"context"
	"errors"
	"fmt"

	"github.com/damonto/euicc-go/driver"
	"github.com/damonto/wwan-go/qcom"
	wwanqmi "github.com/damonto/wwan-go/qcom/qmi"
)

// QMI implements driver.SmartCardChannel over a QMI connection. It is not safe
// for concurrent use.
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
		return nil, fmt.Errorf("open QMI transport: %w", err)
	}
	reader, err := qcom.NewClient(transport, qcom.WithSlot(slot))
	if err != nil {
		closeErr := transport.Close()
		if closeErr != nil {
			closeErr = fmt.Errorf("close QMI transport: %w", closeErr)
		}
		return nil, errors.Join(fmt.Errorf("create QMI client: %w", err), closeErr)
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
