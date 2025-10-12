package qrtr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/damonto/euicc-go/driver/qmi/core"
)

type Transport struct {
	conn net.Conn
}

func New(conn net.Conn) core.Transport {
	return &Transport{conn: conn}
}

func (t *Transport) toBytes(r *core.Request) ([]byte, error) {
	value := bytes.NewBuffer(nil)
	if _, err := r.Value.WriteTo(value); err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, core.QMIMessageTypeRequest)
	binary.Write(buf, binary.LittleEndian, r.TransactionID)
	binary.Write(buf, binary.LittleEndian, r.MessageID)
	binary.Write(buf, binary.LittleEndian, uint16(value.Len()))
	buf.Write(value.Bytes())
	return buf.Bytes(), nil
}

// Read reads a response from the connection and unmarshals it into the Request's Response field
func (t *Transport) Read(c net.Conn, r *core.Request) (int, error) {
	if r.ReadTimeout == 0 {
		r.ReadTimeout = 30 * time.Second
	}
	deadline := time.Now().Add(r.ReadTimeout)
	for time.Now().Before(deadline) {
		c.SetReadDeadline(time.Now().Add(1 * time.Second))

		buf := make([]byte, 512)
		n, err := c.Read(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
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
	bs, err := t.toBytes(request)
	if err != nil {
		return err
	}
	if _, err = t.conn.Write(bs); err != nil {
		return err
	}
	_, err = t.Read(t.conn, request)
	return err
}
