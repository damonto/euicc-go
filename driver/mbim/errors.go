package mbim

import "fmt"

// MBIMProtocolError represents protocol-level errors from MBIM error messages.
type MBIMProtocolError uint32

const (
	MBIMProtocolErrorInvalid MBIMProtocolError = iota
	MBIMProtocolErrorTimeoutFragment
	MBIMProtocolErrorFragmentOutOfSequence
	MBIMProtocolErrorLengthMismatch
	MBIMProtocolErrorDuplicatedTID
	MBIMProtocolErrorNotOpened
	MBIMProtocolErrorUnknown
	MBIMProtocolErrorCancel
	MBIMProtocolErrorMaxTransfer
)

var mbimProtocolErrorText = map[MBIMProtocolError]string{
	MBIMProtocolErrorInvalid:               "Invalid MBIM protocol error",
	MBIMProtocolErrorTimeoutFragment:       "MBIM protocol error: Timeout Fragment",
	MBIMProtocolErrorFragmentOutOfSequence: "MBIM protocol error: Fragment Out Of Sequence",
	MBIMProtocolErrorLengthMismatch:        "MBIM protocol error: Length Mismatch",
	MBIMProtocolErrorDuplicatedTID:         "MBIM protocol error: Duplicated TID",
	MBIMProtocolErrorNotOpened:             "MBIM protocol error: Not Opened",
	MBIMProtocolErrorUnknown:               "MBIM protocol error: Unknown",
	MBIMProtocolErrorCancel:                "MBIM protocol error: Cancel",
	MBIMProtocolErrorMaxTransfer:           "MBIM protocol error: Max Transfer",
}

func (e MBIMProtocolError) Error() string {
	if text, ok := mbimProtocolErrorText[e]; ok {
		return text
	}
	return fmt.Sprintf("Unknown MBIM Protocol Error: %d", e)
}

// MBIMStatus represents status errors from the MBIM device.
type MBIMStatus uint32

const (
	MBIMStatusNone MBIMStatus = iota // 0x00000000
	MBIMStatusBusy
	MBIMStatusFailure
	MBIMStatusSimNotInserted
	MBIMStatusBadSim
	MBIMStatusPinRequired
	MBIMStatusPinDisabled
	MBIMStatusNotRegistered
	MBIMStatusProvidersNotFound
	MBIMStatusNoDeviceSupport
	MBIMStatusProviderNotVisible
	MBIMStatusDataClassNotAvailable
	MBIMStatusPacketServiceDetached
	MBIMStatusMaxActivatedContexts
	MBIMStatusNotInitialized
	MBIMStatusVoiceCallInProgress
	MBIMStatusContextNotActivated
	MBIMStatusServiceNotActivated
	MBIMStatusInvalidAccessString
	MBIMStatusInvalidUserNamePwd
	MBIMStatusRadioPowerOff
	MBIMStatusInvalidParameters
	MBIMStatusReadFailure
	MBIMStatusWriteFailure
	MBIMStatusReserved
	MBIMStatusNoPhonebook
	MBIMStatusParameterTooLong
	MBIMStatusStkBusy
	MBIMStatusOperationNotAllowed
	MBIMStatusMemoryFailure
	MBIMStatusInvalidMemoryIndex
	MBIMStatusMemoryFull
	MBIMStatusFilterNotSupported
	MBIMStatusDssInstanceLimit
	MBIMStatusInvalidDeviceServiceOperation
	MBIMStatusAuthIncorrectAutn
	MBIMStatusAuthSyncFailure
	MBIMStatusAuthAmfNotSet
	MBIMStatusContextNotSupported
	MBIMStatusSmsUnknownSmscAddress          MBIMStatus = 0x00000064 // 100
	MBIMStatusSmsNetworkTimeout              MBIMStatus = 0x00000065 // 101
	MBIMStatusSmsLangNotSupported            MBIMStatus = 0x00000066 // 102
	MBIMStatusSmsEncodingNotSupported        MBIMStatus = 0x00000067 // 103
	MBIMStatusSmsFormatNotSupported          MBIMStatus = 0x00000068 // 104
	MBIMStatusMsNoLogicalChannels            MBIMStatus = 0x87430001
	MBIMStatusMsSelectFailed                 MBIMStatus = 0x87430002
	MBIMStatusMsInvalidLogicalChannel        MBIMStatus = 0x87430003
	MBIMStatusInvalidSignature               MBIMStatus = 0x91000001
	MBIMStatusInvalidImei                    MBIMStatus = 0x91000002
	MBIMStatusInvalidTimestamp               MBIMStatus = 0x91000003
	MBIMStatusNetworkListTooLarge            MBIMStatus = 0x91000004
	MBIMStatusSignatureAlgorithmNotSupported MBIMStatus = 0x91000005
	MBIMStatusFeatureNotSupported            MBIMStatus = 0x91000006
	MBIMStatusDecodeOrParsingError           MBIMStatus = 0x91000007
)

var mbimStatusText = map[MBIMStatus]string{
	MBIMStatusNone:                           "Success",
	MBIMStatusBusy:                           "Busy",
	MBIMStatusFailure:                        "Failure",
	MBIMStatusSimNotInserted:                 "Sim Not Inserted",
	MBIMStatusBadSim:                         "Bad Sim",
	MBIMStatusPinRequired:                    "Pin Required",
	MBIMStatusPinDisabled:                    "Pin Disabled",
	MBIMStatusNotRegistered:                  "Not Registered",
	MBIMStatusProvidersNotFound:              "Providers Not Found",
	MBIMStatusNoDeviceSupport:                "No Device Support",
	MBIMStatusProviderNotVisible:             "Provider Not Visible",
	MBIMStatusDataClassNotAvailable:          "Data Class Not Available",
	MBIMStatusPacketServiceDetached:          "Packet Service Detached",
	MBIMStatusMaxActivatedContexts:           "Max Activated Contexts",
	MBIMStatusNotInitialized:                 "Not Initialized",
	MBIMStatusVoiceCallInProgress:            "Voice Call In Progress",
	MBIMStatusContextNotActivated:            "Context Not Activated",
	MBIMStatusServiceNotActivated:            "Service Not Activated",
	MBIMStatusInvalidAccessString:            "Invalid Access String",
	MBIMStatusInvalidUserNamePwd:             "Invalid User Name Pwd",
	MBIMStatusRadioPowerOff:                  "Radio Power Off",
	MBIMStatusInvalidParameters:              "Invalid Parameters",
	MBIMStatusReadFailure:                    "Read Failure",
	MBIMStatusWriteFailure:                   "Write Failure",
	MBIMStatusReserved:                       "Reserved",
	MBIMStatusNoPhonebook:                    "No Phonebook",
	MBIMStatusParameterTooLong:               "Parameter Too Long",
	MBIMStatusStkBusy:                        "Stk Busy",
	MBIMStatusOperationNotAllowed:            "Operation Not Allowed",
	MBIMStatusMemoryFailure:                  "Memory Failure",
	MBIMStatusInvalidMemoryIndex:             "Invalid Memory Index",
	MBIMStatusMemoryFull:                     "Memory Full",
	MBIMStatusFilterNotSupported:             "Filter Not Supported",
	MBIMStatusDssInstanceLimit:               "Dss Instance Limit",
	MBIMStatusInvalidDeviceServiceOperation:  "Invalid Device Service Operation",
	MBIMStatusAuthIncorrectAutn:              "Auth Incorrect Autn",
	MBIMStatusAuthSyncFailure:                "Auth Sync Failure",
	MBIMStatusAuthAmfNotSet:                  "Auth Amf Not Set",
	MBIMStatusContextNotSupported:            "Context Not Supported",
	MBIMStatusSmsUnknownSmscAddress:          "Sms Unknown Smsc Address",
	MBIMStatusSmsNetworkTimeout:              "Sms Network Timeout",
	MBIMStatusSmsLangNotSupported:            "Sms Lang Not Supported",
	MBIMStatusSmsEncodingNotSupported:        "Sms Encoding Not Supported",
	MBIMStatusSmsFormatNotSupported:          "Sms Format Not Supported",
	MBIMStatusMsNoLogicalChannels:            "Ms No Logical Channels",
	MBIMStatusMsSelectFailed:                 "Ms Select Failed",
	MBIMStatusMsInvalidLogicalChannel:        "Ms Invalid Logical Channel",
	MBIMStatusInvalidSignature:               "Invalid Signature",
	MBIMStatusInvalidImei:                    "Invalid Imei",
	MBIMStatusInvalidTimestamp:               "Invalid Timestamp",
	MBIMStatusNetworkListTooLarge:            "Network List Too Large",
	MBIMStatusSignatureAlgorithmNotSupported: "Signature Algorithm Not Supported",
	MBIMStatusFeatureNotSupported:            "Feature Not Supported",
	MBIMStatusDecodeOrParsingError:           "Decode Or Parsing Error",
}

func (e MBIMStatus) Error() string {
	if text, ok := mbimStatusText[e]; ok {
		return text
	}
	return fmt.Sprintf("Unknown MBIM Status Error: %d", e)
}
