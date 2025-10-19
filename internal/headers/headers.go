package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

// Parse parses the provided data and returns the number of bytes consumed,
// whether the parsing is done, and any error encountered.
// Parse is done when it encounters a blank line.
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIndex := bytes.Index(data, []byte("\r\n"))

	if crlfIndex == -1 {
		return 0, false, nil
	}

	if crlfIndex == 0 {
		return 2, true, nil
	}

	headerLine := data[:crlfIndex]

	bytesConsumed := crlfIndex + 2

	colonIndex := bytes.Index(headerLine, []byte(":"))
	if colonIndex == -1 {
		return 0, false, fmt.Errorf("invalid header format: missing colon")
	}

	key := headerLine[:colonIndex]

	if bytes.HasSuffix(key, []byte(" ")) {
		return 0, false, fmt.Errorf("invalid header format: space before colon")
	}
	if bytes.Contains(key, []byte(" ")) {
		return 0, false, fmt.Errorf("invalid header format: space in key")
	}
	if !validateHeaderKey(key) {
		return 0, false, fmt.Errorf("invalid header format: invalid key")
	}

	lowercaseKey := string(bytes.ToLower(key))
	value := bytes.TrimSpace(headerLine[colonIndex+1:])
	if existingVal, ok := h[lowercaseKey]; ok {
		h[lowercaseKey] = existingVal + "," + string(value)
	} else {
		h[lowercaseKey] = string(value)
	}

	return bytesConsumed, false, nil
}

// Based on RFC 9110 (5.6.2)
func isTokenChar(b byte) bool {
	if b >= '0' && b <= '9' {
		return true
	}
	if b >= 'A' && b <= 'Z' {
		return true
	}
	if b >= 'a' && b <= 'z' {
		return true
	}
	switch b {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}

func validateHeaderKey(key []byte) bool {
	if len(key) == 0 {
		return false
	}
	for _, b := range key {
		if !isTokenChar(b) {
			return false
		}
	}
	return true
}
