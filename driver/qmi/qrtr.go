package qmi

import (
	"errors"

	"github.com/damonto/euicc-go/apdu"
	"github.com/damonto/euicc-go/driver/qmi/protocol"
	"github.com/damonto/euicc-go/driver/qmi/qrtr"
	transport "github.com/damonto/euicc-go/driver/qmi/transport/qrtr"
	"github.com/damonto/euicc-go/driver/qmi/uim"
)

// QRTR implements the apdu.SmartCardChannel interface using QRTR protocol.
type QRTR struct {
	conn *qrtr.Conn
	uim.Client
}

// NewQRTR creates a new QRTR connection to the UIM service.
func NewQRTR(slot uint8) (apdu.SmartCardChannel, error) {
	if slot == 0 {
		return nil, errors.New("slot must be >= 1")
	}

	conn, err := qrtr.Open(protocol.QMIServiceUIM)
	if err != nil {
		return nil, err
	}
	q := &QRTR{
		conn: conn,
		Client: uim.Client{
			Transport: transport.New(conn),
			Slot:      slot,
		},
	}
	return q, nil
}

func (c *QRTR) Disconnect() error {
	return c.conn.Close()
}
