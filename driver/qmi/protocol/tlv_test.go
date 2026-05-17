package protocol

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestTLVReadFromReturnsWireLength(t *testing.T) {
	data := []byte{
		0x02, 0x04, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x10, 0x01, 0x00,
		0xaa,
	}
	var tlvs TLVs

	n, err := tlvs.ReadFrom(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ReadFrom failed: %v", err)
	}
	if n != int64(len(data)) {
		t.Fatalf("ReadFrom length = %d, want %d", n, len(data))
	}
}

func TestTLVWriteToReturnsWireLength(t *testing.T) {
	tlvs := TLVs{
		{Type: 0x02, Len: 4, Value: []byte{0x00, 0x00, 0x00, 0x00}},
		{Type: 0x10, Len: 1, Value: []byte{0xaa}},
	}
	var buf bytes.Buffer

	n, err := tlvs.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if n != int64(buf.Len()) {
		t.Fatalf("WriteTo length = %d, want %d", n, buf.Len())
	}
}

func TestTLVWriteToRejectsLengthMismatch(t *testing.T) {
	tlvs := TLVs{{Type: 0x10, Len: 2, Value: []byte{0xaa}}}

	_, err := tlvs.WriteTo(io.Discard)
	if err == nil {
		t.Fatal("WriteTo error = nil, want length mismatch")
	}
}

func TestQMIErrorFallbackIncludesCode(t *testing.T) {
	err := QMIError(65000)

	if got, want := err.Error(), "QMI error 65000"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, QMIError(65000)) {
		t.Fatal("QMIError should remain comparable through errors.Is")
	}
}
