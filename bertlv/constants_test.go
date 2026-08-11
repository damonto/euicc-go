package bertlv

import (
	"bytes"
	"testing"
)

func TestConstants(t *testing.T) {
	type Fixture struct {
		Tag   Tag
		Class func(uint64) Tag
		Form  func(uint64) Tag
	}
	fixtures := []Fixture{
		{Tag{0x00}, Universal.Primitive, Primitive.Universal},
		{Tag{0x20}, Universal.Constructed, Constructed.Universal},
		{Tag{0x40}, Application.Primitive, Primitive.Application},
		{Tag{0x60}, Application.Constructed, Constructed.Application},
		{Tag{0x80}, ContextSpecific.Primitive, Primitive.ContextSpecific},
		{Tag{0xa0}, ContextSpecific.Constructed, Constructed.ContextSpecific},
		{Tag{0xc0}, Private.Primitive, Primitive.Private},
		{Tag{0xe0}, Private.Constructed, Constructed.Private},
	}
	for _, fixture := range fixtures {
		if got := fixture.Class(0); !bytes.Equal(got, fixture.Tag) {
			t.Errorf("Class(0) = % X, want % X", got, fixture.Tag)
		}
		if got := fixture.Form(0); !bytes.Equal(got, fixture.Tag) {
			t.Errorf("Form(0) = % X, want % X", got, fixture.Tag)
		}
	}
}
