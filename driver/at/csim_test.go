package at

import (
	"bytes"
	"strings"
	"testing"
)

func TestDecodeCSIMResponse(t *testing.T) {
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
			got, err := decodeCSIMResponse(tt.response)
			if err != nil {
				t.Fatalf("decodeCSIMResponse failed: %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("decodeCSIMResponse = % X, want % X", got, tt.want)
			}
		})
	}
}

func TestDecodeCSIMResponseRejectsLengthMismatch(t *testing.T) {
	_, err := decodeCSIMResponse("+CSIM: 2,\"9000\"")
	if err == nil {
		t.Fatal("decodeCSIMResponse error = nil, want error")
	}
}

func TestDecodeCSIMResponseReportsBareDecodeError(t *testing.T) {
	_, err := decodeCSIMResponse("not-hex")
	if err == nil {
		t.Fatal("decodeCSIMResponse error = nil, want error")
	}
	if !strings.Contains(err.Error(), `"not-hex"`) {
		t.Fatalf("decodeCSIMResponse error = %q, want line context", err.Error())
	}
	if !strings.Contains(err.Error(), "invalid CSIM response data") {
		t.Fatalf("decodeCSIMResponse error = %q, want data error", err.Error())
	}
}

func TestDecodeCSIMResponseRejectsBareShortResponse(t *testing.T) {
	_, err := decodeCSIMResponse("90")
	if err == nil {
		t.Fatal("decodeCSIMResponse error = nil, want error")
	}
	if !strings.Contains(err.Error(), "missing response status word") {
		t.Fatalf("decodeCSIMResponse error = %q, want status word error", err.Error())
	}
}

func TestDecodeCSIMResponseRejectsUnknownBareStatusWord(t *testing.T) {
	_, err := decodeCSIMResponse("DEADBEEF")
	if err == nil {
		t.Fatal("decodeCSIMResponse error = nil, want error")
	}
	if !strings.Contains(err.Error(), "unrecognized APDU status word") {
		t.Fatalf("decodeCSIMResponse error = %q, want status word error", err.Error())
	}
}
