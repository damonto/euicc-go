package driver

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/damonto/euicc-go/apdu"
	"github.com/damonto/euicc-go/bertlv"
	sgp22 "github.com/damonto/euicc-go/v2"
)

type Transmitter interface {
	sgp22.Transmitter
	Close() error
}

type transmitter struct {
	card   io.ReadWriteCloser
	logger *slog.Logger
}

func NewTransmitter(logger *slog.Logger, channel apdu.SmartCardChannel, AID []byte, MSS int) (Transmitter, error) {
	t, err := apdu.NewTransmitter(channel, AID, MSS)
	if err != nil {
		return nil, err
	}
	return &transmitter{card: t, logger: logger}, nil
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
	t.logger.Debug("[APDU] sending", "command", fmt.Sprintf("%X", command))
	if _, err := t.card.Write(command); err != nil {
		return nil, err
	}
	bs, err := io.ReadAll(t.card)
	if err != nil {
		return nil, err
	}
	t.logger.Debug("[APDU] received", "response", fmt.Sprintf("%X", bs))
	return bs, err
}

func (t *transmitter) Close() error {
	return t.card.Close()
}
