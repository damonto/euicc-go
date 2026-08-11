package sgp22

import (
	"bytes"
	"testing"
)

func TestICCID_String(t *testing.T) {
	var iccid ICCID
	var parsed ICCID
	var err error

	// Standard ICCID
	iccid = ICCID{0x98, 0x44, 0x74, 0x68, 0x00, 0x00, 0x54, 0x37, 0x21, 0xF8}
	if got, want := iccid.String(), "8944478600004573128"; got != want {
		t.Errorf("ICCID.String() = %q, want %q", got, want)
	}
	parsed, err = NewICCID(iccid.String())
	if err != nil {
		t.Fatalf("NewICCID() error = %v", err)
	}
	if !bytes.Equal(parsed, iccid) {
		t.Errorf("NewICCID() = % X, want % X", parsed, iccid)
	}

	// Non-standard ICCID
	// 89860110F9900160570
	iccid = ICCID{0x98, 0x68, 0x10, 0x01, 0x9F, 0x09, 0x10, 0x06, 0x75, 0xF0}
	if got, want := iccid.String(), "89860110f9900160570"; got != want {
		t.Errorf("ICCID.String() = %q, want %q", got, want)
	}
	parsed, err = NewICCID(iccid.String())
	if err != nil {
		t.Fatalf("NewICCID() error = %v", err)
	}
	if !bytes.Equal(parsed, iccid) {
		t.Errorf("NewICCID() = % X, want % X", parsed, iccid)
	}
}
