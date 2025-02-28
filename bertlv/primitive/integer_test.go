package primitive

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestInteger(t *testing.T) {
	testInt(t, map[int64][][]byte{
		0:             {{0x00}},
		127:           {{0x7f}, {0x00, 0x7f}},
		128:           {{0x00, 0x80}},
		256:           {{0x01, 0x00}},
		-1:            {{0xff}, {0xff, 0xff}, {0xff, 0xff, 0xff, 0xff}},
		-128:          {{0x80}, {0xff, 0x80}},
		-129:          {{0xff, 0x7f}, {0xff, 0xff, 0xff, 0x7f}},
		-1000:         {{0xfc, 0x18}, {0xff, 0xff, 0xfc, 0x18}},
		-8388607:      {{0x80, 0x00, 0x01}},
		math.MaxInt64: {{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		math.MinInt64: {{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	})
	testInt(t, map[int32][][]byte{
		math.MinInt32: {{0x80, 0x00, 0x00, 0x00}},
		math.MaxInt32: {{0x7f, 0xff, 0xff, 0xff}},
	})
	testInt(t, map[int16][][]byte{
		math.MinInt16: {{0x80, 0x00}},
		math.MaxInt16: {{0x7f, 0xff}},
	})
	testInt(t, map[int8][][]byte{
		0:            {{0x00}},
		1:            {{0x01}},
		math.MaxInt8: {{0x7f}},
		math.MinInt8: {{0x80}},
		-1:           {{0xff}},
	})
}

func TestIntegerError(t *testing.T) {
	var value int8
	assert.NoError(t, UnmarshalInt(&value).UnmarshalBinary(nil))
	assert.Error(t, UnmarshalInt(&value).UnmarshalBinary([]byte{0xff, 0xff}))
}

func testInt[T signedInt](t *testing.T, fixtures map[T][][]byte) {
	for expected, variants := range fixtures {
		var value T
		for _, variant := range variants {
			assert.NoError(t, UnmarshalInt(&value).UnmarshalBinary(variant))
			assert.Equal(t, expected, value, "UnmarshalBinary")
		}
		actual, err := MarshalInt(expected).MarshalBinary()
		assert.NoError(t, err)
		assert.Equal(t, variants[0], actual, "MarshalBinary")
	}
}
