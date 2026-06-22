// Package qmi provides deprecated compatibility aliases for the Qualcomm driver.
//
// Deprecated: use github.com/damonto/euicc-go/driver/qcom.
package qmi

import (
	"github.com/damonto/euicc-go/driver"
	"github.com/damonto/euicc-go/driver/qcom"
)

// QMI is an alias for qcom.QMI.
//
// Deprecated: use qcom.QMI.
type QMI = qcom.QMI

// QRTR is an alias for qcom.QRTR.
//
// Deprecated: use qcom.QRTR.
type QRTR = qcom.QRTR

// New creates a new QMI connection to the specified device.
//
// Deprecated: use qcom.NewQMI.
func New(device string, slot uint8) (driver.SmartCardChannel, error) {
	return qcom.NewQMI(device, slot)
}

// NewQRTR creates a new QRTR connection to the UIM service.
//
// Deprecated: use qcom.NewQRTR.
func NewQRTR(slot uint8) (driver.SmartCardChannel, error) {
	return qcom.NewQRTR(slot)
}
