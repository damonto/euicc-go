package mbim

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// Request represents a standard MBIM request
type Request struct {
	MessageType   MessageType
	MessageLength uint32
	TransactionID uint32
	ReadTimeout   time.Duration // Timeout for reading response
	Command       encoding.BinaryMarshaler
	Response      encoding.BinaryUnmarshaler
}

var mutex sync.Mutex

func (r *Request) WriteTo(w net.Conn) (int, error) {
	data, err := r.MarshalBinary()
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

func (r *Request) ReadFrom(c net.Conn) (int, error) {
	if r.ReadTimeout == 0 {
		r.ReadTimeout = 30 * time.Second // Default timeout if not set
	}
	deadline := time.Now().Add(r.ReadTimeout)
	for time.Now().Before(deadline) {
		buf := make([]byte, 4096)
		c.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, err := c.Read(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue // Timeout, try again
			}
			return 0, err
		}
		response := CommandResponse{Response: r.Response}
		if err := response.UnmarshalBinary(buf[:n]); err != nil {
			if r.MessageType != response.MessageType&^0x80000000 {
				continue // Ignore messages from other sources
			}
			return 0, err
		}
		if r.MessageType != response.MessageType&^0x80000000 && r.TransactionID != response.TransactionID {
			continue
		}
		return n, nil
	}
	return 0, fmt.Errorf("transaction ID %d not found in response", r.TransactionID)
}

// Transmit sends the MBIM message and waits for a response
func (r *Request) Transmit(conn net.Conn) error {
	if _, err := r.WriteTo(conn); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	if _, err := r.ReadFrom(conn); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	return nil
}

// MarshalBinary creates binary representation of the MBIM message
func (r *Request) MarshalBinary() ([]byte, error) {
	command, err := r.Command.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	r.MessageLength = uint32(12 + len(command))
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, r.MessageType)
	binary.Write(buf, binary.LittleEndian, r.MessageLength)
	binary.Write(buf, binary.LittleEndian, r.TransactionID)
	if len(command) > 0 {
		buf.Write(command)
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

// CommandResponse represents the response to a command
type CommandResponse struct {
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

// UnmarshalBinary parses binary data into MBIM command response
func (r *CommandResponse) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	binary.Read(buf, binary.LittleEndian, &r.MessageType)
	binary.Read(buf, binary.LittleEndian, &r.MessageLength)
	binary.Read(buf, binary.LittleEndian, &r.TransactionID)
	// If the message length is larger than 16, it contains the fragment and service information (command-done, indicate, etc.)
	if r.MessageLength > 16 {
		binary.Read(buf, binary.LittleEndian, &r.FragmentTotal)
		binary.Read(buf, binary.LittleEndian, &r.FragmentCurrent)
		binary.Read(buf, binary.LittleEndian, &r.ServiceID)
		binary.Read(buf, binary.LittleEndian, &r.CommandID)
	}
	binary.Read(buf, binary.LittleEndian, &r.Status)
	if r.Status != MBIMStatusNone {
		return r.Status
	}
	binary.Read(buf, binary.LittleEndian, &r.ResponseLength)
	r.ResponseBuffer = make([]byte, r.ResponseLength)
	binary.Read(buf, binary.LittleEndian, &r.ResponseBuffer)
	return r.Response.UnmarshalBinary(r.ResponseBuffer)
}
