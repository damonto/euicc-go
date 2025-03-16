//go:build linux

package mbim

/*
#cgo pkg-config: glib-2.0 mbim-glib

#include <stdint.h>
#include <string.h>

#include "mbim.h"
*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/damonto/euicc-go/apdu"
)

type mbim struct {
	device string
	mbim   *C.struct_mbim_data
}

func New(device string, slot uint8, useProxy bool) (apdu.SmartCardChannel, error) {
	m := (*C.struct_mbim_data)(C.calloc(1, C.sizeof_struct_mbim_data))
	if m == nil {
		return nil, errors.New("failed to allocate memory for MBIM data")
	}
	// MBIM uses 0-based indexing
	m.uim_slot = C.guint32(slot - 1)
	// Try to open the port through the 'mbim-proxy'.
	if useProxy {
		m.use_proxy = C.gboolean(1)
	}
	return &mbim{
		device: device,
		mbim:   m,
	}, nil
}

func (m *mbim) Connect() error {
	cDevice := C.CString(m.device)
	defer C.free(unsafe.Pointer(cDevice))
	cErr := (*C.char)(C.calloc(256, C.sizeof_char))
	if cErr == nil {
		return errors.New("failed to allocate memory for error message")
	}
	defer C.free(unsafe.Pointer(cErr))
	if C.go_mbim_apdu_connect(m.mbim, cDevice, cErr) == -1 {
		return errors.New(C.GoString(cErr))
	}
	return nil
}

func (m *mbim) Disconnect() error {
	defer C.free(unsafe.Pointer(m.mbim))
	cErr := (*C.char)(C.calloc(256, C.sizeof_char))
	if cErr == nil {
		return errors.New("failed to allocate memory for error message")
	}
	defer C.free(unsafe.Pointer(cErr))
	if C.go_mbim_apdu_disconnect(m.mbim, cErr) == -1 {
		return errors.New(C.GoString(cErr))
	}
	return nil
}

func (m *mbim) Transmit(command []byte) ([]byte, error) {
	cCommand := C.CBytes(command)
	var cResponse *C.uint8_t
	var cResponseLen C.uint32_t
	defer C.free(unsafe.Pointer(cCommand))
	cErr := (*C.char)(C.calloc(256, C.sizeof_char))
	if cErr == nil {
		return nil, errors.New("failed to allocate memory for error message")
	}
	defer C.free(unsafe.Pointer(cErr))
	if C.go_mbim_apdu_transmit(m.mbim, &cResponse, &cResponseLen, (*C.uchar)(cCommand), C.uint(len(command)), cErr) == -1 {
		return nil, errors.New(C.GoString(cErr))
	}
	defer C.free(unsafe.Pointer(cResponse))
	response := C.GoBytes(unsafe.Pointer(cResponse), C.int(cResponseLen))
	return response, nil
}

func (m *mbim) OpenLogicalChannel(aid []byte) (byte, error) {
	cAID := C.CBytes(aid)
	defer C.free(unsafe.Pointer(cAID))
	cErr := (*C.char)(C.calloc(256, C.sizeof_char))
	if cErr == nil {
		return 0, errors.New("failed to allocate memory for error message")
	}
	defer C.free(unsafe.Pointer(cErr))
	channel := C.go_mbim_apdu_open_logical_channel(m.mbim, (*C.uchar)(cAID), C.uint8_t(len(aid)), cErr)
	if channel < 1 {
		return 0, errors.New(C.GoString(cErr))
	}
	return byte(channel), nil
}

func (m *mbim) CloseLogicalChannel(channel byte) error {
	cErr := (*C.char)(C.calloc(256, C.sizeof_char))
	if cErr == nil {
		return errors.New("failed to allocate memory for error message")
	}
	defer C.free(unsafe.Pointer(cErr))
	if C.go_mbim_apdu_close_logical_channel(m.mbim, C.uint8_t(channel), cErr) == -1 {
		return errors.New(C.GoString(cErr))
	}
	return nil
}
