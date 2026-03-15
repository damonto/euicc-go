package at

import (
	"bufio"
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

func TestRunKeepsBufferedDataAcrossCalls(t *testing.T) {
	serial := newScriptedSerial("\r\nAT+CSIM=?\r\n\r\nOK\r\nAT+CSIM=4,\"00\"\r\n+CSIM: 4,\"9000\"\r\nOK\r\n")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	got, err := at.run("AT+CSIM=?")
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if got != "" {
		t.Fatalf("first run = %q, want empty response", got)
	}

	got, err = at.run(`AT+CSIM=4,"00"`)
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if got != `+CSIM: 4,"9000"` {
		t.Fatalf("second run = %q, want %q", got, `+CSIM: 4,"9000"`)
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

func TestRunMatchesCompleteResultCodes(t *testing.T) {
	serial := newScriptedSerial("AT+TEST\r\n+INFO: LOOKUP\r\nOK\r\n")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	got, err := at.run("AT+TEST")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if got != "+INFO: LOOKUP" {
		t.Fatalf("run = %q, want %q", got, "+INFO: LOOKUP")
	}
}

func TestRunReturnsStructuredATErrors(t *testing.T) {
	serial := newScriptedSerial("AT+TEST\r\n+CME ERROR: 42\r\n")
	at := &AT{
		s:      serial,
		reader: bufio.NewReader(serial),
	}

	_, err := at.run("AT+TEST")
	if err == nil {
		t.Fatal("run error = nil, want error")
	}
	if err.Error() != "+CME ERROR: 42" {
		t.Fatalf("run error = %q, want %q", err.Error(), "+CME ERROR: 42")
	}
}

var _ io.ReadWriteCloser = (*scriptedSerial)(nil)
