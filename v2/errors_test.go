package sgp22

import (
	"testing"
)

func TestLoadBoundProfilePackageError(t *testing.T) {
	err := LoadBoundProfilePackageError{
		BPPCommandID: BPPCommandIDLoadProfileElements,
		ErrorReason:  BPPErrorReasonPPRNotAllowed,
	}

	if got, want := err.CommandID(), "loadProfileElements"; got != want {
		t.Errorf("CommandID() = %q, want %q", got, want)
	}
	if got, want := err.String(), "pprNotAllowed"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	if got, want := err.Error(), "loadProfileElements,pprNotAllowed"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestLoadBoundProfilePackageErrorUnknownCommand(t *testing.T) {
	err := LoadBoundProfilePackageError{
		BPPCommandID: BPPCommandID(99),
		ErrorReason:  BPPErrorReasonPPRNotAllowed,
	}

	if got, want := err.CommandID(), "unknown(99)"; got != want {
		t.Errorf("CommandID() = %q, want %q", got, want)
	}
	if got, want := err.Error(), "unknown(99),pprNotAllowed"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
