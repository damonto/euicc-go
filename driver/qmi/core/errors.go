package core

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

func (q QMIError) Error() string {
	switch q {
	case QMIErrorNone:
		return "No error"
	case QMIErrorMalformedMessage:
		return "Malformed message"
	case QMIErrorNoMemory:
		return "No memory"
	case QMIErrorInternal:
		return "Internal error"
	case QMIErrorAborted:
		return "Aborted"
	case QMIErrorClientIdsExhausted:
		return "Client IDs exhausted"
	case QMIErrorUnabortableTransaction:
		return "Unabortable transaction"
	case QMIErrorInvalidClientId:
		return "Invalid client ID"
	case QMIErrorNoThresholdsProvided:
		return "No thresholds provided"
	case QMIErrorInvalidHandle:
		return "Invalid handle"
	case QMIErrorInvalidProfile:
		return "Invalid profile"
	case QMIErrorInvalidPinId:
		return "Invalid PIN ID"
	case QMIErrorIncorrectPin:
		return "Incorrect PIN"
	case QMIErrorNoNetworkFound:
		return "No network found"
	case QMIErrorCallFailed:
		return "Call failed"
	case QMIErrorOutOfCall:
		return "Out of call"
	case QMIErrorNotProvisioned:
		return "Not provisioned"
	case QMIErrorMissingArgument:
		return "Missing argument"
	case QMIErrorArgumentTooLong:
		return "Argument too long"
	case QMIErrorInvalidTransactionId:
		return "Invalid transaction ID"
	case QMIErrorDeviceInUse:
		return "Device in use"
	case QMIErrorNetworkUnsupported:
		return "Network unsupported"
	case QMIErrorDeviceUnsupported:
		return "Device unsupported"
	case QMIErrorNoEffect:
		return "No effect"
	case QMIErrorNoFreeProfile:
		return "No free profile"
	case QMIErrorInvalidPdpType:
		return "Invalid PDP type"
	case QMIErrorInvalidTechnologyPreference:
		return "Invalid technology preference"
	case QMIErrorInvalidProfileType:
		return "Invalid profile type"
	case QMIErrorInvalidServiceType:
		return "Invalid service type"
	case QMIErrorInvalidRegisterAction:
		return "Invalid register action"
	case QMIErrorInvalidPsAttachAction:
		return "Invalid PS attach action"
	case QMIErrorAuthenticationFailed:
		return "Authentication failed"
	case QMIErrorPinBlocked:
		return "PIN blocked"
	case QMIErrorPinAlwaysBlocked:
		return "PIN always blocked"
	case QMIErrorUimUninitialized:
		return "UIM uninitialized"
	case QMIErrorNotSupported:
		return "Not supported"
	case QMIErrorNoSim:
		return "No SIM"
	default:
		return "Unknown error"
	}
}
