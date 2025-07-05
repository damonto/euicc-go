package mbim

type MessageType uint32

// MBIM Message Types
const (
	// MBIM Core Message Types
	MessageTypeOpen           MessageType = 0x00000001
	MessageTypeClose          MessageType = 0x00000002
	MessageTypeCommand        MessageType = 0x00000003
	MessageTypeHostError      MessageType = 0x00000004
	MessageTypeFunctionError  MessageType = 0x00000005
	MessageTypeIndicateStatus MessageType = 0x00000007

	// MBIM Response Message Types
	MessageTypeOpenDone           MessageType = 0x80000001
	MessageTypeCloseDone          MessageType = 0x80000002
	MessageTypeCommandDone        MessageType = 0x80000003
	MessageTypeFunctionErrorDone  MessageType = 0x80000005
	MessageTypeIndicateStatusDone MessageType = 0x80000007
)

// MBIM Services (UUIDs)
var (
	// Basic Connect Service
	ServiceBasicConnect = [16]byte{0xA2, 0x89, 0xCC, 0x33, 0xBC, 0xBB, 0x8B, 0x4F, 0xB6, 0xB0, 0x13, 0x3E, 0xC2, 0xAA, 0xE6, 0xDF}
	// MS UICC Low Level Access Service
	ServiceMsUiccLowLevelAccess = [16]byte{0xC2, 0xF6, 0x58, 0x8E, 0xF0, 0x37, 0x4B, 0xC9, 0x86, 0x65, 0xF4, 0xD4, 0x4B, 0xD0, 0x93, 0x67}
	// MS Basic Connect Extensions Service
	ServiceMsBasicConnectExtensions = [16]byte{0x3D, 0x01, 0xDC, 0xC5, 0xFE, 0xF5, 0x4D, 0x05, 0x0D, 0x3A, 0xBE, 0xF7, 0x05, 0x8E, 0x9A, 0xAF}
	// MBIM Proxy Control Service (internal libmbim service)
	ServiceMbimProxyControl = [16]byte{0x83, 0x8C, 0xF7, 0xFB, 0x8D, 0x0D, 0x4D, 0x7F, 0x87, 0x1E, 0xD7, 0x1D, 0xBE, 0xFB, 0xB3, 0x9B}
)

// MBIM Basic Connect CIDs
const (
	CIDDeviceCaps            = 0x00000001
	CIDSubscriberReadyStatus = 0x00000002
	CIDRadioState            = 0x00000003
	CIDPinState              = 0x00000004
	CIDPinList               = 0x00000005
	CIDHomeProvider          = 0x00000006
	CIDPreferredProviders    = 0x00000007
	CIDVisibleProviders      = 0x00000008
	CIDRegisterState         = 0x00000009
)

// MBIM UICC Low Level Access CIDs
const (
	CIDUiccATR                = 0x00000001
	CIDUiccOpenChannel        = 0x00000002
	CIDUiccCloseChannel       = 0x00000003
	CIDUiccAPDU               = 0x00000004
	CIDUiccTerminalCapability = 0x00000005
	CIDUiccReset              = 0x00000006
)

// MBIM Proxy Control CIDs
const (
	CIDProxyControlConfiguration = 0x00000001
	CIDProxyControlVersion       = 0x00000002
	CIDDeviceSlotMappings        = 0x00000007
	CIDSlotInfoStatus            = 0x00000008
)

// MBIM Subscriber Ready States
const (
	MBIMSubscriberReadyStateNotInitialized = 0x00000000
	MBIMSubscriberReadyStateInitialized    = 0x00000001
	MBIMSubscriberReadyStateSimNotInserted = 0x00000002
	MBIMSubscriberReadyStateBadSim         = 0x00000003
	MBIMSubscriberReadyStateFailure        = 0x00000004
	MBIMSubscriberReadyStateNotActivated   = 0x00000005
	MBIMSubscriberReadyStateDeviceLocked   = 0x00000006
	MBIMSubscriberReadyStateNoEsimProfile  = 0x00000007
)

// MBIM Command Types
const (
	CommandTypeQuery = 0x00000000
	CommandTypeSet   = 0x00000001
)
