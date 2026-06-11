package qmi

import (
	"testing"

	"github.com/damonto/euicc-go/driver/qcom"
)

func TestNewMatchesNewQMIInvalidSlot(t *testing.T) {
	for _, slot := range []uint8{0, 6} {
		_, deprecatedErr := New("/dev/cdc-wdm1", slot)
		_, currentErr := qcom.NewQMI("/dev/cdc-wdm1", slot)
		if deprecatedErr == nil {
			t.Fatalf("New() error = nil for slot %d, want invalid slot error", slot)
		}
		if currentErr == nil {
			t.Fatalf("NewQMI() error = nil for slot %d, want invalid slot error", slot)
		}
		if deprecatedErr.Error() != currentErr.Error() {
			t.Fatalf("New() error = %q, want %q", deprecatedErr.Error(), currentErr.Error())
		}
	}
}

func TestNewQRTRMatchesQCOMInvalidSlot(t *testing.T) {
	for _, slot := range []uint8{0, 6} {
		_, deprecatedErr := NewQRTR(slot)
		_, currentErr := qcom.NewQRTR(slot)
		if deprecatedErr == nil {
			t.Fatalf("NewQRTR() error = nil for slot %d, want invalid slot error", slot)
		}
		if currentErr == nil {
			t.Fatalf("qcom.NewQRTR() error = nil for slot %d, want invalid slot error", slot)
		}
		if deprecatedErr.Error() != currentErr.Error() {
			t.Fatalf("NewQRTR() error = %q, want %q", deprecatedErr.Error(), currentErr.Error())
		}
	}
}
