package qmi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/damonto/euicc-go/driver/qmi/core"
)

type captureResponse struct {
	payload []byte
}

func (r *captureResponse) UnmarshalResponse(tlvs *core.TLVs) error {
	value, ok := tlvs.Find(0x10)
	if !ok {
		return errors.New("missing payload TLV")
	}
	r.payload = append([]byte(nil), value.Value...)
	return nil
}

func TestReadSkipsResponsesForOtherTransactions(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	go func() {
		defer server.Close()
		_, _ = server.Write(encodeResponse(t, core.QMIServiceUIM, 7, 41, 0xAA))
		_, _ = server.Write(encodeResponse(t, core.QMIServiceUIM, 7, 42, 0xBB))
	}()

	response := &captureResponse{}
	request := &core.Request{
		ClientID:      7,
		TransactionID: 42,
		ServiceType:   core.QMIServiceUIM,
		Response:      response,
		ReadTimeout:   100 * time.Millisecond,
	}

	transport := &Transport{}
	if _, err := transport.Read(client, request); err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if !bytes.Equal(response.payload, []byte{0xBB}) {
		t.Fatalf("payload = %X, want BB", response.payload)
	}
}

func encodeResponse(t *testing.T, serviceType core.ServiceType, clientID uint8, txnID uint16, payload byte) []byte {
	t.Helper()

	value := new(bytes.Buffer)
	tlvs := core.TLVs{
		{Type: 0x02, Len: 4, Value: []byte{0x00, 0x00, 0x00, 0x00}},
		{Type: 0x10, Len: 1, Value: []byte{payload}},
	}
	if _, err := tlvs.WriteTo(value); err != nil {
		t.Fatalf("write TLVs: %v", err)
	}

	sdu := new(bytes.Buffer)
	if err := binary.Write(sdu, binary.LittleEndian, Header[uint16]{
		MessageType:   core.QMIMessageTypeResponse,
		TransactionID: txnID,
		MessageID:     core.QMIUIMSendAPDU,
		MessageLength: uint16(value.Len()),
	}); err != nil {
		t.Fatalf("write SDU header: %v", err)
	}
	if _, err := sdu.Write(value.Bytes()); err != nil {
		t.Fatalf("write SDU payload: %v", err)
	}

	packet := new(bytes.Buffer)
	if err := binary.Write(packet, binary.LittleEndian, QMUXHeader{
		IfType:       core.QMUXHeaderIfType,
		Length:       uint16(sdu.Len() + 5),
		ControlFlags: core.QMUXHeaderControlFlagRequest,
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
