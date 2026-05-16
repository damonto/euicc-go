package ccid

import (
	"bytes"
	"testing"
)

func TestClassByteForChannel(t *testing.T) {
	tests := []struct {
		name    string
		cla     byte
		channel byte
		want    byte
		wantErr bool
	}{
		{name: "basic channel", cla: 0x00, channel: 0, want: 0x00},
		{name: "channel one", cla: 0x00, channel: 1, want: 0x01},
		{name: "preserves proprietary class", cla: 0x80, channel: 2, want: 0x82},
		{name: "first extended channel", cla: 0x00, channel: 4, want: 0x40},
		{name: "last extended channel", cla: 0x00, channel: 19, want: 0x4F},
		{name: "extended proprietary class", cla: 0x80, channel: 4, want: 0xC0},
		{name: "too high", cla: 0x00, channel: 20, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := classByteForChannel(tt.cla, tt.channel)
			if tt.wantErr {
				if err == nil {
					t.Fatal("classByteForChannel error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("classByteForChannel failed: %v", err)
			}
			if got != tt.want {
				t.Fatalf("classByteForChannel = %02X, want %02X", got, tt.want)
			}
		})
	}
}

func TestCloseLogicalChannelRejectsInvalidChannel(t *testing.T) {
	reader := &CCIDReader{}

	if err := reader.closeLogicalChannel(0); err == nil {
		t.Fatal("closeLogicalChannel(0) error = nil, want invalid channel error")
	}
	if err := reader.closeLogicalChannel(maxLogicalChannel + 1); err == nil {
		t.Fatal("closeLogicalChannel(max+1) error = nil, want invalid channel error")
	}
}

func TestSetReaderReturnsStateErrors(t *testing.T) {
	reader := &CCIDReader{closed: true}
	if err := reader.SetReader("reader"); err == nil {
		t.Fatal("SetReader error = nil, want closed reader error")
	}

	reader = &CCIDReader{open: true}
	if err := reader.SetReader("reader"); err == nil {
		t.Fatal("SetReader error = nil, want connected reader error")
	}
}

func TestSelectAIDCommandRejectsOversizedAID(t *testing.T) {
	_, err := selectAIDCommand(1, make([]byte, 256))
	if err == nil {
		t.Fatal("selectAIDCommand error = nil, want oversized AID error")
	}
	if err.Error() != "AID length 256 exceeds short APDU limit" {
		t.Fatalf("selectAIDCommand error = %q, want oversized AID error", err.Error())
	}
}

func TestSelectAIDCommand(t *testing.T) {
	aid := []byte{0xA0, 0x00, 0x00, 0x05, 0x59}
	got, err := selectAIDCommand(4, aid)
	if err != nil {
		t.Fatalf("selectAIDCommand failed: %v", err)
	}
	want := []byte{0x40, 0xA4, 0x04, 0x00, byte(len(aid)), 0xA0, 0x00, 0x00, 0x05, 0x59}
	if !bytes.Equal(got, want) {
		t.Fatalf("selectAIDCommand = % X, want % X", got, want)
	}
}

func TestStatusWords(t *testing.T) {
	if !statusOK([]byte{0xDE, 0xAD, 0x90, 0x00}) {
		t.Fatal("statusOK returned false for 9000")
	}
	if statusOK([]byte{0x90}) {
		t.Fatal("statusOK returned true for short response")
	}
	if !statusHasMore([]byte{0x61, 0x10}) {
		t.Fatal("statusHasMore returned false for 61xx")
	}
	if statusHasMore([]byte{0x90, 0x00}) {
		t.Fatal("statusHasMore returned true for 9000")
	}
}
