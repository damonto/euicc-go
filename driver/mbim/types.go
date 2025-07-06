package mbim

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type Payload interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// Message represents a standard MBIM message
type Message struct {
	MessageType   MessageType
	MessageLength uint32
	TransactionID uint32
	ReadTimeout   time.Duration // Timeout for reading response
	Payload       Payload
}

var mutex sync.Mutex

func (m *Message) WriteTo(w net.Conn) (int, error) {
	data, err := m.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("failed to marshal message: %w", err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	n, err := w.Write(data)
	if err != nil {
		return n, fmt.Errorf("failed to write message: %w", err)
	}
	return n, nil
}

func (m *Message) ReadFrom(r net.Conn) (int, error) {
	sourceTransactionID := m.TransactionID
	sourceMessageType := m.MessageType
	if m.ReadTimeout == 0 {
		m.ReadTimeout = 30 * time.Second // Default timeout if not set
	}
	deadline := time.Now().Add(m.ReadTimeout)
	for time.Now().Before(deadline) {
		buf := make([]byte, 4096)
		r.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, err := r.Read(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue // Timeout, try again
			}
			return 0, err
		}
		if err := m.UnmarshalBinary(buf[:n]); err != nil {
			if sourceMessageType != m.MessageType&^0x80000000 {
				continue // Ignore messages from other sources
			}
			return 0, err
		}
		if m.TransactionID != sourceTransactionID {
			continue
		}
		return n, nil
	}
	return 0, fmt.Errorf("transaction ID %d not found in response", sourceTransactionID)
}

// Transmit sends the MBIM message and waits for a response
func (m *Message) Transmit(conn net.Conn) error {
	if _, err := m.WriteTo(conn); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	if _, err := m.ReadFrom(conn); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	return nil
}

// UnmarshalBinary parses a binary MBIM message
func (m *Message) UnmarshalBinary(data []byte) error {
	if len(data) < 12 {
		return errors.New("message too short for MBIM header")
	}
	buf := bytes.NewReader(data)
	binary.Read(buf, binary.LittleEndian, &m.MessageType)
	binary.Read(buf, binary.LittleEndian, &m.MessageLength)
	binary.Read(buf, binary.LittleEndian, &m.TransactionID)
	return m.Payload.UnmarshalBinary(data)
}

// MarshalBinary creates binary representation of the MBIM message
func (m *Message) MarshalBinary() ([]byte, error) {
	payload, err := m.Payload.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	m.MessageLength = uint32(12 + len(payload))
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, m.MessageType)
	binary.Write(buf, binary.LittleEndian, m.MessageLength)
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
	ServiceID       [16]byte
	CommandID       uint32
	CommandType     uint32 // 0=Query, 1=Set
	DataLength      uint32
	Data            []byte
	Response        encoding.BinaryUnmarshaler
}

// MarshalBinary creates binary representation of the MBIM command
func (c *Command) MarshalBinary() ([]byte, error) {
	c.DataLength = uint32(len(c.Data))
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, c.FragmentTotal)
	binary.Write(buf, binary.LittleEndian, c.FragmentCurrent)
	binary.Write(buf, binary.LittleEndian, c.ServiceID)
	binary.Write(buf, binary.LittleEndian, c.CommandID)
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

// CommandDoneResponse represents the response to a command
type CommandDoneResponse struct {
	MessageType     MessageType
	MessageLength   uint32
	TransactionID   uint32
	FragmentTotal   uint32
	FragmentCurrent uint32
	ServiceID       [16]byte
	CommandID       uint32
	Status          MBIMStatus
	ResponseLength  uint32
	ResponseBuffer  []byte
	Response        encoding.BinaryUnmarshaler
}

// UnmarshalBinary parses binary data into MBIM command done response
func (r *CommandDoneResponse) UnmarshalBinary(data []byte) error {
	if len(data) < 36 {
		return errors.New("command done response data too short")
	}
	buf := bytes.NewReader(data)
	binary.Read(buf, binary.LittleEndian, &r.MessageType)
	binary.Read(buf, binary.LittleEndian, &r.MessageLength)
	binary.Read(buf, binary.LittleEndian, &r.TransactionID)
	binary.Read(buf, binary.LittleEndian, &r.FragmentTotal)
	binary.Read(buf, binary.LittleEndian, &r.FragmentCurrent)
	binary.Read(buf, binary.LittleEndian, &r.ServiceID)
	binary.Read(buf, binary.LittleEndian, &r.CommandID)
	binary.Read(buf, binary.LittleEndian, &r.Status)
	if r.Status != MBIMStatusNone {
		return r.Status
	}
	binary.Read(buf, binary.LittleEndian, &r.ResponseLength)
	r.ResponseBuffer = make([]byte, r.ResponseLength)
	binary.Read(buf, binary.LittleEndian, &r.ResponseBuffer)
	return r.Response.UnmarshalBinary(r.ResponseBuffer)
}
