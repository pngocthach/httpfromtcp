package request

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"regexp"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       RequestStatus
}

type RequestStatus int

const (
	StateInit RequestStatus = iota
	StateHeaders
	StateDone
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) done() bool {
	return r.state == StateDone
}

// RequestFromReader reads and parses an HTTP request from the provided reader.
// It incrementally reads data and parses the request line and headers.
// Returns a fully parsed Request or an error if parsing fails.
func RequestFromReader(reader io.Reader) (*Request, error) {
	const bufferSize = 8
	buf := make([]byte, bufferSize)

	readToIndex := 0

	request := &Request{
		state: StateInit,
	}

	for !request.done() {
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, readErr := reader.Read(buf[readToIndex:])
		readToIndex += n

		bytesParsed, err := request.parse(buf[:readToIndex])

		if bytesParsed > 0 {
			remainingBytes := readToIndex - bytesParsed
			copy(buf, buf[bytesParsed:readToIndex])
			readToIndex = remainingBytes
		}

		if err != nil {
			return nil, err
		}

		if readErr != nil {
			if readErr == io.EOF {
				if !request.done() {
					return nil, fmt.Errorf("connection closed before request was fully parsed")
				}
				break
			}
			return nil, readErr
		}
	}

	return request, nil
}

// Parse processes the provided data buffer and advances the request parsing state.
// Returns the number of bytes consumed and any error encountered during parsing.
func (r *Request) parse(data []byte) (int, error) {
	bytesConsumed := 0

	for {
		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(data[bytesConsumed:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				return bytesConsumed, nil
			}

			r.RequestLine = *rl
			bytesConsumed += n
			r.state = StateHeaders

		case StateHeaders:
			if r.Headers == nil {
				r.Headers = headers.NewHeaders()
			}
			n, done, err := r.Headers.Parse(data[bytesConsumed:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				return bytesConsumed, nil
			}

			bytesConsumed += n

			if done {
				r.state = StateDone
			} else {
				return bytesConsumed, nil
			}

		case StateDone:
			return bytesConsumed, nil

		default:
			return 0, fmt.Errorf("unknown parser state")
		}
	}
}

var validMethod = regexp.MustCompile("^[A-Z]+$")

// parseRequestLine extracts the HTTP method, request target, and version from the first line.
// Returns the parsed RequestLine, number of bytes consumed (including CRLF), and any parsing error.
// Returns (nil, 0, nil) if there is insufficient data to parse a complete line.
func parseRequestLine(data []byte) (*RequestLine, int, error) {
	crlfIndex := bytes.Index(data, []byte("\r\n"))
	if crlfIndex == -1 {
		return nil, 0, nil
	}

	requestLineData := data[:crlfIndex]

	bytesConsumed := len(requestLineData) + 2

	parts := bytes.Split(requestLineData, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, fmt.Errorf("request line must have 3 parts, got %d", len(parts))
	}

	method, target, versionData := parts[0], parts[1], parts[2]

	versionParts := bytes.Split(versionData, []byte("/"))
	if len(versionParts) != 2 || !bytes.Equal(versionParts[0], []byte("HTTP")) {
		return nil, 0, fmt.Errorf("invalid HTTP version format: %s", versionData)
	}

	version := versionParts[1]
	if !bytes.Equal(version, []byte("1.1")) {
		return nil, 0, fmt.Errorf("invalid HTTP version: %s, only 1.1 is supported", version)
	}

	if !validMethod.Match(method) {
		return nil, 0, fmt.Errorf("invalid method: %s", method)
	}

	return &RequestLine{
		HttpVersion:   string(version),
		RequestTarget: string(target),
		Method:        string(method),
	}, bytesConsumed, nil
}
