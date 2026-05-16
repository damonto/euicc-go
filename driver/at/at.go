package at

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/damonto/euicc-go/apdu"
)

const (
	maxLogicalChannel      = 19
	maxShortAPDUDataLength = 255
)

type AT struct {
	mu      sync.Mutex
	s       io.ReadWriteCloser
	reader  *bufio.Reader
	channel byte
}

var (
	errReadTimeout = errors.New("read timeout")
)

func New(device string) (apdu.SmartCardChannel, error) {
	s, err := Open(device)
	if err != nil {
		return nil, fmt.Errorf("open serial port %s: %w", device, err)
	}
	return &AT{
		s:      s,
		reader: bufio.NewReader(s),
	}, nil
}

func (a *AT) sendCommand(command string) (string, error) {
	if err := writeFull(a.s, []byte(command+"\r\n")); err != nil {
		return "", err
	}
	var sb strings.Builder
	for {
		line, err := a.reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)
		switch {
		case line == "", line == command:
			continue
		case line == "OK":
			return strings.TrimSpace(sb.String()), nil
		case line == "ERROR", strings.HasPrefix(line, "+CME ERROR:"), strings.HasPrefix(line, "+CMS ERROR:"):
			return "", errors.New(line)
		default:
			if sb.Len() > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(line)
		}
	}
}

func (a *AT) Transmit(command []byte) ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.transmitAPDU(command)
}

func (a *AT) transmitAPDU(command []byte) ([]byte, error) {
	cmd := fmt.Sprintf("%X", command)
	cmd = fmt.Sprintf("AT+CSIM=%d,%q", len(cmd), cmd)
	rawResponse, err := a.sendCommand(cmd)
	if err != nil {
		return nil, err
	}
	response, err := decodeCSIMResponse(rawResponse)
	if err != nil {
		return nil, err
	}
	if len(response) < 2 {
		return nil, errors.New("invalid response status word")
	}
	return response, nil
}

func writeFull(w io.Writer, p []byte) error {
	for len(p) > 0 {
		n, err := w.Write(p)
		if err != nil {
			return err
		}
		if n <= 0 {
			return io.ErrShortWrite
		}
		p = p[n:]
	}
	return nil
}

func (a *AT) Connect() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, err := a.sendCommand("AT+CSIM=?"); err != nil {
		return err
	}
	response, err := a.transmitAPDU([]byte{0x80, 0xAA, 0x00, 0x00, 0x0A, 0xA9, 0x08, 0x81, 0x00, 0x82, 0x01, 0x01, 0x83, 0x01, 0x07})
	if err != nil {
		return err
	}
	// The initialization APDU only needs an accepted status; response data is ignored.
	if !statusOK(response) && !statusHasMore(response) {
		return fmt.Errorf("connect APDU: %X", response)
	}
	return nil
}

func (a *AT) OpenLogicalChannel(AID []byte) (byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(AID) > maxShortAPDUDataLength {
		return 0, fmt.Errorf("AID length %d exceeds short APDU limit", len(AID))
	}
	channel, err := a.openChannel()
	if err != nil {
		return 0, err
	}
	if err := a.selectAID(channel, AID); err != nil {
		return 0, errors.Join(err, a.closeLogicalChannel(channel))
	}
	a.channel = channel
	return channel, nil
}

func (a *AT) openChannel() (byte, error) {
	response, err := a.transmitAPDU([]byte{0x00, 0x70, 0x00, 0x00, 0x01})
	if err != nil {
		return 0, err
	}
	if len(response) < 3 {
		return 0, fmt.Errorf("open logical channel returned short response: %X", response)
	}
	if !statusOK(response) {
		return 0, fmt.Errorf("open logical channel: %X", response)
	}
	channel := response[0]
	if err := validateLogicalChannel(channel); err != nil {
		return 0, fmt.Errorf("open logical channel returned %w", err)
	}
	return channel, nil
}

func (a *AT) selectAID(channel byte, AID []byte) error {
	command, err := selectAIDCommand(channel, AID)
	if err != nil {
		return err
	}
	response, err := a.transmitAPDU(command)
	if err != nil {
		return err
	}
	if len(response) < 2 {
		return fmt.Errorf("select AID returned short response: %X", response)
	}
	if !statusOK(response) && !statusHasMore(response) {
		return fmt.Errorf("select AID: %X", response)
	}
	return nil
}

func (a *AT) CloseLogicalChannel(channel byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.closeLogicalChannel(channel)
}

func (a *AT) closeLogicalChannel(channel byte) error {
	if err := validateLogicalChannel(channel); err != nil {
		return err
	}
	response, err := a.transmitAPDU([]byte{0x00, 0x70, 0x80, channel, 0x00})
	if err != nil {
		return err
	}
	if len(response) < 2 {
		return fmt.Errorf("close logical channel returned short response: %X", response)
	}
	if !statusOK(response) {
		return fmt.Errorf("close logical channel: %X", response)
	}
	if a.channel == channel {
		a.channel = 0
	}
	return nil
}

func (a *AT) Disconnect() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.s.Close()
}

func selectAIDCommand(channel byte, AID []byte) ([]byte, error) {
	cla, err := classByteForChannel(0x00, channel)
	if err != nil {
		return nil, err
	}
	if len(AID) > maxShortAPDUDataLength {
		return nil, fmt.Errorf("AID length %d exceeds short APDU limit", len(AID))
	}
	command := make([]byte, 0, 5+len(AID))
	command = append(command, cla, 0xA4, 0x04, 0x00, byte(len(AID)))
	command = append(command, AID...)
	return command, nil
}

func classByteForChannel(cla, channel byte) (byte, error) {
	if channel < 4 {
		return (cla & 0x9C) | channel, nil
	}
	if channel <= maxLogicalChannel {
		return (cla & 0xB0) | 0x40 | (channel - 4), nil
	}
	return 0, fmt.Errorf("logical channel %d exceeds maximum %d", channel, maxLogicalChannel)
}

func validateLogicalChannel(channel byte) error {
	if channel == 0 || channel > maxLogicalChannel {
		return fmt.Errorf("invalid logical channel %d", channel)
	}
	return nil
}
