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

const maxControlTransfer = 4096

// Request represents a standard MBIM request
type Request struct {
	MessageType   MessageType
	MessageLength uint32
	TransactionID uint32
	ReadTimeout   time.Duration
	Command       encoding.BinaryMarshaler
	Response      encoding.BinaryUnmarshaler
}

func (r *Request) WriteTo(w net.Conn) (int, error) {
	data, err := r.MarshalBinary()
	if err != nil {
		return 0, err
	}
	frames, err := fragmentedMessage{
		data:         data,
		maxFrameSize: maxControlTransfer,
	}.Frames()
	if err != nil {
		return 0, err
	}
	var written int
	for _, frame := range frames {
		n, err := writeFull(w, frame)
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func (r *Request) ReadFrom(c net.Conn) (int, error) {
	if r.ReadTimeout == 0 {
		r.ReadTimeout = 30 * time.Second
	}
	deadline := time.Now().Add(r.ReadTimeout)
	var collector *fragmentCollector
	for time.Now().Before(deadline) {
		if err := c.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
			return 0, err
		}

		header := make([]byte, 12)
		if _, err := io.ReadAtLeast(c, header, 12); err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return 0, err
		}

		length := binary.LittleEndian.Uint32(header[4:8])
		if length < 12 {
			return 0, fmt.Errorf("invalid MBIM message length %d", length)
		}
		if length > maxControlTransfer {
			return 0, fmt.Errorf("MBIM frame length %d exceeds max control transfer %d", length, maxControlTransfer)
		}
		buf := make([]byte, length)
		copy(buf[:12], header)
		if _, err := io.ReadFull(c, buf[12:]); err != nil {
			return 0, err
		}

		messageType := MessageType(binary.LittleEndian.Uint32(header[0:4]))
		transactionID := binary.LittleEndian.Uint32(header[8:12])
		if transactionID != r.TransactionID {
			continue
		}
		expectedMessageType, ok := responseMessageType(r.MessageType)
		if !ok {
			return 0, fmt.Errorf("unsupported MBIM request message type %#x", r.MessageType)
		}
		if messageType != expectedMessageType && messageType != MessageTypeFunctionError {
			continue
		}

		if isFragmentMessage(messageType) {
			if collector != nil {
				if err := collector.add(buf); err != nil {
					return 0, err
				}
				if !collector.complete() {
					continue
				}
				completeFrame, err := collector.MarshalBinary()
				if err != nil {
					return 0, err
				}
				buf = completeFrame
			} else if len(buf) >= 20 && binary.LittleEndian.Uint32(buf[12:16]) > 1 {
				var err error
				collector, err = newFragmentCollector(buf)
				if err != nil {
					return 0, err
				}
				if !collector.complete() {
					continue
				}
				buf, err = collector.MarshalBinary()
				if err != nil {
					return 0, err
				}
			}
		}

		response := CommandResponse{Response: r.Response}
		if err := response.UnmarshalBinary(buf); err != nil {
			return 0, err
		}
		return len(buf), nil
	}
	return 0, fmt.Errorf("transaction ID %d not found in response", r.TransactionID)
}

func writeFull(w io.Writer, data []byte) (int, error) {
	var written int
	for len(data) > 0 {
		n, err := w.Write(data)
		written += n
		if err != nil {
			return written, err
		}
		if n <= 0 {
			return written, io.ErrShortWrite
		}
		data = data[n:]
	}
	return written, nil
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

func responseMessageType(requestType MessageType) (MessageType, bool) {
	switch requestType {
	case MessageTypeOpen:
		return MessageTypeOpenDone, true
	case MessageTypeClose:
		return MessageTypeCloseDone, true
	case MessageTypeCommand:
		return MessageTypeCommandDone, true
	default:
		return 0, false
	}
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
	if len(data) < 12 {
		return fmt.Errorf("MBIM response too short: %d", len(data))
	}
	r.MessageType = MessageType(binary.LittleEndian.Uint32(data[0:4]))
	r.MessageLength = binary.LittleEndian.Uint32(data[4:8])
	r.TransactionID = binary.LittleEndian.Uint32(data[8:12])
	if r.MessageLength != uint32(len(data)) {
		return fmt.Errorf("MBIM response length mismatch: header=%d actual=%d", r.MessageLength, len(data))
	}

	switch r.MessageType {
	case MessageTypeOpenDone, MessageTypeCloseDone:
		return r.unmarshalStatusDone(data)
	case MessageTypeCommandDone:
		return r.unmarshalCommandDone(data)
	case MessageTypeFunctionError:
		return r.unmarshalFunctionError(data)
	default:
		return fmt.Errorf("unexpected MBIM response message type %#x", r.MessageType)
	}
}

func (r *CommandResponse) unmarshalStatusDone(data []byte) error {
	if len(data) != 16 {
		return fmt.Errorf("MBIM status response length %d, want 16", len(data))
	}
	r.Status = MBIMStatus(binary.LittleEndian.Uint32(data[12:16]))
	if r.Status != MBIMStatusNone {
		return r.Status
	}
	if r.Response == nil {
		return nil
	}
	return r.Response.UnmarshalBinary(nil)
}

func (r *CommandResponse) unmarshalCommandDone(data []byte) error {
	if len(data) < 48 {
		return fmt.Errorf("MBIM command response too short: %d", len(data))
	}
	r.FragmentTotal = binary.LittleEndian.Uint32(data[12:16])
	r.FragmentCurrent = binary.LittleEndian.Uint32(data[16:20])
	copy(r.ServiceID[:], data[20:36])
	r.CommandID = binary.LittleEndian.Uint32(data[36:40])
	r.Status = MBIMStatus(binary.LittleEndian.Uint32(data[40:44]))
	if r.Status != MBIMStatusNone {
		return r.Status
	}
	if r.FragmentTotal != 1 || r.FragmentCurrent != 0 {
		return fmt.Errorf("unsupported MBIM fragmented response: fragment %d of %d", r.FragmentCurrent, r.FragmentTotal)
	}
	r.ResponseLength = binary.LittleEndian.Uint32(data[44:48])
	if r.ResponseLength > uint32(len(data)-48) {
		return fmt.Errorf("MBIM command response buffer length %d exceeds remaining %d", r.ResponseLength, len(data)-48)
	}
	r.ResponseBuffer = data[48 : 48+r.ResponseLength]
	if r.Response == nil {
		return nil
	}
	return r.Response.UnmarshalBinary(r.ResponseBuffer)
}

func (r *CommandResponse) unmarshalFunctionError(data []byte) error {
	if len(data) != 16 {
		return fmt.Errorf("MBIM function error response length %d, want 16", len(data))
	}
	return MBIMProtocolError(binary.LittleEndian.Uint32(data[12:16]))
}
