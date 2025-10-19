package response

import (
	"fmt"
	"httpfromtcp/internal/headers" // Import headers
	"io"
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

type Writer struct {
	conn           io.Writer
	statusWritten  bool
	headersWritten bool
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{conn: w}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.statusWritten {
		return fmt.Errorf("status line already written")
	}

	reasonPhrase, ok := reasonPhrases[statusCode]
	if !ok {
		reasonPhrase = ""
	}
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)

	_, err := w.conn.Write([]byte(statusLine))
	if err == nil {
		w.statusWritten = true
	}
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if !w.statusWritten {
		return fmt.Errorf("must write status line before writing headers")
	}
	if w.headersWritten {
		return fmt.Errorf("headers already written")
	}

	for key, value := range h {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.conn.Write([]byte(headerLine))
		if err != nil {
			return fmt.Errorf("error writing header '%s': %w", key, err)
		}
	}

	_, err := w.conn.Write([]byte("\r\n"))
	if err == nil {
		w.headersWritten = true
	}
	return err
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	if !w.headersWritten {
		return 0, fmt.Errorf("must write headers (including blank line) before writing body")
	}

	n, err := w.conn.Write(body)
	return n, err
}
