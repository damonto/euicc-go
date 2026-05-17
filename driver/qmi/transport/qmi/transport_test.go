package qmi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/damonto/euicc-go/driver/qmi/protocol"
)

type captureResponse struct {
	payload []byte
}

func (r *captureResponse) UnmarshalResponse(tlvs *protocol.TLVs) error {
	value, ok := tlvs.Find(0x10)
	if !ok {
		return errors.New("missing payload TLV")
	}
	r.payload = append([]byte(nil), value.Value...)
	return nil
}

func TestReadAcceptsFirstResponseForSynchronousTransport(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	go func() {
		defer server.Close()
		_, _ = server.Write(encodeResponse(t, protocol.QMIServiceUIM, 8, 41, 0xAA))
		_, _ = server.Write(encodeResponse(t, protocol.QMIServiceUIM, 7, 42, 0xBB))
	}()

	response := &captureResponse{}
	request := &protocol.Request{
		ClientID:      7,
		TransactionID: 42,
		ServiceType:   protocol.QMIServiceUIM,
		Response:      response,
		ReadTimeout:   100 * time.Millisecond,
	}

	transport := &Transport{}
	if _, err := transport.Read(client, request); err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if !bytes.Equal(response.payload, []byte{0xAA}) {
		t.Fatalf("payload = %X, want AA", response.payload)
	}
}

func TestReadUsesRequestDeadlineAndClearsIt(t *testing.T) {
	stopErr := errors.New("stop read")
	conn := &fakeDeadlineNetConn{readErr: stopErr}
	request := &protocol.Request{
		Response:    &captureResponse{},
		ReadTimeout: 25 * time.Millisecond,
	}

	transport := &Transport{}
	if _, err := transport.Read(conn, request); !errors.Is(err, stopErr) {
		t.Fatalf("Read error = %v, want %v", err, stopErr)
	}
	if len(conn.readDeadlines) != 2 {
		t.Fatalf("SetReadDeadline calls = %d, want 2", len(conn.readDeadlines))
	}
	if conn.readDeadlines[0].IsZero() {
		t.Fatal("first deadline is zero, want request deadline")
	}
	if conn.readDeadlines[0].After(time.Now().Add(request.ReadTimeout)) {
		t.Fatalf("deadline = %s, want within request timeout %s", conn.readDeadlines[0], request.ReadTimeout)
	}
	if !conn.readDeadlines[1].IsZero() {
		t.Fatalf("second deadline = %s, want cleared deadline", conn.readDeadlines[1])
	}
}

type fakeDeadlineNetConn struct {
	readErr       error
	readDeadlines []time.Time
}

func (c *fakeDeadlineNetConn) Read([]byte) (int, error)         { return 0, c.readErr }
func (c *fakeDeadlineNetConn) Write(p []byte) (int, error)      { return len(p), nil }
func (c *fakeDeadlineNetConn) Close() error                     { return nil }
func (c *fakeDeadlineNetConn) LocalAddr() net.Addr              { return &net.IPAddr{} }
func (c *fakeDeadlineNetConn) RemoteAddr() net.Addr             { return &net.IPAddr{} }
func (c *fakeDeadlineNetConn) SetDeadline(time.Time) error      { return nil }
func (c *fakeDeadlineNetConn) SetWriteDeadline(time.Time) error { return nil }

func (c *fakeDeadlineNetConn) SetReadDeadline(t time.Time) error {
	c.readDeadlines = append(c.readDeadlines, t)
	return nil
}

func encodeResponse(t *testing.T, serviceType protocol.ServiceType, clientID uint8, txnID uint16, payload byte) []byte {
	t.Helper()

	value := new(bytes.Buffer)
	tlvs := protocol.TLVs{
		{Type: 0x02, Len: 4, Value: []byte{0x00, 0x00, 0x00, 0x00}},
		{Type: 0x10, Len: 1, Value: []byte{payload}},
	}
	if _, err := tlvs.WriteTo(value); err != nil {
		t.Fatalf("write TLVs: %v", err)
	}

	sdu := new(bytes.Buffer)
	if err := binary.Write(sdu, binary.LittleEndian, Header[uint16]{
		MessageType:   protocol.QMIMessageTypeResponse,
		TransactionID: txnID,
		MessageID:     protocol.QMIUIMSendAPDU,
		MessageLength: uint16(value.Len()),
	}); err != nil {
		t.Fatalf("write SDU header: %v", err)
	}
	if _, err := sdu.Write(value.Bytes()); err != nil {
		t.Fatalf("write SDU payload: %v", err)
	}

	packet := new(bytes.Buffer)
	if err := binary.Write(packet, binary.LittleEndian, QMUXHeader{
		IfType:       protocol.QMUXHeaderIfType,
		Length:       uint16(sdu.Len() + 5),
		ControlFlags: protocol.QMUXHeaderControlFlagRequest,
		ServiceType:  serviceType,
		ClientID:     clientID,
	}); err != nil {
		t.Fatalf("write QMUX header: %v", err)
	}
	if _, err := packet.Write(sdu.Bytes()); err != nil {
		t.Fatalf("write packet payload: %v", err)
	}

	return packet.Bytes()
}
