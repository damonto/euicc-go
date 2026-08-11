package primitive

import (
	"bytes"
	"testing"
)

func TestBitString(t *testing.T) {
	var bits BitString
	if err := UnmarshalBitString((*[]bool)(&bits)).UnmarshalBinary([]byte{0x06, 0x6E, 0x5D, 0xC0}); err != nil {
		t.Fatalf("UnmarshalBitString() error = %v", err)
	}
	if got, want := bits.String(), "011011100101110111"; got != want {
		t.Errorf("BitString.String() = %q, want %q", got, want)
	}
	data, err := MarshalBitString(bits).MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBitString() error = %v", err)
	}
	if want := []byte{0x06, 0x6E, 0x5D, 0xC0}; !bytes.Equal(data, want) {
		t.Errorf("MarshalBitString() = % X, want % X", data, want)
	}
}

func TestBitStringMultipleOfEightBits(t *testing.T) {
	data, err := MarshalBitString([]bool{true, false, true, false, true, false, true, false}).MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBitString() error = %v", err)
	}
	if want := []byte{0x00, 0xaa}; !bytes.Equal(data, want) {
		t.Errorf("MarshalBitString() = % X, want % X", data, want)
	}
}

func TestBitStringError(t *testing.T) {
	var bits BitString
	if err := UnmarshalBitString((*[]bool)(&bits)).UnmarshalBinary([]byte{0x08, 0x6E, 0x5D, 0xC0}); err == nil {
		t.Error("UnmarshalBitString() error = nil for invalid padding")
	}
	if err := UnmarshalBitString((*[]bool)(&bits)).UnmarshalBinary(nil); err == nil {
		t.Error("UnmarshalBitString(nil) error = nil")
	}
}
