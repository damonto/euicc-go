package qmi

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync/atomic"
	"syscall"

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
	conn, err := newQMIConn()
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

// newQMIConn establishes connection to qmi-proxy
func newQMIConn() (net.Conn, error) {
	fd, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket: %w", err)
	}
	if err := syscall.Connect(fd, &syscall.SockaddrUnix{Name: "\x00qmi-proxy"}); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("failed to connect to qmi-proxy: %w", err)
	}
	conn, err := net.FileConn(os.NewFile(uintptr(fd), "euicc-go-qmi-proxy"))
	if err != nil {
		return nil, fmt.Errorf("failed to create net.Conn: %w", err)
	}
	return conn, err
}

// openProxyConnection sends a request to the qmi-proxy to open a connection
func (q *QMI) openProxyConnection() error {
	request := core.InternalOpenRequest{
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
		DevicePath:    []byte(q.device),
	}
	err := core.Transmit(q.Transport, request.Request())
	if err == io.EOF {
		return fmt.Errorf("device %s is not connected", q.device)
	}
	return err
}

// allocateClientID sends a request to allocate a client ID for UIM service
func (q *QMI) allocateClientID() error {
	request := core.AllocateClientIDRequest{
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
	}
	err := core.Transmit(q.Transport, request.Request())
	if err == io.EOF {
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
	err := core.Transmit(q.Transport, request.Request())
	if err != nil {
		return err
	}
	return nil
}

// Disconnect releases the client ID and closes the connection
func (q *QMI) Disconnect() error {
	if err := q.releaseClientID(); err != nil {
		return err
	}
	return q.conn.Close()
}
