package ccid

import (
	"errors"

	"github.com/ElMostafaIdrassi/goscard"
	"github.com/damonto/euicc-go/apdu"
)

type CCID interface {
	apdu.SmartCardChannel
	ListReaders() ([]string, error)
	SetReader(reader string)
}

type CCIDReader struct {
	context goscard.Context
	card    goscard.Card
	channel byte
	reader  string
}

func New() (CCID, error) {
	if err := goscard.Initialize(goscard.NewDefaultLogger(goscard.LogLevelNone)); err != nil {
		return nil, err
	}
	context, _, err := goscard.NewContext(goscard.SCardScopeSystem, nil, nil)
	if err != nil {
		return nil, err
	}
	ccid := &CCIDReader{context: context}
	return ccid, nil
}

func (c *CCIDReader) ListReaders() ([]string, error) {
	readers, _, err := c.context.ListReaders(nil)
	if err != nil {
		return nil, err
	}
	if len(readers) == 0 {
		return nil, errors.New("no readers found")
	}
	return readers, nil
}

func (c *CCIDReader) SetReader(reader string) {
	c.reader = reader
}

func (c *CCIDReader) Connect() error {
	card, _, err := c.context.Connect(c.reader, goscard.SCardShareExclusive, goscard.SCardProtocolT0)
	if err != nil {
		return err
	}
	c.card = card
	_, err = c.Transmit([]byte{0x80, 0xAA, 0x00, 0x00, 0x0A, 0xA9, 0x08, 0x81, 0x00, 0x82, 0x01, 0x01, 0x83, 0x01, 0x07})
	return err
}

func (c *CCIDReader) Disconnect() error {
	defer goscard.Finalize()
	if _, err := c.card.Disconnect(goscard.SCardLeaveCard); err != nil {
		return err
	}
	if _, err := c.context.Release(); err != nil {
		return err
	}
	return nil
}

func (c *CCIDReader) Transmit(command []byte) ([]byte, error) {
	r, _, err := c.card.Transmit(&goscard.SCardIoRequestT0, command, nil)
	return r, err
}

func (c *CCIDReader) OpenLogicalChannel(aid []byte) (byte, error) {
	channel, err := c.Transmit([]byte{0x00, 0x70, 0x00, 0x00, 0x01})
	if err != nil {
		return 0, err
	}
	if channel[1] != 0x90 {
		return 0, errors.New("failed to open logical channel")
	}
	c.channel = channel[0]
	command := []byte{c.channel, 0xA4, 0x04, 0x00, byte(len(aid))}
	command = append(command, aid...)
	sw, err := c.Transmit(command)
	if err != nil {
		return 0, err
	}
	if sw[len(sw)-2] != 0x90 && sw[len(sw)-2] != 0x61 {
		return 0, errors.New("failed to select AID")
	}
	return c.channel, nil
}

func (c *CCIDReader) CloseLogicalChannel(channel byte) error {
	_, err := c.Transmit([]byte{0x00, 0x70, 0x80, channel, 0x00})
	return err
}
