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
	pcsc := &CCIDReader{}
	if err := goscard.Initialize(goscard.NewDefaultLogger(goscard.LogLevelNone)); err != nil {
		return nil, err
	}

	context, _, err := goscard.NewContext(goscard.SCardScopeSystem, nil, nil)
	if err != nil {
		return nil, err
	}
	pcsc.context = context
	readers, err := pcsc.ListReaders()
	if err != nil {
		return nil, err
	}
	pcsc.SetReader(readers[0])
	return pcsc, nil
}

func (p *CCIDReader) ListReaders() ([]string, error) {
	readers, _, err := p.context.ListReaders(nil)
	if err != nil {
		return nil, err
	}
	if len(readers) == 0 {
		return nil, errors.New("no readers found")
	}
	return readers, nil
}

func (p *CCIDReader) SetReader(reader string) {
	p.reader = reader
}

func (p *CCIDReader) Connect() error {
	card, _, err := p.context.Connect(p.reader, goscard.SCardShareExclusive, goscard.SCardProtocolT0)
	if err != nil {
		return err
	}
	p.card = card
	_, err = p.Transmit([]byte{0x80, 0xAA, 0x00, 0x00, 0x0A, 0xA9, 0x08, 0x81, 0x00, 0x82, 0x01, 0x01, 0x83, 0x01, 0x07})
	return err
}

func (p *CCIDReader) Disconnect() error {
	defer goscard.Finalize()
	if _, err := p.card.Disconnect(goscard.SCardLeaveCard); err != nil {
		return err
	}
	if _, err := p.context.Release(); err != nil {
		return err
	}
	return nil
}

func (p *CCIDReader) Transmit(command []byte) ([]byte, error) {
	resp, _, err := p.card.Transmit(&goscard.SCardIoRequestT0, command, nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *CCIDReader) OpenLogicalChannel(aid []byte) (byte, error) {
	channel, err := p.Transmit([]byte{0x00, 0x70, 0x00, 0x00, 0x01})
	if err != nil {
		return 0, err
	}
	if channel[1] != 0x90 {
		return 0, errors.New("failed to open logical channel")
	}
	p.channel = channel[0]
	command := []byte{p.channel, 0xA4, 0x04, 0x00, byte(len(aid))}
	command = append(command, aid...)
	sw, err := p.Transmit(command)
	if err != nil {
		return 0, err
	}
	if sw[len(sw)-2] != 0x90 && sw[len(sw)-2] != 0x61 {
		return 0, errors.New("failed to select AID")
	}
	return p.channel, nil
}

func (p *CCIDReader) CloseLogicalChannel(channel byte) error {
	command := []byte{0x00, 0x70, 0x80, channel, 0x00}
	_, err := p.Transmit(command)
	return err
}
