package primitive

import (
	"bytes"
	"math"
	"testing"
)

func TestInteger(t *testing.T) {
	testInt(t, map[int64][][]byte{
		0:             {{0x00}, {0x00, 0x00}},
		127:           {{0x7f}, {0x00, 0x7f}},
		128:           {{0x00, 0x80}},
		1000:          {{0x03, 0xe8}, {0x00, 0x00, 0x03, 0xe8}},
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
	if err := UnmarshalInt(&value).UnmarshalBinary(nil); err == nil {
		t.Error("UnmarshalInt(nil) error = nil")
	}
	for _, test := range []struct {
		input []byte
		want  int8
	}{
		{[]byte{0x00, 0x7f}, 127},
		{[]byte{0xff, 0x80}, -128},
		{[]byte{0xff, 0xff}, -1},
	} {
		if err := UnmarshalInt(&value).UnmarshalBinary(test.input); err != nil {
			t.Errorf("UnmarshalInt(% X) error = %v", test.input, err)
		} else if value != test.want {
			t.Errorf("UnmarshalInt(% X) = %d, want %d", test.input, value, test.want)
		}
	}
	for _, input := range [][]byte{{0x00, 0x80}, {0xff, 0x7f}} {
		if err := UnmarshalInt(&value).UnmarshalBinary(input); err == nil {
			t.Errorf("UnmarshalInt(% X) error = nil", input)
		}
	}
}

func testInt[T signedInt](t *testing.T, fixtures map[T][][]byte) {
	t.Helper()
	for expected, variants := range fixtures {
		var value T
		for _, variant := range variants {
			if err := UnmarshalInt(&value).UnmarshalBinary(variant); err != nil {
				t.Errorf("UnmarshalInt(% X) error = %v", variant, err)
				continue
			}
			if value != expected {
				t.Errorf("UnmarshalInt(% X) = %v, want %v", variant, value, expected)
			}
		}
		actual, err := MarshalInt(expected).MarshalBinary()
		if err != nil {
			t.Errorf("MarshalInt(%v) error = %v", expected, err)
			continue
		}
		if want := variants[0]; !bytes.Equal(actual, want) {
			t.Errorf("MarshalInt(%v) = % X, want % X", expected, actual, want)
		}
	}
}
