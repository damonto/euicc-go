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
	connectErr     error
	openErr        error
	closeErr       error
	disconnectErr  error
	connected      bool
	disconnected   bool
	openedAID      []byte
	closedChannel  byte
}

func (f *fakeSmartCardChannel) Connect() error {
	f.connected = true
	return f.connectErr
}

func (f *fakeSmartCardChannel) Disconnect() error {
	f.disconnected = true
	return f.disconnectErr
}

func (f *fakeSmartCardChannel) OpenLogicalChannel(AID []byte) (byte, error) {
	f.openedAID = append([]byte(nil), AID...)
	return f.logicalChannel, f.openErr
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
	return f.closeErr
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

func TestNewTransmitterValidatesBeforeConnecting(t *testing.T) {
	if _, err := NewTransmitter(discardLogger(), nil, nil, 254); err == nil {
		t.Fatal("NewTransmitter() error = nil for nil channel")
	}

	for _, MSS := range []int{-1, 0, 255} {
		channel := &fakeSmartCardChannel{logicalChannel: 1}
		if _, err := NewTransmitter(discardLogger(), channel, nil, MSS); err == nil {
			t.Fatalf("NewTransmitter() error = nil for MSS %d", MSS)
		}
		if channel.connected {
			t.Fatalf("NewTransmitter() connected channel for invalid MSS %d", MSS)
		}
	}
}

func TestNewTransmitterCleansUpFailedInitialization(t *testing.T) {
	connectErr := errors.New("connect")
	channel := &fakeSmartCardChannel{connectErr: connectErr}
	if _, err := NewTransmitter(discardLogger(), channel, nil, 254); !errors.Is(err, connectErr) {
		t.Fatalf("NewTransmitter() error = %v, want connect error", err)
	}
	if !channel.disconnected {
		t.Fatal("NewTransmitter() did not disconnect after Connect failure")
	}

	openErr := errors.New("open")
	channel = &fakeSmartCardChannel{logicalChannel: 3, openErr: openErr}
	if _, err := NewTransmitter(discardLogger(), channel, nil, 254); !errors.Is(err, openErr) {
		t.Fatalf("NewTransmitter() error = %v, want open error", err)
	}
	if channel.closedChannel != 0 || !channel.disconnected {
		t.Fatalf("NewTransmitter() cleanup = close %d, disconnect %t; want disconnect only", channel.closedChannel, channel.disconnected)
	}
}

func TestNewTransmitterRejectsInvalidLogicalChannelAndDisconnects(t *testing.T) {
	channel := &fakeSmartCardChannel{logicalChannel: 20}
	_, err := NewTransmitter(discardLogger(), channel, nil, 254)
	if err == nil || !strings.Contains(err.Error(), "outside 1..19") {
		t.Fatalf("NewTransmitter() error = %v, want invalid logical channel", err)
	}
	if channel.closedChannel != 20 || !channel.disconnected {
		t.Fatalf("NewTransmitter() cleanup = close %d, disconnect %t; want close 20 and disconnect", channel.closedChannel, channel.disconnected)
	}
}

func TestTransmitterCloseAlwaysDisconnects(t *testing.T) {
	closeErr := errors.New("close")
	disconnectErr := errors.New("disconnect")
	channel := &fakeSmartCardChannel{
		logicalChannel: 2,
		closeErr:       closeErr,
		disconnectErr:  disconnectErr,
	}
	tx, err := NewTransmitter(discardLogger(), channel, nil, 254)
	if err != nil {
		t.Fatalf("NewTransmitter() error = %v", err)
	}
	err = tx.Close()
	if !errors.Is(err, closeErr) || !errors.Is(err, disconnectErr) {
		t.Fatalf("Close() error = %v, want close and disconnect errors", err)
	}
	if !channel.disconnected {
		t.Fatal("Close() did not disconnect after logical channel close failure")
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

func TestTransmitterMarksFinalBlockWhenCommandIsExactMultipleOfMSS(t *testing.T) {
	channel := &fakeSmartCardChannel{
		logicalChannel: 1,
		responses:      [][]byte{{0x90, 0x00}, {0x90, 0x00}},
	}
	tx, err := NewTransmitter(discardLogger(), channel, nil, 3)
	if err != nil {
		t.Fatalf("NewTransmitter() error = %v", err)
	}

	if _, err := tx.TransmitRaw([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}); err != nil {
		t.Fatalf("TransmitRaw() error = %v", err)
	}
	want := [][]byte{
		{0x81, 0xE2, 0x11, 0x00, 0x03, 0x01, 0x02, 0x03},
		{0x81, 0xE2, 0x91, 0x01, 0x03, 0x04, 0x05, 0x06},
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

func TestTransmitterRejectsMoreThan256StoreDataBlocks(t *testing.T) {
	channel := &fakeSmartCardChannel{logicalChannel: 1}
	tx, err := NewTransmitter(discardLogger(), channel, nil, 1)
	if err != nil {
		t.Fatalf("NewTransmitter() error = %v", err)
	}

	_, err = tx.TransmitRaw(make([]byte, maxStoreDataBlocks+1))
	if err == nil || !strings.Contains(err.Error(), "maximum is 256") {
		t.Fatalf("TransmitRaw() error = %v, want block count error", err)
	}
	if len(channel.requests) != 0 {
		t.Fatalf("TransmitRaw() sent %d requests before rejecting block count", len(channel.requests))
	}
}

func TestTransmitterLogsRawAPDUBodiesAtDebugLevel(t *testing.T) {
	var logs bytes.Buffer
	channel := &fakeSmartCardChannel{
		logicalChannel: 1,
		responses: [][]byte{
			{0x61, 0x02},
			{0xBE, 0xEF, 0x90, 0x00},
		},
	}
	tx, err := NewTransmitter(debugLogger(&logs), channel, nil, 254)
	if err != nil {
		t.Fatalf("NewTransmitter() error = %v", err)
	}
	if _, err := tx.TransmitRaw([]byte{0xDE, 0xAD}); err != nil {
		t.Fatalf("TransmitRaw() error = %v", err)
	}
	output := logs.String()
	for _, raw := range []string{
		"command=81E2910002DEAD",
		"response=6102",
		"command=81C0000002",
		"response=BEEF9000",
	} {
		if !strings.Contains(output, raw) {
			t.Fatalf("debug logs do not contain %q: %s", raw, output)
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
	var logs bytes.Buffer
	channel := &fakeSmartCardChannel{
		logicalChannel: 1,
		responses:      [][]byte{{0x6A, 0x82}},
	}
	tx, err := NewTransmitter(debugLogger(&logs), channel, []byte{0xA0}, 254)
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
	if output := logs.String(); !strings.Contains(output, "response=6A82") {
		t.Fatalf("debug logs do not contain raw error response: %s", output)
	}
}
