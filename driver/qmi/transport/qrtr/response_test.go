package qrtr

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/damonto/euicc-go/driver/qmi/core"
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
	request := &core.Request{
		TransactionID: 42,
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

func encodeResponse(t *testing.T, txnID uint16, payload byte) []byte {
	t.Helper()

	value := new(bytes.Buffer)
	tlvs := core.TLVs{
		{Type: 0x02, Len: 4, Value: []byte{0x00, 0x00, 0x00, 0x00}},
		{Type: 0x10, Len: 1, Value: []byte{payload}},
	}
	if _, err := tlvs.WriteTo(value); err != nil {
		t.Fatalf("write TLVs: %v", err)
	}

	packet := new(bytes.Buffer)
	mustWrite(t, packet, core.QMIMessageTypeResponse)
	mustWrite(t, packet, txnID)
	mustWrite(t, packet, core.QMIUIMSendAPDU)
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
