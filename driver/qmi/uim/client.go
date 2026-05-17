package uim

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/damonto/euicc-go/driver/qmi/protocol"
)

// Client implements UIM smart card operations over a QMI transport.
type Client struct {
	mu        sync.Mutex
	Transport protocol.Transport
	Slot      uint8
	ClientID  uint8
	TxnID     uint32
	channel   byte
}

const (
	maxAIDLength = 0xff

	sendAPDUFixedTLVLength       = 4 + 5 + 4
	maxTransmitAPDUCommandLength = protocol.MaxQMUXServiceTLVLength - sendAPDUFixedTLVLength
)

// Connect establishes QMI session and allocates UIM client ID
func (q *Client) Connect() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if err := q.ensureSlotActivated(); err != nil {
		return err
	}
	// In QMI mode, we need to keep the SIM slot set to 1, because once the
	// configured slot becomes active, it will be assigned as slot 1.
	q.Slot = 1
	return nil
}

// ensureSlotActivated checks if the desired slot is activated and activates it if necessary
func (q *Client) ensureSlotActivated() error {
	slot, err := q.currentActivatedSlot()
	if err != nil {
		// Some older devices do not support the GetSlotStatusRequest QMI command
		if errors.Is(err, protocol.QMIErrorNotSupported) {
			return nil
		}
		return err
	}
	if slot == q.Slot {
		return nil
	}
	if err := q.switchSlot(); err != nil {
		return err
	}
	return q.waitForSlotActivation()
}

// waitForSlotActivation waits for the specified slot to be activated
func (q *Client) waitForSlotActivation() error {
	var err error
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for attempt := range 10 {
		if attempt > 0 {
			<-ticker.C
		}
		request := GetCardStatusRequest{
			ClientID:      q.ClientID,
			TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
		}
		err = q.Transport.Transmit(request.Request())
		if err == nil && request.Response.Ready() {
			return nil
		}
	}
	return fmt.Errorf("sim did not become available after slot %d activation err: %w", q.Slot, err)
}

// currentActivatedSlot returns the currently active logical slot
func (q *Client) currentActivatedSlot() (uint8, error) {
	request := GetSlotStatusRequest{
		ClientID:      q.ClientID,
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
	}
	if err := q.Transport.Transmit(request.Request()); err != nil {
		return 0, err
	}
	return request.Response.ActivatedSlot, nil
}

// switchSlot switches to the specified logical and physical slot
func (q *Client) switchSlot() error {
	request := SwitchSlotRequest{
		ClientID:      q.ClientID,
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
		LogicalSlot:   1,
		PhysicalSlot:  uint32(q.Slot),
	}
	return q.Transport.Transmit(request.Request())
}

// OpenLogicalChannel opens a logical channel with the specified AID
func (q *Client) OpenLogicalChannel(AID []byte) (byte, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(AID) > maxAIDLength {
		return 0, fmt.Errorf("AID length %d exceeds QMI limit %d", len(AID), maxAIDLength)
	}

	request := OpenLogicalChannelRequest{
		ClientID:      q.ClientID,
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
		Slot:          q.Slot,
		AID:           AID,
	}
	if err := q.Transport.Transmit(request.Request()); err != nil {
		return 0, err
	}
	q.channel = request.Response.Channel
	return q.channel, nil
}

// CloseLogicalChannel closes the specified logical channel
func (q *Client) CloseLogicalChannel(channel byte) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	request := CloseLogicalChannelRequest{
		ClientID:      q.ClientID,
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
		Channel:       channel,
		Slot:          q.Slot,
	}
	return q.Transport.Transmit(request.Request())
}

// Transmit sends an APDU command (basic channel implementation)
func (q *Client) Transmit(command []byte) ([]byte, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(command) > maxTransmitAPDUCommandLength {
		return nil, fmt.Errorf("APDU command length %d exceeds QMI limit %d", len(command), maxTransmitAPDUCommandLength)
	}

	request := TransmitAPDURequest{
		ClientID:      q.ClientID,
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
		Slot:          q.Slot,
		Channel:       q.channel,
		Command:       command,
	}
	if err := q.Transport.Transmit(request.Request()); err != nil {
		return nil, err
	}
	return request.Response.Response, nil
}
