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
	access accessMode
	device string
	slot   uint8
}

var _ driver.SmartCardChannel = (*QMI)(nil)

// NewQMI creates a QMI channel. Configure its connection using WithAutoDetect,
// WithDirect, WithProxy, or WithClient.
func NewQMI(options ...Option) (*QMI, error) {
	config := applyOptions(options)
	if err := config.validateQMI(); err != nil {
		return nil, err
	}
	if config.client != nil {
		return &QMI{channel: newChannel(config.client, config.timeout)}, nil
	}
	return &QMI{
		channel: newChannel(nil, config.timeout),
		access:  config.access,
		device:  config.device,
		slot:    config.slot,
	}, nil
}

// Connect opens the QMI transport and activates the configured slot.
func (q *QMI) Connect() error {
	if q.closed {
		return errors.New("smart card channel is closed")
	}
	if q.reader != nil {
		return q.channel.Connect()
	}

	ctx, cancel := context.WithTimeout(context.Background(), q.timeout)
	defer cancel()
	var accessOption wwanqmi.Option
	switch q.access {
	case accessAutoDetect:
		accessOption = wwanqmi.WithAutoDetect(q.device)
	case accessDirect:
		accessOption = wwanqmi.WithDirect(q.device)
	case accessProxy:
		accessOption = wwanqmi.WithProxy(q.device)
	}
	transport, err := wwanqmi.Open(ctx, accessOption)
	if err != nil {
		return fmt.Errorf("open QMI transport: %w", err)
	}
	reader, err := qcom.NewClient(transport, qcom.WithSlot(q.slot))
	if err != nil {
		closeErr := transport.Close()
		if closeErr != nil {
			closeErr = fmt.Errorf("close QMI transport: %w", closeErr)
		}
		return errors.Join(fmt.Errorf("create QMI client: %w", err), closeErr)
	}
	q.reader = reader
	if err := q.channel.Connect(); err != nil {
		return errors.Join(err, q.releaseReader())
	}
	return nil
}
