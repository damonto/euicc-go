package bertlv

import (
	"bytes"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTag(t *testing.T) {
	fixtures := map[uint64]Tag{
		0x00:           {0xa0},
		0x1e:           {0xbe},
		0x1f:           {0xbf, 0x1f},
		0x7f:           {0xbf, 0x80, 0x7f},
		0x80:           {0xbf, 0x81, 0x00},
		math.MaxUint8:  {0xbf, 0x81, 0x7f},
		math.MaxUint16: {0xbf, 0x83, 0xff, 0x7f},
		math.MaxUint32: {0xbf, 0x8f, 0xff, 0xff, 0xff, 0x7f},
		math.MaxUint64: {0xbf, 0x81, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
	}
	var err error
	var tag = Tag{}
	for value, expected := range fixtures {
		assert.Equal(t, expected, NewTag(ContextSpecific, Constructed, value))

		_, err = tag.ReadFrom(bytes.NewReader(expected))
		assert.NoError(t, err)
		assert.Equal(t, value, tag.Value())
		assert.Equal(t, expected, tag)
	}
}

func TestTag_Class(t *testing.T) {
	type Fixture struct {
		Tag      *Tag
		Class    Class
		ToString string
		Verifier func(*Tag) bool
	}
	fixtures := []*Fixture{
		{&Tag{0b00_0_0_0000}, Universal, "[UNIVERSAL 0]", (*Tag).Universal},
		{&Tag{0b01_0_0_0000}, Application, "[APPLICATION 0]", (*Tag).Application},
		{&Tag{0b10_0_0_0000}, ContextSpecific, "[0]", (*Tag).ContextSpecific},
		{&Tag{0b11_0_0_0000}, Private, "[PRIVATE 0]", (*Tag).Private},
	}
	for _, fixture := range fixtures {
		assert.Equal(t, fixture.ToString, fixture.Tag.String())
		assert.Equal(t, fixture.Class, fixture.Tag.Class())
		assert.True(t, fixture.Verifier(fixture.Tag))
	}
}

func TestTag_Form(t *testing.T) {
	type Fixture struct {
		Tag      *Tag
		Form     Form
		Verifier func(*Tag) bool
	}
	fixtures := []*Fixture{
		{&Tag{0b00_0_0_0000}, Primitive, (*Tag).Primitive},
		{&Tag{0b00_1_0_0000}, Constructed, (*Tag).Constructed},
	}
	for _, fixture := range fixtures {
		assert.Equal(t, fixture.Form, fixture.Tag.Form())
		assert.True(t, fixture.Verifier(fixture.Tag))
	}
}

func TestTag_If(t *testing.T) {
	var tag Tag
	tag = Tag{0b00_0_0_0000}
	assert.True(t, tag.If(Universal, Primitive, 0))
	tag = Tag{0b01_1_0_0000}
	assert.True(t, tag.If(Application, Constructed, 0))
}

func TestTag_Equal(t *testing.T) {
	tag := Tag{0b01_1_0_0000}
	assert.True(t, tag.Equal(Tag{0b01_1_0_0000}))
	assert.False(t, tag.Equal(Tag{0b00_0_0_0001}))
}

func TestTag_UnmarshalBinary_Error(t *testing.T) {
	type Fixture struct {
		Tag   Tag
		Error string
	}
	fixtures := []*Fixture{
		{Tag{}, "tag encoding with less than one byte\nEOF"},
		{Tag{0xbf}, "tag encoding with more than 2 bytes\nEOF"},
		{Tag{0xbf, 0x80}, "tag encoding with more than 3 bytes\nEOF"},
		{Tag{0xbf, 0x80, 0x80}, "tag encoding with more than 4 bytes\nEOF"},
	}
	var err error
	var tag = Tag{}
	for _, fixture := range fixtures {
		_, err = tag.ReadFrom(bytes.NewReader(fixture.Tag))
		assert.EqualError(t, err, fixture.Error)
	}
}
