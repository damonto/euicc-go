package driver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/damonto/euicc-go/bertlv"
	sgp22 "github.com/damonto/euicc-go/v2"
	wwanapdu "github.com/damonto/wwan-go/apdu"
)

const (
	maxLogicalChannel  = 19
	maxMSS             = 254
	maxStoreDataBlocks = 256
)

// SmartCardChannel provides serialized access to a smart card. Driver
// constructors configure channels without opening their transports; Connect
// performs the I/O needed to establish a session. SmartCardChannel is not safe
// for concurrent use; callers must serialize all operations, including
// Disconnect.
type SmartCardChannel interface {
	Connect() error
	Disconnect() error
	OpenLogicalChannel(aid []byte) (byte, error)
	Transmit(command []byte) ([]byte, error)
	CloseLogicalChannel(channel byte) error
}

// Transmitter exchanges BER-TLV commands with an eUICC. It is not safe for
// concurrent use; callers must serialize Transmit, TransmitRaw, and Close.
type Transmitter interface {
	sgp22.Transmitter
	Close() error
}

type transmitter struct {
	card *cardTransmitter
}

// NewTransmitter connects to channel and opens a logical channel for AID.
// The transmitter takes ownership of channel and disconnects it on failure or
// when Close is called. Logger must not be nil.
func NewTransmitter(logger *slog.Logger, channel SmartCardChannel, aid []byte, mss int) (Transmitter, error) {
	t, err := newCardTransmitter(logger, channel, aid, mss)
	if err != nil {
		return nil, err
	}
	return &transmitter{card: t}, nil
}

func (t *transmitter) Transmit(request bertlv.Marshaler, response bertlv.Unmarshaler) error {
	req, err := request.MarshalBERTLV()
	if err != nil {
		return err
	}
	bs, err := t.TransmitRaw(req.Bytes())
	if err != nil {
		return err
	}
	var tlv bertlv.TLV
	if err := tlv.UnmarshalBinary(bs); err != nil {
		return err
	}
	return response.UnmarshalBERTLV(&tlv)
}

func (t *transmitter) TransmitRaw(command []byte) ([]byte, error) {
	return t.card.exchange(command)
}

func (t *transmitter) Close() error {
	return t.card.Close()
}

type cardTransmitter struct {
	mss            int
	channel        SmartCardChannel
	logicalChannel byte
	logger         *slog.Logger
}

func newCardTransmitter(logger *slog.Logger, channel SmartCardChannel, aid []byte, mss int) (*cardTransmitter, error) {
	if channel == nil {
		return nil, errors.New("smart card channel is nil")
	}
	if mss < 1 || mss > maxMSS {
		return nil, fmt.Errorf("MSS must be between 1 and %d: got %d", maxMSS, mss)
	}
	if err := channel.Connect(); err != nil {
		return nil, errors.Join(
			fmt.Errorf("connect smart card channel: %w", err),
			disconnectChannel(channel),
		)
	}

	logicalChannel, err := channel.OpenLogicalChannel(aid)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("open logical channel: %w", err),
			disconnectChannel(channel),
		)
	}
	if logicalChannel == 0 || logicalChannel > maxLogicalChannel {
		return nil, errors.Join(
			fmt.Errorf("logical channel %d is outside 1..%d", logicalChannel, maxLogicalChannel),
			closeChannel(channel, logicalChannel),
			disconnectChannel(channel),
		)
	}

	return &cardTransmitter{
		mss:            mss,
		channel:        channel,
		logicalChannel: logicalChannel,
		logger:         logger,
	}, nil
}

func closeChannel(channel SmartCardChannel, logicalChannel byte) error {
	if logicalChannel == 0 {
		return nil
	}
	if err := channel.CloseLogicalChannel(logicalChannel); err != nil {
		return fmt.Errorf("close logical channel %d: %w", logicalChannel, err)
	}
	return nil
}

func disconnectChannel(channel SmartCardChannel) error {
	if err := channel.Disconnect(); err != nil {
		return fmt.Errorf("disconnect smart card channel: %w", err)
	}
	return nil
}

func (t *cardTransmitter) exchange(command []byte) ([]byte, error) {
	var responseData bytes.Buffer
	request := wwanapdu.Request{CLA: 0x80, INS: 0xE2}
	var response wwanapdu.Response
	var err error
	blockCount := 0
	if len(command) > 0 {
		blockCount = 1 + (len(command)-1)/t.mss
	}
	if blockCount > maxStoreDataBlocks {
		return nil, fmt.Errorf("command requires %d STORE DATA blocks; maximum is %d", blockCount, maxStoreDataBlocks)
	}
	block := 0
	for data := range slices.Chunk(command, t.mss) {
		request.Data = data
		request.P1 = 0x11
		request.P2 = byte(block)
		if block == blockCount-1 {
			request.P1 = 0x91
		}
		if response, err = t.transmitAPDU(&request); err != nil {
			break
		}
		block++
		if !response.HasMore() {
			responseData.Write(response.Data())
			continue
		}
		if err = t.readCommandResponse(&responseData, response.SW2()); err != nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return responseData.Bytes(), nil
}

func (t *cardTransmitter) transmitAPDU(request *wwanapdu.Request) (wwanapdu.Response, error) {
	t.setChannelToCLA(request, t.logicalChannel)
	command, err := request.MarshalBinary()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	debug := t.logger.Enabled(ctx, slog.LevelDebug)
	if debug {
		t.logger.DebugContext(ctx, "[APDU] sending", "command", fmt.Sprintf("%X", command))
	}
	b, err := t.channel.Transmit(command)
	if debug {
		if err != nil {
			t.logger.DebugContext(ctx, "[APDU] received", "response", fmt.Sprintf("%X", b), "error", err)
		} else {
			t.logger.DebugContext(ctx, "[APDU] received", "response", fmt.Sprintf("%X", b))
		}
	}
	if err != nil {
		return nil, err
	}
	response := wwanapdu.Response(b)
	if !response.OK() && !response.HasMore() {
		err = fmt.Errorf("returned an unexpected response with status %04X", response.SW())
	}
	return response, err
}

func (t *cardTransmitter) setChannelToCLA(request *wwanapdu.Request, channel byte) {
	if channel < 4 {
		request.CLA = (request.CLA & 0x9C) | channel
	} else if channel < 20 {
		request.CLA = (request.CLA & 0xB0) | 0x40 | (channel - 4)
	}
}

func (t *cardTransmitter) readCommandResponse(responseData *bytes.Buffer, le byte) error {
	var err error
	var request wwanapdu.Request
	var response wwanapdu.Response
	request.CLA = 0x80
	request.INS = 0xC0
	request.Le = &le
	for {
		if response, err = t.transmitAPDU(&request); err != nil {
			return err
		}
		responseData.Write(response.Data())
		if !response.HasMore() {
			break
		}
		*request.Le = response.SW2()
	}
	return nil
}

func (t *cardTransmitter) Close() error {
	return errors.Join(
		closeChannel(t.channel, t.logicalChannel),
		disconnectChannel(t.channel),
	)
}
