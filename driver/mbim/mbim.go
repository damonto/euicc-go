package mbim

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"syscall"
	"time"

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
	if slot == 0 {
		return nil, fmt.Errorf("slot must be >= 1")
	}
	m := &MBIM{
		device: device,
		slot:   slot - 1, // Convert to 0-based
	}
	if err := m.connectToProxy(); err != nil {
		return nil, fmt.Errorf("failed to connect to mbim-proxy: %w", err)
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
	file := os.NewFile(uintptr(fd), "euicc-go-mbim-proxy")
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
	if err := m.ensureSlotActivated(); err != nil {
		return fmt.Errorf("failed to ensure slot is activated: %w", err)
	}
	return nil
}

// ensureSlotActivated checks if the desired slot is activated and activates it if necessary
func (m *MBIM) ensureSlotActivated() error {
	slot, err := m.currentActivatedSlot()
	if err != nil {
		return fmt.Errorf("failed to get current active slot: %w", err)
	}
	if slot == m.slot {
		return nil
	}
	if err := m.activateSlot(m.slot); err != nil {
		return fmt.Errorf("failed to activate slot %d: %w", m.slot, err)
	}
	if err := m.waitForSlotActivation(); err != nil {
		return fmt.Errorf("failed to wait for slot activation: %w", err)
	}
	return nil
}

// currentActivatedSlot queries the current active slot mapping
func (m *MBIM) currentActivatedSlot() (uint8, error) {
	request := DeviceSlotMappingsRequest{
		TransactionID: atomic.AddUint32(&m.txnID, 1),
		MapCount:      0, // Query operation
	}
	if err := request.Request().Transmit(m.conn); err != nil {
		return 0, fmt.Errorf("failed to query slot mappings: %w", err)
	}
	if len(request.Response.SlotMappings) == 0 {
		return 0, fmt.Errorf("no slot mappings found")
	}
	return uint8(request.Response.SlotMappings[0].Slot), nil
}

// activateSlot sets the device to use the specified slot
func (m *MBIM) activateSlot(slot uint8) error {
	request := DeviceSlotMappingsRequest{
		TransactionID: atomic.AddUint32(&m.txnID, 1),
		MapCount:      1,
		SlotMappings: []SlotMapping{
			{Slot: uint32(slot)},
		},
	}
	if err := request.Request().Transmit(m.conn); err != nil {
		return fmt.Errorf("failed to set slot mapping: %w", err)
	}
	return nil
}

// waitForSlotActivation waits for the slot to become active by checking subscriber ready status
func (m *MBIM) waitForSlotActivation() error {
	for range 30 {
		request := SubscriberReadyStatusRequest{
			TransactionID: atomic.AddUint32(&m.txnID, 1),
		}
		if err := request.Request().Transmit(m.conn); err != nil {
			continue // Ignore errors, retry
		}
		readyState := request.Response.ReadyState
		if readyState == MBIMSubscriberReadyStateInitialized || readyState == MBIMSubscriberReadyStateNoEsimProfile {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("sim did not become available after slot %d activation", m.slot)
}

// configureProxy sends proxy configuration request with device path using the libmbim proxy protocol
func (m *MBIM) configureProxy() error {
	request := ProxyConfigRequest{
		TransactionID: atomic.AddUint32(&m.txnID, 1),
		DevicePath:    m.device,
		Timeout:       30,
	}
	return request.Request().Transmit(m.conn)
}

// openDevice sends MBIM Open message to establish connection
func (m *MBIM) openDevice() error {
	request := OpenDeviceRequest{
		TransactionID: atomic.AddUint32(&m.txnID, 1),
	}
	return request.Request().Transmit(m.conn)
}

// OpenLogicalChannel opens a logical channel for the specified Application ID
func (m *MBIM) OpenLogicalChannel(AID []byte) (byte, error) {
	request := OpenLogicalChannelRequest{
		TransactionID: atomic.AddUint32(&m.txnID, 1),
		AppId:         AID,
		SelectP2Arg:   0,
		Group:         1,
	}
	if err := request.Request().Transmit(m.conn); err != nil {
		return 0, fmt.Errorf("failed to open logical channel: %w", err)
	}
	m.channel = request.Response.Channel
	return byte(m.channel), nil
}

// Transmit implements apdu.SmartCardChannel.
func (m *MBIM) Transmit(command []byte) ([]byte, error) {
	request := TransmitAPDURequest{
		TransactionID:   atomic.AddUint32(&m.txnID, 1),
		Channel:         m.channel,
		SecureMessaging: 0,
		ClassByteType:   0,
		APDU:            command,
	}
	if err := request.Request().Transmit(m.conn); err != nil {
		return nil, fmt.Errorf("failed to transmit APDU: %w", err)
	}
	sw := make([]byte, 2)
	binary.LittleEndian.PutUint16(sw, uint16(request.Response.Status&0xFFFF))
	response := append(request.Response.Response, sw...)
	return response, nil
}

// CloseLogicalChannel closes the specified logical channel
func (m *MBIM) CloseLogicalChannel(channel byte) error {
	request := CloseLogicalChannelRequest{
		TransactionID: atomic.AddUint32(&m.txnID, 1),
		Channel:       uint32(channel),
		Group:         1,
	}
	return request.Request().Transmit(m.conn)
}

// Disconnect closes the MBIM connection and releases resources
func (m *MBIM) Disconnect() error {
	return m.conn.Close()
}
