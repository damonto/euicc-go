package mbim

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestMBIMProtocolErrorText(t *testing.T) {
	tests := map[MBIMProtocolError]string{
		MBIMProtocolErrorInvalid:       "Invalid MBIM protocol error",
		MBIMProtocolErrorNotOpened:     "MBIM protocol error: Not Opened",
		MBIMProtocolErrorMaxTransfer:   "MBIM protocol error: Max Transfer",
		MBIMProtocolError(0xffff_ffff): "Unknown MBIM Protocol Error: 4294967295",
	}

	for code, want := range tests {
		if got := code.Error(); got != want {
			t.Fatalf("%d Error() = %q, want %q", code, got, want)
		}
	}
}

func TestMBIMStatusText(t *testing.T) {
	tests := map[MBIMStatus]string{
		MBIMStatusNone:                    "Success",
		MBIMStatusReserved:                "Reserved",
		MBIMStatusMsInvalidLogicalChannel: "Ms Invalid Logical Channel",
		MBIMStatusDecodeOrParsingError:    "Decode Or Parsing Error",
		MBIMStatus(0xffff_ffff):           "Unknown MBIM Status Error: 4294967295",
	}

	for code, want := range tests {
		if got := code.Error(); got != want {
			t.Fatalf("%d Error() = %q, want %q", code, got, want)
		}
	}
}

func TestMBIMErrorTextCoversDeclaredErrors(t *testing.T) {
	data, err := os.ReadFile("errors.go")
	if err != nil {
		t.Fatalf("read errors.go: %v", err)
	}

	assertMBIMErrorTextCoversDeclaredConstants(t, string(data), "MBIMProtocolError")
	assertMBIMErrorTextCoversDeclaredConstants(t, string(data), "MBIMStatus")
}

func assertMBIMErrorTextCoversDeclaredConstants(t *testing.T, data, prefix string) {
	t.Helper()

	mapped := make(map[string]bool)
	mapRe := regexp.MustCompile(`(?m)^\s*(` + prefix + `[A-Za-z0-9]+):\s*"`)
	for _, match := range mapRe.FindAllStringSubmatch(data, -1) {
		mapped[match[1]] = true
	}

	for _, name := range declaredMBIMConstants(data, prefix) {
		if !mapped[name] {
			t.Fatalf("%s is declared but not mapped to text", name)
		}
	}
}

func declaredMBIMConstants(data, prefix string) []string {
	var names []string
	inConst := false
	for _, line := range strings.Split(data, "\n") {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "const (":
			inConst = true
			continue
		case ")":
			inConst = false
			continue
		}
		if !inConst {
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) == 0 || !strings.HasPrefix(fields[0], prefix) {
			continue
		}
		names = append(names, fields[0])
	}
	return names
}
