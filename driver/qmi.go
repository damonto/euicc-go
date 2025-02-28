//go:build linux

package driver

/*
#cgo pkg-config: glib-2.0 qmi-glib

#include <stdint.h>
#include <string.h>

#include "qmi.h"
*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/damonto/euicc-go/apdu"
)

type qmi struct {
	device string
	qmi    *C.struct_qmi_data
}

func NewQMI(device string, uimSlot uint8) (apdu.SmartCardChannel, error) {
	q := (*C.struct_qmi_data)(C.malloc(C.sizeof_struct_qmi_data))
	if q == nil {
		return nil, errors.New("failed to allocate memory for QMI data")
	}
	C.memset(unsafe.Pointer(q), 0, C.sizeof_struct_qmi_data)
	q.uimSlot = C.uint8_t(uimSlot)
	return &qmi{
		device: device,
		qmi:    q,
	}, nil
}

func (q *qmi) Connect() error {
	cDevice := C.CString(q.device)
	defer C.free(unsafe.Pointer(cDevice))
	if C.libeuicc_qmi_apdu_connect(q.qmi, cDevice) == -1 {
		return errors.New("failed to connect to QMI")
	}
	return nil
}

func (q *qmi) Disconnect() error {
	C.libeuicc_qmi_apdu_disconnect(q.qmi)
	if q.qmi != nil {
		C.free(unsafe.Pointer(q.qmi))
		q.qmi = nil
	}
	return nil
}

func (q *qmi) Transmit(command []byte) ([]byte, error) {
	cCommand := C.CBytes(command)
	var cResponse *C.uint8_t
	var cResponseLen C.uint32_t
	defer C.free(unsafe.Pointer(cCommand))
	if C.libeuicc_qmi_apdu_transmit(q.qmi, &cResponse, &cResponseLen, (*C.uchar)(cCommand), C.uint(len(command))) == -1 {
		return nil, errors.New("failed to transmit APDU")
	}
	defer C.free(unsafe.Pointer(cResponse))
	response := C.GoBytes(unsafe.Pointer(cResponse), C.int(cResponseLen))
	return response, nil
}

func (q *qmi) OpenLogicalChannel(aid []byte) (byte, error) {
	cAID := C.CBytes(aid)
	defer C.free(unsafe.Pointer(cAID))
	channel := C.libeuicc_qmi_apdu_open_logical_channel(q.qmi, (*C.uchar)(cAID), C.uint8_t(len(aid)))
	if channel < 1 {
		return 0, errors.New("failed to open logical channel")
	}
	return byte(channel), nil
}

func (q *qmi) CloseLogicalChannel(channel byte) error {
	if C.libeuicc_qmi_apdu_close_logical_channel(q.qmi, C.uint8_t(channel)) == -1 {
		return errors.New("failed to close logical channel")
	}
	return nil
}
