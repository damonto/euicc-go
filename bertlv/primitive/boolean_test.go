package primitive

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBoolean(t *testing.T) {
	type Fixture struct {
		Expected bool
		Variants [][]byte
	}
	fixtures := []*Fixture{
		{false, [][]byte{{0x00}}},
		{true, [][]byte{{0xff}, {0x01}}},
	}
	var err error
	var output []byte
	for _, fixture := range fixtures {
		for _, variant := range fixture.Variants {
			var parsed bool
			assert.NoError(t, UnmarshalBool(&parsed).UnmarshalBinary(variant))
			assert.Equal(t, fixture.Expected, parsed)
		}
		output, err = MarshalBool(fixture.Expected).MarshalBinary()
		assert.NoError(t, err)
		assert.Equal(t, fixture.Variants[0], output)
	}
}

func TestBooleanRejectsInvalidLength(t *testing.T) {
	var parsed bool

	assert.Error(t, UnmarshalBool(&parsed).UnmarshalBinary(nil))
	assert.Error(t, UnmarshalBool(&parsed).UnmarshalBinary([]byte{0x00, 0x00}))
}
