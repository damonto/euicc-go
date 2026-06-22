package driver

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
)

type fakeSmartCardChannel struct {
	logicalChannel byte
	responses      [][]byte
	requests       [][]byte
	connected      bool
	disconnected   bool
	openedAID      []byte
	closedChannel  byte
}

func (f *fakeSmartCardChannel) Connect() error {
	f.connected = true
	return nil
}

func (f *fakeSmartCardChannel) Disconnect() error {
	f.disconnected = true
	return nil
}

func (f *fakeSmartCardChannel) OpenLogicalChannel(AID []byte) (byte, error) {
	f.openedAID = append([]byte(nil), AID...)
	return f.logicalChannel, nil
}

func (f *fakeSmartCardChannel) Transmit(command []byte) ([]byte, error) {
	f.requests = append(f.requests, append([]byte(nil), command...))
	if len(f.responses) == 0 {
		return nil, errors.New("unexpected transmit")
	}
	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func (f *fakeSmartCardChannel) CloseLogicalChannel(channel byte) error {
	f.closedChannel = channel
	return nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewTransmitterConnectsAndClosesChannel(t *testing.T) {
	channel := &fakeSmartCardChannel{logicalChannel: 4}
	aid := []byte{0xA0, 0x00}

	tx, err := NewTransmitter(discardLogger(), channel, aid, 254)
	if err != nil {
		t.Fatalf("NewTransmitter() error = %v", err)
	}
	if !channel.connected {
		t.Fatal("NewTransmitter() did not connect channel")
	}
	if !bytes.Equal(channel.openedAID, aid) {
		t.Fatalf("OpenLogicalChannel() AID = % X, want % X", channel.openedAID, aid)
	}

	if err := tx.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if channel.closedChannel != 4 {
		t.Fatalf("CloseLogicalChannel() channel = %d, want 4", channel.closedChannel)
	}
	if !channel.disconnected {
		t.Fatal("Close() did not disconnect channel")
	}
}

func TestTransmitterSplitsStoreDataAPDUByMSS(t *testing.T) {
	channel := &fakeSmartCardChannel{
		logicalChannel: 1,
		responses: [][]byte{
			{0x90, 0x00},
			{0x90, 0x00},
		},
	}
	tx, err := NewTransmitter(discardLogger(), channel, []byte{0xA0}, 3)
	if err != nil {
		t.Fatalf("NewTransmitter() error = %v", err)
	}

	got, err := tx.TransmitRaw([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	if err != nil {
		t.Fatalf("TransmitRaw() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("TransmitRaw() = % X, want empty response", got)
	}

	want := [][]byte{
		{0x81, 0xE2, 0x11, 0x00, 0x03, 0x01, 0x02, 0x03},
		{0x81, 0xE2, 0x91, 0x01, 0x02, 0x04, 0x05},
	}
	if len(channel.requests) != len(want) {
		t.Fatalf("Transmit() request count = %d, want %d", len(channel.requests), len(want))
	}
	for i := range want {
		if !bytes.Equal(channel.requests[i], want[i]) {
			t.Fatalf("request %d = % X, want % X", i, channel.requests[i], want[i])
		}
	}
}

func TestTransmitterReadsCommandResponseWhenStatusHasMore(t *testing.T) {
	channel := &fakeSmartCardChannel{
		logicalChannel: 2,
		responses: [][]byte{
			{0x61, 0x02},
			{0xDE, 0xAD, 0x90, 0x00},
		},
	}
	tx, err := NewTransmitter(discardLogger(), channel, []byte{0xA0}, 254)
	if err != nil {
		t.Fatalf("NewTransmitter() error = %v", err)
	}

	got, err := tx.TransmitRaw([]byte{0x01, 0x02})
	if err != nil {
		t.Fatalf("TransmitRaw() error = %v", err)
	}
	if want := []byte{0xDE, 0xAD}; !bytes.Equal(got, want) {
		t.Fatalf("TransmitRaw() = % X, want % X", got, want)
	}

	want := [][]byte{
		{0x82, 0xE2, 0x91, 0x00, 0x02, 0x01, 0x02},
		{0x82, 0xC0, 0x00, 0x00, 0x02},
	}
	if len(channel.requests) != len(want) {
		t.Fatalf("Transmit() request count = %d, want %d", len(channel.requests), len(want))
	}
	for i := range want {
		if !bytes.Equal(channel.requests[i], want[i]) {
			t.Fatalf("request %d = % X, want % X", i, channel.requests[i], want[i])
		}
	}
}

func TestTransmitterReturnsErrorForUnexpectedStatus(t *testing.T) {
	channel := &fakeSmartCardChannel{
		logicalChannel: 1,
		responses:      [][]byte{{0x6A, 0x82}},
	}
	tx, err := NewTransmitter(discardLogger(), channel, []byte{0xA0}, 254)
	if err != nil {
		t.Fatalf("NewTransmitter() error = %v", err)
	}

	_, err = tx.TransmitRaw([]byte{0x01})
	if err == nil {
		t.Fatal("TransmitRaw() error = nil, want unexpected status error")
	}
	if !strings.Contains(err.Error(), "6A82") {
		t.Fatalf("TransmitRaw() error = %q, want status 6A82", err.Error())
	}
}
