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

type RequestStatus string

const InitState RequestStatus = "init"
const DoneState RequestStatus = "done"

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) done() bool {
	return r.state == DoneState
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffSize := 8
	buf := make([]byte, buffSize)
	request := &Request{
		state: InitState,
	}
	readToIndex := 0

	for !request.done() {
		// read data into the buffer
		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			request.state = DoneState
			return nil, err
		}

		// 2 case: 1. no buffer -> n = bufSize   2. read all data -> n = total read size
		readToIndex += n

		// parse data from the buffer
		readN, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		if readN != 0 {
			readToIndex -= readN
			copy(buf, buf[:readToIndex])
		}
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

	for {
		if r.state == InitState {
			rl, readN, err := parseRequestLine(data)
			if err != nil {
				return 0, err
			}
			if readN == 0 {
				return 0, nil
			}

			r.RequestLine = *rl
			read += readN
			r.state = DoneState

		} else if r.state == DoneState {
			break
		}
	}

	return read, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	lines := bytes.Split(data, []byte("\r\n"))
	parts := bytes.Split(lines[0], []byte(" "))
	read := 0

	if len(lines) == 1 {
		return nil, 0, nil
	}

	if len(parts) != 3 {
		return nil, read, fmt.Errorf("request line must be 3 parts")
	}

	method := parts[0]
	target := parts[1]
	httpVersion := parts[2]
	versionParts := bytes.Split(httpVersion, []byte("/"))

	if len(versionParts) != 2 || !bytes.Equal(versionParts[0], []byte("HTTP")) {
		return nil, read, fmt.Errorf("http version error")
	}
	version := versionParts[1]

	if !bytes.Equal(bytes.ToUpper(method), method) {
		return nil, read, fmt.Errorf("method must be capitalize")
	}
	if !bytes.Equal(version, []byte("1.1")) {
		return nil, read, fmt.Errorf("version must be 1.1")
	}

	read = len(lines) + len("\r\n")

	return &RequestLine{
		HttpVersion:   string(version),
		RequestTarget: string(target),
		Method:        string(method),
	}, read, nil
}
