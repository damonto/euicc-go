package bertlv

import (
	"testing"

	"github.com/damonto/euicc-go/bertlv/primitive"
	"github.com/stretchr/testify/assert"
)

func TestNewValue(t *testing.T) {
	tlv := NewValue(Primitive.ContextSpecific(0), []byte{0xff})
	assert.Equal(t, Tag{0x80}, tlv.Tag)
	assert.Len(t, tlv.Value, 1)
	assert.Len(t, tlv.Children, 0)
	assert.Equal(t, []byte{0x80, 0x01, 0xff}, tlv.Bytes())
	assert.Panics(t, func() { NewValue(Constructed.ContextSpecific(0), nil) })
}

func TestNewChildren(t *testing.T) {
	tlv := NewChildren(Constructed.ContextSpecific(0))
	assert.Equal(t, Tag{0xa0}, tlv.Tag)
	assert.Len(t, tlv.Value, 0)
	assert.Equal(t, []byte{0xa0, 0x00}, tlv.Bytes())
	assert.Panics(t, func() { NewChildren(Primitive.ContextSpecific(0)) })
}

func TestNewChildrenIter(t *testing.T) {
	tlv := NewChildrenIter(Constructed.ContextSpecific(0), func(yield func(*TLV) bool) {
		if !yield(NewValue(Primitive.ContextSpecific(0), []byte{0xff})) {
			return
		}
		yield(NewValue(Primitive.ContextSpecific(1), []byte{0xff}))
	})
	assert.Equal(t, Tag{0xa0}, tlv.Tag)
	assert.Len(t, tlv.Value, 0)
	assert.Equal(t, []byte{0xa0, 0x6, 0x80, 0x1, 0xff, 0x81, 0x1, 0xff}, tlv.Bytes())
}

func TestMarshalValue(t *testing.T) {
	tlv, err := MarshalValue(
		Primitive.ContextSpecific(0),
		primitive.MarshalInt[int8](-1),
	)
	assert.NoError(t, err)
	assert.Equal(t, Tag{0x80}, tlv.Tag)
	assert.Len(t, tlv.Value, 1)
	assert.Len(t, tlv.Children, 0)
	assert.Equal(t, []byte{0x80, 0x01, 0xff}, tlv.Bytes())
}

func TestTLV_MarshalValue(t *testing.T) {
	tlv := NewValue(Primitive.ContextSpecific(0), nil)
	assert.NoError(t, tlv.MarshalValue(primitive.MarshalInt[int8](-1)))
	assert.Equal(t, []byte{0x80, 0x01, 0xff}, tlv.Bytes())
	var value int8
	assert.NoError(t, tlv.UnmarshalValue(primitive.UnmarshalInt(&value)))
	assert.Equal(t, int8(-1), value)
}

func TestTLV_MarshalValueError(t *testing.T) {
	tlv := NewChildren(Constructed.ContextSpecific(0))
	var value int8
	assert.Error(t, tlv.MarshalValue(primitive.MarshalInt(value)))
	assert.Error(t, tlv.UnmarshalValue(primitive.UnmarshalInt(&value)))
}

func TestTLV_String(t *testing.T) {
	tlv := NewValue(Primitive.ContextSpecific(0), []byte{0xff})
	assert.Equal(t, "[0] (1 byte)", tlv.String())
	tlv = NewChildren(Constructed.ContextSpecific(0))
	assert.Equal(t, "[0] (0 elem)", tlv.String())
}
