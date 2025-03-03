package apdu

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"sync"
)

type Transmitter struct {
	MSS            int
	mutex          sync.Mutex
	channel        SmartCardChannel
	logicalChannel byte
	response       *bytes.Buffer
}

func NewTransmitter(channel SmartCardChannel, AID []byte, MSS int) (io.ReadWriteCloser, error) {
	var err error
	if err = channel.Connect(); err != nil {
		return nil, err
	}
	var transmitter Transmitter
	transmitter.channel = channel
	if transmitter.logicalChannel, err = channel.OpenLogicalChannel(AID); err != nil {
		return nil, err
	}
	transmitter.MSS = MSS
	return &transmitter, nil
}

func (t *Transmitter) Read(p []byte) (n int, err error) {
	return t.response.Read(p)
}

func (t *Transmitter) Write(command []byte) (n int, err error) {
	t.response = new(bytes.Buffer)
	request := Request{CLA: 0x80, INS: 0xE2}
	var response Response
	chunks := byte(len(command) / t.MSS)
	for request.Data = range slices.Chunk(command, t.MSS) {
		if request.P1 = 0x11; request.P2 == chunks {
			request.P1 = 0x91
		}
		if response, err = t.transmit(&request); err != nil {
			break
		}
		request.P2++
		n += len(request.Data)
		if !response.HasMore() {
			t.response.Write(response.Data())
			continue
		}
		if err = t.readCommandResponse(t.response, response.SW2()); err != nil {
			break
		}
	}
	return
}

func (t *Transmitter) transmit(request *Request) (response Response, err error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	request.CLA = (request.CLA & 0xF0) | (t.logicalChannel & 0x0F)
	if response, err = t.channel.Transmit(request.APDU()); err != nil {
		return
	}
	if !response.OK() && !response.HasMore() {
		err = fmt.Errorf("returned an unexpected response with status %04X", response.SW())
	}
	return
}

func (t *Transmitter) readCommandResponse(w io.Writer, le byte) error {
	var err error
	var request Request
	var response Response
	request.CLA = 0x80
	request.INS = 0xC0
	request.Le = &le
	for {
		if response, err = t.transmit(&request); err != nil {
			return err
		}
		if _, err = w.Write(response.Data()); err != nil {
			return err
		}
		if !response.HasMore() {
			break
		}
		*request.Le = response.SW2()
	}
	return nil
}

func (t *Transmitter) Close() error {
	if err := t.channel.CloseLogicalChannel(t.logicalChannel); err != nil {
		return err
	}
	return t.channel.Disconnect()
}
