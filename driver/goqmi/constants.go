package goqmi

import "fmt"

// ServiceType represents QMI service types
type ServiceType uint8

const (
	QMIServiceCtl ServiceType = 0x00 // Control service
	QMIServiceUIM ServiceType = 0x0B // User Identity Module service
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
)

// QMUX header constants
const (
	QMUXHeaderIfType             = 0x01
	QMUXHeaderControlFlagRequest = 0x00

	// TLV type constants
	TLVTypeResult = 0x02 // Result TLV type (mandatory in most responses)
)

// QMIResult represents the result code in QMI responses
type QMIResult uint16

const (
	QMIResultSuccess QMIResult = 0x0000 // Success
	QMIResultFailure QMIResult = 0x0001 // Failure
)

// String returns a human-readable description of the QMI result
func (r QMIResult) String() string {
	switch r {
	case QMIResultSuccess:
		return "Success"
	case QMIResultFailure:
		return "Failure"
	default:
		return fmt.Sprintf("Unknown result %d", uint16(r))
	}
}
