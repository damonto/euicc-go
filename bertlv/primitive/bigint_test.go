package primitive

import (
	"bytes"
	"math/big"
	"testing"
)

func TestBigInt(t *testing.T) {
	var n big.Int
	if err := UnmarshalBigInt(&n).UnmarshalBinary([]byte{0x7f}); err != nil {
		t.Fatalf("UnmarshalBigInt() error = %v", err)
	}
	if got := n.Int64(); got != 127 {
		t.Errorf("UnmarshalBigInt() = %d, want 127", got)
	}
	data, err := MarshalBigInt(&n).MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBigInt() error = %v", err)
	}
	if want := []byte{0x7f}; !bytes.Equal(data, want) {
		t.Errorf("MarshalBigInt() = % X, want % X", data, want)
	}
}
