package mbim

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"syscall"

	"github.com/damonto/euicc-go/apdu"
)

type MBIM struct {
	device  string
	slot    uint8
	conn    net.Conn
	txnID   uint32
	channel uint32
}

// New creates a new MBIM proxy connection to the specified device
func New(device string, slot uint8) (apdu.SmartCardChannel, error) {
	m := &MBIM{
		device: device,
		slot:   slot - 1,
	}
	if err := m.connectToProxy(); err != nil {
		return nil, fmt.Errorf("failed to connect to mbim-proxy: %w", err)
	}
	if err := m.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to device: %w", err)
	}
	return m, nil
}

// connectToProxy establishes connection to mbim-proxy using abstract Unix socket
func (m *MBIM) connectToProxy() error {
	fd, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return fmt.Errorf("failed to create socket: %w", err)
	}
	if err := syscall.Connect(fd, &syscall.SockaddrUnix{Name: "\x00mbim-proxy"}); err != nil {
		syscall.Close(fd)
		return fmt.Errorf("failed to connect to mbim-proxy: %w", err)
	}
	file := os.NewFile(uintptr(fd), "euicc-go")
	m.conn, err = net.FileConn(file)
	if err != nil {
		return fmt.Errorf("failed to create net.Conn: %w", err)
	}
	return nil
}

// Connect establishes MBIM session and opens device
func (m *MBIM) Connect() error {
	if err := m.configureProxy(); err != nil {
		return fmt.Errorf("failed to configure proxy: %w", err)
	}
	if err := m.openDevice(); err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}
	return nil
}

// configureProxy sends proxy configuration request with device path using the libmbim proxy protocol
func (m *MBIM) configureProxy() error {
	request := ProxyConfigRequest{
		TxnID:      atomic.AddUint32(&m.txnID, 1),
		DevicePath: m.device,
		Timeout:    30,
	}
	message := request.Message()
	if _, err := message.WriteTo(m.conn); err != nil {
		return err
	}
	if _, err := message.ReadFrom(m.conn); err != nil {
		return err
	}
	return nil
}

// openDevice sends MBIM Open message to establish connection
func (m *MBIM) openDevice() error {
	request := OpenDeviceRequest{
		TxnID: atomic.AddUint32(&m.txnID, 1),
	}
	message := request.Message()
	if _, err := message.WriteTo(m.conn); err != nil {
		return err
	}
	if _, err := message.ReadFrom(m.conn); err != nil {
		return err
	}
	return nil
}

// OpenLogicalChannel opens a logical channel for the specified Application ID
func (m *MBIM) OpenLogicalChannel(aid []byte) (byte, error) {
	request := OpenLogicalChannelRequest{
		TxnID:       atomic.AddUint32(&m.txnID, 1),
		AppId:       aid,
		SelectP2Arg: 0x00,
		Group:       0x01,
	}
	message := request.Message()
	if _, err := message.WriteTo(m.conn); err != nil {
		return 0, err
	}
	if _, err := message.ReadFrom(m.conn); err != nil {
		return 0, err
	}
	m.channel = request.Response.Channel
	return byte(m.channel), nil
}

// Transmit implements apdu.SmartCardChannel.
func (m *MBIM) Transmit(command []byte) ([]byte, error) {
	request := TransmitAPDURequest{
		TxnID:           atomic.AddUint32(&m.txnID, 1),
		Channel:         m.channel,
		SecureMessaging: 0,
		ClassByteType:   0,
		APDU:            command,
	}
	message := request.Message()
	if _, err := message.WriteTo(m.conn); err != nil {
		return nil, err
	}
	if _, err := message.ReadFrom(m.conn); err != nil {
		return nil, err
	}
	sw := make([]byte, 2)
	binary.LittleEndian.PutUint16(sw, uint16(request.Response.Status&0xFFFF))
	response := append(request.Response.APDU, sw...)
	return response, nil
}

// CloseLogicalChannel closes the specified logical channel
func (m *MBIM) CloseLogicalChannel(channel byte) error {
	request := CloseLogicalChannelRequest{
		TxnID:   atomic.AddUint32(&m.txnID, 1),
		Channel: uint32(channel),
		Group:   1,
	}
	message := request.Message()
	if _, err := message.WriteTo(m.conn); err != nil {
		return err
	}
	if _, err := message.ReadFrom(m.conn); err != nil {
		return err
	}
	return nil
}

// Disconnect closes the MBIM connection and releases resources
func (m *MBIM) Disconnect() error {
	return m.conn.Close()
}
