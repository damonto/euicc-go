package sgp22

import (
	"errors"
	"fmt"
)

var (
	ErrUnexpectedTag        = errors.New("unexpected tag")
	ErrIncorrectInputValues = errors.New("incorrect input values")
	ErrNothingToDelete      = errors.New("nothing to delete")
	ErrICCIDNotFound        = errors.New("iccid not found")
	ErrCatBusy              = errors.New("cat busy")
	ErrUndefined            = errors.New("undefined error")
)

type BPPCommandID int8

const (
	BPPCommandIDInitialiseSecureChannel BPPCommandID = 0
	BPPCommandIDConfigureISDP           BPPCommandID = 1
	BPPCommandIDStoreMetadata           BPPCommandID = 2
	BPPCommandIDStoreMetadata2          BPPCommandID = 3
	BPPCommandIDReplaceSessionKeys      BPPCommandID = 4
	BPPCommandIDLoadProfileElements     BPPCommandID = 5
)

func (id BPPCommandID) String() string {
	switch id {
	case BPPCommandIDInitialiseSecureChannel:
		return "initialiseSecureChannel"
	case BPPCommandIDConfigureISDP:
		return "configureISDP"
	case BPPCommandIDStoreMetadata:
		return "storeMetadata"
	case BPPCommandIDStoreMetadata2:
		return "storeMetadata2"
	case BPPCommandIDReplaceSessionKeys:
		return "replaceSessionKeys"
	case BPPCommandIDLoadProfileElements:
		return "loadProfileElements"
	}
	return fmt.Sprintf("unknown(%d)", id)
}

type BPPErrorReason int8

const (
	BPPErrorReasonIncorrectInputValues                           BPPErrorReason = 1
	BPPErrorReasonInvalidSignature                               BPPErrorReason = 2
	BPPErrorReasonInvalidTransactionID                           BPPErrorReason = 3
	BPPErrorReasonUnsupportedCRTValues                           BPPErrorReason = 4
	BPPErrorReasonUnsupportedRemoteOperationType                 BPPErrorReason = 5
	BPPErrorReasonUnsupportedProfileClass                        BPPErrorReason = 6
	BPPErrorReasonSCP03tStructureError                           BPPErrorReason = 7
	BPPErrorReasonSCP03tSecurityError                            BPPErrorReason = 8
	BPPErrorReasonInstallFailedDueToICCIDAlreadyExistsOnEUICC    BPPErrorReason = 9
	BPPErrorReasonInstallFailedDueToInsufficientMemoryForProfile BPPErrorReason = 10
	BPPErrorReasonInstallFailedDueToInterruption                 BPPErrorReason = 11
	BPPErrorReasonInstallFailedDueToPEProcessingError            BPPErrorReason = 12
	BPPErrorReasonInstallFailedDueToDataMismatch                 BPPErrorReason = 13
	BPPErrorReasonTestProfileInstallFailedDueToInvalidNAAKey     BPPErrorReason = 14
	BPPErrorReasonPPRNotAllowed                                  BPPErrorReason = 15
	BPPErrorReasonInstallFailedDueToUnknownError                 BPPErrorReason = 127
)

func (r BPPErrorReason) String() string {
	switch r {
	case BPPErrorReasonIncorrectInputValues:
		return "incorrectInputValues"
	case BPPErrorReasonInvalidSignature:
		return "invalidSignature"
	case BPPErrorReasonInvalidTransactionID:
		return "invalidTransactionId"
	case BPPErrorReasonUnsupportedCRTValues:
		return "unsupportedCrtValues"
	case BPPErrorReasonUnsupportedRemoteOperationType:
		return "unsupportedRemoteOperationType"
	case BPPErrorReasonUnsupportedProfileClass:
		return "unsupportedProfileClass"
	case BPPErrorReasonSCP03tStructureError:
		return "scp03tStructureError"
	case BPPErrorReasonSCP03tSecurityError:
		return "scp03tSecurityError"
	case BPPErrorReasonInstallFailedDueToICCIDAlreadyExistsOnEUICC:
		return "installFailedDueToIccidAlreadyExistsOnEuicc"
	case BPPErrorReasonInstallFailedDueToInsufficientMemoryForProfile:
		return "installFailedDueToInsufficientMemoryForProfile"
	case BPPErrorReasonInstallFailedDueToInterruption:
		return "installFailedDueToInterruption"
	case BPPErrorReasonInstallFailedDueToPEProcessingError:
		return "installFailedDueToPEProcessingError"
	case BPPErrorReasonInstallFailedDueToDataMismatch:
		return "installFailedDueToDataMismatch"
	case BPPErrorReasonTestProfileInstallFailedDueToInvalidNAAKey:
		return "testProfileInstallFailedDueToInvalidNaaKey"
	case BPPErrorReasonPPRNotAllowed:
		return "pprNotAllowed"
	case BPPErrorReasonInstallFailedDueToUnknownError:
		return "installFailedDueToUnknownError"
	}
	return fmt.Sprintf("unknown(%d)", r)
}

type LoadBoundProfilePackageError struct {
	BPPCommandID BPPCommandID
	ErrorReason  BPPErrorReason
}

type SetDefaultDPAddressResult int8

const (
	SetDefaultDPAddressResultOK             SetDefaultDPAddressResult = 0
	SetDefaultDPAddressResultUndefinedError SetDefaultDPAddressResult = 127
)

type ProfileInfoListErrorCode int8

const (
	ProfileInfoListErrorIncorrectInputValues ProfileInfoListErrorCode = 1
	ProfileInfoListErrorUndefinedError       ProfileInfoListErrorCode = 127
)

type EuiccMemoryResetResult int8

const (
	EuiccMemoryResetResultOK              EuiccMemoryResetResult = 0
	EuiccMemoryResetResultNothingToDelete EuiccMemoryResetResult = 1
	EuiccMemoryResetResultCATBusy         EuiccMemoryResetResult = 5
	EuiccMemoryResetResultUndefinedError  EuiccMemoryResetResult = 127
)

type SetNicknameResult int8

const (
	SetNicknameResultOK             SetNicknameResult = 0
	SetNicknameResultICCIDNotFound  SetNicknameResult = 1
	SetNicknameResultUndefinedError SetNicknameResult = 127
)

type DeleteNotificationStatus int8

const (
	DeleteNotificationStatusOK              DeleteNotificationStatus = 0
	DeleteNotificationStatusNothingToDelete DeleteNotificationStatus = 1
	DeleteNotificationStatusUndefinedError  DeleteNotificationStatus = 127
)

// ProfileOperationResult identifies EnableProfile, DisableProfile, or DeleteProfile result codes.
type ProfileOperationResult int8

const (
	ProfileOperationResultOK                        ProfileOperationResult = 0
	ProfileOperationResultICCIDOrAIDNotFound        ProfileOperationResult = 1
	ProfileOperationResultProfileNotInDisabledState ProfileOperationResult = 2
	ProfileOperationResultProfileNotInEnabledState  ProfileOperationResult = ProfileOperationResultProfileNotInDisabledState
	ProfileOperationResultDisallowedByPolicy        ProfileOperationResult = 3
	ProfileOperationResultWrongProfileReenabling    ProfileOperationResult = 4
	ProfileOperationResultCATBusy                   ProfileOperationResult = 5
	ProfileOperationResultUndefinedError            ProfileOperationResult = 127
)

// ProfileOperationError represents a non-ok profile operation result.
type ProfileOperationError struct {
	Operation ProfileOperation
	Result    ProfileOperationResult
}

func (o ProfileOperation) String() string {
	switch o {
	case EnableProfile:
		return "enableProfile"
	case DisableProfile:
		return "disableProfile"
	case DeleteProfile:
		return "deleteProfile"
	}
	return fmt.Sprintf("profileOperation(%d)", o)
}

func (e ProfileOperationError) Error() string {
	return fmt.Sprintf("%s,%s", e.Operation, e.ResultName())
}

func (e ProfileOperationError) ResultName() string {
	switch e.Result {
	case ProfileOperationResultICCIDOrAIDNotFound:
		return "iccidOrAidNotFound"
	case ProfileOperationResultProfileNotInDisabledState:
		if e.Operation == DisableProfile {
			return "profileNotInEnabledState"
		}
		return "profileNotInDisabledState"
	case ProfileOperationResultDisallowedByPolicy:
		return "disallowedByPolicy"
	case ProfileOperationResultWrongProfileReenabling:
		return "wrongProfileReenabling"
	case ProfileOperationResultCATBusy:
		return "catBusy"
	case ProfileOperationResultUndefinedError:
		return "undefinedError"
	}
	return fmt.Sprintf("unknown(%d)", e.Result)
}

func (e ProfileOperationError) Unwrap() error {
	switch e.Result {
	case ProfileOperationResultCATBusy:
		return ErrCatBusy
	case ProfileOperationResultUndefinedError:
		return ErrUndefined
	}
	return nil
}

// AuthenticateErrorCode identifies an ES10b.AuthenticateServer authenticateResponseError code.
type AuthenticateErrorCode int

const (
	AuthenticateErrorCodeInvalidCertificate     AuthenticateErrorCode = 1
	AuthenticateErrorCodeInvalidSignature       AuthenticateErrorCode = 2
	AuthenticateErrorCodeUnsupportedCurve       AuthenticateErrorCode = 3
	AuthenticateErrorCodeNoSessionContext       AuthenticateErrorCode = 4
	AuthenticateErrorCodeInvalidOID             AuthenticateErrorCode = 5
	AuthenticateErrorCodeEuiccChallengeMismatch AuthenticateErrorCode = 6
	AuthenticateErrorCodeCIPKUnknown            AuthenticateErrorCode = 7
	AuthenticateErrorCodeUndefinedError         AuthenticateErrorCode = 127
)

func (c AuthenticateErrorCode) String() string {
	switch c {
	case AuthenticateErrorCodeInvalidCertificate:
		return "invalidCertificate"
	case AuthenticateErrorCodeInvalidSignature:
		return "invalidSignature"
	case AuthenticateErrorCodeUnsupportedCurve:
		return "unsupportedCurve"
	case AuthenticateErrorCodeNoSessionContext:
		return "noSessionContext"
	case AuthenticateErrorCodeInvalidOID:
		return "invalidOid"
	case AuthenticateErrorCodeEuiccChallengeMismatch:
		return "euiccChallengeMismatch"
	case AuthenticateErrorCodeCIPKUnknown:
		return "ciPKUnknown"
	case AuthenticateErrorCodeUndefinedError:
		return "undefinedError"
	}
	return fmt.Sprintf("unknown(%d)", c)
}

// AuthenticateResponseError represents the authenticateResponseError branch of AuthenticateServerResponse.
type AuthenticateResponseError struct {
	TransactionID HexString
	ErrorCode     AuthenticateErrorCode
}

func (e AuthenticateResponseError) Error() string {
	return fmt.Sprintf("authenticateServer,%s", e.ErrorCode)
}

func (e LoadBoundProfilePackageError) CommandID() string {
	return e.BPPCommandID.String()
}

func (e LoadBoundProfilePackageError) String() string {
	return e.ErrorReason.String()
}

func (e LoadBoundProfilePackageError) Error() string {
	return fmt.Sprintf("%s,%s", e.CommandID(), e.String())
}
