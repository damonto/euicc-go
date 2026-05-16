package qmi

import (
	"bytes"
	"strings"
	"testing"

	"github.com/damonto/euicc-go/driver/qmi/core"
)

func TestResponseRejectsQMUXLengthMismatch(t *testing.T) {
	packet := encodeResponse(t, core.QMIServiceUIM, 7, 42, 0xAA)
	packet[1]--

	var response Response
	err := response.UnmarshalBinary(packet)
	if err == nil || !strings.Contains(err.Error(), "QMUX length mismatch") {
		t.Fatalf("UnmarshalBinary error = %v, want QMUX length mismatch", err)
	}
}

func TestResponseRejectsTLVLengthMismatch(t *testing.T) {
	packet := encodeResponse(t, core.QMIServiceUIM, 7, 42, 0xAA)
	packet[11]++

	var response Response
	err := response.UnmarshalBinary(packet)
	if err == nil || !strings.Contains(err.Error(), "QMI TLV length mismatch") {
		t.Fatalf("UnmarshalBinary error = %v, want QMI TLV length mismatch", err)
	}
}

func TestBytesRejectsOversizedQMUXPacket(t *testing.T) {
	request := &core.Request{
		ClientID:      7,
		TransactionID: 42,
		ServiceType:   core.QMIServiceUIM,
		MessageID:     core.QMIUIMSendAPDU,
		Value: core.TLVs{
			{Type: 0x10, Len: core.MaxEncodedMessageLength, Value: bytes.Repeat([]byte{0xAA}, core.MaxEncodedMessageLength)},
		},
	}

	transport := &Transport{}
	if _, err := transport.bytes(request); err == nil {
		t.Fatal("bytes error = nil, want oversized message error")
	}
}
