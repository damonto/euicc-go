package bertlv

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
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
		{[]byte{0x80, 0x81}, "tag 80: invalid length encoding\nfailed to read length: expected 1 bytes, got 0"},
		{[]byte{0xA0, 0x03, 0x00, 0x02}, "tag A0: invalid child object\ntag 00: invalid length encoding\nEOF"},
	}
	var err error
	var tlv TLV
	for _, fixture := range fixtures {
		_, err = tlv.ReadFrom(bytes.NewReader(fixture.TLV))
		assert.EqualError(t, err, fixture.Error)
	}
}

func TestTLV_FilterInvalidChildren(t *testing.T) {
	tlv := NewChildren(
		Constructed.ContextSpecific(0),
		NewValue(Primitive.ContextSpecific(1), []byte{0x01}),
		nil,
	)
	assert.Equal(t, []byte{0xa0, 0x03, 0x81, 0x01, 0x01}, tlv.Bytes())
	assert.NoError(t, tlv.UnmarshalBinary(tlv.Bytes()))
	assert.Len(t, tlv.Children, 1)
}

func TestTLV_UnmarshalJSONWithNewline(t *testing.T) {
	var tlv TLV
	assert.NoError(t, json.Unmarshal([]byte(`"gA\nH/"`), &tlv))
	assert.Equal(t, []byte{0xff}, tlv.Value)
}

func TestTLV_UnmarshalBERTLV(t *testing.T) {
	original := NewChildren(
		Constructed.Application(0),
		NewValue(Primitive.Application(1), []byte{0x01}),
	)
	var cloned TLV
	assert.NoError(t, cloned.UnmarshalBERTLV(original))
	assert.Equal(t, original, &cloned)
	assert.Equal(t, original.Bytes(), cloned.Bytes())
}
