package bertlv

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestTLV_ReadFrom(t *testing.T) {
	type Fixture struct {
		TLV   []byte
		Error string
	}
	fixtures := []*Fixture{
		{[]byte{}, "tag encoding with less than one byte\nEOF"},
		{[]byte{0x80, 0x01}, "tag 80: invalid length encoding\nEOF"},
		{[]byte{0x80, 0x81}, "tag 80: invalid length encoding\nread length: expected 1 bytes, got 0"},
		{[]byte{0xA0, 0x03, 0x00, 0x02}, "tag A0: invalid child object\ntag 00: invalid length encoding\nEOF"},
	}
	var err error
	var tlv TLV
	for _, fixture := range fixtures {
		_, err = tlv.ReadFrom(bytes.NewReader(fixture.TLV))
		if err == nil || err.Error() != fixture.Error {
			t.Errorf("TLV.ReadFrom(% X) error = %v, want %q", fixture.TLV, err, fixture.Error)
		}
	}
}

func TestTLV_ReadFromRejectsChildPastConstructedLength(t *testing.T) {
	var tlv TLV
	_, err := tlv.ReadFrom(bytes.NewReader([]byte{0xa0, 0x02, 0x80, 0x03, 0xff, 0xee, 0xdd}))

	if err == nil {
		t.Error("TLV.ReadFrom() error = nil for child past constructed length")
	}
}

func TestTLV_UnmarshalBinaryRejectsTrailingData(t *testing.T) {
	var tlv TLV

	err := tlv.UnmarshalBinary([]byte{0x80, 0x01, 0xff, 0x00})

	if want := "trailing data after TLV: 1 bytes"; err == nil || err.Error() != want {
		t.Errorf("TLV.UnmarshalBinary() error = %v, want %q", err, want)
	}
}

func TestTLV_FilterInvalidChildren(t *testing.T) {
	tlv := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.ContextSpecific(1), []byte{0x01}),
		nil,
	)
	encoded, err := tlv.Bytes()
	if err != nil {
		t.Fatalf("TLV.Bytes() error = %v", err)
	}
	if want := []byte{0xa0, 0x03, 0x81, 0x01, 0x01}; !bytes.Equal(encoded, want) {
		t.Errorf("TLV.Bytes() = % X, want % X", encoded, want)
	}
	if err := tlv.UnmarshalBinary(encoded); err != nil {
		t.Fatalf("TLV.UnmarshalBinary() error = %v", err)
	}
	if got := len(tlv.Children); got != 1 {
		t.Errorf("len(TLV.Children) = %d, want 1", got)
	}
}

func TestTLV_UnmarshalJSONWithNewline(t *testing.T) {
	var tlv TLV
	if err := json.Unmarshal([]byte(`"gA\nH/"`), &tlv); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if want := []byte{0xff}; !bytes.Equal(tlv.Value, want) {
		t.Errorf("TLV.Value = % X, want % X", tlv.Value, want)
	}
}

func TestTLV_UnmarshalJSONWithoutPadding(t *testing.T) {
	var tlv TLV
	if err := json.Unmarshal([]byte(`"gAL/7g"`), &tlv); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if want := Primitive.ContextSpecific(0); !bytes.Equal(tlv.Tag, want) {
		t.Errorf("TLV.Tag = % X, want % X", tlv.Tag, want)
	}
	if want := []byte{0xff, 0xee}; !bytes.Equal(tlv.Value, want) {
		t.Errorf("TLV.Value = % X, want % X", tlv.Value, want)
	}
}

func TestTLV_UnmarshalBERTLV(t *testing.T) {
	original := NewChildren(
		Constructed.Application(0),
		NewValue(Primitive.Application(1), []byte{0x01}),
	)
	var cloned TLV
	if err := cloned.UnmarshalBERTLV(original); err != nil {
		t.Fatalf("TLV.UnmarshalBERTLV() error = %v", err)
	}
	if !reflect.DeepEqual(&cloned, original) {
		t.Errorf("TLV.UnmarshalBERTLV() = %#v, want %#v", &cloned, original)
	}
	originalBytes, err := original.Bytes()
	if err != nil {
		t.Fatalf("original.Bytes() error = %v", err)
	}
	clonedBytes, err := cloned.Bytes()
	if err != nil {
		t.Fatalf("cloned.Bytes() error = %v", err)
	}
	if !bytes.Equal(clonedBytes, originalBytes) {
		t.Errorf("cloned.Bytes() = % X, want % X", clonedBytes, originalBytes)
	}
}
