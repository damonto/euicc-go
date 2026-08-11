package ccid

import (
	"testing"
)

func TestSetReaderReturnsStateErrors(t *testing.T) {
	reader := &CCIDReader{closed: true}
	if err := reader.SetReader("reader"); err == nil {
		t.Fatal("SetReader() error = nil, want closed reader error")
	}

	reader = &CCIDReader{connected: true}
	if err := reader.SetReader("reader"); err == nil {
		t.Fatal("SetReader() error = nil, want connected reader error")
	}
}

func TestConnectRequiresReader(t *testing.T) {
	reader := New()
	err := reader.Connect()
	if err == nil {
		t.Fatal("Connect() error = nil, want reader required error")
	}
	if err.Error() != "ccid reader is required" {
		t.Fatalf("Connect() error = %q, want reader required", err.Error())
	}
}
