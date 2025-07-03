package mbim

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"slices"
	"time"
)

type Marshaller interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
}

type Unmarshaller interface {
	UnmarshalBinary(data []byte) error
}

// Message represents a standard MBIM message
type Message struct {
	Type          MessageType
	Length        uint32
	TransactionID uint32
	Payload       Marshaller
}

func (m *Message) WriteTo(w net.Conn) (int, error) {
	data, err := m.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("failed to marshal message: %w", err)
	}
	n, err := w.Write(data)
	if err != nil {
		return n, fmt.Errorf("failed to write message: %w", err)
	}
	return n, nil
}

func (m *Message) ReadFrom(r net.Conn) (int, error) {
	txnId := m.TransactionID
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		buf := make([]byte, 4096)
		n, err := r.Read(buf)
		if err != nil {
			return 0, fmt.Errorf("failed to read message: %w", err)
		}
		if err := m.UnmarshalBinary(buf[:n]); err != nil {
			return 0, fmt.Errorf("failed to unmarshal message: %w", err)
		}
		if m.TransactionID != txnId {
			continue
		}
		return n, nil
	}
	return 0, fmt.Errorf("transaction ID %d not found in response", txnId)
}

// UnmarshalBinary parses a binary MBIM message
func (m *Message) UnmarshalBinary(data []byte) error {
	if len(data) < 12 {
		return errors.New("message too short for MBIM header")
	}
	buf := bytes.NewReader(data)
	var msgType uint32
	if err := binary.Read(buf, binary.LittleEndian, &msgType); err != nil {
		return err
	}
	m.Type = MessageType(msgType)

	// Parse Length
	binary.Read(buf, binary.LittleEndian, &m.Length)
	// The Length field includes the 12-byte header, so payload = Length - 12
	if m.Length < 12 {
		return fmt.Errorf("invalid message length: %d (must be at least 12 for header)", m.Length)
	}
	// Parse Transaction ID
	binary.Read(buf, binary.LittleEndian, &m.TransactionID)

	return m.Payload.UnmarshalBinary(data[12 : 12+int(m.Length)-12])
}

// MarshalBinary creates binary representation of the MBIM message
func (m *Message) MarshalBinary() ([]byte, error) {
	payload, err := m.Payload.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	m.Length = uint32(12 + len(payload))
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, m.Type)
	binary.Write(buf, binary.LittleEndian, m.Length)
	binary.Write(buf, binary.LittleEndian, m.TransactionID)
	if len(payload) > 0 {
		buf.Write(payload)
	}
	return buf.Bytes(), nil
}

// Command represents an MBIM command message payload
type Command struct {
	FragmentTotal   uint32
	FragmentCurrent uint32
	Service         [16]byte
	CID             uint32
	CommandType     uint32 // 0=Query, 1=Set
	DataLength      uint32
	Data            []byte
	Response        Unmarshaller
}

// MarshalBinary creates binary representation of the MBIM command
func (c *Command) MarshalBinary() ([]byte, error) {
	c.DataLength = uint32(len(c.Data))
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, c.FragmentTotal)
	binary.Write(buf, binary.LittleEndian, c.FragmentCurrent)
	binary.Write(buf, binary.LittleEndian, c.Service)
	binary.Write(buf, binary.LittleEndian, c.CID)
	binary.Write(buf, binary.LittleEndian, c.CommandType)
	binary.Write(buf, binary.LittleEndian, c.DataLength)
	if len(c.Data) > 0 {
		buf.Write(c.Data)
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary parses binary data into MBIM command
func (c *Command) UnmarshalBinary(data []byte) error {
	response := CommandDoneResponse{
		Response: c.Response,
	}
	return response.UnmarshalBinary(data)
}

type CommandDoneResponse struct {
	Type            MessageType
	Length          uint32
	TransactionID   uint32
	FragmentTotal   uint32
	FragmentCurrent uint32
	Service         [16]byte
	CID             uint32
	Status          MBIMStatusError
	Response        Unmarshaller
}

func (r *CommandDoneResponse) UnmarshalBinary(data []byte) error {
	if len(data) < 40 {
		return errors.New("command done response data too short")
	}

	buf := bytes.NewReader(data)

	var messageType uint32
	binary.Read(buf, binary.LittleEndian, &messageType)
	r.Type = MessageType(messageType)

	binary.Read(buf, binary.LittleEndian, &r.Length)
	binary.Read(buf, binary.LittleEndian, &r.TransactionID)
	binary.Read(buf, binary.LittleEndian, &r.FragmentTotal)
	binary.Read(buf, binary.LittleEndian, &r.FragmentCurrent)
	binary.Read(buf, binary.LittleEndian, &r.Service)
	binary.Read(buf, binary.LittleEndian, &r.CID)

	if isExtendedService(r.Service) {
		// Extended service: read 16-byte GUID
		if len(data) < 48 {
			return errors.New("extended command done message too short")
		}
		if r.Status = MBIMStatusError(binary.LittleEndian.Uint32(data[40:44])); r.Status != MBIMStatusErrorNone {
			return r.Status
		}
		valueLen := binary.LittleEndian.Uint32(data[44:48])
		if int(48+valueLen) > len(data) {
			return errors.New("extended response buffer too short")
		}
		return r.Response.UnmarshalBinary(data[48 : 48+valueLen])
	} else {
		// Basic service: 4-byte Service ID
		if len(data) < 36 {
			return errors.New("basic command done message too short")
		}
		if r.Status = MBIMStatusError(binary.LittleEndian.Uint32(data[28:32])); r.Status != MBIMStatusErrorNone {
			return r.Status
		}
		valueLen := binary.LittleEndian.Uint32(data[32:36])
		if int(36+valueLen) > len(data) {
			return errors.New("basic response buffer too short")
		}
		return r.Response.UnmarshalBinary(data[36 : 36+valueLen])
	}
}

func isExtendedService(service [16]byte) bool {
	known := [][16]byte{
		ServiceMsUiccLowLevelAccess,
		ServiceMsBasicConnectExtensions,
	}
	return slices.Contains(known, service)
}
