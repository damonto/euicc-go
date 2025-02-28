package sgp22

import (
	"errors"
	"fmt"
)

var (
	ErrUnexpectedTag   = errors.New("unexpected tag")
	ErrNothingToDelete = errors.New("nothing to delete")
	ErrICCIDNotFound   = errors.New("iccid not found")
	ErrCatBusy         = errors.New("cat busy")
	ErrUndefined       = errors.New("undefined error")
)

type LoadBoundProfilePackageError struct{ BPPCommandID, ErrorReason byte }

func (e LoadBoundProfilePackageError) CommandID() string {
	switch e.BPPCommandID {
	case 0:
		return "initialiseSecureChannel"
	case 1:
		return "configureISDP"
	case 2:
		return "storeMetadata"
	case 3:
		return "storeMetadata2"
	case 4:
		return "replaceSessionKeys"
	case 5:
		return "loadProfileElements"
	}
	return ""
}

func (e LoadBoundProfilePackageError) String() string {
	switch e.ErrorReason {
	case 1:
		return "incorrectInputValues"
	case 2:
		return "invalidSignature"
	case 3:
		return "invalidTransactionId"
	case 4:
		return "unsupportedCrtValues"
	case 5:
		return "unsupportedRemoteOperationType"
	case 6:
		return "unsupportedProfileClass"
	case 7:
		return "scp03tStructureError"
	case 8:
		return "scp03tSecurityError"
	case 9:
		return "installFailedDueToIccidAlreadyExistsOnEuicc"
	case 10:
		return "installFailedDueToInsufficientMemoryForProfile"
	case 11:
		return "installFailedDueToInterruption"
	case 12:
		return "installFailedDueToPEProcessingError "
	case 13:
		return "installFailedDueToDataMismatch"
	case 14:
		return "testProfileInstallFailedDueToInvalidNaaKey"
	case 15:
		return "pprNotAllowed"
	}
	return "installFailedDueToUnknownError"
}

func (e LoadBoundProfilePackageError) Error() string {
	return fmt.Sprintf("%s,%s", e.CommandID(), e.String())
}
