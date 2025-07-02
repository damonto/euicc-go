package qmi

import (
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"syscall"

	"github.com/damonto/euicc-go/apdu"
)

// QMI implements the apdu.SmartCardChannel interface using QMI protocol
type QMI struct {
	device    string
	slot      uint8
	conn      net.Conn
	cid       uint8
	txnID     uint32
	channelID byte
}

// New creates a new QMI connection to the specified device
func New(device string, slot uint8) (apdu.SmartCardChannel, error) {
	q := &QMI{
		device: device,
		slot:   slot,
	}
	if err := q.connectToProxy(); err != nil {
		return nil, fmt.Errorf("failed to connect to qmi-proxy: %w", err)
	}
	return q, nil
}

// connectToProxy establishes connection to qmi-proxy
func (q *QMI) connectToProxy() error {
	fd, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return fmt.Errorf("failed to create socket: %w", err)
	}

	if err := syscall.Connect(fd, &syscall.SockaddrUnix{Name: "\x00qmi-proxy"}); err != nil {
		syscall.Close(fd)
		return fmt.Errorf("failed to connect to qmi-proxy: %w", err)
	}

	file := os.NewFile(uintptr(fd), "qmi-connection")
	q.conn, err = net.FileConn(file)
	if err != nil {
		return fmt.Errorf("failed to create net.Conn: %w", err)
	}
	return nil
}

// Connect establishes QMI session and allocates UIM client ID
func (q *QMI) Connect() error {
	if err := q.openProxyConnection(); err != nil {
		return fmt.Errorf("failed to open proxy connection: %w", err)
	}
	if err := q.allocateClientID(); err != nil {
		return fmt.Errorf("failed to allocate client ID: %w", err)
	}
	return nil
}

// openProxyConnection sends a request to the qmi-proxy to open a connection
func (q *QMI) openProxyConnection() error {
	txnID := uint16(atomic.AddUint32(&q.txnID, 1))
	_, err := sendRequest(q.conn, txnID, &InternalOpenRequest{
		TxnID:      uint8(txnID),
		DevicePath: []byte(q.device),
	})
	return err
}

// allocateClientID sends a request to allocate a client ID for UIM service
func (q *QMI) allocateClientID() error {
	txnID := uint16(atomic.AddUint32(&q.txnID, 1))
	response, err := sendRequest(q.conn, txnID, &AllocateClientIDRequest{
		TxnID: uint8(txnID),
	})
	if err != nil {
		return fmt.Errorf("failed to allocate client ID: %w", err)
	}
	if len(response) < 1 {
		return fmt.Errorf("invalid response for client ID allocation")
	}
	q.cid = response[0]
	return nil
}

// Disconnect releases the client ID and closes the connection
func (q *QMI) Disconnect() error {
	if err := q.releaseClientID(); err != nil {
		return fmt.Errorf("failed to release client ID: %w", err)
	}
	return q.conn.Close()
}

// releaseClientID sends a request to release the allocated client ID
func (q *QMI) releaseClientID() error {
	txnID := uint16(atomic.AddUint32(&q.txnID, 1))
	_, err := sendRequest(q.conn, txnID, &ReleaseClientIDRequest{
		ClientID: q.cid,
		TxnID:    uint8(txnID),
	})
	return err
}

// OpenLogicalChannel opens a logical channel with the specified AID
func (q *QMI) OpenLogicalChannel(aid []byte) (byte, error) {
	txnID := uint16(atomic.AddUint32(&q.txnID, 1))
	request := OpenLogicalChannelRequest{
		ClientID: q.cid,
		TxnID:    txnID,
		Slot:     q.slot,
		AID:      aid,
	}
	response, err := sendRequest(q.conn, txnID, &request)
	if err != nil {
		return 0, fmt.Errorf("failed to open logical channel: %w", err)
	}
	if len(response) < 1 {
		return 0, fmt.Errorf("invalid response for logical channel open")
	}
	q.channelID = response[0]
	return q.channelID, nil
}

// CloseLogicalChannel closes the specified logical channel
func (q *QMI) CloseLogicalChannel(channelID byte) error {
	txnID := uint16(atomic.AddUint32(&q.txnID, 1))
	request := CloseLogicalChannelRequest{
		ClientID:  q.cid,
		TxnID:     txnID,
		ChannelID: channelID,
		Slot:      q.slot,
	}
	if _, err := sendRequest(q.conn, txnID, &request); err != nil {
		return fmt.Errorf("failed to close logical channel: %w", err)
	}
	return nil
}

// Transmit sends an APDU command (basic channel implementation)
func (q *QMI) Transmit(command []byte) ([]byte, error) {
	txnID := uint16(atomic.AddUint32(&q.txnID, 1))
	request := TransmitAPDURequest{
		ClientID:  q.cid,
		TxnID:     txnID,
		Slot:      q.slot,
		ChannelID: q.channelID,
		Command:   command,
	}
	response, err := sendRequest(q.conn, txnID, &request)
	if err != nil {
		return nil, fmt.Errorf("failed to transmit APDU: %w", err)
	}
	if len(response) < 2 {
		return nil, fmt.Errorf("invalid response for APDU transmission")
	}
	return response, nil
}
