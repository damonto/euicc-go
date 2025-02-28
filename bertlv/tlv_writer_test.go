package bertlv

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestTLV_Len(t *testing.T) {
	var tlv TLV
	tlv.Tag = Primitive.ContextSpecific(0)
	tlv.Value = make([]byte, 127)
	assert.Equal(t, 129, tlv.Len())
	tlv.Value = make([]byte, 255)
	assert.Equal(t, 258, tlv.Len())
	tlv.Value = make([]byte, 256)
	assert.Equal(t, 260, tlv.Len())
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
	assert.ErrorIs(t, err, io.ErrClosedPipe)
	w = &limitWriter{Limited: 1}
	_, err = tlv.WriteTo(w)
	assert.ErrorIs(t, err, io.ErrClosedPipe)
	w = &limitWriter{Limited: 3}
	_, err = tlv.WriteTo(w)
	assert.ErrorIs(t, err, io.ErrClosedPipe)
}

func TestTLV_MarshalText(t *testing.T) {
	tlv := NewValue(Primitive.ContextSpecific(0), []byte{0xff})
	text, err := tlv.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte("gAH/"), text)
	err = tlv.UnmarshalText(text)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xff}, tlv.Value)
}

func TestTLV_MarshalBinary(t *testing.T) {
	tlv := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.Universal(0), []byte{0xff}),
	)
	binary, err := tlv.MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xa0, 0x03, 0x00, 0x01, 0xff}, binary)
	err = tlv.UnmarshalBinary(binary)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xff}, tlv.Children[0].Value)
}

func TestTLV_MarshalJSON(t *testing.T) {
	tlv := NewValue(Primitive.ContextSpecific(0), []byte{0xff})
	encoded, err := json.Marshal(tlv)
	assert.NoError(t, err)
	assert.Equal(t, `"gAH/"`, string(encoded))
	assert.NoError(t, json.Unmarshal(encoded, &tlv))
}

func TestTLV_MarshalBERTLV(t *testing.T) {
	original := NewChildren(
		Constructed.Application(0),
		NewValue(Primitive.Application(1), []byte{0x01}),
	)
	cloned, err := original.MarshalBERTLV()
	assert.NoError(t, err)
	assert.Equal(t, original, cloned)
	assert.Equal(t, original.Bytes(), cloned.Bytes())
}

func TestTLV_Clone(t *testing.T) {
	original := NewChildren(
		Constructed.Application(0),
		NewValue(Primitive.Application(0), []byte{0x01}),
		nil,
		NewValue(Primitive.Application(1), []byte{0x01}),
	)
	cloned := original.Clone()
	assert.Equal(t, original.Bytes(), cloned.Bytes())
}

func TestTLV_LargeValue(t *testing.T) {
	tlv := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.ContextSpecific(0), make([]byte, 0x10000)),
	)
	_, err := tlv.MarshalBinary()
	assert.EqualError(t, err, "tlv: length exceeds maximum (65535), got 65540")
}

func TestTLV_InvalidConstructedTag(t *testing.T) {
	tlv := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.ContextSpecific(1), []byte{0x01}),
	)
	tlv.Tag = Primitive.ContextSpecific(0)
	assert.Panics(t, func() { tlv.Bytes() })
}

func TestTLV_InvalidValueTag(t *testing.T) {
	tlv := NewValue(
		Primitive.ContextSpecific(0),
		[]byte{0xff},
	)
	tlv.Tag = Constructed.ContextSpecific(0)
	assert.Panics(t, func() { tlv.Bytes() })
}

type limitWriter struct {
	n       int
	Limited int
}

func (w *limitWriter) Write(p []byte) (n int, err error) {
	w.n += len(p)
	if w.n > w.Limited {
		return 0, io.ErrClosedPipe
	}
	return n, nil
}
