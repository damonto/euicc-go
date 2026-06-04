package mbim

import (
	"bytes"
	"context"
	"errors"
	"testing"

	uiccmbim "github.com/damonto/uicc-go/mbim"
)

type fakeMBIMReader struct {
	openChannel uint32
	response    []byte
	status      uint32
	closed      bool
}

func (f *fakeMBIMReader) OpenChannel(context.Context, []byte) (uint32, error) {
	return f.openChannel, nil
}

func (f *fakeMBIMReader) TransmitAPDU(context.Context, uint32, []byte) ([]byte, uint32, error) {
	return append([]byte(nil), f.response...), f.status, nil
}

func (f *fakeMBIMReader) CloseChannel(context.Context, uint32) error { return nil }

func (f *fakeMBIMReader) Close() error {
	f.closed = true
	return nil
}

func TestNewRejectsInvalidSlot(t *testing.T) {
	if _, err := New("/dev/cdc-wdm1", 0); err == nil {
		t.Fatal("New() error = nil, want invalid slot error")
	}
}

func TestConnectOpensReaderLazily(t *testing.T) {
	oldOpenReader := openReader
	defer func() { openReader = oldOpenReader }()

	called := false
	fake := &fakeMBIMReader{}
	openReader = func(_ context.Context, _ ...uiccmbim.Option) (reader, error) {
		called = true
		return fake, nil
	}

	ch, err := New("/dev/cdc-wdm1", 1)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if called {
		t.Fatal("New() called opener before Connect()")
	}
	if err := ch.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if !called {
		t.Fatal("Connect() did not call opener")
	}
}

func TestConnectReturnsOpenError(t *testing.T) {
	oldOpenReader := openReader
	defer func() { openReader = oldOpenReader }()

	openErr := errors.New("open")
	openReader = func(_ context.Context, _ ...uiccmbim.Option) (reader, error) {
		return nil, openErr
	}

	ch, err := New("/dev/cdc-wdm1", 1)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := ch.Connect(); !errors.Is(err, openErr) {
		t.Fatalf("Connect() error = %v, want open error", err)
	}
}

func TestTransmitAppendsAPDUStatusWord(t *testing.T) {
	fake := &fakeMBIMReader{
		openChannel: 3,
		response:    []byte{0xDE, 0xAD},
		status:      0x0090,
	}
	m := &MBIM{reader: fake}
	if _, err := m.OpenLogicalChannel([]byte{0xA0, 0x00}); err != nil {
		t.Fatalf("OpenLogicalChannel() error = %v", err)
	}

	got, err := m.Transmit([]byte{0x80, 0xE2})
	if err != nil {
		t.Fatalf("Transmit() error = %v", err)
	}
	want := []byte{0xDE, 0xAD, 0x90, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("Transmit() = % X, want % X", got, want)
	}
}

func TestDisconnectClosesReader(t *testing.T) {
	fake := &fakeMBIMReader{}
	m := &MBIM{reader: fake}

	if err := m.Disconnect(); err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}
	if !fake.closed {
		t.Fatal("Disconnect() did not close reader")
	}
}
