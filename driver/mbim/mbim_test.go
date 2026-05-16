package mbim

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

type fakeConn struct {
	read   *bytes.Reader
	writes [][]byte
	closed bool
}

func newFakeConn(responses ...[]byte) *fakeConn {
	return &fakeConn{read: bytes.NewReader(bytes.Join(responses, nil))}
}

func (c *fakeConn) Read(p []byte) (int, error) {
	return c.read.Read(p)
}

func (c *fakeConn) Write(p []byte) (int, error) {
	c.writes = append(c.writes, append([]byte(nil), p...))
	return len(p), nil
}

func (c *fakeConn) Close() error {
	c.closed = true
	return nil
}

func (c *fakeConn) LocalAddr() net.Addr              { return nil }
func (c *fakeConn) RemoteAddr() net.Addr             { return nil }
func (c *fakeConn) SetDeadline(time.Time) error      { return nil }
func (c *fakeConn) SetReadDeadline(time.Time) error  { return nil }
func (c *fakeConn) SetWriteDeadline(time.Time) error { return nil }

func commandDone(transactionID uint32, serviceID [16]byte, commandID uint32, payload []byte) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, MessageTypeCommandDone)
	_ = binary.Write(buf, binary.LittleEndian, uint32(48+len(payload)))
	_ = binary.Write(buf, binary.LittleEndian, transactionID)
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, serviceID)
	_ = binary.Write(buf, binary.LittleEndian, commandID)
	_ = binary.Write(buf, binary.LittleEndian, MBIMStatusNone)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(payload)))
	_, _ = buf.Write(payload)
	return buf.Bytes()
}

func fragmentFrame(message []byte, total, current uint32, payload []byte) []byte {
	buf := make([]byte, 20+len(payload))
	copy(buf[0:4], message[0:4])
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(buf)))
	copy(buf[8:12], message[8:12])
	binary.LittleEndian.PutUint32(buf[12:16], total)
	binary.LittleEndian.PutUint32(buf[16:20], current)
	copy(buf[20:], payload)
	return buf
}

func fragmentCommandDone(transactionID uint32, serviceID [16]byte, commandID uint32, payload []byte, chunks ...int) [][]byte {
	message := commandDone(transactionID, serviceID, commandID, payload)
	fragmentPayload := message[20:]
	fragments := make([][]byte, 0, len(chunks))
	offset := 0
	for i, chunk := range chunks {
		fragments = append(fragments, fragmentFrame(message, uint32(len(chunks)), uint32(i), fragmentPayload[offset:offset+chunk]))
		offset += chunk
	}
	if offset != len(fragmentPayload) {
		panic("fragment chunks do not cover the full MBIM fragment payload")
	}
	return fragments
}

func closeDone(transactionID uint32) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, MessageTypeCloseDone)
	_ = binary.Write(buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(buf, binary.LittleEndian, transactionID)
	_ = binary.Write(buf, binary.LittleEndian, MBIMStatusNone)
	return buf.Bytes()
}

func functionError(transactionID uint32, err MBIMProtocolError) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, MessageTypeFunctionError)
	_ = binary.Write(buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(buf, binary.LittleEndian, transactionID)
	_ = binary.Write(buf, binary.LittleEndian, err)
	return buf.Bytes()
}

func TestRequestMarshalMatchesLibmbimWireLayout(t *testing.T) {
	tests := []struct {
		name string
		req  *Request
		want []byte
	}{
		{
			name: "open",
			req:  (&OpenDeviceRequest{TransactionID: 1}).Request(),
			want: []byte{
				0x01, 0x00, 0x00, 0x00,
				0x10, 0x00, 0x00, 0x00,
				0x01, 0x00, 0x00, 0x00,
				0x00, 0x10, 0x00, 0x00,
			},
		},
		{
			name: "close",
			req:  (&CloseRequest{TransactionID: 2}).Request(),
			want: []byte{
				0x02, 0x00, 0x00, 0x00,
				0x0c, 0x00, 0x00, 0x00,
				0x02, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "empty command",
			req: (&SubscriberReadyStatusRequest{
				TransactionID: 6,
			}).Request(),
			want: []byte{
				0x03, 0x00, 0x00, 0x00,
				0x30, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00,
				0x01, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0xa2, 0x89, 0xcc, 0x33,
				0xbc, 0xbb, 0x8b, 0x4f,
				0xb6, 0xb0, 0x13, 0x3e,
				0xc2, 0xaa, 0xe6, 0xdf,
				0x02, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.req.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary failed: %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("MarshalBinary = % X, want % X", got, tt.want)
			}
		})
	}
}

func TestWriteToSplitsLargeCommandLikeLibmbim(t *testing.T) {
	apdu := make([]byte, 5000)
	request := (&TransmitAPDURequest{
		TransactionID: 7,
		APDU:          apdu,
	}).Request()
	raw, err := request.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}
	conn := newFakeConn()

	n, err := request.WriteTo(conn)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if len(conn.writes) != 2 {
		t.Fatalf("writes = %d, want 2", len(conn.writes))
	}
	if n != len(conn.writes[0])+len(conn.writes[1]) {
		t.Fatalf("WriteTo n = %d, want sum of fragment sizes", n)
	}

	var gotPayload []byte
	for i, fragment := range conn.writes {
		if len(fragment) > maxControlTransfer {
			t.Fatalf("fragment %d length = %d, want <= %d", i, len(fragment), maxControlTransfer)
		}
		if got := MessageType(binary.LittleEndian.Uint32(fragment[0:4])); got != MessageTypeCommand {
			t.Fatalf("fragment %d message type = %#x, want command", i, got)
		}
		if got := binary.LittleEndian.Uint32(fragment[8:12]); got != 7 {
			t.Fatalf("fragment %d transaction ID = %d, want 7", i, got)
		}
		if got := binary.LittleEndian.Uint32(fragment[12:16]); got != 2 {
			t.Fatalf("fragment %d total = %d, want 2", i, got)
		}
		if got := binary.LittleEndian.Uint32(fragment[16:20]); got != uint32(i) {
			t.Fatalf("fragment %d current = %d, want %d", i, got, i)
		}
		gotPayload = append(gotPayload, fragment[20:]...)
	}
	if !bytes.Equal(gotPayload, raw[20:]) {
		t.Fatalf("reassembled payload length = %d, want %d", len(gotPayload), len(raw[20:]))
	}
}

func TestSplitFragmentsMatchesLibmbimFixture(t *testing.T) {
	message := []byte{
		0x07, 0x00, 0x00, 0x80,
		0x30, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x02, 0x03,
		0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B,
		0x0C, 0x0D, 0x0E, 0x0F,
		0x10, 0x11, 0x12, 0x13,
		0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1A, 0x1B,
	}
	expected := [][]byte{
		{
			0x07, 0x00, 0x00, 0x80,
			0x1C, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00,
			0x04, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x01, 0x02, 0x03,
			0x04, 0x05, 0x06, 0x07,
		},
		{
			0x07, 0x00, 0x00, 0x80,
			0x1C, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00,
			0x04, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00,
			0x08, 0x09, 0x0A, 0x0B,
			0x0C, 0x0D, 0x0E, 0x0F,
		},
		{
			0x07, 0x00, 0x00, 0x80,
			0x1C, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00,
			0x04, 0x00, 0x00, 0x00,
			0x02, 0x00, 0x00, 0x00,
			0x10, 0x11, 0x12, 0x13,
			0x14, 0x15, 0x16, 0x17,
		},
		{
			0x07, 0x00, 0x00, 0x80,
			0x18, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00,
			0x04, 0x00, 0x00, 0x00,
			0x03, 0x00, 0x00, 0x00,
			0x18, 0x19, 0x1A, 0x1B,
		},
	}

	got, err := fragmentedMessage{
		data:         message,
		maxFrameSize: 28,
	}.Frames()
	if err != nil {
		t.Fatalf("Frames failed: %v", err)
	}
	if len(got) != len(expected) {
		t.Fatalf("fragments = %d, want %d", len(got), len(expected))
	}
	for i := range expected {
		if !bytes.Equal(got[i], expected[i]) {
			t.Fatalf("fragment %d = % X, want % X", i, got[i], expected[i])
		}
	}
}

func TestReadFromRejectsInvalidMessageLength(t *testing.T) {
	header := make([]byte, 12)
	binary.LittleEndian.PutUint32(header[0:4], uint32(MessageTypeCommandDone))
	binary.LittleEndian.PutUint32(header[4:8], 8)
	binary.LittleEndian.PutUint32(header[8:12], 1)
	conn := newFakeConn(header)
	apduRequest := TransmitAPDURequest{TransactionID: 1}
	request := apduRequest.Request()

	_, err := request.ReadFrom(conn)
	if err == nil {
		t.Fatal("ReadFrom error = nil, want invalid length error")
	}
	if !strings.Contains(err.Error(), "invalid MBIM message length 8") {
		t.Fatalf("ReadFrom error = %q, want invalid length", err.Error())
	}
}

func TestReadFromRejectsOversizedFrameLength(t *testing.T) {
	header := make([]byte, 12)
	binary.LittleEndian.PutUint32(header[0:4], uint32(MessageTypeCommandDone))
	binary.LittleEndian.PutUint32(header[4:8], maxControlTransfer+1)
	binary.LittleEndian.PutUint32(header[8:12], 1)
	conn := newFakeConn(header)
	apduRequest := TransmitAPDURequest{TransactionID: 1}
	request := apduRequest.Request()

	_, err := request.ReadFrom(conn)
	if err == nil {
		t.Fatal("ReadFrom error = nil, want oversized frame error")
	}
	if !strings.Contains(err.Error(), "exceeds max control transfer") {
		t.Fatalf("ReadFrom error = %q, want max control transfer error", err.Error())
	}
}

func TestCommandResponseRejectsTruncatedCommandDone(t *testing.T) {
	data := make([]byte, 16)
	binary.LittleEndian.PutUint32(data[0:4], uint32(MessageTypeCommandDone))
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(data)))
	binary.LittleEndian.PutUint32(data[8:12], 1)

	var response CommandResponse
	err := response.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("UnmarshalBinary error = nil, want short response error")
	}
	if !strings.Contains(err.Error(), "MBIM command response too short") {
		t.Fatalf("UnmarshalBinary error = %q, want short response", err.Error())
	}
}

func TestCommandResponseRejectsInformationBufferOverflow(t *testing.T) {
	data := commandDone(1, ServiceMsUiccLowLevelAccess, CIDUiccAPDU, []byte{0x00})
	binary.LittleEndian.PutUint32(data[44:48], 2)

	var response CommandResponse
	err := response.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("UnmarshalBinary error = nil, want buffer length error")
	}
	if !strings.Contains(err.Error(), "exceeds remaining") {
		t.Fatalf("UnmarshalBinary error = %q, want buffer length error", err.Error())
	}
}

func TestReadFromReturnsFunctionError(t *testing.T) {
	conn := newFakeConn(functionError(1, MBIMProtocolErrorNotOpened))
	apduRequest := TransmitAPDURequest{TransactionID: 1}
	request := apduRequest.Request()

	err := request.Transmit(conn)
	if err == nil {
		t.Fatal("Transmit error = nil, want protocol error")
	}
	var protocolErr MBIMProtocolError
	if !errors.As(err, &protocolErr) {
		t.Fatalf("Transmit error = %T %[1]v, want MBIMProtocolError", err)
	}
	if protocolErr != MBIMProtocolErrorNotOpened {
		t.Fatalf("protocol error = %v, want %v", protocolErr, MBIMProtocolErrorNotOpened)
	}
}

func TestReadFromReassemblesCommandDoneFragments(t *testing.T) {
	apduPayload := make([]byte, 12)
	binary.LittleEndian.PutUint32(apduPayload[0:4], 0x9000)
	binary.LittleEndian.PutUint32(apduPayload[4:8], 4)
	apduPayload = append(apduPayload, 0xde, 0xad, 0xbe, 0xef)
	fragments := fragmentCommandDone(1, ServiceMsUiccLowLevelAccess, CIDUiccAPDU, apduPayload, 12, 12, 12, 8)

	response := new(TransmitAPDUResponse)
	request := &Request{
		MessageType:   MessageTypeCommand,
		TransactionID: 1,
		Response:      response,
	}
	n, err := request.ReadFrom(newFakeConn(fragments...))
	if err != nil {
		t.Fatalf("ReadFrom failed: %v", err)
	}
	if n != 48+len(apduPayload) {
		t.Fatalf("ReadFrom n = %d, want %d", n, 48+len(apduPayload))
	}
	if response.Status != 0x9000 {
		t.Fatalf("Status = %#x, want 0x9000", response.Status)
	}
	if !bytes.Equal(response.Response, []byte{0xde, 0xad, 0xbe, 0xef}) {
		t.Fatalf("Response = % X, want DE AD BE EF", response.Response)
	}
}

func TestReadFromRejectsFragmentThatDoesNotStartAtZero(t *testing.T) {
	apduPayload := make([]byte, 12)
	fragments := fragmentCommandDone(1, ServiceMsUiccLowLevelAccess, CIDUiccAPDU, apduPayload, 14, 13, 13)
	conn := newFakeConn(fragments[1])
	request := &Request{
		MessageType:   MessageTypeCommand,
		TransactionID: 1,
		Response:      new(TransmitAPDUResponse),
	}

	_, err := request.ReadFrom(conn)
	if err == nil {
		t.Fatal("ReadFrom error = nil, want out-of-sequence error")
	}
	if !strings.Contains(err.Error(), "expecting MBIM fragment 0/3") {
		t.Fatalf("ReadFrom error = %q, want fragment 0 error", err.Error())
	}
}

func TestReadFromRejectsOutOfOrderFragment(t *testing.T) {
	apduPayload := make([]byte, 12)
	fragments := fragmentCommandDone(1, ServiceMsUiccLowLevelAccess, CIDUiccAPDU, apduPayload, 14, 13, 13)
	conn := newFakeConn(fragments[0], fragments[2])
	request := &Request{
		MessageType:   MessageTypeCommand,
		TransactionID: 1,
		Response:      new(TransmitAPDUResponse),
	}

	_, err := request.ReadFrom(conn)
	if err == nil {
		t.Fatal("ReadFrom error = nil, want out-of-sequence error")
	}
	if !strings.Contains(err.Error(), "expecting MBIM fragment 1/3") {
		t.Fatalf("ReadFrom error = %q, want fragment 1 error", err.Error())
	}
}

func TestReadFromRejectsFragmentTotalMismatch(t *testing.T) {
	apduPayload := make([]byte, 12)
	fragments := fragmentCommandDone(1, ServiceMsUiccLowLevelAccess, CIDUiccAPDU, apduPayload, 14, 13, 13)
	binary.LittleEndian.PutUint32(fragments[1][12:16], 4)
	conn := newFakeConn(fragments[0], fragments[1])
	request := &Request{
		MessageType:   MessageTypeCommand,
		TransactionID: 1,
		Response:      new(TransmitAPDUResponse),
	}

	_, err := request.ReadFrom(conn)
	if err == nil {
		t.Fatal("ReadFrom error = nil, want total mismatch error")
	}
	if !strings.Contains(err.Error(), "MBIM fragment total mismatch") {
		t.Fatalf("ReadFrom error = %q, want total mismatch", err.Error())
	}
}

func TestReadFromRejectsSingleFragmentWhileCollecting(t *testing.T) {
	apduPayload := make([]byte, 12)
	fragments := fragmentCommandDone(1, ServiceMsUiccLowLevelAccess, CIDUiccAPDU, apduPayload, 14, 13, 13)
	single := commandDone(1, ServiceMsUiccLowLevelAccess, CIDUiccAPDU, apduPayload)
	conn := newFakeConn(fragments[0], single)
	request := &Request{
		MessageType:   MessageTypeCommand,
		TransactionID: 1,
		Response:      new(TransmitAPDUResponse),
	}

	_, err := request.ReadFrom(conn)
	if err == nil {
		t.Fatal("ReadFrom error = nil, want out-of-sequence error")
	}
	if !strings.Contains(err.Error(), "expecting MBIM fragment 1/3") {
		t.Fatalf("ReadFrom error = %q, want fragment 1 error", err.Error())
	}
}

func TestDeviceSlotMappingsResponseUsesOffsetTable(t *testing.T) {
	data := new(bytes.Buffer)
	_ = binary.Write(data, binary.LittleEndian, uint32(2))
	_ = binary.Write(data, binary.LittleEndian, uint32(24))
	_ = binary.Write(data, binary.LittleEndian, uint32(4))
	_ = binary.Write(data, binary.LittleEndian, uint32(20))
	_ = binary.Write(data, binary.LittleEndian, uint32(4))
	_ = binary.Write(data, binary.LittleEndian, uint32(7))
	_ = binary.Write(data, binary.LittleEndian, uint32(3))

	var response DeviceSlotMappingsResponse
	if err := response.UnmarshalBinary(data.Bytes()); err != nil {
		t.Fatalf("UnmarshalBinary failed: %v", err)
	}
	if got := response.SlotMappings[0].Slot; got != 3 {
		t.Fatalf("SlotMappings[0].Slot = %d, want 3", got)
	}
	if got := response.SlotMappings[1].Slot; got != 7 {
		t.Fatalf("SlotMappings[1].Slot = %d, want 7", got)
	}
}

func TestDeviceSlotMappingsResponseRejectsInvalidSlotSize(t *testing.T) {
	data := []byte{
		0x01, 0x00, 0x00, 0x00,
		0x0c, 0x00, 0x00, 0x00,
		0x08, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00,
	}

	var response DeviceSlotMappingsResponse
	if err := response.UnmarshalBinary(data); err == nil {
		t.Fatal("UnmarshalBinary error = nil, want invalid size error")
	}
}

func TestOpenLogicalChannelRejectsPayloadStatus(t *testing.T) {
	payload := make([]byte, 16)
	binary.LittleEndian.PutUint32(payload[0:4], 1)
	conn := newFakeConn(commandDone(1, ServiceMsUiccLowLevelAccess, CIDUiccOpenChannel, payload))
	m := &MBIM{conn: conn}

	_, err := m.OpenLogicalChannel([]byte{0xA0, 0x00})
	if err == nil {
		t.Fatal("OpenLogicalChannel error = nil, want status error")
	}
	if !strings.Contains(err.Error(), "status 0x1") {
		t.Fatalf("OpenLogicalChannel error = %q, want status", err.Error())
	}
	if m.channel != 0 {
		t.Fatalf("channel = %d, want unchanged", m.channel)
	}
}

func TestCloseLogicalChannelRejectsPayloadStatus(t *testing.T) {
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, 2)
	conn := newFakeConn(commandDone(1, ServiceMsUiccLowLevelAccess, CIDUiccCloseChannel, payload))
	m := &MBIM{conn: conn}

	err := m.CloseLogicalChannel(1)
	if err == nil {
		t.Fatal("CloseLogicalChannel error = nil, want status error")
	}
	if !strings.Contains(err.Error(), "status 0x2") {
		t.Fatalf("CloseLogicalChannel error = %q, want status", err.Error())
	}
}

func TestConnectClosesConnectionOnFailure(t *testing.T) {
	conn := newFakeConn()
	m := &MBIM{conn: conn}

	err := m.Connect()
	if err == nil {
		t.Fatal("Connect error = nil, want read failure")
	}
	if !conn.closed {
		t.Fatal("Connect did not close connection after failure")
	}
}

func TestDisconnectSendsCloseMessageBeforeClosingConnection(t *testing.T) {
	conn := newFakeConn(closeDone(1))
	m := &MBIM{conn: conn}

	if err := m.Disconnect(); err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}
	if !conn.closed {
		t.Fatal("Disconnect did not close connection")
	}
	if len(conn.writes) != 1 {
		t.Fatalf("writes = %d, want 1", len(conn.writes))
	}
	if got := MessageType(binary.LittleEndian.Uint32(conn.writes[0][0:4])); got != MessageTypeClose {
		t.Fatalf("written message type = %#x, want close", got)
	}
}

func TestWaitForSlotActivationReportsLastReadyState(t *testing.T) {
	oldPollInterval := slotActivationPollInterval
	slotActivationPollInterval = time.Nanosecond
	t.Cleanup(func() {
		slotActivationPollInterval = oldPollInterval
	})

	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, MBIMSubscriberReadyStateNotActivated)
	responses := make([][]byte, 10)
	for i := range responses {
		responses[i] = commandDone(uint32(i+1), ServiceBasicConnect, CIDSubscriberReadyStatus, payload)
	}
	m := &MBIM{conn: newFakeConn(responses...)}

	err := m.waitForSlotActivation()
	if err == nil {
		t.Fatal("waitForSlotActivation error = nil, want not ready error")
	}
	if !strings.Contains(err.Error(), "last ready state 0x5") {
		t.Fatalf("waitForSlotActivation error = %q, want last ready state", err.Error())
	}
	if errors.Is(err, io.EOF) {
		t.Fatalf("waitForSlotActivation error wraps EOF despite successful status reads: %v", err)
	}
}
