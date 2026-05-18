package protocol

import "fmt"

// QMIError represents QMI protocol errors as defined in libqmi
// These correspond to the "Error" field in QMI Result TLVs
type QMIError uint16

const (
	QMIErrorNone                        QMIError = 0     /*< nick=None >*/
	QMIErrorMalformedMessage            QMIError = 1     /*< nick=MalformedMessage >*/
	QMIErrorNoMemory                    QMIError = 2     /*< nick=NoMemory >*/
	QMIErrorInternal                    QMIError = 3     /*< nick=Internal >*/
	QMIErrorAborted                     QMIError = 4     /*< nick=Aborted >*/
	QMIErrorClientIdsExhausted          QMIError = 5     /*< nick=ClientIdsExhausted >*/
	QMIErrorUnabortableTransaction      QMIError = 6     /*< nick=UnabortableTransaction >*/
	QMIErrorInvalidClientId             QMIError = 7     /*< nick=InvalidClientId >*/
	QMIErrorNoThresholdsProvided        QMIError = 8     /*< nick=NoThresholdsProvided >*/
	QMIErrorInvalidHandle               QMIError = 9     /*< nick=InvalidHandle >*/
	QMIErrorInvalidProfile              QMIError = 10    /*< nick=InvalidProfile >*/
	QMIErrorInvalidPinId                QMIError = 11    /*< nick=InvalidPinId >*/
	QMIErrorIncorrectPin                QMIError = 12    /*< nick=IncorrectPin >*/
	QMIErrorNoNetworkFound              QMIError = 13    /*< nick=NoNetworkFound >*/
	QMIErrorCallFailed                  QMIError = 14    /*< nick=CallFailed >*/
	QMIErrorOutOfCall                   QMIError = 15    /*< nick=OutOfCall >*/
	QMIErrorNotProvisioned              QMIError = 16    /*< nick=NotProvisioned >*/
	QMIErrorMissingArgument             QMIError = 17    /*< nick=MissingArgument >*/
	QMIErrorArgumentTooLong             QMIError = 19    /*< nick=ArgumentTooLong >*/
	QMIErrorInvalidTransactionId        QMIError = 22    /*< nick=InvalidTransactionId >*/
	QMIErrorDeviceInUse                 QMIError = 23    /*< nick=DeviceInUse >*/
	QMIErrorNetworkUnsupported          QMIError = 24    /*< nick=NetworkUnsupported >*/
	QMIErrorDeviceUnsupported           QMIError = 25    /*< nick=DeviceUnsupported >*/
	QMIErrorNoEffect                    QMIError = 26    /*< nick=NoEffect >*/
	QMIErrorNoFreeProfile               QMIError = 27    /*< nick=NoFreeProfile >*/
	QMIErrorInvalidPdpType              QMIError = 28    /*< nick=InvalidPdpType >*/
	QMIErrorInvalidTechnologyPreference QMIError = 29    /*< nick=InvalidTechnologyPreference >*/
	QMIErrorInvalidProfileType          QMIError = 30    /*< nick=InvalidProfileType >*/
	QMIErrorInvalidServiceType          QMIError = 31    /*< nick=InvalidServiceType >*/
	QMIErrorInvalidRegisterAction       QMIError = 32    /*< nick=InvalidRegisterAction >*/
	QMIErrorInvalidPsAttachAction       QMIError = 33    /*< nick=InvalidPsAttachAction >*/
	QMIErrorAuthenticationFailed        QMIError = 34    /*< nick=AuthenticationFailed >*/
	QMIErrorPinBlocked                  QMIError = 35    /*< nick=PinBlocked >*/
	QMIErrorPinAlwaysBlocked            QMIError = 36    /*< nick=PinAlwaysBlocked >*/
	QMIErrorUimUninitialized            QMIError = 37    /*< nick=UimUninitialized >*/
	QMIErrorMaximumQosRequestsInUse     QMIError = 38    /*< nick=MaximumQosRequestsInUse >*/
	QMIErrorIncorrectFlowFilter         QMIError = 39    /*< nick=IncorrectFlowFilter >*/
	QMIErrorNetworkQosUnaware           QMIError = 40    /*< nick=NetworkQosUnaware >*/
	QMIErrorInvalidQosId                QMIError = 41    /*< nick=InvalidQosId >*/
	QMIErrorRequestedNumberUnsupported  QMIError = 42    /*< nick=RequestedNumberUnsupported >*/
	QMIErrorInterfaceNotFound           QMIError = 43    /*< nick=InterfaceNotFound >*/
	QMIErrorFlowSuspended               QMIError = 44    /*< nick=FlowSuspended >*/
	QMIErrorInvalidDataFormat           QMIError = 45    /*< nick=InvalidDataFormat >*/
	QMIErrorGeneralError                QMIError = 46    /*< nick=GeneralError >*/
	QMIErrorUnknownError                QMIError = 47    /*< nick=UnknownError >*/
	QMIErrorInvalidArgument             QMIError = 48    /*< nick=InvalidArgument >*/
	QMIErrorInvalidIndex                QMIError = 49    /*< nick=InvalidIndex >*/
	QMIErrorNoEntry                     QMIError = 50    /*< nick=NoEntry >*/
	QMIErrorDeviceStorageFull           QMIError = 51    /*< nick=DeviceStorageFull >*/
	QMIErrorDeviceNotReady              QMIError = 52    /*< nick=DeviceNotReady >*/
	QMIErrorNetworkNotReady             QMIError = 53    /*< nick=NetworkNotReady >*/
	QMIErrorWmsCauseCode                QMIError = 54    /*< nick=WmsCauseCode >*/
	QMIErrorWmsMessageNotSent           QMIError = 55    /*< nick=WmsMessageNotSent >*/
	QMIErrorWmsMessageDeliveryFailure   QMIError = 56    /*< nick=WmsMessageDeliveryFailure >*/
	QMIErrorWmsInvalidMessageId         QMIError = 57    /*< nick=WmsInvalidMessageId >*/
	QMIErrorWmsEncoding                 QMIError = 58    /*< nick=WmsEncoding >*/
	QMIErrorAuthenticationLock          QMIError = 59    /*< nick=AuthenticationLock >*/
	QMIErrorInvalidTransition           QMIError = 60    /*< nick=InvalidTransition >*/
	QMIErrorNotMcastInterface           QMIError = 61    /*< nick=NotMcastInterface >*/
	QMIErrorMaximumMcastRequestsInUse   QMIError = 62    /*< nick=MaximumMcastRequestsInUse >*/
	QMIErrorInvalidMcastHandle          QMIError = 63    /*< nick=InvalidMcastHandle >*/
	QMIErrorInvalidIpFamilyPreference   QMIError = 64    /*< nick=InvalidIpFamilyPreference >*/
	QMIErrorSessionInactive             QMIError = 65    /*< nick=SessionInactive >*/
	QMIErrorSessionInvalid              QMIError = 66    /*< nick=SessionInvalid >*/
	QMIErrorSessionOwnership            QMIError = 67    /*< nick=SessionOwnership >*/
	QMIErrorInsufficientResources       QMIError = 68    /*< nick=InsufficientResources >*/
	QMIErrorDisabled                    QMIError = 69    /*< nick=Disabled >*/
	QMIErrorInvalidOperation            QMIError = 70    /*< nick=InvalidOperation >*/
	QMIErrorInvalidQmiCommand           QMIError = 71    /*< nick=InvalidQmiCommand >*/
	QMIErrorWmsTPduType                 QMIError = 72    /*< nick=WmsTPduType >*/
	QMIErrorWmsSmscAddress              QMIError = 73    /*< nick=WmsSmscAddress >*/
	QMIErrorInformationUnavailable      QMIError = 74    /*< nick=InformationUnavailable >*/
	QMIErrorSegmentTooLong              QMIError = 75    /*< nick=SegmentTooLong >*/
	QMIErrorSegmentOrder                QMIError = 76    /*< nick=SegmentOrder >*/
	QMIErrorBundlingNotSupported        QMIError = 77    /*< nick=BundlingNotSupported >*/
	QMIErrorOperationPartialFailure     QMIError = 78    /*< nick=OperationPartialFailure >*/
	QMIErrorPolicyMismatch              QMIError = 79    /*< nick=PolicyMismatch >*/
	QMIErrorSimFileNotFound             QMIError = 80    /*< nick=SimFileNotFound >*/
	QMIErrorExtendedInternal            QMIError = 81    /*< nick=ExtendedInternal >*/
	QMIErrorAccessDenied                QMIError = 82    /*< nick=AccessDenied >*/
	QMIErrorHardwareRestricted          QMIError = 83    /*< nick=HardwareRestricted >*/
	QMIErrorAckNotSent                  QMIError = 84    /*< nick=AckNotSent >*/
	QMIErrorInjectTimeout               QMIError = 85    /*< nick=InjectTimeout >*/
	QMIErrorIncompatibleState           QMIError = 90    /*< nick=IncompatibleState >*/
	QMIErrorFdnRestrict                 QMIError = 91    /*< nick=FdnRestrict >*/
	QMIErrorSupsFailureCase             QMIError = 92    /*< nick=SupsFailureCase >*/
	QMIErrorNoRadio                     QMIError = 93    /*< nick=NoRadio >*/
	QMIErrorNotSupported                QMIError = 94    /*< nick=NotSupported >*/
	QMIErrorNoSubscription              QMIError = 95    /*< nick=NoSubscription >*/
	QMIErrorCardCallControlFailed       QMIError = 96    /*< nick=CardCallControlFailed >*/
	QMIErrorNetworkAborted              QMIError = 97    /*< nick=NetworkAborted >*/
	QMIErrorMsgBlocked                  QMIError = 98    /*< nick=MsgBlocked >*/
	QMIErrorInvalidSessionType          QMIError = 100   /*< nick=InvalidSessionType >*/
	QMIErrorInvalidPbType               QMIError = 101   /*< nick=InvalidPbType >*/
	QMIErrorNoSim                       QMIError = 102   /*< nick=NoSim >*/
	QMIErrorPbNotReady                  QMIError = 103   /*< nick=PbNotReady >*/
	QMIErrorPinRestriction              QMIError = 104   /*< nick=PinRestriction >*/
	QMIErrorPin2Restriction             QMIError = 105   /*< nick=Pin1Restriction >*/
	QMIErrorPukRestriction              QMIError = 106   /*< nick=PukRestriction >*/
	QMIErrorPuk2Restriction             QMIError = 107   /*< nick=Puk2Restriction >*/
	QMIErrorPbAccessRestricted          QMIError = 108   /*< nick=PbAccessRestricted >*/
	QMIErrorPbDeleteInProgress          QMIError = 109   /*< nick=PbDeleteInProgress >*/
	QMIErrorPbTextTooLong               QMIError = 110   /*< nick=PbTextTooLong >*/
	QMIErrorPbNumberTooLong             QMIError = 111   /*< nick=PbNumberTooLong >*/
	QMIErrorPbHiddenKeyRestriction      QMIError = 112   /*< nick=PbHiddenKeyRestriction >*/
	QMIErrorPbNotAvailable              QMIError = 113   /*< nick=PbNotAvailable >*/
	QMIErrorDeviceMemoryError           QMIError = 114   /*< nick=DeviceMemoryError >*/
	QMIErrorNoPermission                QMIError = 115   /*< nick=NoPermission >*/
	QMIErrorTooSoon                     QMIError = 116   /*< nick=TooSoon >*/
	QMIErrorTimeNotAcquired             QMIError = 117   /*< nick=TimeNotAcquired >*/
	QMIErrorOperationInProgress         QMIError = 118   /*< nick=OperationInProgress >*/
	QMIErrorFwWriteFailed               QMIError = 388   /*< nick=FwWriteFailed >*/
	QMIErrorFwInfoReadFailed            QMIError = 389   /*< nick=FwInfoReadFailed >*/
	QMIErrorFwFileNotFound              QMIError = 390   /*< nick=FwFileNotFound >*/
	QMIErrorFwDirNotFound               QMIError = 391   /*< nick=FwDirNotFound >*/
	QMIErrorFwAlreadyActivated          QMIError = 392   /*< nick=FwAlreadyActivated >*/
	QMIErrorFwCannotGenericImage        QMIError = 393   /*< nick=FwCannotGenericImage >*/
	QMIErrorFwFileOpenFailed            QMIError = 400   /*< nick=FwFileOpenFailed >*/
	QMIErrorFwUpdateDiscontinuousFrame  QMIError = 401   /*< nick=FwUpdateDiscontinuousFrame >*/
	QMIErrorFwUpdateFailed              QMIError = 402   /*< nick=FwUpdateFailed >*/
	QMIErrorCatEventRegistrationFailed  QMIError = 61441 /*< nick=CatEventRegistrationFailed >*/
	QMIErrorCatInvalidTerminalResponse  QMIError = 61442 /*< nick=CatInvalidTerminalResponse >*/
	QMIErrorCatInvalidEnvelopeCommand   QMIError = 61443 /*< nick=CatInvalidEnvelopeCommand >*/
	QMIErrorCatEnvelopeCommandBusy      QMIError = 61444 /*< nick=CatEnvelopeCommandBusy >*/
	QMIErrorCatEnvelopeCommandFailed    QMIError = 61445 /*< nick=CatEnvelopeCommandFailed >*/
)

var qmiErrorText = map[QMIError]string{
	QMIErrorNone:                        "No error",
	QMIErrorMalformedMessage:            "Malformed message",
	QMIErrorNoMemory:                    "No memory",
	QMIErrorInternal:                    "Internal error",
	QMIErrorAborted:                     "Aborted",
	QMIErrorClientIdsExhausted:          "Client IDs exhausted",
	QMIErrorUnabortableTransaction:      "Unabortable transaction",
	QMIErrorInvalidClientId:             "Invalid client ID",
	QMIErrorNoThresholdsProvided:        "No thresholds provided",
	QMIErrorInvalidHandle:               "Invalid handle",
	QMIErrorInvalidProfile:              "Invalid profile",
	QMIErrorInvalidPinId:                "Invalid PIN ID",
	QMIErrorIncorrectPin:                "Incorrect PIN",
	QMIErrorNoNetworkFound:              "No network found",
	QMIErrorCallFailed:                  "Call failed",
	QMIErrorOutOfCall:                   "Out of call",
	QMIErrorNotProvisioned:              "Not provisioned",
	QMIErrorMissingArgument:             "Missing argument",
	QMIErrorArgumentTooLong:             "Argument too long",
	QMIErrorInvalidTransactionId:        "Invalid transaction ID",
	QMIErrorDeviceInUse:                 "Device in use",
	QMIErrorNetworkUnsupported:          "Network unsupported",
	QMIErrorDeviceUnsupported:           "Device unsupported",
	QMIErrorNoEffect:                    "No effect",
	QMIErrorNoFreeProfile:               "No free profile",
	QMIErrorInvalidPdpType:              "Invalid PDP type",
	QMIErrorInvalidTechnologyPreference: "Invalid technology preference",
	QMIErrorInvalidProfileType:          "Invalid profile type",
	QMIErrorInvalidServiceType:          "Invalid service type",
	QMIErrorInvalidRegisterAction:       "Invalid register action",
	QMIErrorInvalidPsAttachAction:       "Invalid PS attach action",
	QMIErrorAuthenticationFailed:        "Authentication failed",
	QMIErrorPinBlocked:                  "PIN blocked",
	QMIErrorPinAlwaysBlocked:            "PIN always blocked",
	QMIErrorUimUninitialized:            "UIM uninitialized",
	QMIErrorMaximumQosRequestsInUse:     "Maximum QoS requests in use",
	QMIErrorIncorrectFlowFilter:         "Incorrect flow filter",
	QMIErrorNetworkQosUnaware:           "Network QoS unaware",
	QMIErrorInvalidQosId:                "Invalid QoS ID",
	QMIErrorRequestedNumberUnsupported:  "Requested number unsupported",
	QMIErrorInterfaceNotFound:           "Interface not found",
	QMIErrorFlowSuspended:               "Flow suspended",
	QMIErrorInvalidDataFormat:           "Invalid data format",
	QMIErrorGeneralError:                "General error",
	QMIErrorUnknownError:                "Unknown error",
	QMIErrorInvalidArgument:             "Invalid argument",
	QMIErrorInvalidIndex:                "Invalid index",
	QMIErrorNoEntry:                     "No entry",
	QMIErrorDeviceStorageFull:           "Device storage full",
	QMIErrorDeviceNotReady:              "Device not ready",
	QMIErrorNetworkNotReady:             "Network not ready",
	QMIErrorWmsCauseCode:                "WMS cause code",
	QMIErrorWmsMessageNotSent:           "WMS message not sent",
	QMIErrorWmsMessageDeliveryFailure:   "WMS message delivery failure",
	QMIErrorWmsInvalidMessageId:         "WMS invalid message ID",
	QMIErrorWmsEncoding:                 "WMS encoding",
	QMIErrorAuthenticationLock:          "Authentication lock",
	QMIErrorInvalidTransition:           "Invalid transition",
	QMIErrorNotMcastInterface:           "Not multicast interface",
	QMIErrorMaximumMcastRequestsInUse:   "Maximum multicast requests in use",
	QMIErrorInvalidMcastHandle:          "Invalid multicast handle",
	QMIErrorInvalidIpFamilyPreference:   "Invalid IP family preference",
	QMIErrorSessionInactive:             "Session inactive",
	QMIErrorSessionInvalid:              "Session invalid",
	QMIErrorSessionOwnership:            "Session ownership",
	QMIErrorInsufficientResources:       "Insufficient resources",
	QMIErrorDisabled:                    "Disabled",
	QMIErrorInvalidOperation:            "Invalid operation",
	QMIErrorInvalidQmiCommand:           "Invalid QMI command",
	QMIErrorWmsTPduType:                 "WMS TPDU type",
	QMIErrorWmsSmscAddress:              "WMS SMSC address",
	QMIErrorInformationUnavailable:      "Information unavailable",
	QMIErrorSegmentTooLong:              "Segment too long",
	QMIErrorSegmentOrder:                "Segment order",
	QMIErrorBundlingNotSupported:        "Bundling not supported",
	QMIErrorOperationPartialFailure:     "Operation partial failure",
	QMIErrorPolicyMismatch:              "Policy mismatch",
	QMIErrorSimFileNotFound:             "SIM file not found",
	QMIErrorExtendedInternal:            "Extended internal error",
	QMIErrorAccessDenied:                "Access denied",
	QMIErrorHardwareRestricted:          "Hardware restricted",
	QMIErrorAckNotSent:                  "ACK not sent",
	QMIErrorInjectTimeout:               "Inject timeout",
	QMIErrorIncompatibleState:           "Incompatible state",
	QMIErrorFdnRestrict:                 "FDN restrict",
	QMIErrorSupsFailureCase:             "SUPS failure case",
	QMIErrorNoRadio:                     "No radio",
	QMIErrorNotSupported:                "Not supported",
	QMIErrorNoSubscription:              "No subscription",
	QMIErrorCardCallControlFailed:       "Card call control failed",
	QMIErrorNetworkAborted:              "Network aborted",
	QMIErrorMsgBlocked:                  "Message blocked",
	QMIErrorInvalidSessionType:          "Invalid session type",
	QMIErrorInvalidPbType:               "Invalid phonebook type",
	QMIErrorNoSim:                       "No SIM",
	QMIErrorPbNotReady:                  "Phonebook not ready",
	QMIErrorPinRestriction:              "PIN restriction",
	QMIErrorPin2Restriction:             "PIN2 restriction",
	QMIErrorPukRestriction:              "PUK restriction",
	QMIErrorPuk2Restriction:             "PUK2 restriction",
	QMIErrorPbAccessRestricted:          "Phonebook access restricted",
	QMIErrorPbDeleteInProgress:          "Phonebook delete in progress",
	QMIErrorPbTextTooLong:               "Phonebook text too long",
	QMIErrorPbNumberTooLong:             "Phonebook number too long",
	QMIErrorPbHiddenKeyRestriction:      "Phonebook hidden key restriction",
	QMIErrorPbNotAvailable:              "Phonebook not available",
	QMIErrorDeviceMemoryError:           "Device memory error",
	QMIErrorNoPermission:                "No permission",
	QMIErrorTooSoon:                     "Too soon",
	QMIErrorTimeNotAcquired:             "Time not acquired",
	QMIErrorOperationInProgress:         "Operation in progress",
	QMIErrorFwWriteFailed:               "Firmware write failed",
	QMIErrorFwInfoReadFailed:            "Firmware info read failed",
	QMIErrorFwFileNotFound:              "Firmware file not found",
	QMIErrorFwDirNotFound:               "Firmware directory not found",
	QMIErrorFwAlreadyActivated:          "Firmware already activated",
	QMIErrorFwCannotGenericImage:        "Firmware cannot generic image",
	QMIErrorFwFileOpenFailed:            "Firmware file open failed",
	QMIErrorFwUpdateDiscontinuousFrame:  "Firmware update discontinuous frame",
	QMIErrorFwUpdateFailed:              "Firmware update failed",
	QMIErrorCatEventRegistrationFailed:  "CAT event registration failed",
	QMIErrorCatInvalidTerminalResponse:  "CAT invalid terminal response",
	QMIErrorCatInvalidEnvelopeCommand:   "CAT invalid envelope command",
	QMIErrorCatEnvelopeCommandBusy:      "CAT envelope command busy",
	QMIErrorCatEnvelopeCommandFailed:    "CAT envelope command failed",
}

func (q QMIError) Error() string {
	if text, ok := qmiErrorText[q]; ok {
		return text
	}
	return fmt.Sprintf("QMI error %d", q)
}
