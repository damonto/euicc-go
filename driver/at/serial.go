//go:build linux || unix
// +build linux unix

package at

import (
	"os"

	"golang.org/x/sys/unix"
)

type SerialPort struct {
	f          *os.File
	oldTermios *unix.Termios
}

func OpenSerialPort(name string) (*SerialPort, error) {
	f, err := os.OpenFile(name, os.O_RDWR|unix.O_NOCTTY, 0666)
	if err != nil {
		return nil, err
	}
	sp := &SerialPort{f: f}
	if err := sp.setTermios(unix.B9600); err != nil {
		f.Close()
		return nil, err
	}
	return sp, nil
}

func (sp *SerialPort) setTermios(baudRate uint32) error {
	fd := int(sp.f.Fd())
	var err error
	if sp.oldTermios, err = unix.IoctlGetTermios(fd, unix.TCGETS); err != nil {
		return err
	}
	t := unix.Termios{
		Ispeed: baudRate,
		Ospeed: baudRate,
	}
	t.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	t.Oflag &^= unix.OPOST
	t.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	t.Cflag &^= unix.CSIZE | unix.PARENB
	t.Cflag |= unix.CS8
	t.Cc[unix.VMIN] = 1
	t.Cc[unix.VTIME] = 0
	return unix.IoctlSetTermios(fd, unix.TCSETS, &t)
}

func (sp *SerialPort) Read(buf []byte) (int, error) {
	n, err := sp.f.Read(buf)
	return n, err
}

func (sp *SerialPort) Write(data []byte) (int, error) {
	n, err := sp.f.Write(data)
	return n, err
}

func (sp *SerialPort) Close() error {
	if err := unix.IoctlSetTermios(int(sp.f.Fd()), unix.TCSETS, sp.oldTermios); err != nil {
		return err
	}
	return sp.f.Close()
}
