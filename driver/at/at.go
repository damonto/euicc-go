package at

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/damonto/euicc-go/apdu"
)

type AT struct {
	s       io.ReadWriteCloser
	channel byte
}

func New(device string) (apdu.SmartCardChannel, error) {
	var at AT
	var err error
	if at.s, err = Open(device); err != nil {
		return nil, fmt.Errorf("open serial port %s: %w", device, err)
	}
	return &at, nil
}

func (a *AT) run(command string) (string, error) {
	if _, err := a.s.Write([]byte(command + "\r\n")); err != nil {
		return "", err
	}
	reader := bufio.NewReader(a.s)
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
	cmd = fmt.Sprintf("AT+CSIM=%d,%q", len(cmd), cmd)
	r, err := a.run(cmd)
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
	if _, err := a.run("AT+CSIM=?"); err != nil {
		return err
	}
	_, err := a.Transmit([]byte{0x80, 0xAA, 0x00, 0x00, 0x0A, 0xA9, 0x08, 0x81, 0x00, 0x82, 0x01, 0x01, 0x83, 0x01, 0x07})
	return err
}

func (a *AT) OpenLogicalChannel(AID []byte) (byte, error) {
	channel, err := a.Transmit([]byte{0x00, 0x70, 0x00, 0x00, 0x01})
	if err != nil {
		return 0, err
	}
	if channel[len(channel)-2] != 0x90 {
		return 0, fmt.Errorf("open logical channel: %X", channel)
	}
	a.channel = channel[0]
	sw, err := a.Transmit(append([]byte{a.channel, 0xA4, 0x04, 0x00, byte(len(AID))}, AID...))
	if err != nil {
		return 0, err
	}
	if sw[len(sw)-2] != 0x90 && sw[len(sw)-2] != 0x61 {
		return 0, fmt.Errorf("select AID: %X", sw)
	}
	return a.channel, nil
}

func (a *AT) CloseLogicalChannel(channel byte) error {
	_, err := a.Transmit([]byte{0x00, 0x70, 0x80, channel, 0x00})
	return err
}

func (a *AT) Disconnect() error {
	return a.s.Close()
}
