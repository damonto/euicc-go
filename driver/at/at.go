package at

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/damonto/euicc-go/apdu"
	"golang.org/x/sys/unix"
)

type AT struct {
	f          *os.File
	oldTermios *unix.Termios
	channel    byte
}

func New(device string) (apdu.SmartCardChannel, error) {
	var at AT
	var err error
	if at.f, err = os.OpenFile(device, os.O_RDWR|unix.O_NOCTTY, 0666); err != nil {
		return nil, err
	}
	if err := at.setTermios(); err != nil {
		return nil, err
	}
	return &at, nil
}

func (a *AT) setTermios() error {
	fd := int(a.f.Fd())
	var err error
	if a.oldTermios, err = unix.IoctlGetTermios(fd, unix.TCGETS); err != nil {
		return err
	}
	t := unix.Termios{
		Ispeed: unix.B9600,
		Ospeed: unix.B9600,
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

func (a *AT) execute(command string) (string, error) {
	if _, err := a.f.WriteString(command + "\r\n"); err != nil {
		return "", err
	}
	reader := bufio.NewReader(a.f)
	var sb strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)
		switch {
		case strings.Contains(line, "OK"):
			return strings.TrimSpace(sb.String()), nil
		case strings.Contains(line, "ERR"):
			return "", errors.New(line)
		default:
			sb.WriteString(line + "\n")
		}
	}
}

func (a *AT) Transmit(command []byte) ([]byte, error) {
	cmd := fmt.Sprintf("%X", command)
	cmd = fmt.Sprintf("AT+CSIM=%d,\"%s\"", len(cmd), cmd)
	r, err := a.execute(cmd)
	if err != nil {
		return nil, err
	}
	sw, err := a.sw(r)
	if err != nil {
		return nil, err
	}
	if sw[len(sw)-2] != 0x90 && sw[len(sw)-2] != 0x61 {
		return sw, fmt.Errorf("unexpected response: %X", sw)
	}
	return sw, nil
}

func (a *AT) sw(sw string) ([]byte, error) {
	lastIdx := strings.LastIndex(sw, ",")
	if lastIdx == -1 {
		return nil, errors.New("invalid response")
	}
	return hex.DecodeString(sw[lastIdx+2 : len(sw)-1])
}

func (a *AT) Connect() error {
	_, err := a.execute("AT+CSIM=?")
	return err
}

func (a *AT) OpenLogicalChannel(aid []byte) (byte, error) {
	channel, err := a.Transmit([]byte{0x00, 0x70, 0x00, 0x00, 0x01})
	if err != nil {
		return 0, err
	}
	if channel[1] != 0x90 {
		return 0, errors.New("failed to open logical channel")
	}
	a.channel = channel[0]
	sw, err := a.Transmit(append([]byte{a.channel, 0xA4, 0x04, 0x00, byte(len(aid))}, aid...))
	if err != nil {
		return 0, err
	}
	if sw[len(sw)-2] != 0x90 && sw[len(sw)-2] != 0x61 {
		return 0, errors.New("failed to select AID")
	}
	return a.channel, nil
}

func (a *AT) CloseLogicalChannel(channel byte) error {
	_, err := a.Transmit([]byte{0x00, 0x70, 0x80, channel, 0x00})
	return err
}

func (a *AT) Disconnect() error {
	if err := unix.IoctlSetTermios(int(a.f.Fd()), unix.TCSETS, a.oldTermios); err != nil {
		return err
	}
	return a.f.Close()
}
