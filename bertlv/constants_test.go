package bertlv

import (
	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, fixture.Tag, fixture.Class(0))
		assert.Equal(t, fixture.Tag, fixture.Form(0))
	}
}
