package mbim

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	wwanmbim "github.com/damonto/wwan-go/mbim"
)

type fakeMBIMReader struct {
	openChannel    uint32
	openCalls      int
	openContextErr error
	response       []byte
	status         uint32
	closedChannel  []uint32
	closed         bool
}

func (f *fakeMBIMReader) OpenChannel(ctx context.Context, _ []byte) (uint32, error) {
	f.openCalls++
	f.openContextErr = ctx.Err()
	return f.openChannel, nil
}

func (f *fakeMBIMReader) TransmitAPDU(context.Context, uint32, []byte) ([]byte, uint32, error) {
	return append([]byte(nil), f.response...), f.status, nil
}

func (f *fakeMBIMReader) CloseChannel(_ context.Context, channel uint32) error {
	f.closedChannel = append(f.closedChannel, channel)
	return nil
}

func (f *fakeMBIMReader) Close() error {
	f.closed = true
	return nil
}

func TestNewRejectsInvalidSlot(t *testing.T) {
	if _, err := New(WithAutoDetect("/dev/cdc-wdm1"), WithSlot(0)); err == nil {
		t.Fatal("New() error = nil, want invalid slot error")
	}
}

func TestNewValidatesOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
	}{
		{name: "missing access method"},
		{name: "missing device", options: []Option{WithDirect("")}},
		{name: "client with access", options: []Option{WithClient(new(wwanmbim.Client)), WithDirect("/dev/cdc-wdm1")}},
		{name: "client with slot", options: []Option{WithClient(new(wwanmbim.Client)), WithSlot(2)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New(tt.options...); err == nil {
				t.Fatal("New() error = nil, want invalid options error")
			}
		})
	}
}

func TestNewAcceptsNonPositiveTimeout(t *testing.T) {
	for _, timeout := range []time.Duration{0, -time.Second} {
		channel, err := New(WithClient(new(wwanmbim.Client)), WithTimeout(timeout))
		if err != nil {
			t.Fatalf("New(WithTimeout(%s)) error = %v", timeout, err)
		}
		if got := channel.timeout; got != timeout {
			t.Fatalf("New(WithTimeout(%s)) timeout = %s", timeout, got)
		}
	}
}

func TestNewUsesClient(t *testing.T) {
	tests := []struct {
		name    string
		client  *wwanmbim.Client
		wantErr bool
	}{
		{name: "nil client", wantErr: true},
		{name: "connected client", client: new(wwanmbim.Client)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel, err := New(WithClient(tt.client), WithTimeout(time.Second))
			if (err != nil) != tt.wantErr {
				t.Fatalf("New() error = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if channel.reader != tt.client {
				t.Fatal("New() did not retain the injected MBIM client")
			}
			if channel.timeout != time.Second {
				t.Fatalf("New() timeout = %s, want %s", channel.timeout, time.Second)
			}
			if err := channel.Connect(); err != nil {
				t.Fatalf("Connect() error = %v", err)
			}
		})
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

func TestOpenLogicalChannelRejectsInvalidChannel(t *testing.T) {
	for _, logicalChannel := range []uint32{0, 20, 256} {
		fake := &fakeMBIMReader{openChannel: logicalChannel}
		m := &MBIM{reader: fake}

		if _, err := m.OpenLogicalChannel(nil); err == nil {
			t.Fatalf("OpenLogicalChannel() error = nil for channel %d", logicalChannel)
		}
		if logicalChannel == 0 && len(fake.closedChannel) != 0 {
			t.Fatal("OpenLogicalChannel() tried to close channel 0")
		}
		if logicalChannel != 0 && (len(fake.closedChannel) != 1 || fake.closedChannel[0] != logicalChannel) {
			t.Fatalf("OpenLogicalChannel() cleanup = %v, want [%d]", fake.closedChannel, logicalChannel)
		}
	}
}

func TestOpenLogicalChannelRejectsSecondChannel(t *testing.T) {
	fake := &fakeMBIMReader{openChannel: 3}
	m := &MBIM{reader: fake, timeout: defaultTimeout}
	if _, err := m.OpenLogicalChannel(nil); err != nil {
		t.Fatalf("first OpenLogicalChannel() error = %v", err)
	}
	if _, err := m.OpenLogicalChannel(nil); err == nil {
		t.Fatal("second OpenLogicalChannel() error = nil")
	}
	if fake.openCalls != 1 {
		t.Fatalf("OpenChannel() calls = %d, want 1", fake.openCalls)
	}
}

func TestNonPositiveTimeoutExpiresContextImmediately(t *testing.T) {
	fake := &fakeMBIMReader{openChannel: 3}
	m := &MBIM{reader: fake, timeout: 0}
	if _, err := m.OpenLogicalChannel(nil); err != nil {
		t.Fatalf("OpenLogicalChannel() error = %v", err)
	}
	if !errors.Is(fake.openContextErr, context.DeadlineExceeded) {
		t.Fatalf("OpenChannel() context error = %v, want deadline exceeded", fake.openContextErr)
	}
}

func TestDisconnectClosesLogicalChannelAndReader(t *testing.T) {
	fake := &fakeMBIMReader{}
	m := &MBIM{reader: fake, channel: 3}

	if err := m.Disconnect(); err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}
	if len(fake.closedChannel) != 1 || fake.closedChannel[0] != 3 {
		t.Fatalf("Disconnect() closed channels = %v, want [3]", fake.closedChannel)
	}
	if !fake.closed {
		t.Fatal("Disconnect() did not close reader")
	}
}
