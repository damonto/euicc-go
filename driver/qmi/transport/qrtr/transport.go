package qrtr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/damonto/euicc-go/driver/qmi/core"
)

type Transport struct {
	conn io.ReadWriter
}

type Header struct {
	MessageType   core.MessageType
	TransactionID uint16
	MessageID     core.MessageID
	MessageLength uint16
}

func New(conn io.ReadWriter) core.Transport {
	return &Transport{conn: conn}
}

func (t *Transport) bytes(r *core.Request) ([]byte, error) {
	value := new(bytes.Buffer)
	if _, err := r.Value.WriteTo(value); err != nil {
		return nil, err
	}
	if value.Len() > core.MaxQRTRServiceTLVLength {
		return nil, fmt.Errorf("QRTR QMI message TLVs length %d exceeds limit %d", value.Len(), core.MaxQRTRServiceTLVLength)
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, Header{
		MessageType:   core.QMIMessageTypeRequest,
		TransactionID: r.TransactionID,
		MessageID:     r.MessageID,
		MessageLength: uint16(value.Len()),
	}); err != nil {
		return nil, fmt.Errorf("write QRTR QMI header: %w", err)
	}
	buf.Write(value.Bytes())
	return buf.Bytes(), nil
}

// Read reads a response from the connection and unmarshals it into the Request's Response field
func (t *Transport) Read(c io.Reader, r *core.Request) (int, error) {
	if r.ReadTimeout == 0 {
		r.ReadTimeout = 30 * time.Second
	}
	deadline := time.Now().Add(r.ReadTimeout)
	for time.Now().Before(deadline) {
		buf := make([]byte, core.MaxQRTRQMIMessageLength)
		n, err := c.Read(buf)
		if err != nil {
			return 0, err
		}

		var response Response
		if err := response.UnmarshalBinary(buf[:n]); err != nil {
			return 0, err
		}
		if response.TransactionID != r.TransactionID {
			continue
		}
		if err := r.Response.UnmarshalResponse(&response.Value); err != nil {
			return 0, err
		}
		return n, nil
	}
	return 0, fmt.Errorf("timed out waiting for response for transaction ID %d", r.TransactionID)
}

func (t *Transport) Transmit(request *core.Request) error {
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
