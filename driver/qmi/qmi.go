package qmi

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/damonto/euicc-go/apdu"
	"github.com/damonto/euicc-go/driver/qmi/core"
	transport "github.com/damonto/euicc-go/driver/qmi/transport/qmi"
)

// QMI implements the apdu.SmartCardChannel interface using QMI protocol
type QMI struct {
	core.QMIClient
	conn   net.Conn
	device string
}

// New creates a new QMI connection to the specified device
func New(device string, slot uint8) (apdu.SmartCardChannel, error) {
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: "\x00qmi-proxy", Net: "unix"})
	if err != nil {
		return nil, err
	}
	q := &QMI{
		conn:   conn,
		device: device,
		QMIClient: core.QMIClient{
			Transport: transport.New(conn),
			Slot:      slot,
		},
	}
	if err := q.openProxyConnection(); err != nil {
		q.conn.Close()
		return nil, err
	}
	if err := q.allocateClientID(); err != nil {
		q.conn.Close()
		return nil, err
	}
	return q, nil
}

// openProxyConnection sends a request to the qmi-proxy to open a connection
func (q *QMI) openProxyConnection() error {
	request := core.InternalOpenRequest{
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
		DevicePath:    []byte(q.device),
	}
	err := q.Transport.Transmit(request.Request())
	if errors.Is(err, io.EOF) {
		return fmt.Errorf("device %s is not connected", q.device)
	}
	return err
}

// allocateClientID sends a request to allocate a client ID for UIM service
func (q *QMI) allocateClientID() error {
	request := core.AllocateClientIDRequest{
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
	}
	err := q.Transport.Transmit(request.Request())
	if errors.Is(err, io.EOF) {
		return fmt.Errorf("device %s doesn't support QMI protocol", q.device)
	}
	if err != nil {
		return err
	}
	q.ClientID = request.Response.ClientID
	return nil
}

// releaseClientID sends a request to release the allocated client ID
func (q *QMI) releaseClientID() error {
	request := core.ReleaseClientIDRequest{
		ClientID:      q.ClientID,
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
	}
	return q.Transport.Transmit(request.Request())
}

// Disconnect releases the client ID and closes the connection
func (q *QMI) Disconnect() error {
	var errs []error
	if q.ClientID != 0 {
		if err := q.releaseClientID(); err != nil {
			errs = append(errs, err)
		} else {
			q.ClientID = 0
		}
	}
	if err := q.conn.Close(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
