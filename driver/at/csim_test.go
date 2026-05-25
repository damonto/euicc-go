package at

import (
	"bytes"
	"encoding"
	"strings"
	"testing"
)

var (
	_ encoding.TextMarshaler   = CSIMCommand(nil)
	_ encoding.TextMarshaler   = CSIMResponse(nil)
	_ encoding.TextUnmarshaler = (*CSIMResponse)(nil)
)

func TestCSIMCommandMarshalText(t *testing.T) {
	command := CSIMCommand{0x00, 0xA4, 0x04, 0x00}
	got, err := command.MarshalText()
	if err != nil {
		t.Fatalf("CSIMCommand.MarshalText failed: %v", err)
	}
	if string(got) != `AT+CSIM=8,"00A40400"` {
		t.Fatalf("CSIMCommand.MarshalText = %q, want %q", got, `AT+CSIM=8,"00A40400"`)
	}
}

func TestCSIMResponseMarshalText(t *testing.T) {
	response := CSIMResponse{0x90, 0x00}
	got, err := response.MarshalText()
	if err != nil {
		t.Fatalf("CSIMResponse.MarshalText failed: %v", err)
	}
	if string(got) != `4,"9000"` {
		t.Fatalf("CSIMResponse.MarshalText = %q, want %q", got, `4,"9000"`)
	}
}

func TestCSIMResponseUnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     []byte
	}{
		{
			name:     "prefixed",
			response: "+CMTI: \"SM\",1\n+CSIM: 4,\"9000\"",
			want:     []byte{0x90, 0x00},
		},
		{
			name:     "bare hex",
			response: "+CMTI: \"SM\",1\n9000",
			want:     []byte{0x90, 0x00},
		},
		{
			name:     "bare quoted hex",
			response: `"9000"`,
			want:     []byte{0x90, 0x00},
		},
		{
			name:     "bare length and quoted hex",
			response: `4,"9000"`,
			want:     []byte{0x90, 0x00},
		},
		{
			name:     "bare length and hex",
			response: `4,9000`,
			want:     []byte{0x90, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got CSIMResponse
			if err := got.UnmarshalText([]byte(tt.response)); err != nil {
				t.Fatalf("CSIMResponse.UnmarshalText failed: %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("CSIMResponse.UnmarshalText = % X, want % X", got, tt.want)
			}
		})
	}
}

func TestCSIMResponseUnmarshalTextRejectsLengthMismatch(t *testing.T) {
	var response CSIMResponse
	if err := response.UnmarshalText([]byte("+CSIM: 2,\"9000\"")); err == nil {
		t.Fatal("CSIMResponse.UnmarshalText error = nil, want error")
	}
}

func TestCSIMResponseUnmarshalTextReportsBareDecodeError(t *testing.T) {
	var response CSIMResponse
	err := response.UnmarshalText([]byte("not-hex"))
	if err == nil {
		t.Fatal("CSIMResponse.UnmarshalText error = nil, want error")
	}
	if !strings.Contains(err.Error(), `"not-hex"`) {
		t.Fatalf("CSIMResponse.UnmarshalText error = %q, want line context", err.Error())
	}
	if !strings.Contains(err.Error(), "invalid CSIM response data") {
		t.Fatalf("CSIMResponse.UnmarshalText error = %q, want data error", err.Error())
	}
}

func TestCSIMResponseUnmarshalTextRejectsBareShortResponse(t *testing.T) {
	var response CSIMResponse
	err := response.UnmarshalText([]byte("90"))
	if err == nil {
		t.Fatal("CSIMResponse.UnmarshalText error = nil, want error")
	}
	if !strings.Contains(err.Error(), "missing response status word") {
		t.Fatalf("CSIMResponse.UnmarshalText error = %q, want status word error", err.Error())
	}
}

func TestCSIMResponseUnmarshalTextRejectsUnknownBareStatusWord(t *testing.T) {
	var response CSIMResponse
	err := response.UnmarshalText([]byte("DEADBEEF"))
	if err == nil {
		t.Fatal("CSIMResponse.UnmarshalText error = nil, want error")
	}
	if !strings.Contains(err.Error(), "unrecognized APDU status word") {
		t.Fatalf("CSIMResponse.UnmarshalText error = %q, want status word error", err.Error())
	}
}
