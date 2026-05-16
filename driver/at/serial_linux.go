//go:build linux
// +build linux

package at

import (
	"errors"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

const linuxReadTimeoutDeciseconds = 50

type SerialPort struct {
	f          *os.File
	oldTermios *unix.Termios
}

func Open(name string) (io.ReadWriteCloser, error) {
	fd, err := unix.Open(name, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	oldTermios, err := setTermios(fd, unix.B115200)
	if err != nil {
		return nil, errors.Join(err, unix.Close(fd))
	}
	if err := flushSerial(fd); err != nil {
		return nil, errors.Join(err, unix.IoctlSetTermios(fd, unix.TCSETS, oldTermios), unix.Close(fd))
	}
	if err := unix.SetNonblock(fd, false); err != nil {
		return nil, errors.Join(err, unix.IoctlSetTermios(fd, unix.TCSETS, oldTermios), unix.Close(fd))
	}

	return &SerialPort{
		f:          os.NewFile(uintptr(fd), name),
		oldTermios: oldTermios,
	}, nil
}

func setTermios(fd int, baudRate uint32) (*unix.Termios, error) {
	oldTermios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return nil, err
	}
	t := *oldTermios
	t.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON | unix.IXOFF | unix.IXANY
	t.Oflag &^= unix.OPOST
	t.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	t.Cflag &^= unix.CSIZE | unix.PARENB | unix.CSTOPB | unix.CRTSCTS | unix.CBAUD
	t.Cflag |= unix.CS8 | unix.CREAD | unix.CLOCAL | baudRate
	t.Ispeed = baudRate
	t.Ospeed = baudRate
	t.Cc[unix.VMIN] = 0
	t.Cc[unix.VTIME] = linuxReadTimeoutDeciseconds
	if err := unix.IoctlSetTermios(fd, unix.TCSETS, &t); err != nil {
		return nil, err
	}
	return oldTermios, nil
}

func flushSerial(fd int) error {
	return unix.IoctlSetInt(fd, unix.TCFLSH, unix.TCIOFLUSH)
}

func (sp *SerialPort) Read(buf []byte) (int, error) {
	n, err := sp.f.Read(buf)
	if n == 0 && err == nil {
		return 0, errReadTimeout
	}
	return n, err
}

func (sp *SerialPort) Write(data []byte) (int, error) {
	n, err := sp.f.Write(data)
	return n, err
}

func (sp *SerialPort) Close() error {
	var errs []error
	if sp.oldTermios != nil {
		if err := unix.IoctlSetTermios(int(sp.f.Fd()), unix.TCSETS, sp.oldTermios); err != nil {
			errs = append(errs, err)
		}
	}
	if err := sp.f.Close(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
