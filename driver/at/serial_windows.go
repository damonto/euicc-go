//go:build windows
// +build windows

package at

import (
	"fmt"
	"io"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type DCB struct {
	DCBlength  uint32
	BaudRate   uint32
	Flags      uint32
	wReserved  uint16
	XonLim     uint16
	XoffLim    uint16
	ByteSize   byte
	Parity     byte
	StopBits   byte
	XonChar    byte
	XoffChar   byte
	ErrorChar  byte
	EofChar    byte
	EvtChar    byte
	wReserved1 uint16
}

type COMMTIMEOUTS struct {
	ReadIntervalTimeout         uint32
	ReadTotalTimeoutMultiplier  uint32
	ReadTotalTimeoutConstant    uint32
	WriteTotalTimeoutMultiplier uint32
	WriteTotalTimeoutConstant   uint32
}

var (
	modKernel32         = windows.NewLazySystemDLL("kernel32.dll")
	procGetCommState    = modKernel32.NewProc("GetCommState")
	procSetCommState    = modKernel32.NewProc("SetCommState")
	procSetCommTimeouts = modKernel32.NewProc("SetCommTimeouts")
)

type SerialPort struct {
	handle windows.Handle
}

func OpenSerialPort(name string) (io.ReadWriteCloser, error) {
	path := `\\.\` + name
	ptr, _ := syscall.UTF16PtrFromString(path)

	h, err := windows.CreateFile(
		ptr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("CreateFile: %w", err)
	}

	sp := &SerialPort{handle: h}

	if err := sp.setCommState(9600); err != nil {
		sp.Close()
		return nil, err
	}
	if err := sp.setCommTimeouts(); err != nil {
		sp.Close()
		return nil, err
	}

	return sp, nil
}

func (sp *SerialPort) setCommState(baudRate uint32) error {
	var dcb DCB
	dcb.DCBlength = uint32(unsafe.Sizeof(dcb))
	ret, _, err := procGetCommState.Call(uintptr(sp.handle), uintptr(unsafe.Pointer(&dcb)))
	if ret == 0 {
		return fmt.Errorf("GetCommState failed: %w", err)
	}
	dcb.BaudRate = baudRate
	dcb.ByteSize = 8
	dcb.Parity = 0
	dcb.StopBits = 0
	dcb.Flags = 0x0001 // fBinary
	ret, _, err = procSetCommState.Call(uintptr(sp.handle), uintptr(unsafe.Pointer(&dcb)))
	if ret == 0 {
		return fmt.Errorf("SetCommState failed: %w", err)
	}
	return nil
}

func (sp *SerialPort) setCommTimeouts() error {
	timeouts := COMMTIMEOUTS{
		ReadIntervalTimeout:         50,
		ReadTotalTimeoutMultiplier:  10,
		ReadTotalTimeoutConstant:    100,
		WriteTotalTimeoutMultiplier: 10,
		WriteTotalTimeoutConstant:   100,
	}
	ret, _, err := procSetCommTimeouts.Call(uintptr(sp.handle), uintptr(unsafe.Pointer(&timeouts)))
	if ret == 0 {
		return fmt.Errorf("SetCommTimeouts failed: %w", err)
	}
	return nil
}

func (sp *SerialPort) Read(buf []byte) (int, error) {
	var n uint32
	err := windows.ReadFile(sp.handle, buf, &n, nil)
	if err != nil {
		return 0, fmt.Errorf("ReadFile: %w", err)
	}
	return int(n), nil
}

func (sp *SerialPort) Write(data []byte) (int, error) {
	var n uint32
	err := windows.WriteFile(sp.handle, data, &n, nil)
	if err != nil {
		return 0, fmt.Errorf("WriteFile: %w", err)
	}
	return int(n), nil
}

func (sp *SerialPort) Close() error {
	return windows.CloseHandle(sp.handle)
}
