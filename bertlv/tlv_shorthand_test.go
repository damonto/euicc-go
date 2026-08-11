package bertlv

import (
	"bytes"
	"testing"

	"github.com/damonto/euicc-go/bertlv/primitive"
)

func TestNewValue(t *testing.T) {
	tlv := NewValue(Primitive.ContextSpecific(0), []byte{0xff})
	if want := (Tag{0x80}); !bytes.Equal(tlv.Tag, want) {
		t.Errorf("TLV.Tag = % X, want % X", tlv.Tag, want)
	}
	if len(tlv.Value) != 1 || len(tlv.Children) != 0 {
		t.Errorf("TLV lengths = value:%d children:%d, want 1 and 0", len(tlv.Value), len(tlv.Children))
	}
	encoded, err := tlv.Bytes()
	if err != nil {
		t.Fatalf("TLV.Bytes() error = %v", err)
	}
	if want := []byte{0x80, 0x01, 0xff}; !bytes.Equal(encoded, want) {
		t.Errorf("TLV.Bytes() = % X, want % X", encoded, want)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("NewValue() did not panic for constructed tag")
			}
		}()
		NewValue(Constructed.ContextSpecific(0), nil)
	}()
}

func TestNewChildren(t *testing.T) {
	tlv := NewChildren(Constructed.ContextSpecific(0))
	if want := (Tag{0xa0}); !bytes.Equal(tlv.Tag, want) {
		t.Errorf("TLV.Tag = % X, want % X", tlv.Tag, want)
	}
	if len(tlv.Value) != 0 {
		t.Errorf("len(TLV.Value) = %d, want 0", len(tlv.Value))
	}
	encoded, err := tlv.Bytes()
	if err != nil {
		t.Fatalf("TLV.Bytes() error = %v", err)
	}
	if want := []byte{0xa0, 0x00}; !bytes.Equal(encoded, want) {
		t.Errorf("TLV.Bytes() = % X, want % X", encoded, want)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("NewChildren() did not panic for primitive tag")
			}
		}()
		NewChildren(Primitive.ContextSpecific(0))
	}()
}

func TestNewChildrenIter(t *testing.T) {
	tlv := NewChildrenIter(Constructed.ContextSpecific(0), func(yield func(*TLV) bool) {
		if !yield(NewValue(Primitive.ContextSpecific(0), []byte{0xff})) {
			return
		}
		yield(NewValue(Primitive.ContextSpecific(1), []byte{0xff}))
	})
	if want := (Tag{0xa0}); !bytes.Equal(tlv.Tag, want) {
		t.Errorf("TLV.Tag = % X, want % X", tlv.Tag, want)
	}
	if len(tlv.Value) != 0 {
		t.Errorf("len(TLV.Value) = %d, want 0", len(tlv.Value))
	}
	encoded, err := tlv.Bytes()
	if err != nil {
		t.Fatalf("TLV.Bytes() error = %v", err)
	}
	if want := []byte{0xa0, 0x6, 0x80, 0x1, 0xff, 0x81, 0x1, 0xff}; !bytes.Equal(encoded, want) {
		t.Errorf("TLV.Bytes() = % X, want % X", encoded, want)
	}
}

func TestMarshalValue(t *testing.T) {
	tlv, err := MarshalValue(
		Primitive.ContextSpecific(0),
		primitive.MarshalInt[int8](-1),
	)
	if err != nil {
		t.Fatalf("MarshalValue() error = %v", err)
	}
	if want := (Tag{0x80}); !bytes.Equal(tlv.Tag, want) {
		t.Errorf("TLV.Tag = % X, want % X", tlv.Tag, want)
	}
	if len(tlv.Value) != 1 || len(tlv.Children) != 0 {
		t.Errorf("TLV lengths = value:%d children:%d, want 1 and 0", len(tlv.Value), len(tlv.Children))
	}
	encoded, err := tlv.Bytes()
	if err != nil {
		t.Fatalf("TLV.Bytes() error = %v", err)
	}
	if want := []byte{0x80, 0x01, 0xff}; !bytes.Equal(encoded, want) {
		t.Errorf("TLV.Bytes() = % X, want % X", encoded, want)
	}
}

func TestTLV_MarshalValue(t *testing.T) {
	tlv := NewValue(Primitive.ContextSpecific(0), nil)
	if err := tlv.MarshalValue(primitive.MarshalInt[int8](-1)); err != nil {
		t.Fatalf("TLV.MarshalValue() error = %v", err)
	}
	encoded, err := tlv.Bytes()
	if err != nil {
		t.Fatalf("TLV.Bytes() error = %v", err)
	}
	if want := []byte{0x80, 0x01, 0xff}; !bytes.Equal(encoded, want) {
		t.Errorf("TLV.Bytes() = % X, want % X", encoded, want)
	}
	var value int8
	if err := tlv.UnmarshalValue(primitive.UnmarshalInt(&value)); err != nil {
		t.Fatalf("TLV.UnmarshalValue() error = %v", err)
	}
	if value != -1 {
		t.Errorf("TLV.UnmarshalValue() = %d, want -1", value)
	}
}

func TestTLV_MarshalValueError(t *testing.T) {
	tlv := NewChildren(Constructed.ContextSpecific(0))
	var value int8
	if err := tlv.MarshalValue(primitive.MarshalInt(value)); err == nil {
		t.Error("TLV.MarshalValue() error = nil for constructed TLV")
	}
	if err := tlv.UnmarshalValue(primitive.UnmarshalInt(&value)); err == nil {
		t.Error("TLV.UnmarshalValue() error = nil for constructed TLV")
	}
}

func TestTLV_String(t *testing.T) {
	tlv := NewValue(Primitive.ContextSpecific(0), []byte{0xff})
	if got, want := tlv.String(), "[0] (1 byte)"; got != want {
		t.Errorf("TLV.String() = %q, want %q", got, want)
	}
	tlv = NewChildren(Constructed.ContextSpecific(0))
	if got, want := tlv.String(), "[0] (0 elem)"; got != want {
		t.Errorf("TLV.String() = %q, want %q", got, want)
	}
}
