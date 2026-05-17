package qmi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/damonto/euicc-go/driver/qmi/protocol"
)

type QMUXHeader struct {
	IfType       uint8
	Length       uint16
	ControlFlags uint8
	ServiceType  protocol.ServiceType
	ClientID     uint8
}

type Header[T uint8 | uint16] struct {
	MessageType   protocol.MessageType
	TransactionID T
	MessageID     protocol.MessageID
	MessageLength uint16
}

type Transport struct {
	conn net.Conn
}

func New(conn net.Conn) protocol.Transport {
	return &Transport{conn: conn}
}

func (t *Transport) bytes(r *protocol.Request) ([]byte, error) {
	value := new(bytes.Buffer)
	if _, err := r.Value.WriteTo(value); err != nil {
		return nil, err
	}
	maxValueLength := protocol.MaxQMUXServiceTLVLength
	if r.ServiceType == protocol.QMIServiceControl {
		maxValueLength = protocol.MaxQMUXControlTLVLength
	}
	if value.Len() > maxValueLength {
		return nil, fmt.Errorf("QMI message TLVs length %d exceeds limit %d", value.Len(), maxValueLength)
	}

	headerBuf := new(bytes.Buffer)
	if r.ServiceType == protocol.QMIServiceControl {
		if err := binary.Write(headerBuf, binary.LittleEndian, Header[uint8]{
			MessageType:   protocol.QMIMessageTypeRequest,
			TransactionID: uint8(r.TransactionID),
			MessageID:     r.MessageID,
			MessageLength: uint16(value.Len()),
		}); err != nil {
			return nil, fmt.Errorf("write control QMI header: %w", err)
		}
	} else {
		if err := binary.Write(headerBuf, binary.LittleEndian, Header[uint16]{
			MessageType:   protocol.QMIMessageTypeRequest,
			TransactionID: r.TransactionID,
			MessageID:     r.MessageID,
			MessageLength: uint16(value.Len()),
		}); err != nil {
			return nil, fmt.Errorf("write service QMI header: %w", err)
		}
	}
	headerBuf.Write(value.Bytes())

	sduBytes := headerBuf.Bytes()
	requestBuf := new(bytes.Buffer)
	if err := binary.Write(requestBuf, binary.LittleEndian, QMUXHeader{
		IfType:       protocol.QMUXHeaderIfType,
		Length:       uint16(len(sduBytes) + 5),
		ControlFlags: protocol.QMUXHeaderControlFlagRequest,
		ServiceType:  r.ServiceType,
		ClientID:     r.ClientID,
	}); err != nil {
		return nil, fmt.Errorf("write QMUX header: %w", err)
	}
	requestBuf.Write(sduBytes)
	return requestBuf.Bytes(), nil
}

// Read reads a response from the connection and unmarshals it into the Request's Response field
func (t *Transport) Read(c net.Conn, r *protocol.Request) (n int, err error) {
	readTimeout := r.ReadTimeout
	if readTimeout == 0 {
		readTimeout = 30 * time.Second
	}
	deadline := time.Now().Add(readTimeout)
	defer func() {
		clearErr := c.SetReadDeadline(time.Time{})
		if clearErr != nil && !errors.Is(clearErr, net.ErrClosed) && err == nil {
			err = clearErr
		}
	}()

	for time.Now().Before(deadline) {
		readDeadline := time.Now().Add(time.Second)
		if readDeadline.After(deadline) {
			readDeadline = deadline
		}
		if err := c.SetReadDeadline(readDeadline); err != nil {
			return 0, err
		}

		header := make([]byte, 3)
		if _, err := io.ReadAtLeast(c, header, 3); err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return 0, err
		}

		length := int(binary.LittleEndian.Uint16(header[1:3])) + 1
		if length < len(header) {
			return 0, fmt.Errorf("invalid QMUX length %d", length)
		}
		buf := make([]byte, length)
		copy(buf[:3], header)
		if _, err := io.ReadFull(c, buf[3:]); err != nil {
			return 0, err
		}

		var response Response
		if err := response.UnmarshalBinary(buf[:length]); err != nil {
			return 0, err
		}
		if response.MessageType != protocol.QMIMessageTypeResponse && r.ClientID != response.ClientID && response.TransactionID != r.TransactionID {
			continue
		}
		if err := r.Response.UnmarshalResponse(&response.Value); err != nil {
			return 0, err
		}
		return length, nil
	}
	return 0, fmt.Errorf("timed out waiting for response for transaction ID %d", r.TransactionID)
}

func (t *Transport) Transmit(request *protocol.Request) error {
	bs, err := t.bytes(request)
	if err != nil {
		return err
	}
	if err := writeFull(t.conn, bs); err != nil {
		return err
	}
	_, err = t.Read(t.conn, request)
	return err
}

func writeFull(w io.Writer, p []byte) error {
	for len(p) > 0 {
		n, err := w.Write(p)
		if err != nil {
			return err
		}
		if n <= 0 {
			return io.ErrShortWrite
		}
		p = p[n:]
	}
	return nil
}
