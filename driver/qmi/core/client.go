package core

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

// QMIClient implements the apdu.SmartCardChannel interface using QMI protocol
type QMIClient struct {
	Transport Transport
	Slot      uint8
	ClientID  uint8
	TxnID     uint32
	channel   byte
}

// Connect establishes QMI session and allocates UIM client ID
func (q *QMIClient) Connect() error {
	if err := q.ensureSlotActivated(); err != nil {
		return err
	}
	// In QMI mode, we need to keep the SIM slot set to 1, because once the
	// configured slot becomes active, it will be assigned as slot 1.
	q.Slot = 1
	return nil
}

// ensureSlotActivated checks if the desired slot is activated and activates it if necessary
func (q *QMIClient) ensureSlotActivated() error {
	slot, err := q.currentActivatedSlot()
	if err != nil {
		// Some older devices do not support the GetSlotStatusRequest QMI command
		if errors.Is(err, QMIErrorNotSupported) {
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
func (q *QMIClient) waitForSlotActivation() error {
	var err error
	for range 10 {
		request := GetCardStatusRequest{
			ClientID:      q.ClientID,
			TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
		}
		err = Transmit(q.Transport, request.Request())
		if err != nil {
			continue
		}
		if request.Response.Ready() {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("sim did not become available after slot %d activation err: %w", q.Slot, err)
}

// currentActivatedSlot returns the currently active logical slot
func (q *QMIClient) currentActivatedSlot() (uint8, error) {
	request := GetSlotStatusRequest{
		ClientID:      q.ClientID,
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
	}
	if err := Transmit(q.Transport, request.Request()); err != nil {
		return 0, err
	}
	return request.Response.ActivatedSlot, nil
}

// switchSlot switches to the specified logical and physical slot
func (q *QMIClient) switchSlot() error {
	request := SwitchSlotRequest{
		ClientID:      q.ClientID,
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
		LogicalSlot:   1,
		PhysicalSlot:  uint32(q.Slot),
	}
	return Transmit(q.Transport, request.Request())
}

// OpenLogicalChannel opens a logical channel with the specified AID
func (q *QMIClient) OpenLogicalChannel(AID []byte) (byte, error) {
	request := OpenLogicalChannelRequest{
		ClientID:      q.ClientID,
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
		Slot:          q.Slot,
		AID:           AID,
	}
	if err := Transmit(q.Transport, request.Request()); err != nil {
		return 0, err
	}
	q.channel = request.Response.Channel
	return q.channel, nil
}

// CloseLogicalChannel closes the specified logical channel
func (q *QMIClient) CloseLogicalChannel(channel byte) error {
	request := CloseLogicalChannelRequest{
		ClientID:      q.ClientID,
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
		Channel:       channel,
		Slot:          q.Slot,
	}
	return Transmit(q.Transport, request.Request())
}

// Transmit sends an APDU command (basic channel implementation)
func (q *QMIClient) Transmit(command []byte) ([]byte, error) {
	request := TransmitAPDURequest{
		ClientID:      q.ClientID,
		TransactionID: uint16(atomic.AddUint32(&q.TxnID, 1)),
		Slot:          q.Slot,
		Channel:       q.channel,
		Command:       command,
	}
	if err := Transmit(q.Transport, request.Request()); err != nil {
		return nil, err
	}
	return request.Response.Response, nil
}
