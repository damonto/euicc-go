package qmi

import "fmt"

// QMIProtocolError represents QMI protocol errors as defined in libqmi
// These correspond to the "Error" field in QMI Result TLVs
type QMIProtocolError uint16

const (
	QMIProtocolErrorNone                        QMIProtocolError = iota  /*< nick=None >*/
	QMIProtocolErrorMalformedMessage                                     /*< nick=MalformedMessage >*/
	QMIProtocolErrorNoMemory                                             /*< nick=NoMemory >*/
	QMIProtocolErrorInternal                                             /*< nick=Internal >*/
	QMIProtocolErrorAborted                                              /*< nick=Aborted >*/
	QMIProtocolErrorClientIdsExhausted                                   /*< nick=ClientIdsExhausted >*/
	QMIProtocolErrorUnabortableTransaction                               /*< nick=UnabortableTransaction >*/
	QMIProtocolErrorInvalidClientId                                      /*< nick=InvalidClientId >*/
	QMIProtocolErrorNoThresholdsProvided                                 /*< nick=NoThresholdsProvided >*/
	QMIProtocolErrorInvalidHandle                                        /*< nick=InvalidHandle >*/
	QMIProtocolErrorInvalidProfile                                       /*< nick=InvalidProfile >*/
	QMIProtocolErrorInvalidPinId                                         /*< nick=InvalidPinId >*/
	QMIProtocolErrorIncorrectPin                                         /*< nick=IncorrectPin >*/
	QMIProtocolErrorNoNetworkFound                                       /*< nick=NoNetworkFound >*/
	QMIProtocolErrorCallFailed                                           /*< nick=CallFailed >*/
	QMIProtocolErrorOutOfCall                                            /*< nick=OutOfCall >*/
	QMIProtocolErrorNotProvisioned                                       /*< nick=NotProvisioned >*/
	QMIProtocolErrorMissingArgument                                      /*< nick=MissingArgument >*/
	QMIProtocolErrorArgumentTooLong                                      /*< nick=ArgumentTooLong >*/
	QMIProtocolErrorInvalidTransactionId                                 /*< nick=InvalidTransactionId >*/
	QMIProtocolErrorDeviceInUse                                          /*< nick=DeviceInUse >*/
	QMIProtocolErrorNetworkUnsupported                                   /*< nick=NetworkUnsupported >*/
	QMIProtocolErrorDeviceUnsupported                                    /*< nick=DeviceUnsupported >*/
	QMIProtocolErrorNoEffect                                             /*< nick=NoEffect >*/
	QMIProtocolErrorNoFreeProfile                                        /*< nick=NoFreeProfile >*/
	QMIProtocolErrorInvalidPdpType                                       /*< nick=InvalidPdpType >*/
	QMIProtocolErrorInvalidTechnologyPreference                          /*< nick=InvalidTechnologyPreference >*/
	QMIProtocolErrorInvalidProfileType                                   /*< nick=InvalidProfileType >*/
	QMIProtocolErrorInvalidServiceType                                   /*< nick=InvalidServiceType >*/
	QMIProtocolErrorInvalidRegisterAction                                /*< nick=InvalidRegisterAction >*/
	QMIProtocolErrorInvalidPsAttachAction                                /*< nick=InvalidPsAttachAction >*/
	QMIProtocolErrorAuthenticationFailed                                 /*< nick=AuthenticationFailed >*/
	QMIProtocolErrorPinBlocked                                           /*< nick=PinBlocked >*/
	QMIProtocolErrorPinAlwaysBlocked                                     /*< nick=PinAlwaysBlocked >*/
	QMIProtocolErrorUimUninitialized                                     /*< nick=UimUninitialized >*/
	QMIProtocolErrorMaximumQosRequestsInUse                              /*< nick=MaximumQosRequestsInUse >*/
	QMIProtocolErrorIncorrectFlowFilter                                  /*< nick=IncorrectFlowFilter >*/
	QMIProtocolErrorNetworkQosUnaware                                    /*< nick=NetworkQosUnaware >*/
	QMIProtocolErrorInvalidQosId                                         /*< nick=InvalidQosId >*/
	QMIProtocolErrorRequestedNumberUnsupported                           /*< nick=RequestedNumberUnsupported >*/
	QMIProtocolErrorInterfaceNotFound                                    /*< nick=InterfaceNotFound >*/
	QMIProtocolErrorFlowSuspended                                        /*< nick=FlowSuspended >*/
	QMIProtocolErrorInvalidDataFormat                                    /*< nick=InvalidDataFormat >*/
	QMIProtocolErrorGeneralError                                         /*< nick=GeneralError >*/
	QMIProtocolErrorUnknownError                                         /*< nick=UnknownError >*/
	QMIProtocolErrorInvalidArgument                                      /*< nick=InvalidArgument >*/
	QMIProtocolErrorInvalidIndex                                         /*< nick=InvalidIndex >*/
	QMIProtocolErrorNoEntry                                              /*< nick=NoEntry >*/
	QMIProtocolErrorDeviceStorageFull                                    /*< nick=DeviceStorageFull >*/
	QMIProtocolErrorDeviceNotReady                                       /*< nick=DeviceNotReady >*/
	QMIProtocolErrorNetworkNotReady                                      /*< nick=NetworkNotReady >*/
	QMIProtocolErrorWmsCauseCode                                         /*< nick=WmsCauseCode >*/
	QMIProtocolErrorWmsMessageNotSent                                    /*< nick=WmsMessageNotSent >*/
	QMIProtocolErrorWmsMessageDeliveryFailure                            /*< nick=WmsMessageDeliveryFailure >*/
	QMIProtocolErrorWmsInvalidMessageId                                  /*< nick=WmsInvalidMessageId >*/
	QMIProtocolErrorWmsEncoding                                          /*< nick=WmsEncoding >*/
	QMIProtocolErrorAuthenticationLock                                   /*< nick=AuthenticationLock >*/
	QMIProtocolErrorInvalidTransition                                    /*< nick=InvalidTransition >*/
	QMIProtocolErrorNotMcastInterface                                    /*< nick=NotMcastInterface >*/
	QMIProtocolErrorMaximumMcastRequestsInUse                            /*< nick=MaximumMcastRequestsInUse >*/
	QMIProtocolErrorInvalidMcastHandle                                   /*< nick=InvalidMcastHandle >*/
	QMIProtocolErrorInvalidIpFamilyPreference                            /*< nick=InvalidIpFamilyPreference >*/
	QMIProtocolErrorSessionInactive                                      /*< nick=SessionInactive >*/
	QMIProtocolErrorSessionInvalid                                       /*< nick=SessionInvalid >*/
	QMIProtocolErrorSessionOwnership                                     /*< nick=SessionOwnership >*/
	QMIProtocolErrorInsufficientResources                                /*< nick=InsufficientResources >*/
	QMIProtocolErrorDisabled                                             /*< nick=Disabled >*/
	QMIProtocolErrorInvalidOperation                                     /*< nick=InvalidOperation >*/
	QMIProtocolErrorInvalidQmiCommand                                    /*< nick=InvalidQmiCommand >*/
	QMIProtocolErrorWmsTPduType                                          /*< nick=WmsTPduType >*/
	QMIProtocolErrorWmsSmscAddress                                       /*< nick=WmsSmscAddress >*/
	QMIProtocolErrorInformationUnavailable                               /*< nick=InformationUnavailable >*/
	QMIProtocolErrorSegmentTooLong                                       /*< nick=SegmentTooLong >*/
	QMIProtocolErrorSegmentOrder                                         /*< nick=SegmentOrder >*/
	QMIProtocolErrorBundlingNotSupported                                 /*< nick=BundlingNotSupported >*/
	QMIProtocolErrorOperationPartialFailure                              /*< nick=OperationPartialFailure >*/
	QMIProtocolErrorPolicyMismatch                                       /*< nick=PolicyMismatch >*/
	QMIProtocolErrorSimFileNotFound                                      /*< nick=SimFileNotFound >*/
	QMIProtocolErrorExtendedInternal                                     /*< nick=ExtendedInternal >*/
	QMIProtocolErrorAccessDenied                                         /*< nick=AccessDenied >*/
	QMIProtocolErrorHardwareRestricted                                   /*< nick=HardwareRestricted >*/
	QMIProtocolErrorAckNotSent                                           /*< nick=AckNotSent >*/
	QMIProtocolErrorInjectTimeout                                        /*< nick=InjectTimeout >*/
	QMIProtocolErrorIncompatibleState                                    /*< nick=IncompatibleState >*/
	QMIProtocolErrorFdnRestrict                                          /*< nick=FdnRestrict >*/
	QMIProtocolErrorSupsFailureCase                                      /*< nick=SupsFailureCase >*/
	QMIProtocolErrorNoRadio                                              /*< nick=NoRadio >*/
	QMIProtocolErrorNotSupported                                         /*< nick=NotSupported >*/
	QMIProtocolErrorNoSubscription                                       /*< nick=NoSubscription >*/
	QMIProtocolErrorCardCallControlFailed                                /*< nick=CardCallControlFailed >*/
	QMIProtocolErrorNetworkAborted                                       /*< nick=NetworkAborted >*/
	QMIProtocolErrorMsgBlocked                                           /*< nick=MsgBlocked >*/
	QMIProtocolErrorInvalidSessionType                                   /*< nick=InvalidSessionType >*/
	QMIProtocolErrorInvalidPbType                                        /*< nick=InvalidPbType >*/
	QMIProtocolErrorNoSim                                                /*< nick=NoSim >*/
	QMIProtocolErrorPbNotReady                                           /*< nick=PbNotReady >*/
	QMIProtocolErrorPinRestriction                                       /*< nick=PinRestriction >*/
	QMIProtocolErrorPin2Restriction                                      /*< nick=Pin1Restriction >*/
	QMIProtocolErrorPukRestriction                                       /*< nick=PukRestriction >*/
	QMIProtocolErrorPuk2Restriction                                      /*< nick=Puk2Restriction >*/
	QMIProtocolErrorPbAccessRestricted                                   /*< nick=PbAccessRestricted >*/
	QMIProtocolErrorPbDeleteInProgress                                   /*< nick=PbDeleteInProgress >*/
	QMIProtocolErrorPbTextTooLong                                        /*< nick=PbTextTooLong >*/
	QMIProtocolErrorPbNumberTooLong                                      /*< nick=PbNumberTooLong >*/
	QMIProtocolErrorPbHiddenKeyRestriction                               /*< nick=PbHiddenKeyRestriction >*/
	QMIProtocolErrorPbNotAvailable                                       /*< nick=PbNotAvailable >*/
	QMIProtocolErrorDeviceMemoryError                                    /*< nick=DeviceMemoryError >*/
	QMIProtocolErrorNoPermission                                         /*< nick=NoPermission >*/
	QMIProtocolErrorTooSoon                                              /*< nick=TooSoon >*/
	QMIProtocolErrorTimeNotAcquired                                      /*< nick=TimeNotAcquired >*/
	QMIProtocolErrorOperationInProgress                                  /*< nick=OperationInProgress >*/
	QMIProtocolErrorFwWriteFailed               QMIProtocolError = 388   /*< nick=FwWriteFailed >*/
	QMIProtocolErrorFwInfoReadFailed            QMIProtocolError = 389   /*< nick=FwInfoReadFailed >*/
	QMIProtocolErrorFwFileNotFound              QMIProtocolError = 390   /*< nick=FwFileNotFound >*/
	QMIProtocolErrorFwDirNotFound               QMIProtocolError = 391   /*< nick=FwDirNotFound >*/
	QMIProtocolErrorFwAlreadyActivated          QMIProtocolError = 392   /*< nick=FwAlreadyActivated >*/
	QMIProtocolErrorFwCannotGenericImage        QMIProtocolError = 393   /*< nick=FwCannotGenericImage >*/
	QMIProtocolErrorFwFileOpenFailed            QMIProtocolError = 400   /*< nick=FwFileOpenFailed >*/
	QMIProtocolErrorFwUpdateDiscontinuousFrame  QMIProtocolError = 401   /*< nick=FwUpdateDiscontinuousFrame >*/
	QMIProtocolErrorFwUpdateFailed              QMIProtocolError = 402   /*< nick=FwUpdateFailed >*/
	QMIProtocolErrorCatEventRegistrationFailed  QMIProtocolError = 61441 /*< nick=CatEventRegistrationFailed >*/
	QMIProtocolErrorCatInvalidTerminalResponse  QMIProtocolError = 61442 /*< nick=CatInvalidTerminalResponse >*/
	QMIProtocolErrorCatInvalidEnvelopeCommand   QMIProtocolError = 61443 /*< nick=CatInvalidEnvelopeCommand >*/
	QMIProtocolErrorCatEnvelopeCommandBusy      QMIProtocolError = 61444 /*< nick=CatEnvelopeCommandBusy >*/
	QMIProtocolErrorCatEnvelopeCommandFailed    QMIProtocolError = 61445 /*< nick=CatEnvelopeCommandFailed >*/
)

func (e QMIProtocolError) Error() string {
	switch e {
	case QMIProtocolErrorNone:
		return "No error"
	case QMIProtocolErrorMalformedMessage:
		return "Malformed message"
	case QMIProtocolErrorNoMemory:
		return "No memory"
	case QMIProtocolErrorInternal:
		return "Internal error"
	case QMIProtocolErrorAborted:
		return "Operation aborted"
	case QMIProtocolErrorClientIdsExhausted:
		return "Client IDs exhausted"
	case QMIProtocolErrorUnabortableTransaction:
		return "Unabortable transaction"
	case QMIProtocolErrorInvalidClientId:
		return "Invalid client ID"
	case QMIProtocolErrorNoThresholdsProvided:
		return "No thresholds provided"
	case QMIProtocolErrorInvalidHandle:
		return "Invalid handle"
	case QMIProtocolErrorInvalidProfile:
		return "Invalid profile"
	case QMIProtocolErrorInvalidPinId:
		return "Invalid PIN ID"
	case QMIProtocolErrorIncorrectPin:
		return "Incorrect PIN"
	case QMIProtocolErrorNoNetworkFound:
		return "No network found"
	case QMIProtocolErrorCallFailed:
		return "Call failed"
	case QMIProtocolErrorOutOfCall:
		return "Out of call"
	case QMIProtocolErrorNotProvisioned:
		return "Not provisioned"
	case QMIProtocolErrorMissingArgument:
		return "Missing argument"
	case QMIProtocolErrorArgumentTooLong:
		return "Argument too long"
	case QMIProtocolErrorInvalidTransactionId:
		return "Invalid transaction ID"
	case QMIProtocolErrorDeviceInUse:
		return "Device in use"
	case QMIProtocolErrorNetworkUnsupported:
		return "Network unsupported"
	case QMIProtocolErrorDeviceUnsupported:
		return "Device unsupported"
	case QMIProtocolErrorNoEffect:
		return "No effect"
	case QMIProtocolErrorNoFreeProfile:
		return "No free profile available"
	case QMIProtocolErrorInvalidPdpType:
		return "Invalid PDP type"
	case QMIProtocolErrorInvalidTechnologyPreference:
		return "Invalid technology preference"
	case QMIProtocolErrorInvalidProfileType:
		return "Invalid profile type"
	case QMIProtocolErrorInvalidServiceType:
		return "Invalid service type"
	case QMIProtocolErrorInvalidRegisterAction:
		return "Invalid register action"
	case QMIProtocolErrorInvalidPsAttachAction:
		return "Invalid PS attach action"
	case QMIProtocolErrorAuthenticationFailed:
		return "Authentication failed"
	case QMIProtocolErrorPinBlocked:
		return "PIN blocked"
	case QMIProtocolErrorPinAlwaysBlocked:
		return "PIN always blocked"
	case QMIProtocolErrorUimUninitialized:
		return "UIM uninitialized"
	case QMIProtocolErrorMaximumQosRequestsInUse:
		return "Maximum QoS requests in use"
	case QMIProtocolErrorIncorrectFlowFilter:
		return "Incorrect flow filter"
	case QMIProtocolErrorNetworkQosUnaware:
		return "Network QoS unaware"
	case QMIProtocolErrorInvalidQosId:
		return "Invalid QoS ID"
	case QMIProtocolErrorRequestedNumberUnsupported:
		return "Requested number unsupported"
	case QMIProtocolErrorInterfaceNotFound:
		return "Interface not found"
	case QMIProtocolErrorFlowSuspended:
		return "Flow suspended"
	case QMIProtocolErrorInvalidDataFormat:
		return "Invalid data format"
	case QMIProtocolErrorGeneralError:
		return "General error"
	case QMIProtocolErrorUnknownError:
		return "Unknown error"
	case QMIProtocolErrorInvalidArgument:
		return "Invalid argument"
	case QMIProtocolErrorInvalidIndex:
		return "Invalid index"
	case QMIProtocolErrorNoEntry:
		return "No entry"
	case QMIProtocolErrorDeviceStorageFull:
		return "Device storage full"
	case QMIProtocolErrorDeviceNotReady:
		return "Device not ready"
	case QMIProtocolErrorNetworkNotReady:
		return "Network not ready"
	case QMIProtocolErrorWmsCauseCode:
		return "WMS cause code"
	case QMIProtocolErrorWmsMessageNotSent:
		return "WMS message not sent"
	case QMIProtocolErrorWmsMessageDeliveryFailure:
		return "WMS message delivery failure"
	case QMIProtocolErrorWmsInvalidMessageId:
		return "WMS invalid message ID"
	case QMIProtocolErrorWmsEncoding:
		return "WMS encoding error"
	case QMIProtocolErrorAuthenticationLock:
		return "Authentication lock"
	case QMIProtocolErrorInvalidTransition:
		return "Invalid transition"
	case QMIProtocolErrorNotMcastInterface:
		return "Not a multicast interface"
	case QMIProtocolErrorMaximumMcastRequestsInUse:
		return "Maximum multicast requests in use"
	case QMIProtocolErrorInvalidMcastHandle:
		return "Invalid multicast handle"
	case QMIProtocolErrorInvalidIpFamilyPreference:
		return "Invalid IP family preference"
	case QMIProtocolErrorSessionInactive:
		return "Session inactive"
	case QMIProtocolErrorSessionInvalid:
		return "Session invalid"
	case QMIProtocolErrorSessionOwnership:
		return "Session ownership error"
	case QMIProtocolErrorInsufficientResources:
		return "Insufficient resources"
	case QMIProtocolErrorDisabled:
		return "Disabled"
	case QMIProtocolErrorInvalidOperation:
		return "Invalid operation"
	case QMIProtocolErrorInvalidQmiCommand:
		return "Invalid QMI command"
	case QMIProtocolErrorWmsTPduType:
		return "WMS TPDU type error"
	case QMIProtocolErrorWmsSmscAddress:
		return "WMS SMSC address error"
	case QMIProtocolErrorInformationUnavailable:
		return "Information unavailable"
	case QMIProtocolErrorSegmentTooLong:
		return "Segment too long"
	case QMIProtocolErrorSegmentOrder:
		return "Segment order error"
	case QMIProtocolErrorBundlingNotSupported:
		return "Bundling not supported"
	case QMIProtocolErrorOperationPartialFailure:
		return "Operation partial failure"
	case QMIProtocolErrorPolicyMismatch:
		return "Policy mismatch"
	case QMIProtocolErrorSimFileNotFound:
		return "SIM file not found"
	case QMIProtocolErrorExtendedInternal:
		return "Extended internal error"
	case QMIProtocolErrorAccessDenied:
		return "Access denied"
	case QMIProtocolErrorHardwareRestricted:
		return "Hardware restricted"
	case QMIProtocolErrorAckNotSent:
		return "Acknowledgment not sent"
	case QMIProtocolErrorInjectTimeout:
		return "Inject timeout"
	case QMIProtocolErrorIncompatibleState:
		return "Incompatible state"
	case QMIProtocolErrorFdnRestrict:
		return "Fdn restrict"
	case QMIProtocolErrorSupsFailureCase:
		return "Sups failure case"
	case QMIProtocolErrorNoRadio:
		return "No radio"
	case QMIProtocolErrorNotSupported:
		return "Not supported"
	case QMIProtocolErrorNoSubscription:
		return "No subscription"
	case QMIProtocolErrorCardCallControlFailed:
		return "Card call control failed"
	case QMIProtocolErrorNetworkAborted:
		return "Network aborted"
	case QMIProtocolErrorMsgBlocked:
		return "Message blocked"
	case QMIProtocolErrorInvalidSessionType:
		return "Invalid session type"
	case QMIProtocolErrorInvalidPbType:
		return "Invalid phonebook type"
	case QMIProtocolErrorNoSim:
		return "No SIM card"
	case QMIProtocolErrorPbNotReady:
		return "Phonebook not ready"
	case QMIProtocolErrorPinRestriction:
		return "PIN restriction"
	case QMIProtocolErrorPin2Restriction:
		return "PIN2 restriction"
	case QMIProtocolErrorPukRestriction:
		return "PUK restriction"
	case QMIProtocolErrorPuk2Restriction:
		return "PUK2 restriction"
	case QMIProtocolErrorPbAccessRestricted:
		return "Phonebook access restricted"
	case QMIProtocolErrorPbDeleteInProgress:
		return "Phonebook delete in progress"
	case QMIProtocolErrorPbTextTooLong:
		return "Phonebook text too long"
	case QMIProtocolErrorPbNumberTooLong:
		return "Phonebook number too long"
	case QMIProtocolErrorPbHiddenKeyRestriction:
		return "Phonebook hidden key restriction"
	case QMIProtocolErrorPbNotAvailable:
		return "Phonebook not available"
	case QMIProtocolErrorDeviceMemoryError:
		return "Device memory error"
	case QMIProtocolErrorNoPermission:
		return "No permission"
	case QMIProtocolErrorTooSoon:
		return "Too soon"
	case QMIProtocolErrorTimeNotAcquired:
		return "Time not acquired"
	case QMIProtocolErrorOperationInProgress:
		return "Operation in progress"
	case QMIProtocolErrorFwWriteFailed:
		return "Firmware write failed"
	case QMIProtocolErrorFwInfoReadFailed:
		return "Firmware info read failed"
	case QMIProtocolErrorFwFileNotFound:
		return "Firmware file not found"
	case QMIProtocolErrorFwDirNotFound:
		return "Firmware directory not found"
	case QMIProtocolErrorFwAlreadyActivated:
		return "Firmware already activated"
	case QMIProtocolErrorFwCannotGenericImage:
		return "Firmware cannot generic image"
	case QMIProtocolErrorFwFileOpenFailed:
		return "Firmware file open failed"
	case QMIProtocolErrorFwUpdateDiscontinuousFrame:
		return "Firmware update discontinuous frame"
	case QMIProtocolErrorFwUpdateFailed:
		return "Firmware update failed"
	case QMIProtocolErrorCatEventRegistrationFailed:
		return "CAT event registration failed"
	case QMIProtocolErrorCatInvalidTerminalResponse:
		return "CAT invalid terminal response"
	case QMIProtocolErrorCatInvalidEnvelopeCommand:
		return "CAT invalid envelope command"
	case QMIProtocolErrorCatEnvelopeCommandBusy:
		return "CAT envelope command busy"
	case QMIProtocolErrorCatEnvelopeCommandFailed:
		return "CAT envelope command failed"
	default:
		return "Unknown QMI protocol error"
	}
}

// QMIError represents a QMI error with result and error codes
type QMIError struct {
	Result    QMIResult
	ErrorCode QMIProtocolError
}

// Error implements the error interface
func (e QMIError) Error() string {
	return fmt.Sprintf("QMI Error: Result=%s (%d), Error=%s (%d)",
		e.Result.String(), uint16(e.Result),
		e.ErrorCode.Error(), uint16(e.ErrorCode))
}
