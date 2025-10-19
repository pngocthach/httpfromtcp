package request

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"regexp"
	"strconv"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       RequestStatus
}

type RequestStatus int

const (
	StateInit RequestStatus = iota
	StateHeaders
	StateBody
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
			copy(newBuf, buf[:readToIndex])
			buf = newBuf
		}

		n, readErr := reader.Read(buf[readToIndex:])
		readToIndex += n

		// Parse all available data in buffer (loop until nothing more can be parsed)
		for {
			bytesParsed, err := request.parse(buf[:readToIndex])
			if err != nil {
				return nil, err
			}

			// If nothing was parsed, we need more data
			if bytesParsed == 0 {
				break
			}

			// Compact buffer after successful parse
			remainingBytes := readToIndex - bytesParsed
			copy(buf, buf[bytesParsed:readToIndex])
			readToIndex = remainingBytes

			// Check if request is complete after parsing
			if request.done() {
				break
			}
		}

		// Handle read errors after parsing
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
				r.state = StateBody
			} else {
				return bytesConsumed, nil
			}

		case StateBody:
			contentLengthStr := r.Headers.Get("Content-Length")
			if contentLengthStr == "" {
				r.state = StateDone
				continue
			}

			contentLength, convErr := strconv.Atoi(contentLengthStr)
			if convErr != nil || contentLength < 0 {
				return 0, fmt.Errorf("invalid Content-Length: %q", contentLengthStr)
			}

			if contentLength == 0 {
				r.state = StateDone
				continue
			}

			bytesNeeded := contentLength - len(r.Body)
			bytesAvailable := len(data) - bytesConsumed
			bytesToConsume := min(bytesNeeded, bytesAvailable)

			if bytesToConsume > 0 {
				r.Body = append(r.Body, data[bytesConsumed:bytesConsumed+bytesToConsume]...)
				bytesConsumed += bytesToConsume
			}

			if len(r.Body) == contentLength {
				r.state = StateDone
			} else if len(r.Body) > contentLength {
				return 0, fmt.Errorf("received more body data than Content-Length specified")
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
