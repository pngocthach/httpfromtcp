package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

var reasonPhrases = map[StatusCode]string{
	StatusOK:                  "OK",
	StatusBadRequest:          "Bad Request",
	StatusInternalServerError: "Internal Server Error",
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	reasonPhrase, ok := reasonPhrases[statusCode]
	if !ok {
		reasonPhrase = ""
	}

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)

	_, err := w.Write([]byte(statusLine))
	if err != nil {
		return fmt.Errorf("error writing status line: %w", err)
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["content-length"] = strconv.Itoa(contentLen)
	h["connection"] = "close"
	h["content-type"] = "text/plain"

	return h
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	for key, value := range h {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.Write([]byte(headerLine))
		if err != nil {
			return fmt.Errorf("error writing header '%s': %w", key, err)
		}
	}

	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("error writing empty line after headers: %w", err)
	}

	return nil
}
