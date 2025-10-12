package mbim

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// Request represents a standard MBIM request
type Request struct {
	MessageType   MessageType
	MessageLength uint32
	TransactionID uint32
	ReadTimeout   time.Duration
	Command       encoding.BinaryMarshaler
	Response      encoding.BinaryUnmarshaler
}

func (r *Request) WriteTo(w net.Conn) (int64, error) {
	data, err := r.MarshalBinary()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	if err != nil {
		return int64(n), err
	}
	return int64(n), nil
}

func (r *Request) ReadFrom(c net.Conn) (int64, error) {
	if r.ReadTimeout == 0 {
		r.ReadTimeout = 30 * time.Second
	}
	deadline := time.Now().Add(r.ReadTimeout)
	for time.Now().Before(deadline) {
		c.SetReadDeadline(time.Now().Add(1 * time.Second))

		header := make([]byte, 12)
		if _, err := io.ReadAtLeast(c, header, 12); err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return 0, err
		}

		length := binary.LittleEndian.Uint32(header[4:8])
		buf := make([]byte, length)
		copy(buf[:12], header)
		if _, err := io.ReadFull(c, buf[12:]); err != nil {
			return 0, err
		}

		messageType := binary.LittleEndian.Uint32(header[0:4])
		transactionID := binary.LittleEndian.Uint32(header[8:12])
		if messageType&^0x80000000 != uint32(r.MessageType) || transactionID != r.TransactionID {
			continue
		}

		response := CommandResponse{Response: r.Response}
		if err := response.UnmarshalBinary(buf); err != nil {
			return 0, err
		}
		return int64(len(buf)), nil
	}
	return 0, fmt.Errorf("transaction ID %d not found in response", r.TransactionID)
}

// Transmit sends the MBIM message and waits for a response
func (r *Request) Transmit(conn net.Conn) error {
	if _, err := r.WriteTo(conn); err != nil {
		return err
	}
	if _, err := r.ReadFrom(conn); err != nil {
		return err
	}
	return nil
}

// MarshalBinary creates binary representation of the MBIM message
func (r *Request) MarshalBinary() ([]byte, error) {
	command, err := r.Command.MarshalBinary()
	if err != nil {
		return nil, err
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
