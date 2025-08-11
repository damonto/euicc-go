package qmi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type TLV struct {
	Type  uint8
	Len   uint16
	Value []byte
}

func (t *TLV) Error() error {
	if len(t.Value) < 4 {
		return fmt.Errorf("result TLV too short, expected 4 bytes, got %d", len(t.Value))
	}
	if binary.LittleEndian.Uint16(t.Value[0:2]) == uint16(QMIResultSuccess) {
		return nil
	}
	return QMIError(binary.LittleEndian.Uint16(t.Value[2:4]))
}

// QMUXHeader represents the header of a QMUX PDU
type QMUXHeader struct {
	IfType       uint8
	Length       uint16
	ControlFlags uint8
	ServiceType  ServiceType
	ClientID     uint8
}

type Header[T uint8 | uint16] struct {
	MessageType   MessageType
	TransactionID T
	MessageID     MessageID
	MessageLength uint16
}

type ResponseUnmarshaler interface {
	UnmarshalResponse(TLVs map[uint8]TLV) error
}

type Request struct {
	ClientID      uint8
	TransactionID uint16
	ServiceType   ServiceType
	ReadTimeout   time.Duration
	MessageID     MessageID
	TLVs          []TLV
	Response      ResponseUnmarshaler
}

func toBuffer(TLVs []TLV) *bytes.Buffer {
	buf := new(bytes.Buffer)
	for _, tlv := range TLVs {
		binary.Write(buf, binary.LittleEndian, tlv.Type)
		binary.Write(buf, binary.LittleEndian, tlv.Len)
		buf.Write(tlv.Value)
	}
	return buf
}

// UnmarshalBinary converts the Request into a binary representation suitable for transmission
func (r *Request) UnmarshalBinary() ([]byte, error) {
	value := toBuffer(r.TLVs)
	headerBuf := new(bytes.Buffer)
	if r.ServiceType == QMIServiceControl {
		binary.Write(headerBuf, binary.LittleEndian, Header[uint8]{
			MessageType:   QMIMessageTypeRequest,
			TransactionID: uint8(r.TransactionID),
			MessageID:     r.MessageID,
			MessageLength: uint16(value.Len()),
		})
	} else {
		binary.Write(headerBuf, binary.LittleEndian, Header[uint16]{
			MessageType:   QMIMessageTypeRequest,
			TransactionID: r.TransactionID,
			MessageID:     r.MessageID,
			MessageLength: uint16(value.Len()),
		})
	}
	headerBuf.Write(value.Bytes())

	sduBytes := headerBuf.Bytes()
	requestBuf := new(bytes.Buffer)
	binary.Write(requestBuf, binary.LittleEndian, QMUXHeader{
		IfType:       QMUXHeaderIfType,
		Length:       uint16(len(sduBytes) + 5),
		ControlFlags: QMUXHeaderControlFlagRequest,
		ServiceType:  r.ServiceType,
		ClientID:     r.ClientID,
	})
	requestBuf.Write(sduBytes)
	return requestBuf.Bytes(), nil
}

// WriteTo writes the Request to the provided connection
func (r *Request) WriteTo(w net.Conn) (int, error) {
	data, err := r.UnmarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}
	n, err := w.Write(data)
	if err != nil {
		return 0, fmt.Errorf("failed to write request: %w", err)
	}
	return n, nil
}

// ReadFrom reads a response from the connection and unmarshals it into the Request's Response field
func (r *Request) ReadFrom(c net.Conn) (int, error) {
	if r.ReadTimeout == 0 {
		r.ReadTimeout = 30 * time.Second
	}
	deadline := time.Now().Add(r.ReadTimeout)
	for time.Now().Before(deadline) {
		c.SetReadDeadline(time.Now().Add(1 * time.Second))

		header := make([]byte, 3)
		if _, err := io.ReadAtLeast(c, header, 3); err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue // Timeout, try again
			}
			return 0, err
		}

		length := int(binary.LittleEndian.Uint16(header[1:3])) + 1
		buf := make([]byte, length)
		copy(buf[:3], header)
		if _, err := io.ReadFull(c, buf[3:]); err != nil {
			return 0, err
		}

		var response Response
		if err := response.UnmarshalBinary(buf[:length]); err != nil {
			return 0, err
		}
		if r.ClientID != response.ClientID && response.TransactionID != r.TransactionID {
			continue
		}
		if err := response.Error(); err != nil {
			return 0, err
		}
		if err := r.Response.UnmarshalResponse(response.TLVs); err != nil {
			return 0, err
		}
		return length, nil
	}
	return 0, fmt.Errorf("timed out waiting for response for transaction ID %d", r.TransactionID)
}

// Transmit sends the request and waits for the response
func (r *Request) Transmit(c net.Conn) error {
	if _, err := r.WriteTo(c); err != nil {
		return err
	}
	if _, err := r.ReadFrom(c); err != nil {
		return err
	}
	return nil
}

// Response represents a complete parsed QMI message
type Response struct {
	QMUXHeader
	TransactionID uint16
	MessageID     MessageID
	MessageType   MessageType
	MessageLength uint16
	TLVs          map[uint8]TLV
}

func (r *Response) UnmarshalBinary(data []byte) error {
	if len(data) < 11 {
		return fmt.Errorf("data too short: got %d bytes", len(data))
	}

	reader := bytes.NewReader(data)
	// Read QMUX header
	if err := binary.Read(reader, binary.LittleEndian, &r.QMUXHeader); err != nil {
		return fmt.Errorf("read QMUX header: %w", err)
	}
	// Read message type
	binary.Read(reader, binary.LittleEndian, &r.MessageType)
	// Read transaction ID
	switch r.QMUXHeader.ServiceType {
	case QMIServiceControl:
		var txnID uint8
		binary.Read(reader, binary.LittleEndian, &txnID)
		r.TransactionID = uint16(txnID)
	default:
		binary.Read(reader, binary.LittleEndian, &r.TransactionID)
	}

	// Read message ID and length
	binary.Read(reader, binary.LittleEndian, &r.MessageID)
	binary.Read(reader, binary.LittleEndian, &r.MessageLength)
	r.TLVs = make(map[uint8]TLV)
	if r.MessageLength > 0 {
		return r.toTVLs(io.LimitReader(reader, int64(r.MessageLength)))
	}
	return nil
}

func (r *Response) toTVLs(reader io.Reader) error {
	for {
		var t uint8
		if err := binary.Read(reader, binary.LittleEndian, &t); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("read TLV type: %w", err)
		}

		var n uint16
		binary.Read(reader, binary.LittleEndian, &n)
		v := make([]byte, n)
		if _, err := io.ReadFull(reader, v); err != nil {
			return err
		}
		r.TLVs[t] = TLV{Type: t, Len: n, Value: v}
	}
	return nil
}

func (r *Response) Error() error {
	tlv, ok := r.TLVs[0x02]
	if !ok {
		return errors.New("no result TLV found")
	}
	return tlv.Error()
}
