package ccid

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ElMostafaIdrassi/goscard"
	"github.com/damonto/euicc-go/apdu"
)

const (
	maxLogicalChannel      = 19
	maxShortAPDUDataLength = 255
)

var pcsc struct {
	mu   sync.Mutex
	refs int
}

type CCID interface {
	apdu.SmartCardChannel
	ListReaders() ([]string, error)
	SetReader(reader string) error
}

type CCIDReader struct {
	mu      sync.Mutex
	context goscard.Context
	card    goscard.Card
	ioSend  *goscard.SCardIORequest
	channel byte
	reader  string
	open    bool
	closed  bool
}

func New() (*CCIDReader, error) {
	return NewWithReader("")
}

// NewWithReader creates a CCID channel with reader preselected.
func NewWithReader(reader string) (*CCIDReader, error) {
	if err := acquirePCSC(); err != nil {
		return nil, err
	}
	context, _, err := goscard.NewContext(goscard.SCardScopeSystem, nil, nil)
	if err != nil {
		releasePCSC()
		return nil, err
	}
	ccid := &CCIDReader{context: context, reader: reader}
	return ccid, nil
}

func (c *CCIDReader) ListReaders() ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, errors.New("ccid reader is closed")
	}
	readers, _, err := c.context.ListReaders(nil)
	if err != nil {
		return nil, err
	}
	return readers, nil
}

// SetReader selects the reader used by Connect.
func (c *CCIDReader) SetReader(reader string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errors.New("ccid reader is closed")
	}
	if c.open {
		return errors.New("ccid reader is connected")
	}
	c.reader = reader
	return nil
}

func (c *CCIDReader) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errors.New("ccid reader is closed")
	}
	if c.reader == "" {
		return errors.New("ccid reader is required")
	}

	var err error
	c.card, _, err = c.context.Connect(c.reader, goscard.SCardShareShared, goscard.SCardProtocolAny)
	if err != nil {
		return err
	}
	c.ioSend, err = ioRequestForProtocol(c.card.ActiveProtocol())
	if err != nil {
		c.open = true
		return errors.Join(err, c.disconnectCard())
	}
	c.open = true

	response, err := c.transmit([]byte{0x80, 0xAA, 0x00, 0x00, 0x0A, 0xA9, 0x08, 0x81, 0x00, 0x82, 0x01, 0x01, 0x83, 0x01, 0x07})
	if err != nil {
		return errors.Join(err, c.disconnectCard())
	}
	if !statusOK(response) && !statusHasMore(response) {
		return errors.Join(fmt.Errorf("connect APDU: %X", response), c.disconnectCard())
	}
	return nil
}

func (c *CCIDReader) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.disconnect()
}

func (c *CCIDReader) Transmit(command []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.transmit(command)
}

func (c *CCIDReader) transmit(command []byte) ([]byte, error) {
	if !c.open {
		if c.closed {
			return nil, errors.New("ccid reader is closed")
		}
		return nil, errors.New("ccid reader is not connected")
	}
	r, _, err := c.card.Transmit(c.ioSend, command, nil)
	return r, err
}

func (c *CCIDReader) OpenLogicalChannel(AID []byte) (byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(AID) > maxShortAPDUDataLength {
		return 0, fmt.Errorf("AID length %d exceeds short APDU limit", len(AID))
	}
	channel, err := c.openChannel()
	if err != nil {
		return 0, err
	}
	if err := c.selectAID(channel, AID); err != nil {
		return 0, errors.Join(err, c.closeLogicalChannel(channel))
	}
	c.channel = channel
	return channel, nil
}

func (c *CCIDReader) CloseLogicalChannel(channel byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closeLogicalChannel(channel)
}

func (c *CCIDReader) openChannel() (byte, error) {
	response, err := c.transmit([]byte{0x00, 0x70, 0x00, 0x00, 0x01})
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
	if channel == 0 || channel > maxLogicalChannel {
		return 0, fmt.Errorf("open logical channel returned invalid logical channel %d", channel)
	}
	return channel, nil
}

func (c *CCIDReader) selectAID(channel byte, AID []byte) error {
	command, err := selectAIDCommand(channel, AID)
	if err != nil {
		return err
	}
	response, err := c.transmit(command)
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

func (c *CCIDReader) closeLogicalChannel(channel byte) error {
	if channel == 0 || channel > maxLogicalChannel {
		return fmt.Errorf("invalid logical channel %d", channel)
	}
	response, err := c.transmit([]byte{0x00, 0x70, 0x80, channel, 0x00})
	if err != nil {
		return err
	}
	if len(response) < 2 {
		return fmt.Errorf("close logical channel returned short response: %X", response)
	}
	if !statusOK(response) {
		return fmt.Errorf("close logical channel: %X", response)
	}
	if c.channel == channel {
		c.channel = 0
	}
	return nil
}

func (c *CCIDReader) disconnect() error {
	if c.closed {
		return nil
	}
	if c.open {
		if err := c.disconnectCard(); err != nil {
			return err
		}
	}
	if _, err := c.context.Release(); err != nil {
		return err
	}
	c.closed = true
	releasePCSC()
	return nil
}

func (c *CCIDReader) disconnectCard() error {
	if !c.open {
		return nil
	}
	if _, err := c.card.Disconnect(goscard.SCardLeaveCard); err != nil {
		return err
	}
	c.open = false
	c.ioSend = nil
	c.channel = 0
	return nil
}

func acquirePCSC() error {
	pcsc.mu.Lock()
	defer pcsc.mu.Unlock()

	if pcsc.refs == 0 {
		if err := goscard.Initialize(goscard.NewDefaultLogger(goscard.LogLevelNone)); err != nil {
			return err
		}
	}
	pcsc.refs++
	return nil
}

func releasePCSC() {
	pcsc.mu.Lock()
	defer pcsc.mu.Unlock()

	if pcsc.refs == 0 {
		return
	}
	pcsc.refs--
	if pcsc.refs == 0 {
		goscard.Finalize()
	}
}

func ioRequestForProtocol(protocol goscard.SCardProtocol) (*goscard.SCardIORequest, error) {
	switch protocol {
	case goscard.SCardProtocolT0:
		return &goscard.SCardIoRequestT0, nil
	case goscard.SCardProtocolT1:
		return &goscard.SCardIoRequestT1, nil
	default:
		return nil, fmt.Errorf("unsupported active PC/SC protocol: %s", protocol.String())
	}
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

func statusOK(response []byte) bool {
	return len(response) >= 2 && response[len(response)-2] == 0x90 && response[len(response)-1] == 0x00
}

func statusHasMore(response []byte) bool {
	return len(response) >= 2 && response[len(response)-2] == 0x61
}
