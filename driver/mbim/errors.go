package mbim

import "fmt"

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

func (e MBIMStatus) Error() string {
	switch e {
	case MBIMStatusNone:
		return "Success"
	case MBIMStatusBusy:
		return "Busy"
	case MBIMStatusFailure:
		return "Failure"
	case MBIMStatusSimNotInserted:
		return "Sim Not Inserted"
	case MBIMStatusBadSim:
		return "Bad Sim"
	case MBIMStatusPinRequired:
		return "Pin Required"
	case MBIMStatusPinDisabled:
		return "Pin Disabled"
	case MBIMStatusNotRegistered:
		return "Not Registered"
	case MBIMStatusProvidersNotFound:
		return "Providers Not Found"
	case MBIMStatusNoDeviceSupport:
		return "No Device Support"
	case MBIMStatusProviderNotVisible:
		return "Provider Not Visible"
	case MBIMStatusDataClassNotAvailable:
		return "Data Class Not Available"
	case MBIMStatusPacketServiceDetached:
		return "Packet Service Detached"
	case MBIMStatusMaxActivatedContexts:
		return "Max Activated Contexts"
	case MBIMStatusNotInitialized:
		return "Not Initialized"
	case MBIMStatusVoiceCallInProgress:
		return "Voice Call In Progress"
	case MBIMStatusContextNotActivated:
		return "Context Not Activated"
	case MBIMStatusServiceNotActivated:
		return "Service Not Activated"
	case MBIMStatusInvalidAccessString:
		return "Invalid Access String"
	case MBIMStatusInvalidUserNamePwd:
		return "Invalid User Name Pwd"
	case MBIMStatusRadioPowerOff:
		return "Radio Power Off"
	case MBIMStatusInvalidParameters:
		return "Invalid Parameters"
	case MBIMStatusReadFailure:
		return "Read Failure"
	case MBIMStatusWriteFailure:
		return "Write Failure"
	case MBIMStatusNoPhonebook:
		return "No Phonebook"
	case MBIMStatusParameterTooLong:
		return "Parameter Too Long"
	case MBIMStatusStkBusy:
		return "Stk Busy"
	case MBIMStatusOperationNotAllowed:
		return "Operation Not Allowed"
	case MBIMStatusMemoryFailure:
		return "Memory Failure"
	case MBIMStatusInvalidMemoryIndex:
		return "Invalid Memory Index"
	case MBIMStatusMemoryFull:
		return "Memory Full"
	case MBIMStatusFilterNotSupported:
		return "Filter Not Supported"
	case MBIMStatusDssInstanceLimit:
		return "Dss Instance Limit"
	case MBIMStatusInvalidDeviceServiceOperation:
		return "Invalid Device Service Operation"
	case MBIMStatusAuthIncorrectAutn:
		return "Auth Incorrect Autn"
	case MBIMStatusAuthSyncFailure:
		return "Auth Sync Failure"
	case MBIMStatusAuthAmfNotSet:
		return "Auth Amf Not Set"
	case MBIMStatusContextNotSupported:
		return "Context Not Supported"
	case MBIMStatusSmsUnknownSmscAddress:
		return "Sms Unknown Smsc Address"
	case MBIMStatusSmsNetworkTimeout:
		return "Sms Network Timeout"
	case MBIMStatusSmsLangNotSupported:
		return "Sms Lang Not Supported"
	case MBIMStatusSmsEncodingNotSupported:
		return "Sms Encoding Not Supported"
	case MBIMStatusSmsFormatNotSupported:
		return "Sms Format Not Supported"
	case MBIMStatusMsNoLogicalChannels:
		return "Ms No Logical Channels"
	case MBIMStatusMsSelectFailed:
		return "Ms Select Failed"
	case MBIMStatusMsInvalidLogicalChannel:
		return "Ms Invalid Logical Channel"
	case MBIMStatusInvalidSignature:
		return "Invalid Signature"
	case MBIMStatusInvalidImei:
		return "Invalid Imei"
	case MBIMStatusInvalidTimestamp:
		return "Invalid Timestamp"
	case MBIMStatusNetworkListTooLarge:
		return "Network List Too Large"
	case MBIMStatusSignatureAlgorithmNotSupported:
		return "Signature Algorithm Not Supported"
	case MBIMStatusFeatureNotSupported:
		return "Feature Not Supported"
	case MBIMStatusDecodeOrParsingError:
		return "Decode Or Parsing Error"
	default:
		return fmt.Sprintf("Unknown MBIM Status Error: %d", e)
	}
}
