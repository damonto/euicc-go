package at

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

type scriptedSerial struct {
	readBuf strings.Reader
	writes  []string
}

func newScriptedSerial(script string) *scriptedSerial {
	return &scriptedSerial{readBuf: *strings.NewReader(script)}
}

func (s *scriptedSerial) Read(p []byte) (int, error) {
	return s.readBuf.Read(p)
}

func (s *scriptedSerial) Write(p []byte) (int, error) {
	s.writes = append(s.writes, string(p))
	return len(p), nil
}

func (s *scriptedSerial) Close() error { return nil }

type partialWriteSerial struct {
	readBuf strings.Reader
	writes  bytes.Buffer
	max     int
}

func newPartialWriteSerial(script string, max int) *partialWriteSerial {
	return &partialWriteSerial{readBuf: *strings.NewReader(script), max: max}
}

func (s *partialWriteSerial) Read(p []byte) (int, error) {
	return s.readBuf.Read(p)
}

func (s *partialWriteSerial) Write(p []byte) (int, error) {
	if len(p) > s.max {
		p = p[:s.max]
	}
	s.writes.Write(p)
	return len(p), nil
}

func (s *partialWriteSerial) Close() error { return nil }

func TestSendCommandKeepsBufferedDataAcrossCalls(t *testing.T) {
	serial := newScriptedSerial("\r\nAT+CSIM=?\r\n\r\nOK\r\nAT+CSIM=4,\"00\"\r\n+CSIM: 4,\"9000\"\r\nOK\r\n")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	got, err := at.sendCommand("AT+CSIM=?")
	if err != nil {
		t.Fatalf("first sendCommand failed: %v", err)
	}
	if got != "" {
		t.Fatalf("first sendCommand = %q, want empty response", got)
	}

	got, err = at.sendCommand(`AT+CSIM=4,"00"`)
	if err != nil {
		t.Fatalf("second sendCommand failed: %v", err)
	}
	if got != `+CSIM: 4,"9000"` {
		t.Fatalf("second sendCommand = %q, want %q", got, `+CSIM: 4,"9000"`)
	}

	if len(serial.writes) != 2 {
		t.Fatalf("writes = %d, want 2", len(serial.writes))
	}
	if serial.writes[0] != "AT+CSIM=?\r\n" {
		t.Fatalf("first write = %q", serial.writes[0])
	}
	if serial.writes[1] != "AT+CSIM=4,\"00\"\r\n" {
		t.Fatalf("second write = %q", serial.writes[1])
	}
}

func TestSendCommandWritesCompleteCommand(t *testing.T) {
	serial := newPartialWriteSerial("AT+TEST\r\nOK\r\n", 3)
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	if _, err := at.sendCommand("AT+TEST"); err != nil {
		t.Fatalf("sendCommand failed: %v", err)
	}
	if got := serial.writes.String(); got != "AT+TEST\r\n" {
		t.Fatalf("write = %q, want %q", got, "AT+TEST\r\n")
	}
}

func TestSendCommandMatchesCompleteResultCodes(t *testing.T) {
	serial := newScriptedSerial("AT+TEST\r\n+INFO: LOOKUP\r\nOK\r\n")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	got, err := at.sendCommand("AT+TEST")
	if err != nil {
		t.Fatalf("sendCommand failed: %v", err)
	}
	if got != "+INFO: LOOKUP" {
		t.Fatalf("sendCommand = %q, want %q", got, "+INFO: LOOKUP")
	}
}

func TestSendCommandReturnsStructuredATErrors(t *testing.T) {
	serial := newScriptedSerial("AT+TEST\r\n+CME ERROR: 42\r\n")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	_, err := at.sendCommand("AT+TEST")
	if err == nil {
		t.Fatal("sendCommand error = nil, want error")
	}
	if err.Error() != "+CME ERROR: 42" {
		t.Fatalf("sendCommand error = %q, want %q", err.Error(), "+CME ERROR: 42")
	}
}

func TestTransmitReturnsNonOKStatusWord(t *testing.T) {
	serial := newScriptedSerial("AT+CSIM=4,\"00\"\r\n+CSIM: 4,\"6A82\"\r\nOK\r\n")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	got, err := at.Transmit([]byte{0x00})
	if err != nil {
		t.Fatalf("Transmit failed: %v", err)
	}
	if !bytes.Equal(got, []byte{0x6A, 0x82}) {
		t.Fatalf("Transmit = % X, want 6A 82", got)
	}
}

func TestCloseLogicalChannelReturnsStatusError(t *testing.T) {
	serial := newScriptedSerial("AT+CSIM=10,\"0070800100\"\r\n+CSIM: 4,\"6A86\"\r\nOK\r\n")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	err := at.CloseLogicalChannel(1)
	if err == nil {
		t.Fatal("CloseLogicalChannel error = nil, want error")
	}
	if err.Error() != "close logical channel: 6A86" {
		t.Fatalf("CloseLogicalChannel error = %q, want %q", err.Error(), "close logical channel: 6A86")
	}
}

func TestCloseLogicalChannelRequiresExactSuccessStatus(t *testing.T) {
	serial := newScriptedSerial("AT+CSIM=10,\"0070800100\"\r\n+CSIM: 4,\"9001\"\r\nOK\r\n")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	err := at.CloseLogicalChannel(1)
	if err == nil {
		t.Fatal("CloseLogicalChannel error = nil, want error")
	}
	if err.Error() != "close logical channel: 9001" {
		t.Fatalf("CloseLogicalChannel error = %q, want %q", err.Error(), "close logical channel: 9001")
	}
}

func TestCloseLogicalChannelRejectsInvalidChannel(t *testing.T) {
	serial := newScriptedSerial("")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	err := at.CloseLogicalChannel(0)
	if err == nil {
		t.Fatal("CloseLogicalChannel error = nil, want invalid channel error")
	}
	if err.Error() != "invalid logical channel 0" {
		t.Fatalf("CloseLogicalChannel error = %q, want invalid channel error", err.Error())
	}
	if len(serial.writes) != 0 {
		t.Fatalf("writes = %d, want no command sent", len(serial.writes))
	}
}

func TestOpenLogicalChannelClosesChannelWhenSelectFails(t *testing.T) {
	serial := newScriptedSerial(
		"AT+CSIM=10,\"0070000001\"\r\n+CSIM: 6,\"019000\"\r\nOK\r\n" +
			"AT+CSIM=14,\"01A4040002A000\"\r\n+CSIM: 4,\"6A82\"\r\nOK\r\n" +
			"AT+CSIM=10,\"0070800100\"\r\n+CSIM: 4,\"9000\"\r\nOK\r\n")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	_, err := at.OpenLogicalChannel([]byte{0xA0, 0x00})
	if err == nil {
		t.Fatal("OpenLogicalChannel error = nil, want error")
	}
	if !strings.Contains(err.Error(), "select AID: 6A82") {
		t.Fatalf("OpenLogicalChannel error = %q, want select failure", err.Error())
	}
	if len(serial.writes) != 3 {
		t.Fatalf("writes = %d, want 3", len(serial.writes))
	}
	if serial.writes[2] != "AT+CSIM=10,\"0070800100\"\r\n" {
		t.Fatalf("close write = %q, want close channel command", serial.writes[2])
	}
}

func TestOpenLogicalChannelUsesExtendedClassByte(t *testing.T) {
	serial := newScriptedSerial(
		"AT+CSIM=10,\"0070000001\"\r\n+CSIM: 6,\"049000\"\r\nOK\r\n" +
			"AT+CSIM=14,\"40A4040002A000\"\r\n+CSIM: 4,\"9000\"\r\nOK\r\n")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	channel, err := at.OpenLogicalChannel([]byte{0xA0, 0x00})
	if err != nil {
		t.Fatalf("OpenLogicalChannel failed: %v", err)
	}
	if channel != 4 {
		t.Fatalf("OpenLogicalChannel = %d, want 4", channel)
	}
	if len(serial.writes) != 2 {
		t.Fatalf("writes = %d, want 2", len(serial.writes))
	}
	if serial.writes[1] != "AT+CSIM=14,\"40A4040002A000\"\r\n" {
		t.Fatalf("select write = %q, want extended channel CLA", serial.writes[1])
	}
}

func TestOpenLogicalChannelRejectsInvalidChannel(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     string
	}{
		{
			name:     "zero",
			response: "009000",
			want:     "open logical channel returned invalid logical channel 0",
		},
		{
			name:     "too high",
			response: "149000",
			want:     "open logical channel returned invalid logical channel 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serial := newScriptedSerial("AT+CSIM=10,\"0070000001\"\r\n+CSIM: 6,\"" + tt.response + "\"\r\nOK\r\n")
			at := &AT{
				s:      serial,
				reader: bufio.NewReader(serial),
			}

			_, err := at.OpenLogicalChannel([]byte{0xA0, 0x00})
			if err == nil {
				t.Fatal("OpenLogicalChannel error = nil, want error")
			}
			if err.Error() != tt.want {
				t.Fatalf("OpenLogicalChannel error = %q, want %q", err.Error(), tt.want)
			}
			if len(serial.writes) != 1 {
				t.Fatalf("writes = %d, want only open channel command", len(serial.writes))
			}
		})
	}
}

func TestOpenLogicalChannelRejectsOversizedAID(t *testing.T) {
	serial := newScriptedSerial("")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	_, err := at.OpenLogicalChannel(make([]byte, 256))
	if err == nil {
		t.Fatal("OpenLogicalChannel error = nil, want error")
	}
	if err.Error() != "AID length 256 exceeds short APDU limit" {
		t.Fatalf("OpenLogicalChannel error = %q, want oversized AID error", err.Error())
	}
	if len(serial.writes) != 0 {
		t.Fatalf("writes = %d, want no command sent", len(serial.writes))
	}
}

var _ io.ReadWriteCloser = (*scriptedSerial)(nil)
var _ io.ReadWriteCloser = (*partialWriteSerial)(nil)
