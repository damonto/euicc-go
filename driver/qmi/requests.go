package qmi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type Request interface {
	Bytes() []byte
	Value(message *Message) ([]byte, error)
}

var mutex sync.Mutex

func sendRequest[I Request](conn net.Conn, txnID uint16, request I) ([]byte, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, err := conn.Write(request.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}
	message, err := waitForResponse(conn, txnID)
	if err != nil {
		return nil, err
	}
	return request.Value(message)
}

func waitForResponse(conn net.Conn, expectedTxnID uint16) (*Message, error) {
	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		buf := make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(100 * time.Microsecond))
		if _, err := conn.Read(buf); err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue // Timeout, try again
			}
			return nil, fmt.Errorf("failed to read from connection: %w", err)
		}
		n := int(binary.LittleEndian.Uint16(buf[1:3])) + 1
		var message Message
		if err := message.UnmarshalBinary(buf[:n]); err != nil {
			return nil, fmt.Errorf("failed to unmarshal message: %w", err)
		}
		if message.TransactionID != expectedTxnID {
			continue // Not the expected transaction ID, keep waiting
		}
		if err := message.Error(); err != nil {
			return nil, err
		}
		return &message, nil
	}
	return nil, fmt.Errorf("timed out waiting for response for transaction ID %d", expectedTxnID)
}

func toBytes(TLVs []TLV) []byte {
	buf := new(bytes.Buffer)
	for _, tlv := range TLVs {
		binary.Write(buf, binary.LittleEndian, tlv.Type)
		binary.Write(buf, binary.LittleEndian, tlv.Len)
		buf.Write(tlv.Value)
	}
	return buf.Bytes()
}

type ControlRequest struct {
	ClientID  uint8
	TxnID     uint8
	MessageID MessageID
	TLVs      []TLV
}

func (r *ControlRequest) Bytes() []byte {
	value := toBytes(r.TLVs)

	headerBuf := new(bytes.Buffer)
	binary.Write(headerBuf, binary.LittleEndian, CTLSDUHeader{
		MessageType:   QMIMessageTypeRequest,
		TransactionID: r.TxnID,
		MessageID:     r.MessageID,
		MessageLength: uint16(len(value)),
	})
	headerBuf.Write(value)

	sduBytes := headerBuf.Bytes()
	requestBuf := new(bytes.Buffer)
	binary.Write(requestBuf, binary.LittleEndian, QMUXHeader{
		IfType:       QMUXHeaderIfType,
		Length:       uint16(len(sduBytes) + 5),
		ControlFlags: QMUXHeaderControlFlagRequest,
		ServiceType:  QMIServiceCtl,
		ClientID:     r.ClientID,
	})
	requestBuf.Write(sduBytes)
	return requestBuf.Bytes()
}

// region Internal Open Request

type InternalOpenRequest struct {
	TxnID      uint8
	DevicePath []byte
}

func (r *InternalOpenRequest) Bytes() []byte {
	request := ControlRequest{
		TxnID:     r.TxnID,
		MessageID: QMICtlInternalProxyOpen,
		TLVs: []TLV{
			{Type: 0x01, Len: uint16(len(r.DevicePath)), Value: r.DevicePath},
		},
	}
	return request.Bytes()
}

func (r *InternalOpenRequest) Value(message *Message) ([]byte, error) {
	return nil, nil
}

// endregion

// region Allocate/Release Client ID Requests

type AllocateClientIDRequest struct {
	TxnID uint8
}

func (r *AllocateClientIDRequest) Bytes() []byte {
	request := ControlRequest{
		TxnID:     r.TxnID,
		MessageID: QMICtlCmdAllocateClientID,
		TLVs: []TLV{
			{Type: 0x01, Len: 1, Value: []byte{byte(QMIServiceUIM)}},
		},
	}
	return request.Bytes()
}

func (r *AllocateClientIDRequest) Value(message *Message) ([]byte, error) {
	if value, ok := message.TLVs[0x01]; ok && len(value.Value) >= 2 {
		return []byte{value.Value[1]}, nil
	}
	return nil, errors.New("could not find allocated client ID in response")
}

// endregion

// region Release Client ID Request

type ReleaseClientIDRequest struct {
	ClientID uint8
	TxnID    uint8
}

func (r *ReleaseClientIDRequest) Bytes() []byte {
	request := ControlRequest{
		TxnID:     r.TxnID,
		MessageID: QMICtlCmdReleaseClientID,
		TLVs: []TLV{
			{Type: 0x01, Len: 2, Value: []byte{byte(QMIServiceUIM), r.ClientID}},
		},
	}
	return request.Bytes()
}

func (r *ReleaseClientIDRequest) Value(message *Message) ([]byte, error) {
	return nil, nil
}

// endregion

type UIMRequest struct {
	ClientID  uint8
	TxnID     uint16
	MessageID MessageID
	TLVs      []TLV
}

func (r *UIMRequest) Bytes() []byte {
	value := toBytes(r.TLVs)
	headerBuf := new(bytes.Buffer)
	binary.Write(headerBuf, binary.LittleEndian, SDUHeader{
		MessageType:   QMIMessageTypeRequest,
		TransactionID: r.TxnID,
		MessageID:     r.MessageID,
		MessageLength: uint16(len(value)),
	})
	headerBuf.Write(value)
	sduBytes := headerBuf.Bytes()

	requestBuf := new(bytes.Buffer)
	binary.Write(requestBuf, binary.LittleEndian, QMUXHeader{
		IfType:       QMUXHeaderIfType,
		Length:       uint16(len(sduBytes) + 5),
		ControlFlags: QMUXHeaderControlFlagRequest,
		ServiceType:  QMIServiceUIM,
		ClientID:     r.ClientID,
	})
	requestBuf.Write(sduBytes)
	return requestBuf.Bytes()
}

// region Open Logical Channel Request

type OpenLogicalChannelRequest struct {
	ClientID uint8
	TxnID    uint16
	Slot     byte
	AID      []byte
}

func (r *OpenLogicalChannelRequest) Bytes() []byte {
	value := append([]byte{byte(len(r.AID))}, r.AID...)
	request := UIMRequest{
		ClientID:  r.ClientID,
		TxnID:     r.TxnID,
		MessageID: QMIUIMOpenLogicalChannel,
		TLVs: []TLV{
			{Type: 0x10, Len: uint16(len(value)), Value: value},
			{Type: 0x01, Len: 1, Value: []byte{r.Slot}},
		},
	}
	return request.Bytes()
}

func (r *OpenLogicalChannelRequest) Value(message *Message) ([]byte, error) {
	value, err := message.Value()
	if err != nil {
		return nil, err
	}
	return value, nil
}

// endregion

// region Close Logical Channel Request

type CloseLogicalChannelRequest struct {
	ClientID  uint8
	TxnID     uint16
	Slot      byte
	ChannelID byte
}

func (r *CloseLogicalChannelRequest) Bytes() []byte {
	request := UIMRequest{
		ClientID:  r.ClientID,
		TxnID:     r.TxnID,
		MessageID: QMIUIMCloseLogicalChannel,
		TLVs: []TLV{
			{Type: 0x01, Len: 1, Value: []byte{r.Slot}},
			{Type: 0x11, Len: 1, Value: []byte{r.ChannelID}},
			{Type: 0x13, Len: 1, Value: []byte{0x01}},
		},
	}
	return request.Bytes()
}

func (r *CloseLogicalChannelRequest) Value(message *Message) ([]byte, error) {
	return nil, nil
}

// endregion

// region Transmit APDU Request

type TransmitAPDURequest struct {
	ClientID  uint8
	TxnID     uint16
	Slot      byte
	ChannelID byte
	Command   []byte
}

func (r *TransmitAPDURequest) Bytes() []byte {
	length := len(r.Command)
	value := append([]byte{byte(length), byte(length >> 8)}, r.Command...)
	request := UIMRequest{
		ClientID:  r.ClientID,
		TxnID:     r.TxnID,
		MessageID: QMIUIMSendAPDU,
		TLVs: []TLV{
			{Type: 0x10, Len: 1, Value: []byte{r.ChannelID}},
			{Type: 0x02, Len: uint16(len(value)), Value: value},
			{Type: 0x01, Len: 1, Value: []byte{r.Slot}},
		},
	}
	return request.Bytes()
}

func (r *TransmitAPDURequest) Value(message *Message) ([]byte, error) {
	value, err := message.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to extract APDU response: %w", err)
	}
	n := int(value[0]) | (int(value[1]) << 8)
	if len(value) >= 2+n {
		return value[2 : 2+n], nil
	}
	return nil, fmt.Errorf("could not find APDU response in message")
}

// endregion
