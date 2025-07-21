package qmi

// QMIError represents QMI protocol errors as defined in libqmi
// These correspond to the "Error" field in QMI Result TLVs
type QMIError uint16

const (
	QMIErrorNone                        QMIError = iota  /*< nick=None >*/
	QMIErrorMalformedMessage                             /*< nick=MalformedMessage >*/
	QMIErrorNoMemory                                     /*< nick=NoMemory >*/
	QMIErrorInternal                                     /*< nick=Internal >*/
	QMIErrorAborted                                      /*< nick=Aborted >*/
	QMIErrorClientIdsExhausted                           /*< nick=ClientIdsExhausted >*/
	QMIErrorUnabortableTransaction                       /*< nick=UnabortableTransaction >*/
	QMIErrorInvalidClientId                              /*< nick=InvalidClientId >*/
	QMIErrorNoThresholdsProvided                         /*< nick=NoThresholdsProvided >*/
	QMIErrorInvalidHandle                                /*< nick=InvalidHandle >*/
	QMIErrorInvalidProfile                               /*< nick=InvalidProfile >*/
	QMIErrorInvalidPinId                                 /*< nick=InvalidPinId >*/
	QMIErrorIncorrectPin                                 /*< nick=IncorrectPin >*/
	QMIErrorNoNetworkFound                               /*< nick=NoNetworkFound >*/
	QMIErrorCallFailed                                   /*< nick=CallFailed >*/
	QMIErrorOutOfCall                                    /*< nick=OutOfCall >*/
	QMIErrorNotProvisioned                               /*< nick=NotProvisioned >*/
	QMIErrorMissingArgument                              /*< nick=MissingArgument >*/
	QMIErrorArgumentTooLong                              /*< nick=ArgumentTooLong >*/
	QMIErrorInvalidTransactionId                         /*< nick=InvalidTransactionId >*/
	QMIErrorDeviceInUse                                  /*< nick=DeviceInUse >*/
	QMIErrorNetworkUnsupported                           /*< nick=NetworkUnsupported >*/
	QMIErrorDeviceUnsupported                            /*< nick=DeviceUnsupported >*/
	QMIErrorNoEffect                                     /*< nick=NoEffect >*/
	QMIErrorNoFreeProfile                                /*< nick=NoFreeProfile >*/
	QMIErrorInvalidPdpType                               /*< nick=InvalidPdpType >*/
	QMIErrorInvalidTechnologyPreference                  /*< nick=InvalidTechnologyPreference >*/
	QMIErrorInvalidProfileType                           /*< nick=InvalidProfileType >*/
	QMIErrorInvalidServiceType                           /*< nick=InvalidServiceType >*/
	QMIErrorInvalidRegisterAction                        /*< nick=InvalidRegisterAction >*/
	QMIErrorInvalidPsAttachAction                        /*< nick=InvalidPsAttachAction >*/
	QMIErrorAuthenticationFailed                         /*< nick=AuthenticationFailed >*/
	QMIErrorPinBlocked                                   /*< nick=PinBlocked >*/
	QMIErrorPinAlwaysBlocked                             /*< nick=PinAlwaysBlocked >*/
	QMIErrorUimUninitialized                             /*< nick=UimUninitialized >*/
	QMIErrorMaximumQosRequestsInUse                      /*< nick=MaximumQosRequestsInUse >*/
	QMIErrorIncorrectFlowFilter                          /*< nick=IncorrectFlowFilter >*/
	QMIErrorNetworkQosUnaware                            /*< nick=NetworkQosUnaware >*/
	QMIErrorInvalidQosId                                 /*< nick=InvalidQosId >*/
	QMIErrorRequestedNumberUnsupported                   /*< nick=RequestedNumberUnsupported >*/
	QMIErrorInterfaceNotFound                            /*< nick=InterfaceNotFound >*/
	QMIErrorFlowSuspended                                /*< nick=FlowSuspended >*/
	QMIErrorInvalidDataFormat                            /*< nick=InvalidDataFormat >*/
	QMIErrorGeneralError                                 /*< nick=GeneralError >*/
	QMIErrorUnknownError                                 /*< nick=UnknownError >*/
	QMIErrorInvalidArgument                              /*< nick=InvalidArgument >*/
	QMIErrorInvalidIndex                                 /*< nick=InvalidIndex >*/
	QMIErrorNoEntry                                      /*< nick=NoEntry >*/
	QMIErrorDeviceStorageFull                            /*< nick=DeviceStorageFull >*/
	QMIErrorDeviceNotReady                               /*< nick=DeviceNotReady >*/
	QMIErrorNetworkNotReady                              /*< nick=NetworkNotReady >*/
	QMIErrorWmsCauseCode                                 /*< nick=WmsCauseCode >*/
	QMIErrorWmsMessageNotSent                            /*< nick=WmsMessageNotSent >*/
	QMIErrorWmsMessageDeliveryFailure                    /*< nick=WmsMessageDeliveryFailure >*/
	QMIErrorWmsInvalidMessageId                          /*< nick=WmsInvalidMessageId >*/
	QMIErrorWmsEncoding                                  /*< nick=WmsEncoding >*/
	QMIErrorAuthenticationLock                           /*< nick=AuthenticationLock >*/
	QMIErrorInvalidTransition                            /*< nick=InvalidTransition >*/
	QMIErrorNotMcastInterface                            /*< nick=NotMcastInterface >*/
	QMIErrorMaximumMcastRequestsInUse                    /*< nick=MaximumMcastRequestsInUse >*/
	QMIErrorInvalidMcastHandle                           /*< nick=InvalidMcastHandle >*/
	QMIErrorInvalidIpFamilyPreference                    /*< nick=InvalidIpFamilyPreference >*/
	QMIErrorSessionInactive                              /*< nick=SessionInactive >*/
	QMIErrorSessionInvalid                               /*< nick=SessionInvalid >*/
	QMIErrorSessionOwnership                             /*< nick=SessionOwnership >*/
	QMIErrorInsufficientResources                        /*< nick=InsufficientResources >*/
	QMIErrorDisabled                                     /*< nick=Disabled >*/
	QMIErrorInvalidOperation                             /*< nick=InvalidOperation >*/
	QMIErrorInvalidQmiCommand                            /*< nick=InvalidQmiCommand >*/
	QMIErrorWmsTPduType                                  /*< nick=WmsTPduType >*/
	QMIErrorWmsSmscAddress                               /*< nick=WmsSmscAddress >*/
	QMIErrorInformationUnavailable                       /*< nick=InformationUnavailable >*/
	QMIErrorSegmentTooLong                               /*< nick=SegmentTooLong >*/
	QMIErrorSegmentOrder                                 /*< nick=SegmentOrder >*/
	QMIErrorBundlingNotSupported                         /*< nick=BundlingNotSupported >*/
	QMIErrorOperationPartialFailure                      /*< nick=OperationPartialFailure >*/
	QMIErrorPolicyMismatch                               /*< nick=PolicyMismatch >*/
	QMIErrorSimFileNotFound                              /*< nick=SimFileNotFound >*/
	QMIErrorExtendedInternal                             /*< nick=ExtendedInternal >*/
	QMIErrorAccessDenied                                 /*< nick=AccessDenied >*/
	QMIErrorHardwareRestricted                           /*< nick=HardwareRestricted >*/
	QMIErrorAckNotSent                                   /*< nick=AckNotSent >*/
	QMIErrorInjectTimeout                                /*< nick=InjectTimeout >*/
	QMIErrorIncompatibleState                            /*< nick=IncompatibleState >*/
	QMIErrorFdnRestrict                                  /*< nick=FdnRestrict >*/
	QMIErrorSupsFailureCase                              /*< nick=SupsFailureCase >*/
	QMIErrorNoRadio                                      /*< nick=NoRadio >*/
	QMIErrorNotSupported                                 /*< nick=NotSupported >*/
	QMIErrorNoSubscription                               /*< nick=NoSubscription >*/
	QMIErrorCardCallControlFailed                        /*< nick=CardCallControlFailed >*/
	QMIErrorNetworkAborted                               /*< nick=NetworkAborted >*/
	QMIErrorMsgBlocked                                   /*< nick=MsgBlocked >*/
	QMIErrorInvalidSessionType                           /*< nick=InvalidSessionType >*/
	QMIErrorInvalidPbType                                /*< nick=InvalidPbType >*/
	QMIErrorNoSim                                        /*< nick=NoSim >*/
	QMIErrorPbNotReady                                   /*< nick=PbNotReady >*/
	QMIErrorPinRestriction                               /*< nick=PinRestriction >*/
	QMIErrorPin2Restriction                              /*< nick=Pin1Restriction >*/
	QMIErrorPukRestriction                               /*< nick=PukRestriction >*/
	QMIErrorPuk2Restriction                              /*< nick=Puk2Restriction >*/
	QMIErrorPbAccessRestricted                           /*< nick=PbAccessRestricted >*/
	QMIErrorPbDeleteInProgress                           /*< nick=PbDeleteInProgress >*/
	QMIErrorPbTextTooLong                                /*< nick=PbTextTooLong >*/
	QMIErrorPbNumberTooLong                              /*< nick=PbNumberTooLong >*/
	QMIErrorPbHiddenKeyRestriction                       /*< nick=PbHiddenKeyRestriction >*/
	QMIErrorPbNotAvailable                               /*< nick=PbNotAvailable >*/
	QMIErrorDeviceMemoryError                            /*< nick=DeviceMemoryError >*/
	QMIErrorNoPermission                                 /*< nick=NoPermission >*/
	QMIErrorTooSoon                                      /*< nick=TooSoon >*/
	QMIErrorTimeNotAcquired                              /*< nick=TimeNotAcquired >*/
	QMIErrorOperationInProgress                          /*< nick=OperationInProgress >*/
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

func (e QMIError) Error() string {
	switch e {
	case QMIErrorNone:
		return "No error"
	case QMIErrorMalformedMessage:
		return "Malformed message"
	case QMIErrorNoMemory:
		return "No memory"
	case QMIErrorInternal:
		return "Internal error"
	case QMIErrorAborted:
		return "Operation aborted"
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
		return "No free profile available"
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
	case QMIErrorMaximumQosRequestsInUse:
		return "Maximum QoS requests in use"
	case QMIErrorIncorrectFlowFilter:
		return "Incorrect flow filter"
	case QMIErrorNetworkQosUnaware:
		return "Network QoS unaware"
	case QMIErrorInvalidQosId:
		return "Invalid QoS ID"
	case QMIErrorRequestedNumberUnsupported:
		return "Requested number unsupported"
	case QMIErrorInterfaceNotFound:
		return "Interface not found"
	case QMIErrorFlowSuspended:
		return "Flow suspended"
	case QMIErrorInvalidDataFormat:
		return "Invalid data format"
	case QMIErrorGeneralError:
		return "General error"
	case QMIErrorUnknownError:
		return "Unknown error"
	case QMIErrorInvalidArgument:
		return "Invalid argument"
	case QMIErrorInvalidIndex:
		return "Invalid index"
	case QMIErrorNoEntry:
		return "No entry"
	case QMIErrorDeviceStorageFull:
		return "Device storage full"
	case QMIErrorDeviceNotReady:
		return "Device not ready"
	case QMIErrorNetworkNotReady:
		return "Network not ready"
	case QMIErrorWmsCauseCode:
		return "WMS cause code"
	case QMIErrorWmsMessageNotSent:
		return "WMS message not sent"
	case QMIErrorWmsMessageDeliveryFailure:
		return "WMS message delivery failure"
	case QMIErrorWmsInvalidMessageId:
		return "WMS invalid message ID"
	case QMIErrorWmsEncoding:
		return "WMS encoding error"
	case QMIErrorAuthenticationLock:
		return "Authentication lock"
	case QMIErrorInvalidTransition:
		return "Invalid transition"
	case QMIErrorNotMcastInterface:
		return "Not a multicast interface"
	case QMIErrorMaximumMcastRequestsInUse:
		return "Maximum multicast requests in use"
	case QMIErrorInvalidMcastHandle:
		return "Invalid multicast handle"
	case QMIErrorInvalidIpFamilyPreference:
		return "Invalid IP family preference"
	case QMIErrorSessionInactive:
		return "Session inactive"
	case QMIErrorSessionInvalid:
		return "Session invalid"
	case QMIErrorSessionOwnership:
		return "Session ownership error"
	case QMIErrorInsufficientResources:
		return "Insufficient resources"
	case QMIErrorDisabled:
		return "Disabled"
	case QMIErrorInvalidOperation:
		return "Invalid operation"
	case QMIErrorInvalidQmiCommand:
		return "Invalid QMI command"
	case QMIErrorWmsTPduType:
		return "WMS TPDU type error"
	case QMIErrorWmsSmscAddress:
		return "WMS SMSC address error"
	case QMIErrorInformationUnavailable:
		return "Information unavailable"
	case QMIErrorSegmentTooLong:
		return "Segment too long"
	case QMIErrorSegmentOrder:
		return "Segment order error"
	case QMIErrorBundlingNotSupported:
		return "Bundling not supported"
	case QMIErrorOperationPartialFailure:
		return "Operation partial failure"
	case QMIErrorPolicyMismatch:
		return "Policy mismatch"
	case QMIErrorSimFileNotFound:
		return "SIM file not found"
	case QMIErrorExtendedInternal:
		return "Extended internal error"
	case QMIErrorAccessDenied:
		return "Access denied"
	case QMIErrorHardwareRestricted:
		return "Hardware restricted"
	case QMIErrorAckNotSent:
		return "Acknowledgment not sent"
	case QMIErrorInjectTimeout:
		return "Inject timeout"
	case QMIErrorIncompatibleState:
		return "Incompatible state"
	case QMIErrorFdnRestrict:
		return "Fdn restrict"
	case QMIErrorSupsFailureCase:
		return "Sups failure case"
	case QMIErrorNoRadio:
		return "No radio"
	case QMIErrorNotSupported:
		return "Not supported"
	case QMIErrorNoSubscription:
		return "No subscription"
	case QMIErrorCardCallControlFailed:
		return "Card call control failed"
	case QMIErrorNetworkAborted:
		return "Network aborted"
	case QMIErrorMsgBlocked:
		return "Message blocked"
	case QMIErrorInvalidSessionType:
		return "Invalid session type"
	case QMIErrorInvalidPbType:
		return "Invalid phonebook type"
	case QMIErrorNoSim:
		return "No SIM card"
	case QMIErrorPbNotReady:
		return "Phonebook not ready"
	case QMIErrorPinRestriction:
		return "PIN restriction"
	case QMIErrorPin2Restriction:
		return "PIN2 restriction"
	case QMIErrorPukRestriction:
		return "PUK restriction"
	case QMIErrorPuk2Restriction:
		return "PUK2 restriction"
	case QMIErrorPbAccessRestricted:
		return "Phonebook access restricted"
	case QMIErrorPbDeleteInProgress:
		return "Phonebook delete in progress"
	case QMIErrorPbTextTooLong:
		return "Phonebook text too long"
	case QMIErrorPbNumberTooLong:
		return "Phonebook number too long"
	case QMIErrorPbHiddenKeyRestriction:
		return "Phonebook hidden key restriction"
	case QMIErrorPbNotAvailable:
		return "Phonebook not available"
	case QMIErrorDeviceMemoryError:
		return "Device memory error"
	case QMIErrorNoPermission:
		return "No permission"
	case QMIErrorTooSoon:
		return "Too soon"
	case QMIErrorTimeNotAcquired:
		return "Time not acquired"
	case QMIErrorOperationInProgress:
		return "Operation in progress"
	case QMIErrorFwWriteFailed:
		return "Firmware write failed"
	case QMIErrorFwInfoReadFailed:
		return "Firmware info read failed"
	case QMIErrorFwFileNotFound:
		return "Firmware file not found"
	case QMIErrorFwDirNotFound:
		return "Firmware directory not found"
	case QMIErrorFwAlreadyActivated:
		return "Firmware already activated"
	case QMIErrorFwCannotGenericImage:
		return "Firmware cannot generic image"
	case QMIErrorFwFileOpenFailed:
		return "Firmware file open failed"
	case QMIErrorFwUpdateDiscontinuousFrame:
		return "Firmware update discontinuous frame"
	case QMIErrorFwUpdateFailed:
		return "Firmware update failed"
	case QMIErrorCatEventRegistrationFailed:
		return "CAT event registration failed"
	case QMIErrorCatInvalidTerminalResponse:
		return "CAT invalid terminal response"
	case QMIErrorCatInvalidEnvelopeCommand:
		return "CAT invalid envelope command"
	case QMIErrorCatEnvelopeCommandBusy:
		return "CAT envelope command busy"
	case QMIErrorCatEnvelopeCommandFailed:
		return "CAT envelope command failed"
	default:
		return "Unknown QMI protocol error"
	}
}
