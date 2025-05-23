//go:build linux

package qmi

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

func New(device string, slot uint8, useProxy bool) (apdu.SmartCardChannel, error) {
	q := (*C.struct_qmi_data)(C.calloc(1, C.sizeof_struct_qmi_data))
	if q == nil {
		return nil, errors.New("failed to allocate memory for QMI data")
	}
	// QMI uses 1-based indexing
	q.uim_slot = C.guint8(slot)
	// Try to open the port through the 'qmi-proxy'.
	if useProxy {
		q.use_proxy = C.gboolean(1)
	}
	return &qmi{
		device: device,
		qmi:    q,
	}, nil
}

func (q *qmi) Connect() error {
	cDevice := C.CString(q.device)
	defer C.free(unsafe.Pointer(cDevice))
	var cErr *C.char
	if C.go_qmi_apdu_connect(q.qmi, cDevice, &cErr) == -1 {
		defer C.free(unsafe.Pointer(cErr))
		defer C.free(unsafe.Pointer(q.qmi))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

func (q *qmi) Disconnect() error {
	defer C.free(unsafe.Pointer(q.qmi))
	var cErr *C.char
	if C.go_qmi_apdu_disconnect(q.qmi, &cErr) == -1 {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

func (q *qmi) Transmit(command []byte) ([]byte, error) {
	cCommand := C.CBytes(command)
	var cResponse *C.uint8_t
	var cResponseLen C.uint32_t
	defer C.free(unsafe.Pointer(cCommand))
	defer C.free(unsafe.Pointer(cResponse))
	var cErr *C.char
	if C.go_qmi_apdu_transmit(q.qmi, &cResponse, &cResponseLen, (*C.uchar)(cCommand), C.uint(len(command)), &cErr) == -1 {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}
	return C.GoBytes(unsafe.Pointer(cResponse), C.int(cResponseLen)), nil
}

func (q *qmi) OpenLogicalChannel(aid []byte) (byte, error) {
	cAID := C.CBytes(aid)
	defer C.free(unsafe.Pointer(cAID))
	var cErr *C.char
	channel := C.go_qmi_apdu_open_logical_channel(q.qmi, (*C.uchar)(cAID), C.uint8_t(len(aid)), &cErr)
	if channel < 1 {
		defer C.free(unsafe.Pointer(cErr))
		return 0, errors.New(C.GoString(cErr))
	}
	return byte(channel), nil
}

func (q *qmi) CloseLogicalChannel(channel byte) error {
	var cErr *C.char
	if C.go_qmi_apdu_close_logical_channel(q.qmi, C.uint8_t(channel), &cErr) == -1 {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}
