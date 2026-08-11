package bertlv

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestTLV_Len(t *testing.T) {
	var tlv TLV
	tlv.Tag = Primitive.ContextSpecific(0)
	tlv.Value = make([]byte, 127)
	if got := tlv.Len(); got != 129 {
		t.Errorf("TLV.Len() = %d, want 129", got)
	}
	tlv.Value = make([]byte, 255)
	if got := tlv.Len(); got != 258 {
		t.Errorf("TLV.Len() = %d, want 258", got)
	}
	tlv.Value = make([]byte, 256)
	if got := tlv.Len(); got != 260 {
		t.Errorf("TLV.Len() = %d, want 260", got)
	}
}

func TestTLV_WriteTo_Error(t *testing.T) {
	tlv := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.Application(0), []byte{0x01}),
		NewValue(Primitive.Application(1), []byte{0x01}),
	)
	var err error
	var w io.Writer
	w = &limitWriter{Limited: 0}
	_, err = tlv.WriteTo(w)
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Errorf("TLV.WriteTo(limit 0) error = %v, want closed pipe", err)
	}
	w = &limitWriter{Limited: 1}
	_, err = tlv.WriteTo(w)
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Errorf("TLV.WriteTo(limit 1) error = %v, want closed pipe", err)
	}
	w = &limitWriter{Limited: 3}
	_, err = tlv.WriteTo(w)
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Errorf("TLV.WriteTo(limit 3) error = %v, want closed pipe", err)
	}
}

func TestTLV_MarshalText(t *testing.T) {
	tlv := NewValue(Primitive.ContextSpecific(0), []byte{0xff})
	text, err := tlv.MarshalText()
	if err != nil {
		t.Fatalf("TLV.MarshalText() error = %v", err)
	}
	if want := []byte("gAH/"); !bytes.Equal(text, want) {
		t.Errorf("TLV.MarshalText() = %q, want %q", text, want)
	}
	err = tlv.UnmarshalText(text)
	if err != nil {
		t.Fatalf("TLV.UnmarshalText() error = %v", err)
	}
	if want := []byte{0xff}; !bytes.Equal(tlv.Value, want) {
		t.Errorf("TLV.Value = % X, want % X", tlv.Value, want)
	}
}

func TestTLVMarshalTextFlushesPadding(t *testing.T) {
	tlv := NewValue(Primitive.ContextSpecific(0), []byte{0xff, 0xee})

	text, err := tlv.MarshalText()
	if err != nil {
		t.Fatalf("TLV.MarshalText() error = %v", err)
	}
	if got, want := string(text), "gAL/7g=="; got != want {
		t.Errorf("TLV.MarshalText() = %q, want %q", got, want)
	}
}

func TestTLV_MarshalBinary(t *testing.T) {
	tlv := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.Universal(0), []byte{0xff}),
	)
	binary, err := tlv.MarshalBinary()
	if err != nil {
		t.Fatalf("TLV.MarshalBinary() error = %v", err)
	}
	if want := []byte{0xa0, 0x03, 0x00, 0x01, 0xff}; !bytes.Equal(binary, want) {
		t.Errorf("TLV.MarshalBinary() = % X, want % X", binary, want)
	}
	err = tlv.UnmarshalBinary(binary)
	if err != nil {
		t.Fatalf("TLV.UnmarshalBinary() error = %v", err)
	}
	if want := []byte{0xff}; !bytes.Equal(tlv.Children[0].Value, want) {
		t.Errorf("child value = % X, want % X", tlv.Children[0].Value, want)
	}
}

func TestTLV_MarshalJSON(t *testing.T) {
	tlv := NewValue(Primitive.ContextSpecific(0), []byte{0xff})
	encoded, err := json.Marshal(tlv)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if got, want := string(encoded), `"gAH/"`; got != want {
		t.Errorf("json.Marshal() = %q, want %q", got, want)
	}
	if err := json.Unmarshal(encoded, &tlv); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
}

func TestTLV_MarshalBERTLV(t *testing.T) {
	original := NewChildren(
		Constructed.Application(0),
		NewValue(Primitive.Application(1), []byte{0x01}),
	)
	cloned, err := original.MarshalBERTLV()
	if err != nil {
		t.Fatalf("TLV.MarshalBERTLV() error = %v", err)
	}
	if !reflect.DeepEqual(cloned, original) {
		t.Errorf("TLV.MarshalBERTLV() = %#v, want %#v", cloned, original)
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

func TestTLV_Clone(t *testing.T) {
	original := NewChildren(
		Constructed.Application(0),
		NewValue(Primitive.Application(0), []byte{0x01}),
		nil,
		NewValue(Primitive.Application(1), []byte{0x01}),
	)
	cloned := original.Clone()
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

func TestTLV_LargeValue(t *testing.T) {
	tlv := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.ContextSpecific(0), make([]byte, 0x10000)),
	)
	_, err := tlv.WriteTo(io.Discard)
	if err != nil {
		t.Fatalf("TLV.WriteTo() error = %v", err)
	}
}

func TestTLVInvalidConstructedTag(t *testing.T) {
	tlv := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.ContextSpecific(1), []byte{0x01}),
	)
	tlv.Tag = Primitive.ContextSpecific(0)
	if _, err := tlv.Bytes(); err == nil {
		t.Error("TLV.Bytes() error = nil for primitive tag with children")
	}
}

func TestTLVInvalidValueTag(t *testing.T) {
	tlv := NewValue(
		Primitive.ContextSpecific(0),
		[]byte{0xff},
	)
	tlv.Tag = Constructed.ContextSpecific(0)
	if _, err := tlv.Bytes(); err == nil {
		t.Error("TLV.Bytes() error = nil for constructed tag with value")
	}
}

type limitWriter struct {
	n       int
	Limited int
}

func (w *limitWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.n += n
	if w.n > w.Limited {
		return 0, io.ErrClosedPipe
	}
	return n, nil
}
