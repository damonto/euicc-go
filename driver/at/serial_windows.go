package at

import (
	"errors"
	"fmt"
	"io"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type SerialPort struct {
	handle     windows.Handle
	portName   string
	readEvent  windows.Handle
	writeEvent windows.Handle
}

func Open(port string) (io.ReadWriteCloser, error) {
	comPort := windows.StringToUTF16Ptr("\\\\.\\" + port)
	handle, err := windows.CreateFile(comPort,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0, nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OVERLAPPED,
		0)
	if err != nil {
		return nil, fmt.Errorf("CreateFile failed: %w", err)
	}

	// Configure DCB
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
	dcb.Flags = dcb.Flags | 0x00000001 // fBinary = 1
	dcb.Flags &^= 0x00000002           // fParity = 0
	// dcb.Flags |= 0x00000010         // fDtrControl = DTR_CONTROL_ENABLE (bit 4)
	// Use DTR_CONTROL_HANDSHAKE or disable it if you control it manually
	dcb.Flags &^= (1 << 4) | (1 << 5) // Clear DTR and RTS control bits before setting

	if err := windows.SetCommState(handle, &dcb); err != nil {
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("SetCommState failed: %w", err)
	}

	// Manually assert DTR after setting state
	if err := windows.EscapeCommFunction(handle, windows.SETDTR); err != nil {
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("SetDTR failed: %w", err)
	}

	// Set timeouts
	timeouts := windows.CommTimeouts{
		ReadIntervalTimeout:         50,
		ReadTotalTimeoutConstant:    50,
		ReadTotalTimeoutMultiplier:  10,
		WriteTotalTimeoutConstant:   50,
		WriteTotalTimeoutMultiplier: 10,
	}
	if err := windows.SetCommTimeouts(handle, &timeouts); err != nil {
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("SetCommTimeouts failed: %w", err)
	}

	// Purge any stale data (important for USB modem)
	windows.PurgeComm(handle, windows.PURGE_RXCLEAR|windows.PURGE_TXCLEAR)

	readEvent, _ := windows.CreateEvent(nil, 1, 0, nil)
	writeEvent, _ := windows.CreateEvent(nil, 1, 0, nil)

	return &SerialPort{
		handle:     handle,
		portName:   port,
		readEvent:  readEvent,
		writeEvent: writeEvent,
	}, nil
}

func (sp *SerialPort) Read(p []byte) (int, error) {
	overlapped := windows.Overlapped{HEvent: sp.readEvent}
	var bytesRead uint32

	err := windows.ReadFile(sp.handle, p, &bytesRead, &overlapped)
	if err != nil && err != windows.ERROR_IO_PENDING {
		return 0, fmt.Errorf("ReadFile failed: %w", err)
	}

	s, _ := windows.WaitForSingleObject(sp.readEvent, 5000) // increase timeout
	switch s {
	case uint32(windows.WAIT_OBJECT_0):
		err = windows.GetOverlappedResult(sp.handle, &overlapped, &bytesRead, false)
		if err != nil {
			return 0, fmt.Errorf("GetOverlappedResult failed: %w", err)
		}
		return int(bytesRead), nil
	case uint32(windows.WAIT_TIMEOUT):
		return 0, errors.New("read timeout")
	default:
		return 0, fmt.Errorf("unexpected wait result: %d", s)
	}
}

func (sp *SerialPort) Write(p []byte) (int, error) {
	overlapped := windows.Overlapped{HEvent: sp.writeEvent}
	var bytesWritten uint32

	err := windows.WriteFile(sp.handle, p, &bytesWritten, &overlapped)
	if err != nil && err != windows.ERROR_IO_PENDING {
		return 0, fmt.Errorf("WriteFile failed: %w", err)
	}

	s, _ := windows.WaitForSingleObject(sp.writeEvent, 5000)
	switch s {
	case uint32(windows.WAIT_OBJECT_0):
		err = windows.GetOverlappedResult(sp.handle, &overlapped, &bytesWritten, false)
		if err != nil {
			return 0, fmt.Errorf("GetOverlappedResult failed: %w", err)
		}
		return int(bytesWritten), nil
	case uint32(windows.WAIT_TIMEOUT):
		return 0, errors.New("write timeout")
	default:
		return 0, fmt.Errorf("unexpected wait result: %d", s)
	}
}

func (sp *SerialPort) Close() error {
	if err := windows.EscapeCommFunction(sp.handle, windows.CLRDTR); err != nil {
		return err
	}
	// Give the modem a moment to process the DTR change
	time.Sleep(100 * time.Millisecond)
	// Cancel any pending I/O operations
	if err := windows.CancelIoEx(sp.handle, nil); err != nil {
		return err
	}
	// Purge all communication buffers and abort pending I/O
	if err := windows.PurgeComm(sp.handle, windows.PURGE_RXABORT|windows.PURGE_TXABORT|windows.PURGE_RXCLEAR|windows.PURGE_TXCLEAR); err != nil {
		return err
	}
	// Close the event handles
	if sp.readEvent != 0 {
		windows.CloseHandle(sp.readEvent)
		sp.readEvent = 0
	}
	if sp.writeEvent != 0 {
		windows.CloseHandle(sp.writeEvent)
		sp.writeEvent = 0
	}
	return windows.CloseHandle(sp.handle)
}
