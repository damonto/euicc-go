package core

// ServiceType represents QMI service types
type ServiceType uint8

const (
	QMIServiceControl ServiceType = 0x00 // Control service
	QMIServiceUIM     ServiceType = 0x0B // UIM service
)

// MessageType represents QMI message types
type MessageType uint8

const (
	QMIMessageTypeRequest    MessageType = 0x00
	QMIMessageTypeResponse   MessageType = 0x02
	QMIMessageTypeIndication MessageType = 0x04
)

// MessageID represents QMI command message IDs
type MessageID uint16

const (
	// CTL service commands
	QMICtlCmdAllocateClientID MessageID = 0x0022
	QMICtlCmdReleaseClientID  MessageID = 0x0023
	QMICtlInternalProxyOpen   MessageID = 0xFF00

	// UIM service commands
	QMIUIMSendAPDU            MessageID = 0x003B
	QMIUIMOpenLogicalChannel  MessageID = 0x0042
	QMIUIMCloseLogicalChannel MessageID = 0x003F
	QMIUIMSwitchSlot          MessageID = 0x0046
	QMIUIMGetSlotStatus       MessageID = 0x0047
	QMIUIMGetCardStatus       MessageID = 0x002F
)

// QMUX header constants
const (
	QMUXHeaderIfType             = 0x01
	QMUXHeaderControlFlagRequest = 0x00
)

// QMIResult represents the result code in QMI responses
type QMIResult uint16

const (
	QMIResultSuccess QMIResult = 0x0000 // Success
	QMIResultFailure QMIResult = 0x0001 // Failure
)

// UIM Physical Card State
type UIMPhysicalCardState uint32

const (
	UIMPhysicalCardStateUnknown UIMPhysicalCardState = 0x00
	UIMPhysicalCardStateAbsent  UIMPhysicalCardState = 0x01
	UIMPhysicalCardStatePresent UIMPhysicalCardState = 0x02
)

// UIM Slot State
type UIMSlotState uint32

const (
	UIMSlotStateInactive UIMSlotState = 0x00
	UIMSlotStateActive   UIMSlotState = 0x01
)

// Card Status

type UIMCardState uint8

const (
	UIMCardStatusAbsent  UIMCardState = 0x00
	UIMCardStatusPresent UIMCardState = 0x01
	UIMCardStatusUnknown UIMCardState = 0x02
)

// UIM Card Application Type
type UIMCardApplicationType uint8

const (
	UIMCardApplicationTypeUnknown UIMCardApplicationType = 0x00
	UIMCardApplicationTypeSIM     UIMCardApplicationType = 0x01
	UIMCardApplicationTypeUSIM    UIMCardApplicationType = 0x02
	UIMCardApplicationTypeRUIM    UIMCardApplicationType = 0x03
	UIMCardApplicationTypeCSIM    UIMCardApplicationType = 0x04
	UIMCardApplicationTypeISIM    UIMCardApplicationType = 0x05
)

// UIM Card Application State
type UIMCardApplicationState uint8

const (
	UIMCardApplicationStateUnknown                   UIMCardApplicationState = 0x00
	UIMCardApplicationStateDetected                  UIMCardApplicationState = 0x01
	UIMCardApplicationStatePIN1OrUPinPinRequired     UIMCardApplicationState = 0x02
	UIMCardApplicationStatePUK1OrUPinPUKRequired     UIMCardApplicationState = 0x03
	UIMCardApplicationStateCheckPersonalizationState UIMCardApplicationState = 0x04
	UIMCardApplicationStatePIN1Blocked               UIMCardApplicationState = 0x05
	UIMCardApplicationStateIllegal                   UIMCardApplicationState = 0x06
	UIMCardApplicationStateReady                     UIMCardApplicationState = 0x07
)
