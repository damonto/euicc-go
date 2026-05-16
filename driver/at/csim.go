package at

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func decodeCSIMResponse(response string) ([]byte, error) {
	for line := range strings.SplitSeq(response, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "+CSIM:") {
			continue
		}
		return decodeCSIMLine(line)
	}

	var decodeErr error
	for line := range strings.SplitSeq(response, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "+") {
			continue
		}
		decoded, err := decodeBareCSIMLine(line)
		if err == nil {
			return decoded, nil
		}
		decodeErr = errors.Join(decodeErr, fmt.Errorf("%q: %w", line, err))
	}
	if decodeErr != nil {
		return nil, fmt.Errorf("invalid CSIM response: %w", decodeErr)
	}

	return nil, fmt.Errorf("invalid CSIM response: %q", response)
}

func decodeCSIMLine(line string) ([]byte, error) {
	body := strings.TrimSpace(strings.TrimPrefix(line, "+CSIM:"))
	return decodeCSIMBody(body)
}

func decodeBareCSIMLine(line string) ([]byte, error) {
	response, err := decodeCSIMBody(line)
	if err != nil {
		return nil, err
	}
	if len(response) < 2 {
		return nil, errors.New("missing response status word")
	}
	if !hasKnownStatusWord(response) {
		return nil, fmt.Errorf("unrecognized APDU status word: %X", response[len(response)-2:])
	}
	return response, nil
}

func decodeCSIMBody(body string) ([]byte, error) {
	lengthField, dataField, ok := strings.Cut(body, ",")
	if ok {
		wantLen, err := strconv.Atoi(strings.TrimSpace(lengthField))
		if err != nil || wantLen < 0 {
			return nil, fmt.Errorf("invalid CSIM response length %q", strings.TrimSpace(lengthField))
		}

		hexData, err := decodeCSIMHexField(dataField)
		if err != nil {
			return nil, err
		}
		if len(hexData) != wantLen {
			return nil, fmt.Errorf("CSIM response length mismatch: got %d, want %d", len(hexData), wantLen)
		}

		return decodeHexResponse(hexData)
	}

	hexData, err := decodeCSIMHexField(body)
	if err != nil {
		return nil, err
	}
	return decodeHexResponse(hexData)
}

func decodeCSIMHexField(field string) (string, error) {
	hexData := strings.TrimSpace(field)
	if strings.HasPrefix(hexData, `"`) {
		var err error
		hexData, err = strconv.Unquote(hexData)
		if err != nil {
			return "", fmt.Errorf("invalid CSIM response data: %w", err)
		}
	}
	return hexData, nil
}

func decodeHexResponse(data string) ([]byte, error) {
	response, err := hex.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("invalid CSIM response data: %w", err)
	}
	return response, nil
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
