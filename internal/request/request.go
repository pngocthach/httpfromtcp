package request

import (
	"bytes"
	"fmt"
	"io"
)

type Request struct {
	RequestLine RequestLine
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
	return r.state == StateDone || r.state == StateHeaders
}

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

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			return nil, err
		}
		readToIndex += n

		bytesParsed, err := request.parse(buf[:readToIndex])

		if bytesParsed > 0 {
			remainingBytes := readToIndex - bytesParsed
			copy(buf, buf[bytesParsed:readToIndex])
			readToIndex = remainingBytes
		}

		if err == io.EOF {
			if !request.done() {
				return nil, fmt.Errorf("connection closed before request was fully parsed")
			}
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return request, nil
}

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
			r.state = StateDone
			return bytesConsumed, nil

		case StateDone:
			return bytesConsumed, nil

		default:
			return 0, fmt.Errorf("unknown parser state")
		}
	}
}

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

	return &RequestLine{
		HttpVersion:   string(version),
		RequestTarget: string(target),
		Method:        string(method),
	}, bytesConsumed, nil
}
