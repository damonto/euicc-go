package primitive

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBitString(t *testing.T) {
	var bits BitString
	assert.NoError(t, UnmarshalBitString((*[]bool)(&bits)).UnmarshalBinary([]byte{0x06, 0x6E, 0x5D, 0xC0}))
	assert.Equal(t, "011011100101110111", bits.String())
	data, err := MarshalBitString(bits).MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x06, 0x6E, 0x5D, 0xC0}, data)
}

func TestBitStringError(t *testing.T) {
	var bits BitString
	assert.Error(t, UnmarshalBitString((*[]bool)(&bits)).UnmarshalBinary([]byte{0x08, 0x6E, 0x5D, 0xC0}))
}
