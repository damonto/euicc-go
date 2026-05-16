//go:build windows
// +build windows

package at

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	windowsSerialTimeoutMS = 5000
	windowsMaxDWORD        = ^uint32(0)

	dcbFlagBinary       = 1 << 0
	dcbFlagParity       = 1 << 1
	dcbFlagOutxCtsFlow  = 1 << 2
	dcbFlagOutxDsrFlow  = 1 << 3
	dcbFlagDtrControl   = 3 << 4
	dcbFlagDsrSensitive = 1 << 6
	dcbFlagOutX         = 1 << 8
	dcbFlagInX          = 1 << 9
	dcbFlagErrorChar    = 1 << 10
	dcbFlagNull         = 1 << 11
	dcbFlagRtsControl   = 3 << 12
	dcbFlagAbortOnError = 1 << 14

	dcbDtrControlEnable = 1 << 4
	dcbRtsControlEnable = 1 << 12
)

var errWriteTimeout = errors.New("write timeout")

type SerialPort struct {
	handle windows.Handle
}

func Open(port string) (io.ReadWriteCloser, error) {
	comPort := windows.StringToUTF16Ptr(windowsPortName(port))
	handle, err := windows.CreateFile(comPort,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0, nil,
		windows.OPEN_EXISTING,
		0,
		0)
	if err != nil {
		return nil, fmt.Errorf("CreateFile failed: %w", err)
	}

	var dcb windows.DCB
	dcb.DCBlength = uint32(unsafe.Sizeof(dcb))
	if err := windows.GetCommState(handle, &dcb); err != nil {
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("GetCommState failed: %w", err)
	}
	dcb.BaudRate = 115200
	dcb.ByteSize = 8
	dcb.Parity = windows.NOPARITY
	dcb.StopBits = windows.ONESTOPBIT
	dcb.Flags |= dcbFlagBinary
	dcb.Flags &^= dcbFlagParity | dcbFlagOutxCtsFlow | dcbFlagOutxDsrFlow |
		dcbFlagDtrControl | dcbFlagDsrSensitive | dcbFlagOutX | dcbFlagInX |
		dcbFlagErrorChar | dcbFlagNull | dcbFlagRtsControl | dcbFlagAbortOnError
	dcb.Flags |= dcbDtrControlEnable | dcbRtsControlEnable

	if err := windows.SetCommState(handle, &dcb); err != nil {
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("SetCommState failed: %w", err)
	}

	if err := windows.EscapeCommFunction(handle, windows.SETDTR); err != nil {
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("SetDTR failed: %w", err)
	}

	timeouts := windows.CommTimeouts{
		ReadIntervalTimeout:       windowsMaxDWORD,
		ReadTotalTimeoutConstant:  windowsSerialTimeoutMS,
		WriteTotalTimeoutConstant: windowsSerialTimeoutMS,
	}
	if err := windows.SetCommTimeouts(handle, &timeouts); err != nil {
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("SetCommTimeouts failed: %w", err)
	}

	if err := windows.PurgeComm(handle, windows.PURGE_RXCLEAR|windows.PURGE_TXCLEAR); err != nil {
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("PurgeComm failed: %w", err)
	}

	return &SerialPort{handle: handle}, nil
}

func (sp *SerialPort) Read(p []byte) (int, error) {
	var bytesRead uint32
	if err := windows.ReadFile(sp.handle, p, &bytesRead, nil); err != nil {
		return 0, fmt.Errorf("ReadFile failed: %w", err)
	}
	if bytesRead == 0 {
		return 0, errReadTimeout
	}
	return int(bytesRead), nil
}

func (sp *SerialPort) Write(p []byte) (int, error) {
	var bytesWritten uint32
	if err := windows.WriteFile(sp.handle, p, &bytesWritten, nil); err != nil {
		return 0, fmt.Errorf("WriteFile failed: %w", err)
	}
	if bytesWritten == 0 && len(p) > 0 {
		return 0, errWriteTimeout
	}
	return int(bytesWritten), nil
}

func (sp *SerialPort) Close() error {
	var errs []error
	if err := windows.EscapeCommFunction(sp.handle, windows.CLRDTR); err != nil {
		errs = append(errs, err)
	}
	if err := windows.PurgeComm(sp.handle, windows.PURGE_RXABORT|windows.PURGE_TXABORT|windows.PURGE_RXCLEAR|windows.PURGE_TXCLEAR); err != nil {
		errs = append(errs, err)
	}
	if err := windows.CloseHandle(sp.handle); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func windowsPortName(port string) string {
	if strings.HasPrefix(port, `\\.\`) {
		return port
	}
	return `\\.\` + port
}
