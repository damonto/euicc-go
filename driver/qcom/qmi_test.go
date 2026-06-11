package qcom

import (
	"context"
	"testing"
)

type fakeUIMReader struct {
	closed bool
}

func (f *fakeUIMReader) ActivateSlot(context.Context) error { return nil }
func (f *fakeUIMReader) OpenLogicalChannel(context.Context, []byte) (uint8, error) {
	return 1, nil
}
func (f *fakeUIMReader) SendAPDU(context.Context, uint8, []byte) ([]byte, error) {
	return []byte{0x90, 0x00}, nil
}
func (f *fakeUIMReader) CloseLogicalChannel(context.Context, uint8) error { return nil }
func (f *fakeUIMReader) Close() error {
	f.closed = true
	return nil
}

func TestDisconnectClosesReader(t *testing.T) {
	reader := &fakeUIMReader{}
	q := &QMI{channel: newChannel(reader)}

	if err := q.Disconnect(); err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}
	if !reader.closed {
		t.Fatal("Disconnect() did not close reader")
	}
}

func TestNewQMIRejectsInvalidSlotBeforeOpen(t *testing.T) {
	for _, slot := range []uint8{0, 6} {
		_, err := NewQMI("/dev/cdc-wdm1", slot)
		if err == nil {
			t.Fatalf("NewQMI() error = nil for slot %d, want invalid slot error", slot)
		}
		if err.Error() != "slot must be between 1 and 5" {
			t.Fatalf("NewQMI() error = %q, want local slot validation", err.Error())
		}
	}
}

func TestNewQRTRRejectsInvalidSlotBeforeOpen(t *testing.T) {
	for _, slot := range []uint8{0, 6} {
		_, err := NewQRTR(slot)
		if err == nil {
			t.Fatalf("NewQRTR() error = nil for slot %d, want invalid slot error", slot)
		}
		if err.Error() != "slot must be between 1 and 5" {
			t.Fatalf("NewQRTR() error = %q, want local slot validation", err.Error())
		}
	}
}
