package mbim

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/damonto/euicc-go/apdu"
)

var (
	slotActivationPollInterval = 500 * time.Millisecond
	slotActivationAttempts     = 10
)

// MBIM implements the apdu.SmartCardChannel interface using MBIM protocol
type MBIM struct {
	mu      sync.Mutex
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
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: "\x00mbim-proxy", Net: "unix"})
	if err != nil {
		return nil, err
	}
	m := &MBIM{
		device: device,
		slot:   slot - 1, // Convert to 0-based
		conn:   conn,
	}
	return m, nil
}

// Connect establishes MBIM session and opens device
func (m *MBIM) Connect() (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	defer func() {
		if err != nil {
			err = errors.Join(err, m.conn.Close())
		}
	}()
	if err := m.configureProxy(); err != nil {
		return fmt.Errorf("configure proxy: %w", err)
	}
	if err := m.openDevice(); err != nil {
		return fmt.Errorf("open device: %w", err)
	}
	if err := m.ensureSlotActivated(); err != nil {
		return fmt.Errorf("ensure slot is activated: %w", err)
	}
	return nil
}

// ensureSlotActivated checks if the desired slot is activated and activates it if necessary
func (m *MBIM) ensureSlotActivated() error {
	slot, err := m.currentActivatedSlot()
	if err != nil {
		return err
	}
	if slot == m.slot {
		return nil
	}
	if err := m.activateSlot(m.slot); err != nil {
		return err
	}
	return m.waitForSlotActivation()
}

// currentActivatedSlot queries the current active slot mapping
func (m *MBIM) currentActivatedSlot() (uint8, error) {
	request := DeviceSlotMappingsRequest{
		TransactionID: atomic.AddUint32(&m.txnID, 1),
		MapCount:      0, // Query operation
	}
	if err := request.Request().Transmit(m.conn); err != nil {
		return 0, err
	}
	if len(request.Response.SlotMappings) == 0 {
		return 0, errors.New("no slot mappings found")
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
		return err
	}
	return nil
}

// waitForSlotActivation waits for the slot to become active by checking subscriber ready status
func (m *MBIM) waitForSlotActivation() error {
	var err error
	var lastReadyState uint32
	var sawReadyState bool
	ticker := time.NewTicker(slotActivationPollInterval)
	defer ticker.Stop()

	for attempt := range slotActivationAttempts {
		if attempt > 0 {
			<-ticker.C
		}
		request := SubscriberReadyStatusRequest{
			TransactionID: atomic.AddUint32(&m.txnID, 1),
		}
		err = request.Request().Transmit(m.conn)
		if err == nil {
			sawReadyState = true
			lastReadyState = request.Response.ReadyState
			if lastReadyState == MBIMSubscriberReadyStateInitialized || lastReadyState == MBIMSubscriberReadyStateNoEsimProfile {
				return nil
			}
		}
	}
	if err != nil {
		if sawReadyState {
			return fmt.Errorf("sim did not become available after slot %d activation, last ready state %#x: %w", m.slot, lastReadyState, err)
		}
		return fmt.Errorf("sim did not become available after slot %d activation: %w", m.slot, err)
	}
	return fmt.Errorf("sim did not become available after slot %d activation, last ready state %#x", m.slot, lastReadyState)
}

// configureProxy sends proxy configuration request with device path using the libmbim proxy protocol
func (m *MBIM) configureProxy() error {
	request := ProxyConfigRequest{
		TransactionID: atomic.AddUint32(&m.txnID, 1),
		DevicePath:    m.device,
		Timeout:       30,
	}
	err := request.Request().Transmit(m.conn)
	if errors.Is(err, io.EOF) {
		return fmt.Errorf("device %s is not connected", m.device)
	}
	return err
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
	m.mu.Lock()
	defer m.mu.Unlock()

	request := OpenLogicalChannelRequest{
		TransactionID: atomic.AddUint32(&m.txnID, 1),
		AppId:         AID,
		SelectP2Arg:   0,
		Group:         1,
	}
	if err := request.Request().Transmit(m.conn); err != nil {
		return 0, err
	}
	if request.Response.Status != 0 {
		return 0, fmt.Errorf("open logical channel status %#x", request.Response.Status)
	}
	m.channel = request.Response.Channel
	return byte(m.channel), nil
}

// Transmit implements apdu.SmartCardChannel.
func (m *MBIM) Transmit(command []byte) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	request := TransmitAPDURequest{
		TransactionID:   atomic.AddUint32(&m.txnID, 1),
		Channel:         m.channel,
		SecureMessaging: 0,
		ClassByteType:   0,
		APDU:            command,
	}
	if err := request.Request().Transmit(m.conn); err != nil {
		return nil, err
	}
	status := []byte{byte(request.Response.Status >> 8), byte(request.Response.Status)}
	response := append(request.Response.Response, status...)
	return response, nil
}

// CloseLogicalChannel closes the specified logical channel
func (m *MBIM) CloseLogicalChannel(channel byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	request := CloseLogicalChannelRequest{
		TransactionID: atomic.AddUint32(&m.txnID, 1),
		Channel:       uint32(channel),
		Group:         1,
	}
	if err := request.Request().Transmit(m.conn); err != nil {
		return err
	}
	if request.Response.Status != 0 {
		return fmt.Errorf("close logical channel status %#x", request.Response.Status)
	}
	return nil
}

// Disconnect closes the MBIM connection and releases resources
func (m *MBIM) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	request := CloseRequest{
		TransactionID: atomic.AddUint32(&m.txnID, 1),
	}
	err := request.Request().Transmit(m.conn)
	return errors.Join(err, m.conn.Close())
}
