package iso7816

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeTransmitter struct {
	requests  [][]byte
	responses [][]byte
	transmit  func(context.Context, []byte) ([]byte, error)
	closed    bool
}

func (f *fakeTransmitter) Transmit(ctx context.Context, request []byte) ([]byte, error) {
	f.requests = append(f.requests, bytes.Clone(request))
	if f.transmit != nil {
		return f.transmit(ctx, request)
	}
	if len(f.responses) == 0 {
		return nil, errors.New("unexpected transmit")
	}
	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func (f *fakeTransmitter) Close() error {
	f.closed = true
	return nil
}

func TestChannelConnectSendsInitializationAPDU(t *testing.T) {
	fake := &fakeTransmitter{responses: [][]byte{{0x90, 0x00}}}
	channel := NewChannel(fake)
	if err := channel.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if want := []byte(connectAPDU); !bytes.Equal(fake.requests[0], want) {
		t.Fatalf("Connect() request = % X, want % X", fake.requests[0], want)
	}
}

func TestChannelOpenLogicalChannelSelectsAID(t *testing.T) {
	fake := &fakeTransmitter{responses: [][]byte{
		{0x04, 0x90, 0x00},
		{0x61, 0x10},
	}}
	channel := NewChannel(fake)

	got, err := channel.OpenLogicalChannel([]byte{0xA0, 0x00})
	if err != nil {
		t.Fatalf("OpenLogicalChannel() error = %v", err)
	}
	if got != 4 {
		t.Fatalf("OpenLogicalChannel() = %d, want 4", got)
	}
	wantSelect := []byte{0x40, 0xA4, 0x04, 0x00, 0x02, 0xA0, 0x00}
	if !bytes.Equal(fake.requests[1], wantSelect) {
		t.Fatalf("select request = % X, want % X", fake.requests[1], wantSelect)
	}
}

func TestChannelOpenLogicalChannelRejectsSecondChannel(t *testing.T) {
	fake := &fakeTransmitter{responses: [][]byte{
		{0x04, 0x90, 0x00},
		{0x90, 0x00},
	}}
	channel := NewChannel(fake)
	if _, err := channel.OpenLogicalChannel([]byte{0xA0}); err != nil {
		t.Fatalf("first OpenLogicalChannel() error = %v", err)
	}
	if _, err := channel.OpenLogicalChannel([]byte{0xA0}); err == nil {
		t.Fatal("second OpenLogicalChannel() error = nil")
	}
	if len(fake.requests) != 2 {
		t.Fatalf("Transmit() calls = %d, want 2", len(fake.requests))
	}
}

func TestChannelOpenLogicalChannelClosesWhenSelectFails(t *testing.T) {
	fake := &fakeTransmitter{responses: [][]byte{
		{0x01, 0x90, 0x00},
		{0x6A, 0x82},
		{0x90, 0x00},
	}}
	channel := NewChannel(fake)

	_, err := channel.OpenLogicalChannel([]byte{0xA0, 0x00})
	if err == nil || !strings.Contains(err.Error(), "select AID: 6A82") {
		t.Fatalf("OpenLogicalChannel() error = %v, want select failure", err)
	}
	wantClose := []byte{0x00, 0x70, 0x80, 0x01, 0x00}
	if !bytes.Equal(fake.requests[2], wantClose) {
		t.Fatalf("close request = % X, want % X", fake.requests[2], wantClose)
	}
}

func TestChannelDisconnectRetriesFailedSelectCleanup(t *testing.T) {
	fake := &fakeTransmitter{responses: [][]byte{
		{0x01, 0x90, 0x00},
		{0x6A, 0x82},
		{0x6F, 0x00},
		{0x90, 0x00},
	}}
	channel := NewChannel(fake)
	if _, err := channel.OpenLogicalChannel([]byte{0xA0}); err == nil {
		t.Fatal("OpenLogicalChannel() error = nil")
	}
	if err := channel.Disconnect(); err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}
	if !fake.closed {
		t.Fatal("Disconnect() did not close transport")
	}
	wantClose := []byte{0x00, 0x70, 0x80, 0x01, 0x00}
	if !bytes.Equal(fake.requests[2], wantClose) || !bytes.Equal(fake.requests[3], wantClose) {
		t.Fatalf("close requests = % X and % X, want % X twice", fake.requests[2], fake.requests[3], wantClose)
	}
}

func TestChannelUsesFreshContextToCleanUpSelectTimeout(t *testing.T) {
	call := 0
	cleanupContextActive := false
	fake := &fakeTransmitter{transmit: func(ctx context.Context, _ []byte) ([]byte, error) {
		call++
		switch call {
		case 1:
			return []byte{0x01, 0x90, 0x00}, nil
		case 2:
			<-ctx.Done()
			return nil, ctx.Err()
		case 3:
			cleanupContextActive = ctx.Err() == nil
			return []byte{0x90, 0x00}, nil
		default:
			return nil, errors.New("unexpected transmit")
		}
	}}
	channel := NewChannel(fake, WithTimeout(20*time.Millisecond))

	_, err := channel.OpenLogicalChannel([]byte{0xA0, 0x00})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("OpenLogicalChannel() error = %v, want deadline exceeded", err)
	}
	if !cleanupContextActive {
		t.Fatal("OpenLogicalChannel() reused expired context for cleanup")
	}
}

func TestChannelNonPositiveTimeoutExpiresContextImmediately(t *testing.T) {
	fake := &fakeTransmitter{transmit: func(ctx context.Context, _ []byte) ([]byte, error) {
		return nil, ctx.Err()
	}}
	channel := NewChannel(fake, WithTimeout(0))
	if err := channel.Connect(); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Connect() error = %v, want deadline exceeded", err)
	}
}

func TestChannelRejectsInvalidLogicalChannel(t *testing.T) {
	tests := []struct {
		name     string
		response []byte
		want     string
	}{
		{name: "zero", response: []byte{0x00, 0x90, 0x00}, want: "open logical channel returned invalid logical channel 0"},
		{name: "too high", response: []byte{0x14, 0x90, 0x00}, want: "open logical channel returned invalid logical channel 20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeTransmitter{responses: [][]byte{tt.response}}
			channel := NewChannel(fake)
			_, err := channel.OpenLogicalChannel([]byte{0xA0, 0x00})
			if err == nil || err.Error() != tt.want {
				t.Fatalf("OpenLogicalChannel() error = %v, want %q", err, tt.want)
			}
			if len(fake.requests) != 1 {
				t.Fatalf("requests = %d, want only open channel command", len(fake.requests))
			}
		})
	}
}

func TestChannelDisconnectClosesLogicalChannelAndTransport(t *testing.T) {
	fake := &fakeTransmitter{responses: [][]byte{
		{0x01, 0x90, 0x00},
		{0x90, 0x00},
		{0x90, 0x00},
	}}
	channel := NewChannel(fake)
	if _, err := channel.OpenLogicalChannel([]byte{0xA0}); err != nil {
		t.Fatalf("OpenLogicalChannel() error = %v", err)
	}
	if err := channel.Disconnect(); err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}
	if !fake.closed {
		t.Fatal("Disconnect() did not close transport")
	}
	wantClose := []byte{0x00, 0x70, 0x80, 0x01, 0x00}
	if !bytes.Equal(fake.requests[2], wantClose) {
		t.Fatalf("disconnect close request = % X, want % X", fake.requests[2], wantClose)
	}
}

func TestChannelAPDUStatusHandling(t *testing.T) {
	fake := &fakeTransmitter{responses: [][]byte{{0x6A, 0x82}}}
	channel := NewChannel(fake)
	got, err := channel.Transmit([]byte{0x00})
	if err != nil {
		t.Fatalf("Transmit() error = %v", err)
	}
	if !bytes.Equal(got, []byte{0x6A, 0x82}) {
		t.Fatalf("Transmit() = % X, want 6A 82", got)
	}
	if err := channel.CloseLogicalChannel(0); err == nil || err.Error() != "invalid logical channel 0" {
		t.Fatalf("CloseLogicalChannel() error = %v, want invalid channel", err)
	}
}
