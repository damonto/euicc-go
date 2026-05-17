package uim

type PhysicalCardState uint32

const (
	PhysicalCardStateUnknown PhysicalCardState = 0x00
	PhysicalCardStateAbsent  PhysicalCardState = 0x01
	PhysicalCardStatePresent PhysicalCardState = 0x02
)

type SlotState uint32

const (
	SlotStateInactive SlotState = 0x00
	SlotStateActive   SlotState = 0x01
)

type CardState uint8

const (
	CardStateAbsent  CardState = 0x00
	CardStatePresent CardState = 0x01
	CardStateUnknown CardState = 0x02
)

type CardApplicationType uint8

const (
	CardApplicationTypeUnknown CardApplicationType = 0x00
	CardApplicationTypeSIM     CardApplicationType = 0x01
	CardApplicationTypeUSIM    CardApplicationType = 0x02
	CardApplicationTypeRUIM    CardApplicationType = 0x03
	CardApplicationTypeCSIM    CardApplicationType = 0x04
	CardApplicationTypeISIM    CardApplicationType = 0x05
)

type CardApplicationState uint8

const (
	CardApplicationStateUnknown                   CardApplicationState = 0x00
	CardApplicationStateDetected                  CardApplicationState = 0x01
	CardApplicationStatePIN1OrUPinPinRequired     CardApplicationState = 0x02
	CardApplicationStatePUK1OrUPinPUKRequired     CardApplicationState = 0x03
	CardApplicationStateCheckPersonalizationState CardApplicationState = 0x04
	CardApplicationStatePIN1Blocked               CardApplicationState = 0x05
	CardApplicationStateIllegal                   CardApplicationState = 0x06
	CardApplicationStateReady                     CardApplicationState = 0x07
)
