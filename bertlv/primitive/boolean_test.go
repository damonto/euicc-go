package primitive

import (
	"bytes"
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
			if err := UnmarshalBool(&parsed).UnmarshalBinary(variant); err != nil {
				t.Errorf("UnmarshalBool(% X) error = %v", variant, err)
				continue
			}
			if parsed != fixture.Expected {
				t.Errorf("UnmarshalBool(% X) = %t, want %t", variant, parsed, fixture.Expected)
			}
		}
		output, err = MarshalBool(fixture.Expected).MarshalBinary()
		if err != nil {
			t.Errorf("MarshalBool(%t) error = %v", fixture.Expected, err)
			continue
		}
		if want := fixture.Variants[0]; !bytes.Equal(output, want) {
			t.Errorf("MarshalBool(%t) = % X, want % X", fixture.Expected, output, want)
		}
	}
}

func TestBooleanRejectsInvalidLength(t *testing.T) {
	var parsed bool

	if err := UnmarshalBool(&parsed).UnmarshalBinary(nil); err == nil {
		t.Error("UnmarshalBool(nil) error = nil")
	}
	if err := UnmarshalBool(&parsed).UnmarshalBinary([]byte{0x00, 0x00}); err == nil {
		t.Error("UnmarshalBool() error = nil for two-byte input")
	}
}
