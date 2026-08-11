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
	slot uint8
}

var _ driver.SmartCardChannel = (*QRTR)(nil)

// NewQRTR creates an unconnected QRTR channel to the UIM service.
func NewQRTR(options ...Option) (*QRTR, error) {
	config := applyOptions(options)
	if err := config.validateQRTR(); err != nil {
		return nil, err
	}

	return &QRTR{
		channel: newChannel(nil, config.timeout),
		slot:    config.slot,
	}, nil
}

// Connect opens the QRTR transport and activates the configured slot.
func (r *QRTR) Connect() error {
	if r.closed {
		return errors.New("smart card channel is closed")
	}
	if r.reader != nil {
		return r.channel.Connect()
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	transport, err := wwanqrtr.Open(ctx)
	if err != nil {
		return fmt.Errorf("open QRTR transport: %w", err)
	}
	reader, err := qcom.NewClient(transport, qcom.WithSlot(r.slot))
	if err != nil {
		closeErr := transport.Close()
		if closeErr != nil {
			closeErr = fmt.Errorf("close QRTR transport: %w", closeErr)
		}
		return errors.Join(fmt.Errorf("create QRTR client: %w", err), closeErr)
	}
	r.reader = reader
	if err := r.channel.Connect(); err != nil {
		return errors.Join(err, r.releaseReader())
	}
	return nil
}
