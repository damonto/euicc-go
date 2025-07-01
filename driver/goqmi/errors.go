package goqmi

// QMIProtocolError represents QMI protocol errors as defined in libqmi
// These correspond to the "Error" field in QMI Result TLVs
type QMIProtocolError uint16

const (
	QMIProtocolErrorNone                        QMIProtocolError = iota // No error
	QMIProtocolErrorMalformedMessage                                    // Malformed message
	QMIProtocolErrorNoMemory                                            // No memory
	QMIProtocolErrorInternal                                            // Internal error
	QMIProtocolErrorAborted                                             // Aborted
	QMIProtocolErrorClientIDsExhausted                                  // Client IDs exhausted
	QMIProtocolErrorUnabortableTransaction                              // Unabortable transaction
	QMIProtocolErrorInvalidClientID                                     // Invalid client ID
	QMIProtocolErrorNoThresholdsProvided                                // No thresholds provided
	QMIProtocolErrorInvalidHandle                                       // Invalid handle
	QMIProtocolErrorInvalidProfile                                      // Invalid profile
	QMIProtocolErrorInvalidPinID                                        // Invalid PIN ID
	QMIProtocolErrorIncorrectPin                                        // Incorrect PIN
	QMIProtocolErrorNoNetworkFound                                      // No network found
	QMIProtocolErrorCallFailed                                          // Call failed
	QMIProtocolErrorOutOfCall                                           // Out of call
	QMIProtocolErrorNotProvisioned                                      // Not provisioned
	QMIProtocolErrorMissingArgument                                     // Missing argument
	QMIProtocolErrorArgumentTooLong                                     // Argument too long
	QMIProtocolErrorInvalidTransactionID                                // Invalid transaction ID
	QMIProtocolErrorDeviceInUse                                         // Device in use
	QMIProtocolErrorNetworkUnsupported                                  // Network unsupported
	QMIProtocolErrorDeviceUnsupported                                   // Device unsupported
	QMIProtocolErrorNoEffect                                            // No effect
	QMIProtocolErrorNoFreeProfile                                       // No free profile
	QMIProtocolErrorInvalidPDPType                                      // Invalid PDP type
	QMIProtocolErrorInvalidTechnologyPreference                         // Invalid technology preference
	QMIProtocolErrorInvalidProfileType                                  // Invalid profile type
	QMIProtocolErrorInvalidServiceType                                  // Invalid service type
	QMIProtocolErrorInvalidRegisterAction                               // Invalid register action
	QMIProtocolErrorInvalidPSAttachAction                               // Invalid PS attach action
	QMIProtocolErrorAuthenticationFailed                                // Authentication failed
	QMIProtocolErrorPinBlocked                                          // PIN blocked
	QMIProtocolErrorPinAlwaysBlocked                                    // PIN always blocked
	QMIProtocolErrorUIMUninitialized                                    // UIM uninitialized
	QMIProtocolErrorMaximumQoSRequestsInUse                             // Maximum QoS requests in use
	QMIProtocolErrorIncorrectFlowFilter                                 // Incorrect flow filter
	QMIProtocolErrorNetworkQoSUnaware                                   // Network QoS unaware
	QMIProtocolErrorInvalidQoSID                                        // Invalid QoS ID
	QMIProtocolErrorRequestedNumberUnsupported                          // Requested number unsupported
	QMIProtocolErrorInterfaceNotFound                                   // Interface not found
	QMIProtocolErrorFlowSuspended                                       // Flow suspended
	QMIProtocolErrorInvalidDataFormat                                   // Invalid data format
	QMIProtocolErrorGeneralError                                        // General error
	QMIProtocolErrorUnknownError                                        // Unknown error
	QMIProtocolErrorInvalidArgument                                     // Invalid argument (commonly seen!)
	QMIProtocolErrorInvalidIndex                                        // Invalid index
	QMIProtocolErrorNoEntry                                             // No entry
	QMIProtocolErrorDeviceStorageFull                                   // Device storage full
	QMIProtocolErrorDeviceNotReady                                      // Device not ready
	QMIProtocolErrorNetworkNotReady                                     // Network not ready
	QMIProtocolErrorWMSCauseCode                                        // WMS cause code
	QMIProtocolErrorWMSMessageNotSent                                   // WMS message not sent
	QMIProtocolErrorWMSMessageDeliveryFailure                           // WMS message delivery failure
	QMIProtocolErrorWMSInvalidMessageID                                 // WMS invalid message ID
	QMIProtocolErrorWMSEncoding                                         // WMS encoding
	QMIProtocolErrorAuthenticationLock                                  // Authentication lock
	QMIProtocolErrorInvalidTransition                                   // Invalid transition
	QMIProtocolErrorSIMFileNotFound                                     // SIM file not found
	QMIProtocolErrorAccessDenied                                        // Access denied
	QMIProtocolErrorHardwareRestricted                                  // Hardware restricted
	QMIProtocolErrorIncompatibleState                                   // Incompatible state
	QMIProtocolErrorFDNRestrict                                         // FDN restrict
	QMIProtocolErrorNotSupported                                        // Not supported
	QMIProtocolErrorNoSubscription                                      // No subscription
	QMIProtocolErrorCardCallControlFailed                               // Card call control failed
	QMIProtocolErrorNetworkAborted                                      // Network aborted
	QMIProtocolErrorMsgBlocked                                          // Message blocked
	QMIProtocolErrorInvalidSessionType                                  // Invalid session type
	QMIProtocolErrorInvalidPBType                                       // Invalid PB type
	QMIProtocolErrorNoSIM                                               // No SIM
	QMIProtocolErrorPBNotReady                                          // PB not ready
	QMIProtocolErrorPinRestriction                                      // PIN restriction
	QMIProtocolErrorPin2Restriction                                     // PIN2 restriction
	QMIProtocolErrorPUKRestriction                                      // PUK restriction
	QMIProtocolErrorPUK2Restriction                                     // PUK2 restriction
	QMIProtocolErrorPBAccessRestricted                                  // PB access restricted
	QMIProtocolErrorPBDeleteInProgress                                  // PB delete in progress
	QMIProtocolErrorPBTextTooLong                                       // PB text too long
	QMIProtocolErrorPBNumberTooLong                                     // PB number too long
	QMIProtocolErrorPBHiddenKeyRestriction                              // PB hidden key restriction
	QMIProtocolErrorPBNotAvailable                                      // PB not available
	QMIProtocolErrorDeviceMemoryError                                   // Device memory error
	QMIProtocolErrorNoPermission                                        // No permission
	QMIProtocolErrorTooSoon                                             // Too soon
	QMIProtocolErrorTimeNotAcquired                                     // Time not acquired
	QMIProtocolErrorOperationInProgress                                 // Operation in progress
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
		return "Aborted"
	case QMIProtocolErrorClientIDsExhausted:
		return "Client IDs exhausted"
	case QMIProtocolErrorUnabortableTransaction:
		return "Unabortable transaction"
	case QMIProtocolErrorInvalidClientID:
		return "Invalid client ID"
	case QMIProtocolErrorNoThresholdsProvided:
		return "No thresholds provided"
	case QMIProtocolErrorInvalidHandle:
		return "Invalid handle"
	case QMIProtocolErrorInvalidProfile:
		return "Invalid profile"
	case QMIProtocolErrorInvalidPinID:
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
	case QMIProtocolErrorInvalidTransactionID:
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
		return "No free profile"
	case QMIProtocolErrorInvalidPDPType:
		return "Invalid PDP type"
	case QMIProtocolErrorInvalidTechnologyPreference:
		return "Invalid technology preference"
	case QMIProtocolErrorInvalidProfileType:
		return "Invalid profile type"
	case QMIProtocolErrorInvalidServiceType:
		return "Invalid service type"
	case QMIProtocolErrorInvalidRegisterAction:
		return "Invalid register action"
	case QMIProtocolErrorInvalidPSAttachAction:
		return "Invalid PS attach action"
	case QMIProtocolErrorAuthenticationFailed:
		return "Authentication failed"
	case QMIProtocolErrorPinBlocked:
		return "PIN blocked"
	case QMIProtocolErrorPinAlwaysBlocked:
		return "PIN always blocked"
	case QMIProtocolErrorUIMUninitialized:
		return "UIM uninitialized"
	case QMIProtocolErrorMaximumQoSRequestsInUse:
		return "Maximum QoS requests in use"
	case QMIProtocolErrorIncorrectFlowFilter:
		return "Incorrect flow filter"
	case QMIProtocolErrorNetworkQoSUnaware:
		return "Network QoS unaware"
	case QMIProtocolErrorInvalidQoSID:
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
	case QMIProtocolErrorWMSCauseCode:
		return "WMS cause code"
	case QMIProtocolErrorWMSMessageNotSent:
		return "WMS message not sent"
	case QMIProtocolErrorWMSMessageDeliveryFailure:
		return "WMS message delivery failure"
	case QMIProtocolErrorWMSInvalidMessageID:
		return "WMS invalid message ID"
	case QMIProtocolErrorWMSEncoding:
		return "WMS encoding"
	case QMIProtocolErrorAuthenticationLock:
		return "Authentication lock"
	case QMIProtocolErrorInvalidTransition:
		return "Invalid transition"
	case QMIProtocolErrorSIMFileNotFound:
		return "SIM file not found"
	case QMIProtocolErrorAccessDenied:
		return "Access denied"
	case QMIProtocolErrorHardwareRestricted:
		return "Hardware restricted"
	case QMIProtocolErrorIncompatibleState:
		return "Incompatible state"
	case QMIProtocolErrorFDNRestrict:
		return "FDN restrict"
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
	case QMIProtocolErrorInvalidPBType:
		return "Invalid PB type"
	case QMIProtocolErrorNoSIM:
		return "No SIM"
	case QMIProtocolErrorPBNotReady:
		return "PB not ready"
	case QMIProtocolErrorPinRestriction:
		return "PIN restriction"
	case QMIProtocolErrorPin2Restriction:
		return "PIN2 restriction"
	case QMIProtocolErrorPUKRestriction:
		return "PUK restriction"
	case QMIProtocolErrorPUK2Restriction:
		return "PUK2 restriction"
	case QMIProtocolErrorPBAccessRestricted:
		return "PB access restricted"
	case QMIProtocolErrorPBDeleteInProgress:
		return "PB delete in progress"
	case QMIProtocolErrorPBTextTooLong:
		return "PB text too long"
	case QMIProtocolErrorPBNumberTooLong:
		return "PB number too long"
	case QMIProtocolErrorPBHiddenKeyRestriction:
		return "PB hidden key restriction"
	case QMIProtocolErrorPBNotAvailable:
		return "PB not available"
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
	default:
		return "Unknown QMI protocol error"
	}
}
