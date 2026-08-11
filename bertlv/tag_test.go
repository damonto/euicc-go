package bertlv

import (
	"bytes"
	"math"
	"testing"
)

func TestNewTag(t *testing.T) {
	fixtures := map[uint64]Tag{
		0x00: {0xa0},
		0x1e: {0xbe},
		0x1f: {0xbf, 0x1f},
		0x7f: {0xbf, 0x7f},
		0x80: {0xbf, 0x81, 0x00},
	}
	for value, expected := range fixtures {
		if got := NewTag(ContextSpecific, Constructed, value); !bytes.Equal(got, expected) {
			t.Errorf("NewTag(%d) = % X, want % X", value, got, expected)
		}
	}
}

func TestTag_ReadFrom_Value(t *testing.T) {
	fixtures := map[uint64]Tag{
		0x00:           {0xa0},
		0x1e:           {0xbe},
		0x1f:           {0xbf, 0x1f},
		0x7f:           {0xbf, 0x7f},
		0x80:           {0xbf, 0x81, 0x00},
		math.MaxUint8:  {0xbf, 0x81, 0x7f},
		math.MaxUint16: {0xbf, 0x83, 0xff, 0x7f},
		math.MaxUint32: {0xbf, 0x8f, 0xff, 0xff, 0xff, 0x7f},
		math.MaxUint64: {0xbf, 0x81, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
	}
	var err error
	var tag = Tag{}
	for value, expected := range fixtures {
		_, err = tag.ReadFrom(bytes.NewReader(expected))
		if err != nil {
			t.Errorf("Tag.ReadFrom(% X) error = %v", expected, err)
			continue
		}
		if got := tag.Value(); got != value {
			t.Errorf("Tag.Value() = %d, want %d", got, value)
		}
		if !bytes.Equal(tag, expected) {
			t.Errorf("Tag.ReadFrom(% X) = % X", expected, tag)
		}
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
		if got := fixture.Tag.String(); got != fixture.ToString {
			t.Errorf("Tag.String() = %q, want %q", got, fixture.ToString)
		}
		if got := fixture.Tag.Class(); got != fixture.Class {
			t.Errorf("Tag.Class() = %d, want %d", got, fixture.Class)
		}
		if !fixture.Verifier(fixture.Tag) {
			t.Errorf("class verifier returned false for % X", *fixture.Tag)
		}
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
		if got := fixture.Tag.Form(); got != fixture.Form {
			t.Errorf("Tag.Form() = %d, want %d", got, fixture.Form)
		}
		if !fixture.Verifier(fixture.Tag) {
			t.Errorf("form verifier returned false for % X", *fixture.Tag)
		}
	}
}

func TestTag_If(t *testing.T) {
	var tag Tag
	tag = Tag{0b00_0_0_0000}
	if !tag.If(Universal, Primitive, 0) {
		t.Error("Tag.If() = false for universal primitive tag")
	}
	tag = Tag{0b01_1_0_0000}
	if !tag.If(Application, Constructed, 0) {
		t.Error("Tag.If() = false for application constructed tag")
	}
}

func TestTag_Equal(t *testing.T) {
	tag := Tag{0b01_1_0_0000}
	if !tag.Equal(Tag{0b01_1_0_0000}) {
		t.Error("Tag.Equal() = false for equal tags")
	}
	if tag.Equal(Tag{0b00_0_0_0001}) {
		t.Error("Tag.Equal() = true for different tags")
	}
}

func TestTag_UnmarshalBinary_Error(t *testing.T) {
	type Fixture struct {
		Tag   Tag
		Error string
	}
	fixtures := []*Fixture{
		{Tag{}, "tag encoding with less than one byte\nEOF"},
		{Tag{0xbf}, "tag encoding with more than 2 bytes\nEOF"},
		{Tag{0xbf, 0x1e}, "invalid high-tag-number encoding"},
		{Tag{0xbf, 0x80}, "invalid high-tag-number encoding"},
		{Tag{0xbf, 0x80, 0x80}, "invalid high-tag-number encoding"},
	}
	var err error
	var tag = Tag{}
	for _, fixture := range fixtures {
		_, err = tag.ReadFrom(bytes.NewReader(fixture.Tag))
		if err == nil || err.Error() != fixture.Error {
			t.Errorf("Tag.ReadFrom(% X) error = %v, want %q", fixture.Tag, err, fixture.Error)
		}
	}
}
