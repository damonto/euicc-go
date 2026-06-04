package ccid

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeReader struct {
	requests  [][]byte
	responses [][]byte
	closed    bool
}

func (f *fakeReader) Transmit(_ context.Context, req []byte) ([]byte, error) {
	f.requests = append(f.requests, append([]byte(nil), req...))
	if len(f.responses) == 0 {
		return nil, errors.New("unexpected transmit")
	}
	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func (f *fakeReader) Close() error {
	f.closed = true
	return nil
}

func TestSetReaderReturnsStateErrors(t *testing.T) {
	reader := &CCIDReader{closed: true}
	if err := reader.SetReader("reader"); err == nil {
		t.Fatal("SetReader() error = nil, want closed reader error")
	}

	reader = &CCIDReader{open: true}
	if err := reader.SetReader("reader"); err == nil {
		t.Fatal("SetReader() error = nil, want connected reader error")
	}
}

func TestNewWithReaderOpensLazilyOnConnect(t *testing.T) {
	oldOpenReader := openReader
	defer func() { openReader = oldOpenReader }()

	called := false
	fake := &fakeReader{responses: [][]byte{{0x90, 0x00}}}
	openReader = func(_ context.Context, reader string) (transmitter, error) {
		called = true
		if reader != "reader-1" {
			t.Fatalf("openReader reader = %q, want reader-1", reader)
		}
		return fake, nil
	}

	reader, err := NewWithReader("reader-1")
	if err != nil {
		t.Fatalf("NewWithReader() error = %v", err)
	}
	if called {
		t.Fatal("NewWithReader() called opener before Connect()")
	}
	if err := reader.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if !called {
		t.Fatal("Connect() did not call opener")
	}
	if err := reader.Disconnect(); err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}
	if !fake.closed {
		t.Fatal("Disconnect() did not close reader")
	}
}

func TestChannelOpenLogicalChannelSelectsAID(t *testing.T) {
	fake := &fakeReader{responses: [][]byte{
		{0x90, 0x00},
		{0x04, 0x90, 0x00},
		{0x61, 0x10},
	}}
	reader := &CCIDReader{reader: "reader-1"}
	oldOpenReader := openReader
	defer func() { openReader = oldOpenReader }()
	openReader = func(context.Context, string) (transmitter, error) {
		return fake, nil
	}

	if err := reader.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	got, err := reader.OpenLogicalChannel([]byte{0xA0, 0x00})
	if err != nil {
		t.Fatalf("OpenLogicalChannel() error = %v", err)
	}
	if got != 4 {
		t.Fatalf("OpenLogicalChannel() = %d, want 4", got)
	}
	wantSelect := []byte{0x40, 0xA4, 0x04, 0x00, 0x02, 0xA0, 0x00}
	if !bytes.Equal(fake.requests[2], wantSelect) {
		t.Fatalf("select request = % X, want % X", fake.requests[2], wantSelect)
	}
}

func TestChannelOpenLogicalChannelClosesWhenSelectFails(t *testing.T) {
	fake := &fakeReader{responses: [][]byte{
		{0x90, 0x00},
		{0x01, 0x90, 0x00},
		{0x6A, 0x82},
		{0x90, 0x00},
	}}
	reader := &CCIDReader{reader: "reader-1"}
	oldOpenReader := openReader
	defer func() { openReader = oldOpenReader }()
	openReader = func(context.Context, string) (transmitter, error) {
		return fake, nil
	}

	if err := reader.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	_, err := reader.OpenLogicalChannel([]byte{0xA0, 0x00})
	if err == nil {
		t.Fatal("OpenLogicalChannel() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "select AID: 6A82") {
		t.Fatalf("OpenLogicalChannel() error = %q, want select failure", err.Error())
	}
	wantClose := []byte{0x00, 0x70, 0x80, 0x01, 0x00}
	if !bytes.Equal(fake.requests[3], wantClose) {
		t.Fatalf("close request = % X, want % X", fake.requests[3], wantClose)
	}
}

func TestConnectRequiresReader(t *testing.T) {
	reader, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = reader.Connect()
	if err == nil {
		t.Fatal("Connect() error = nil, want reader required error")
	}
	if err.Error() != "ccid reader is required" {
		t.Fatalf("Connect() error = %q, want reader required", err.Error())
	}
}

func TestListReadersUsesInjectedLister(t *testing.T) {
	oldListReaders := listReaders
	defer func() { listReaders = oldListReaders }()

	listReaders = func(context.Context) ([]string, error) {
		return []string{"reader-1"}, nil
	}
	reader, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	got, err := reader.ListReaders()
	if err != nil {
		t.Fatalf("ListReaders() error = %v", err)
	}
	if len(got) != 1 || got[0] != "reader-1" {
		t.Fatalf("ListReaders() = %v, want [reader-1]", got)
	}
}
