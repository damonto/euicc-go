package qmi

import (
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/damonto/euicc-go/apdu"
)

// QMI implements the apdu.SmartCardChannel interface using QMI protocol
type QMI struct {
	device  string
	slot    uint8
	conn    net.Conn
	cid     uint8
	txnID   uint32
	channel byte
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
	file := os.NewFile(uintptr(fd), "euicc-go-qmi-proxy")
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
	if err := q.ensureSlotActivated(); err != nil {
		return fmt.Errorf("failed to ensure slot is activated: %w", err)
	}
	return nil
}

// ensureSlotActivated checks if the desired slot is activated and activates it if necessary
func (q *QMI) ensureSlotActivated() error {
	slot, err := q.currentActivatedSlot()
	if err != nil {
		return fmt.Errorf("failed to get active slot: %w", err)
	}
	if slot == q.slot {
		return nil
	}
	if err := q.switchSlot(); err != nil {
		return fmt.Errorf("failed to switch slot: %w", err)
	}
	if err := q.waitForSlotActivation(); err != nil {
		return fmt.Errorf("failed to wait for slot activation: %w", err)
	}
	return nil
}

// waitForSlotActivation waits for the specified slot to be activated
func (q *QMI) waitForSlotActivation() error {
	for range 30 {
		slot, err := q.currentActivatedSlot()
		if err != nil {
			continue
		}
		if slot == q.slot {
			// Wait for a short period to ensure the slot is fully activated.
			time.Sleep(1 * time.Second)
			return nil
		}
	}
	return fmt.Errorf("failed to activate slot %d", q.slot)
}

// currentActivatedSlot returns the currently active logical slot
func (q *QMI) currentActivatedSlot() (uint8, error) {
	request := GetSlotStatusRequest{
		ClientID:      q.cid,
		TransactionID: uint16(atomic.AddUint32(&q.txnID, 1)),
	}
	if err := request.Request().Transmit(q.conn); err != nil {
		return 0, fmt.Errorf("failed to send get slot status request: %w", err)
	}
	return request.Response.ActivatedSlot, nil
}

// switchSlot switches to the specified logical and physical slot
func (q *QMI) switchSlot() error {
	request := SwitchSlotRequest{
		ClientID:      q.cid,
		TransactionID: uint16(atomic.AddUint32(&q.txnID, 1)),
		LogicalSlot:   1,
		PhysicalSlot:  uint32(q.slot),
	}
	return request.Request().Transmit(q.conn)
}

// openProxyConnection sends a request to the qmi-proxy to open a connection
func (q *QMI) openProxyConnection() error {
	request := InternalOpenRequest{
		TransactionID: uint16(atomic.AddUint32(&q.txnID, 1)),
		DevicePath:    []byte(q.device),
	}
	return request.Request().Transmit(q.conn)
}

// allocateClientID sends a request to allocate a client ID for UIM service
func (q *QMI) allocateClientID() error {
	request := AllocateClientIDRequest{
		TransactionID: uint16(atomic.AddUint32(&q.txnID, 1)),
	}
	if err := request.Request().Transmit(q.conn); err != nil {
		return fmt.Errorf("failed to send allocate client ID request: %w", err)
	}
	q.cid = request.Response.ClientID
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
	request := ReleaseClientIDRequest{
		ClientID:      q.cid,
		TransactionID: uint16(atomic.AddUint32(&q.txnID, 1)),
	}
	if err := request.Request().Transmit(q.conn); err != nil {
		return fmt.Errorf("failed to send release client ID request: %w", err)
	}
	return nil
}

// OpenLogicalChannel opens a logical channel with the specified AID
func (q *QMI) OpenLogicalChannel(AID []byte) (byte, error) {
	request := OpenLogicalChannelRequest{
		ClientID:      q.cid,
		TransactionID: uint16(atomic.AddUint32(&q.txnID, 1)),
		Slot:          q.slot,
		AID:           AID,
	}
	if err := request.Request().Transmit(q.conn); err != nil {
		return 0, fmt.Errorf("failed to send open logical channel request: %w", err)
	}
	q.channel = request.Response.Channel
	return q.channel, nil
}

// CloseLogicalChannel closes the specified logical channel
func (q *QMI) CloseLogicalChannel(channel byte) error {
	request := CloseLogicalChannelRequest{
		ClientID:      q.cid,
		TransactionID: uint16(atomic.AddUint32(&q.txnID, 1)),
		Channel:       channel,
		Slot:          q.slot,
	}
	return request.Request().Transmit(q.conn)
}

// Transmit sends an APDU command (basic channel implementation)
func (q *QMI) Transmit(command []byte) ([]byte, error) {
	request := TransmitAPDURequest{
		ClientID:      q.cid,
		TransactionID: uint16(atomic.AddUint32(&q.txnID, 1)),
		Slot:          q.slot,
		Channel:       q.channel,
		Command:       command,
	}
	if err := request.Request().Transmit(q.conn); err != nil {
		return nil, fmt.Errorf("failed to send transmit APDU request: %w", err)
	}
	return request.Response.Response, nil
}
