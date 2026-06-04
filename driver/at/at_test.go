package at

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeTransmitter struct {
	requests  [][]byte
	responses [][]byte
	closed    bool
}

func (f *fakeTransmitter) Transmit(_ context.Context, req []byte) ([]byte, error) {
	f.requests = append(f.requests, append([]byte(nil), req...))
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

func TestNewRejectsEmptyDevice(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("New() error = nil, want empty device error")
	}
}

func TestChannelOpenLogicalChannelSelectsAID(t *testing.T) {
	fake := &fakeTransmitter{responses: [][]byte{
		{0x04, 0x90, 0x00},
		{0x61, 0x10},
	}}
	channel := newChannel(fake)

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

func TestChannelOpenLogicalChannelClosesWhenSelectFails(t *testing.T) {
	fake := &fakeTransmitter{responses: [][]byte{
		{0x01, 0x90, 0x00},
		{0x6A, 0x82},
		{0x90, 0x00},
	}}
	channel := newChannel(fake)

	_, err := channel.OpenLogicalChannel([]byte{0xA0, 0x00})
	if err == nil {
		t.Fatal("OpenLogicalChannel() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "select AID: 6A82") {
		t.Fatalf("OpenLogicalChannel() error = %q, want select failure", err.Error())
	}
	wantClose := []byte{0x00, 0x70, 0x80, 0x01, 0x00}
	if !bytes.Equal(fake.requests[2], wantClose) {
		t.Fatalf("close request = % X, want % X", fake.requests[2], wantClose)
	}
}

func TestChannelRejectsInvalidLogicalChannel(t *testing.T) {
	tests := []struct {
		name     string
		response []byte
		want     string
	}{
		{
			name:     "zero",
			response: []byte{0x00, 0x90, 0x00},
			want:     "open logical channel returned invalid logical channel 0",
		},
		{
			name:     "too high",
			response: []byte{0x14, 0x90, 0x00},
			want:     "open logical channel returned invalid logical channel 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeTransmitter{responses: [][]byte{tt.response}}
			channel := newChannel(fake)

			_, err := channel.OpenLogicalChannel([]byte{0xA0, 0x00})
			if err == nil {
				t.Fatal("OpenLogicalChannel() error = nil, want error")
			}
			if err.Error() != tt.want {
				t.Fatalf("OpenLogicalChannel() error = %q, want %q", err.Error(), tt.want)
			}
			if len(fake.requests) != 1 {
				t.Fatalf("requests = %d, want only open channel command", len(fake.requests))
			}
		})
	}
}

func TestChannelAPDUStatusHandling(t *testing.T) {
	fake := &fakeTransmitter{responses: [][]byte{{0x6A, 0x82}}}
	channel := newChannel(fake)

	got, err := channel.Transmit([]byte{0x00})
	if err != nil {
		t.Fatalf("Transmit() error = %v", err)
	}
	if !bytes.Equal(got, []byte{0x6A, 0x82}) {
		t.Fatalf("Transmit() = % X, want 6A 82", got)
	}

	err = channel.CloseLogicalChannel(0)
	if err == nil {
		t.Fatal("CloseLogicalChannel() error = nil, want invalid channel error")
	}
	if err.Error() != "invalid logical channel 0" {
		t.Fatalf("CloseLogicalChannel() error = %q, want invalid channel", err.Error())
	}
}
