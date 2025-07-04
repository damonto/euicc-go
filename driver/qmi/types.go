package qmi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type TLV struct {
	Type  uint8
	Len   uint16
	Value []byte
}

func (t *TLV) Error() error {
	if t.Type != TLVTypeResult {
		return fmt.Errorf("not a result TLV, got type %d", t.Type)
	}
	if len(t.Value) < 4 {
		return fmt.Errorf("result TLV too short, expected 4 bytes, got %d", len(t.Value))
	}
	result := binary.LittleEndian.Uint16(t.Value[0:2])
	if result == uint16(QMIResultSuccess) {
		return nil // No error, success case
	}
	errorCode := binary.LittleEndian.Uint16(t.Value[2:4])
	return QMIError{
		Result:    QMIResult(result),
		ErrorCode: QMIProtocolError(errorCode),
	}
}

// QMUXHeader represents the header of a QMUX PDU
type QMUXHeader struct {
	IfType       uint8
	Length       uint16
	ControlFlags uint8
	ServiceType  ServiceType
	ClientID     uint8
}

// SDUHeader represents the header for non-CTL service messages (2-byte transaction ID)
type SDUHeader struct {
	MessageType   MessageType
	TransactionID uint16
	MessageID     MessageID
	MessageLength uint16
}

// CTLSDUHeader represents the header for CTL service messages (1-byte transaction ID)
type CTLSDUHeader struct {
	MessageType   MessageType
	TransactionID uint8
	MessageID     MessageID
	MessageLength uint16
}

// Message represents a complete parsed QMI message
type Message struct {
	QMUXHeader
	TransactionID uint16
	MessageID     MessageID
	MessageLength uint16
	TLVs          map[uint8]TLV
}

func (m *Message) UnmarshalBinary(data []byte) error {
	if len(data) < 11 {
		return fmt.Errorf("data too short: got %d bytes", len(data))
	}

	reader := bytes.NewReader(data)
	// Read QMUX header
	if err := binary.Read(reader, binary.LittleEndian, &m.QMUXHeader); err != nil {
		return fmt.Errorf("read QMUX header: %w", err)
	}

	// Read message type (not used yet)
	var messageType MessageType
	if err := binary.Read(reader, binary.LittleEndian, &messageType); err != nil {
		return fmt.Errorf("read message type: %w", err)
	}

	// Read transaction ID
	switch m.QMUXHeader.ServiceType {
	case QMIServiceCtl:
		var txnID uint8
		if err := binary.Read(reader, binary.LittleEndian, &txnID); err != nil {
			return fmt.Errorf("read CTL txn ID: %w", err)
		}
		m.TransactionID = uint16(txnID)
	default:
		if err := binary.Read(reader, binary.LittleEndian, &m.TransactionID); err != nil {
			return fmt.Errorf("read txn ID: %w", err)
		}
	}

	// Read message ID and length
	if err := binary.Read(reader, binary.LittleEndian, &m.MessageID); err != nil {
		return fmt.Errorf("read message ID: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &m.MessageLength); err != nil {
		return fmt.Errorf("read message length: %w", err)
	}
	m.TLVs = make(map[uint8]TLV)
	if m.MessageLength > 0 {
		return m.toTVLs(io.LimitReader(reader, int64(m.MessageLength)))
	}
	return nil
}

func (m *Message) toTVLs(r io.Reader) error {
	for {
		var t uint8
		if err := binary.Read(r, binary.LittleEndian, &t); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("read TLV type: %w", err)
		}

		var n uint16
		if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
			return fmt.Errorf("read TLV length: %w", err)
		}

		v := make([]byte, n)
		if _, err := io.ReadFull(r, v); err != nil {
			return fmt.Errorf("read TLV value: %w", err)
		}

		m.TLVs[t] = TLV{Type: t, Len: n, Value: v}
	}
	return nil
}

func (m *Message) Value() ([]byte, error) {
	tlv, ok := m.TLVs[0x10]
	if !ok {
		return nil, errors.New("no value TLV found")
	}
	if len(tlv.Value) == 0 {
		return nil, errors.New("value TLV is empty")
	}
	return tlv.Value, nil
}

func (m *Message) Error() error {
	tlv, ok := m.TLVs[TLVTypeResult]
	if !ok {
		return errors.New("no result TLV found")
	}
	return tlv.Error()
}
