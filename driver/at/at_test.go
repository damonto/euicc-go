package at

import "testing"

func TestNewRejectsEmptyDevice(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("New() error = nil, want empty device error")
	}
}
