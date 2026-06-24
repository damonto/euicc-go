package sgp22

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadBoundProfilePackageError(t *testing.T) {
	err := LoadBoundProfilePackageError{
		BPPCommandID: BPPCommandIDLoadProfileElements,
		ErrorReason:  BPPErrorReasonPPRNotAllowed,
	}

	assert.Equal(t, "loadProfileElements", err.CommandID())
	assert.Equal(t, "pprNotAllowed", err.String())
	assert.Equal(t, "loadProfileElements,pprNotAllowed", err.Error())
}

func TestLoadBoundProfilePackageErrorUnknownCommand(t *testing.T) {
	err := LoadBoundProfilePackageError{
		BPPCommandID: BPPCommandID(99),
		ErrorReason:  BPPErrorReasonPPRNotAllowed,
	}

	assert.Equal(t, "unknown(99)", err.CommandID())
	assert.Equal(t, "unknown(99),pprNotAllowed", err.Error())
}
