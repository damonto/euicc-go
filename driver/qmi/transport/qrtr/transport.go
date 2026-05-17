package qrtr

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/damonto/euicc-go/driver/qmi/protocol"
)

type Transport struct {
	conn Conn
}

type Conn interface {
	io.ReadWriter
	SetReadDeadline(time.Time) error
}

type Header struct {
	MessageType   protocol.MessageType
	TransactionID uint16
	MessageID     protocol.MessageID
	MessageLength uint16
}

func New(conn Conn) protocol.Transport {
	return &Transport{conn: conn}
}

func (t *Transport) bytes(r *protocol.Request) ([]byte, error) {
	value := new(bytes.Buffer)
	if _, err := r.Value.WriteTo(value); err != nil {
		return nil, err
	}
	if value.Len() > protocol.MaxQRTRServiceTLVLength {
		return nil, fmt.Errorf("QRTR QMI message TLVs length %d exceeds limit %d", value.Len(), protocol.MaxQRTRServiceTLVLength)
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, Header{
		MessageType:   protocol.QMIMessageTypeRequest,
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
func (t *Transport) Read(c Conn, r *protocol.Request) (n int, err error) {
	readTimeout := r.ReadTimeout
	if readTimeout == 0 {
		readTimeout = 30 * time.Second
	}
	deadline := time.Now().Add(readTimeout)
	if err := c.SetReadDeadline(deadline); err != nil {
		return 0, err
	}
	defer func() {
		clearErr := c.SetReadDeadline(time.Time{})
		if clearErr != nil && !errors.Is(clearErr, net.ErrClosed) && err == nil {
			err = clearErr
		}
	}()

	for time.Now().Before(deadline) {
		buf := make([]byte, protocol.MaxQRTRQMIMessageLength)
		n, err := c.Read(buf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				break
			}
			return 0, err
		}

		var response Response
		if err := response.UnmarshalBinary(buf[:n]); err != nil {
			return 0, err
		}
		if response.MessageType != protocol.QMIMessageTypeResponse || response.TransactionID != r.TransactionID {
			continue
		}
		if err := r.Response.UnmarshalResponse(&response.Value); err != nil {
			return 0, err
		}
		return n, nil
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
