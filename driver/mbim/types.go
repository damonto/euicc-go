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
	Type          MessageType
	Length        uint32
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
	sourceType := m.Type
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
			if sourceType != m.Type&^0x80000000 {
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
	binary.Read(buf, binary.LittleEndian, &m.Type)
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
	Response        encoding.BinaryUnmarshaler
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

// CommandDoneResponse represents the response to a command
type CommandDoneResponse struct {
	Type            MessageType
	Length          uint32
	TransactionID   uint32
	FragmentTotal   uint32
	FragmentCurrent uint32
	Service         [16]byte
	CID             uint32
	Status          MBIMStatus
	Response        encoding.BinaryUnmarshaler
}

// UnmarshalBinary parses binary data into MBIM command done response
func (r *CommandDoneResponse) UnmarshalBinary(data []byte) error {
	if len(data) < 36 {
		return errors.New("command done response data too short")
	}

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.LittleEndian, &r.Type)
	binary.Read(buf, binary.LittleEndian, &r.Length)
	binary.Read(buf, binary.LittleEndian, &r.TransactionID)
	binary.Read(buf, binary.LittleEndian, &r.FragmentTotal)
	binary.Read(buf, binary.LittleEndian, &r.FragmentCurrent)
	binary.Read(buf, binary.LittleEndian, &r.Service)
	binary.Read(buf, binary.LittleEndian, &r.CID)

	if r.Status = MBIMStatus(binary.LittleEndian.Uint32(data[28:32])); r.Status != MBIMStatusNone {
		return r.Status
	}
	valueLen := binary.LittleEndian.Uint32(data[32:36])
	if int(36+valueLen) > len(data) {
		return errors.New("basic response buffer too short")
	}
	return r.Response.UnmarshalBinary(data[36 : 36+valueLen])
}
