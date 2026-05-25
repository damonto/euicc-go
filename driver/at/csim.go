package at

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type CSIMCommand []byte

func (c CSIMCommand) MarshalText() ([]byte, error) {
	hexData := strings.ToUpper(hex.EncodeToString(c))
	return fmt.Appendf(nil, "AT+CSIM=%d,%q", len(hexData), hexData), nil
}

type CSIMResponse []byte

func (r CSIMResponse) MarshalText() ([]byte, error) {
	hexData := strings.ToUpper(hex.EncodeToString(r))
	return fmt.Appendf(nil, "%d,%q", len(hexData), hexData), nil
}

func (r *CSIMResponse) UnmarshalText(text []byte) error {
	raw := string(text)
	for line := range strings.SplitSeq(raw, "\n") {
		line = strings.TrimSpace(line)
		if body, ok := strings.CutPrefix(line, "+CSIM:"); ok {
			return r.unmarshalBody(strings.TrimSpace(body), false)
		}
	}

	var err error
	for line := range strings.SplitSeq(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "+") {
			continue
		}
		if unmarshalErr := r.unmarshalBody(line, true); unmarshalErr == nil {
			return nil
		} else {
			err = errors.Join(err, fmt.Errorf("%q: %w", line, unmarshalErr))
		}
	}
	if err != nil {
		return fmt.Errorf("invalid CSIM response: %w", err)
	}

	return fmt.Errorf("invalid CSIM response: %q", raw)
}

func (r *CSIMResponse) unmarshalBody(body string, requireKnownStatusWord bool) error {
	hexData := body
	wantLen := -1
	lengthField, dataField, ok := strings.Cut(body, ",")
	if ok {
		lengthField = strings.TrimSpace(lengthField)
		n, err := strconv.Atoi(lengthField)
		if err != nil || n < 0 {
			return fmt.Errorf("invalid CSIM response length %q", lengthField)
		}
		wantLen = n
		hexData = dataField
	}

	hexData = strings.TrimSpace(hexData)
	if strings.HasPrefix(hexData, `"`) {
		var err error
		hexData, err = strconv.Unquote(hexData)
		if err != nil {
			return fmt.Errorf("invalid CSIM response data: %w", err)
		}
	}
	if wantLen >= 0 {
		if len(hexData) != wantLen {
			return fmt.Errorf("CSIM response length mismatch: got %d, want %d", len(hexData), wantLen)
		}
	}

	response, err := hex.DecodeString(hexData)
	if err != nil {
		return fmt.Errorf("invalid CSIM response data: %w", err)
	}
	if requireKnownStatusWord {
		if len(response) < 2 {
			return errors.New("missing response status word")
		}
		if !hasKnownStatusWord(response) {
			return fmt.Errorf("unrecognized APDU status word: %X", response[len(response)-2:])
		}
	}
	*r = response
	return nil
}

func statusOK(response []byte) bool {
	return len(response) >= 2 && response[len(response)-2] == 0x90 && response[len(response)-1] == 0x00
}

func statusHasMore(response []byte) bool {
	return len(response) >= 2 && response[len(response)-2] == 0x61
}

func hasKnownStatusWord(response []byte) bool {
	if len(response) < 2 {
		return false
	}
	sw1 := response[len(response)-2]
	return (sw1 >= 0x61 && sw1 <= 0x6F) || (sw1 >= 0x90 && sw1 <= 0x9F)
}
