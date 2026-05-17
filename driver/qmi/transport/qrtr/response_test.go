package qrtr

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/damonto/euicc-go/driver/qmi/protocol"
)

func TestResponseRejectsShortHeader(t *testing.T) {
	var response Response
	err := response.UnmarshalBinary([]byte{0x02, 0x01, 0x00, 0x3b, 0x00, 0x00})
	if err == nil || !strings.Contains(err.Error(), "data too short") {
		t.Fatalf("UnmarshalBinary error = %v, want short data error", err)
	}
}

func TestResponseRejectsTLVLengthMismatch(t *testing.T) {
	packet := encodeResponse(t, 42, 0xAA)
	packet[5]++

	var response Response
	err := response.UnmarshalBinary(packet)
	if err == nil || !strings.Contains(err.Error(), "QMI TLV length mismatch") {
		t.Fatalf("UnmarshalBinary error = %v, want QMI TLV length mismatch", err)
	}
}

func TestBytesRejectsOversizedMessage(t *testing.T) {
	request := &protocol.Request{
		TransactionID: 42,
		MessageID:     protocol.QMIUIMSendAPDU,
		Value: protocol.TLVs{
			{Type: 0x10, Len: protocol.MaxEncodedMessageLength, Value: bytes.Repeat([]byte{0xAA}, protocol.MaxEncodedMessageLength)},
		},
	}

	transport := &Transport{}
	if _, err := transport.bytes(request); err == nil {
		t.Fatal("bytes error = nil, want oversized message error")
	}
}

func TestReadSetsRequestDeadline(t *testing.T) {
	conn := &fakeDeadlineConn{packet: encodeResponse(t, 42, 0xAA)}
	response := &captureResponse{}
	request := &protocol.Request{
		TransactionID: 42,
		Response:      response,
		ReadTimeout:   25 * time.Millisecond,
	}

	transport := &Transport{}
	n, err := transport.Read(conn, request)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != len(conn.packet) {
		t.Fatalf("Read length = %d, want %d", n, len(conn.packet))
	}
	if len(conn.deadlines) != 2 {
		t.Fatalf("SetReadDeadline calls = %d, want 2", len(conn.deadlines))
	}
	if conn.deadlines[0].IsZero() {
		t.Fatal("first deadline is zero, want request deadline")
	}
	if conn.deadlines[0].After(time.Now().Add(request.ReadTimeout)) {
		t.Fatalf("deadline = %s, want within request timeout %s", conn.deadlines[0], request.ReadTimeout)
	}
	if !conn.deadlines[1].IsZero() {
		t.Fatalf("second deadline = %s, want cleared deadline", conn.deadlines[1])
	}
	if got, want := response.payload, []byte{0xAA}; !bytes.Equal(got, want) {
		t.Fatalf("payload = %X, want %X", got, want)
	}
}

type fakeDeadlineConn struct {
	packet    []byte
	deadline  time.Time
	deadlines []time.Time
}

func (c *fakeDeadlineConn) Read(b []byte) (int, error) {
	if c.deadline.IsZero() {
		return 0, errors.New("missing read deadline")
	}
	return copy(b, c.packet), nil
}

func (c *fakeDeadlineConn) Write(p []byte) (int, error) {
	return len(p), nil
}

func (c *fakeDeadlineConn) SetReadDeadline(t time.Time) error {
	c.deadline = t
	c.deadlines = append(c.deadlines, t)
	return nil
}

type captureResponse struct {
	payload []byte
}

func (r *captureResponse) UnmarshalResponse(tlvs *protocol.TLVs) error {
	value, ok := tlvs.Find(0x10)
	if !ok {
		return errors.New("missing payload TLV")
	}
	r.payload = value.Value
	return nil
}

func encodeResponse(t *testing.T, txnID uint16, payload byte) []byte {
	t.Helper()

	value := new(bytes.Buffer)
	tlvs := protocol.TLVs{
		{Type: 0x02, Len: 4, Value: []byte{0x00, 0x00, 0x00, 0x00}},
		{Type: 0x10, Len: 1, Value: []byte{payload}},
	}
	if _, err := tlvs.WriteTo(value); err != nil {
		t.Fatalf("write TLVs: %v", err)
	}

	packet := new(bytes.Buffer)
	mustWrite(t, packet, protocol.QMIMessageTypeResponse)
	mustWrite(t, packet, txnID)
	mustWrite(t, packet, protocol.QMIUIMSendAPDU)
	mustWrite(t, packet, uint16(value.Len()))
	if _, err := packet.Write(value.Bytes()); err != nil {
		t.Fatalf("write packet payload: %v", err)
	}
	return packet.Bytes()
}

func mustWrite(t *testing.T, w *bytes.Buffer, value any) {
	t.Helper()
	if err := binary.Write(w, binary.LittleEndian, value); err != nil {
		t.Fatalf("binary.Write failed: %v", err)
	}
}
