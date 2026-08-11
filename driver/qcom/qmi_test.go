package qcom

import (
	"context"
	"errors"
	"testing"
	"time"

	wwanqcom "github.com/damonto/wwan-go/qcom"
)

type fakeQMITransport struct {
	closed bool
}

func (*fakeQMITransport) Do(context.Context, wwanqcom.Request) (wwanqcom.Response, error) {
	return wwanqcom.Response{}, errors.New("unexpected QMI request")
}

func (f *fakeQMITransport) Close() error {
	f.closed = true
	return nil
}

type fakeUIMReader struct {
	openChannel        uint8
	openCalls          int
	activateCalls      int
	activateContextErr error
	closedChannel      []uint8
	closed             bool
}

func (f *fakeUIMReader) ActivateSlot(ctx context.Context) error {
	f.activateCalls++
	f.activateContextErr = ctx.Err()
	return nil
}
func (f *fakeUIMReader) OpenLogicalChannel(context.Context, []byte) (uint8, error) {
	f.openCalls++
	return f.openChannel, nil
}
func (f *fakeUIMReader) SendAPDU(context.Context, uint8, []byte) ([]byte, error) {
	return []byte{0x90, 0x00}, nil
}
func (f *fakeUIMReader) CloseLogicalChannel(_ context.Context, channel uint8) error {
	f.closedChannel = append(f.closedChannel, channel)
	return nil
}
func (f *fakeUIMReader) Close() error {
	f.closed = true
	return nil
}

func TestDisconnectClosesLogicalChannelAndReader(t *testing.T) {
	reader := &fakeUIMReader{}
	qcomChannel := newChannel(reader, defaultTimeout)
	qcomChannel.channel = 2
	q := &QMI{channel: qcomChannel}

	if err := q.Disconnect(); err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}
	if len(reader.closedChannel) != 1 || reader.closedChannel[0] != 2 {
		t.Fatalf("Disconnect() closed channels = %v, want [2]", reader.closedChannel)
	}
	if !reader.closed {
		t.Fatal("Disconnect() did not close reader")
	}
}

func TestOpenLogicalChannelRejectsInvalidChannel(t *testing.T) {
	for _, logicalChannel := range []uint8{0, 20} {
		reader := &fakeUIMReader{openChannel: logicalChannel}
		channel := newChannel(reader, defaultTimeout)
		channel.connected = true

		if _, err := channel.OpenLogicalChannel(nil); err == nil {
			t.Fatalf("OpenLogicalChannel() error = nil for channel %d", logicalChannel)
		}
		if logicalChannel == 0 && len(reader.closedChannel) != 0 {
			t.Fatal("OpenLogicalChannel() tried to close channel 0")
		}
		if logicalChannel != 0 && (len(reader.closedChannel) != 1 || reader.closedChannel[0] != logicalChannel) {
			t.Fatalf("OpenLogicalChannel() cleanup = %v, want [%d]", reader.closedChannel, logicalChannel)
		}
	}
}

func TestOpenLogicalChannelRejectsSecondChannel(t *testing.T) {
	reader := &fakeUIMReader{openChannel: 2}
	channel := newChannel(reader, defaultTimeout)
	channel.connected = true
	if _, err := channel.OpenLogicalChannel(nil); err != nil {
		t.Fatalf("first OpenLogicalChannel() error = %v", err)
	}
	if _, err := channel.OpenLogicalChannel(nil); err == nil {
		t.Fatal("second OpenLogicalChannel() error = nil")
	}
	if reader.openCalls != 1 {
		t.Fatalf("OpenLogicalChannel() calls = %d, want 1", reader.openCalls)
	}
}

func TestNewQMIUsesClient(t *testing.T) {
	tests := []struct {
		name    string
		client  func(t *testing.T) (*wwanqcom.Client, *fakeQMITransport)
		wantErr bool
	}{
		{
			name:    "nil client",
			client:  func(*testing.T) (*wwanqcom.Client, *fakeQMITransport) { return nil, nil },
			wantErr: true,
		},
		{
			name: "owned client",
			client: func(t *testing.T) (*wwanqcom.Client, *fakeQMITransport) {
				t.Helper()
				transport := new(fakeQMITransport)
				client, err := wwanqcom.NewClient(transport)
				if err != nil {
					t.Fatalf("NewClient() error = %v", err)
				}
				return client, transport
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, transport := tt.client(t)
			channel, err := NewQMI(WithClient(client), WithTimeout(time.Second))
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewQMI() error = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if channel.timeout != time.Second {
				t.Fatalf("NewQMI() timeout = %s, want %s", channel.timeout, time.Second)
			}
			if err := channel.Disconnect(); err != nil {
				t.Fatalf("Disconnect() error = %v", err)
			}
			if !transport.closed {
				t.Fatal("Disconnect() did not close the injected QMI client")
			}
		})
	}
}

func TestNewQMIRejectsInvalidSlotBeforeOpen(t *testing.T) {
	for _, slot := range []uint8{0, 6} {
		_, err := NewQMI(WithAutoDetect("/dev/cdc-wdm1"), WithSlot(slot))
		if err == nil {
			t.Fatalf("NewQMI() error = nil for slot %d, want invalid slot error", slot)
		}
		if err.Error() != "slot must be between 1 and 5" {
			t.Fatalf("NewQMI() error = %q, want local slot validation", err.Error())
		}
	}
}

func TestNewQRTRRejectsInvalidSlotBeforeOpen(t *testing.T) {
	for _, slot := range []uint8{0, 6} {
		_, err := NewQRTR(WithSlot(slot))
		if err == nil {
			t.Fatalf("NewQRTR() error = nil for slot %d, want invalid slot error", slot)
		}
		if err.Error() != "slot must be between 1 and 5" {
			t.Fatalf("NewQRTR() error = %q, want local slot validation", err.Error())
		}
	}
}

func TestNewQMIRequiresAccessMethod(t *testing.T) {
	if _, err := NewQMI(); err == nil {
		t.Fatal("NewQMI() error = nil, want access method error")
	}
}

func TestNewQMIRejectsClientConfigurationConflicts(t *testing.T) {
	client := new(wwanqcom.Client)
	for _, options := range [][]Option{
		{WithClient(client), WithDirect("/dev/cdc-wdm1")},
		{WithClient(client), WithSlot(2)},
	} {
		if _, err := NewQMI(options...); err == nil {
			t.Fatal("NewQMI() error = nil, want conflicting options error")
		}
	}
	if _, err := NewQRTR(WithClient(client)); err == nil {
		t.Fatal("NewQRTR() error = nil, want unsupported client option error")
	}
}

func TestNewAcceptsNonPositiveTimeout(t *testing.T) {
	for _, timeout := range []time.Duration{0, -time.Second} {
		channel, err := NewQMI(WithClient(new(wwanqcom.Client)), WithTimeout(timeout))
		if err != nil {
			t.Fatalf("NewQMI(WithTimeout(%s)) error = %v", timeout, err)
		}
		if got := channel.timeout; got != timeout {
			t.Fatalf("NewQMI(WithTimeout(%s)) timeout = %s", timeout, got)
		}
		config := applyOptions([]Option{WithTimeout(timeout)})
		if err := config.validateQRTR(); err != nil {
			t.Fatalf("NewQRTR options with timeout %s: %v", timeout, err)
		}
	}
}

func TestQCOMChannelConnectIsIdempotent(t *testing.T) {
	reader := new(fakeUIMReader)
	channel := newChannel(reader, defaultTimeout)
	if err := channel.Connect(); err != nil {
		t.Fatalf("first Connect() error = %v", err)
	}
	if err := channel.Connect(); err != nil {
		t.Fatalf("second Connect() error = %v", err)
	}
	if reader.activateCalls != 1 {
		t.Fatalf("ActivateSlot() calls = %d, want 1", reader.activateCalls)
	}
}

func TestQCOMChannelReleaseReaderResetsConnection(t *testing.T) {
	reader := new(fakeUIMReader)
	channel := newChannel(reader, defaultTimeout)
	channel.connected = true
	if err := channel.releaseReader(); err != nil {
		t.Fatalf("releaseReader() error = %v", err)
	}
	if !reader.closed {
		t.Fatal("releaseReader() did not close reader")
	}
	if channel.reader != nil || channel.connected {
		t.Fatal("releaseReader() retained connected reader state")
	}
}

func TestConstructorsDoNotOpenTransports(t *testing.T) {
	qmi, err := NewQMI(WithDirect("/does/not/exist"))
	if err != nil {
		t.Fatalf("NewQMI() error = %v", err)
	}
	if qmi.reader != nil || qmi.connected {
		t.Fatal("NewQMI() opened the transport")
	}
	if _, err := qmi.OpenLogicalChannel(nil); err == nil {
		t.Fatal("OpenLogicalChannel() before Connect error = nil")
	}
	if err := qmi.Disconnect(); err != nil {
		t.Fatalf("Disconnect() before Connect error = %v", err)
	}
	if err := qmi.Connect(); err == nil {
		t.Fatal("Connect() after Disconnect error = nil")
	}

	qrtr, err := NewQRTR()
	if err != nil {
		t.Fatalf("NewQRTR() error = %v", err)
	}
	if qrtr.reader != nil || qrtr.connected {
		t.Fatal("NewQRTR() opened the transport")
	}
}

func TestNonPositiveTimeoutExpiresContextImmediately(t *testing.T) {
	reader := new(fakeUIMReader)
	channel := newChannel(reader, 0)
	if err := channel.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if !errors.Is(reader.activateContextErr, context.DeadlineExceeded) {
		t.Fatalf("ActivateSlot() context error = %v, want deadline exceeded", reader.activateContextErr)
	}
}
