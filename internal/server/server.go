package server

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Server struct {
	listener net.Listener
	handler  Handler
	isClosed atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := ":" + strconv.Itoa(port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot listen on port %d: %w", port, err)
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}
	server.isClosed.Store(false)

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	log.Println("Closing server listener...")
	s.isClosed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	defer func() {
		if !s.isClosed.Load() {
			s.Close()
		}
		log.Println("Listening goroutine stopped.")
	}()

	for {
		conn, err := s.listener.Accept()

		if s.isClosed.Load() {
			log.Println("Listener closed, stopping accept loop.")
			if conn != nil {
				conn.Close()
			}
			return
		}

		if err != nil {
			log.Printf("ERROR: Cannot accept connection: %v", err)
			continue
		}

		log.Printf("Accepted connection from %s", conn.RemoteAddr())
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer func() {
		conn.Close()
	}()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("ERROR: Cannot read request: %v", err)
		writeHandlerError(conn, &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    fmt.Sprintf("Bad Request: %v", err),
		})
		return
	}

	responseBodyBuf := new(bytes.Buffer)
	handlerErr := s.handler(responseBodyBuf, req)
	if handlerErr != nil {
		writeHandlerError(conn, handlerErr)
		return
	}

	err = response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		log.Printf("ERROR: Cannot write status line: %v", err)
		return
	}

	headers := response.GetDefaultHeaders(responseBodyBuf.Len())
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Printf("ERROR: Cannot write headers: %v", err)
		return
	}

	if responseBodyBuf.Len() > 0 {
		_, err = conn.Write(responseBodyBuf.Bytes())
		if err != nil {
			log.Printf("ERROR: Cannot write body: %v", err)
			return
		}
	} else {
		_, err = conn.Write([]byte(""))
		if err != nil {
			log.Printf("ERROR: Cannot write empty body: %v", err)
			return
		}
	}

	log.Printf("Sent response to %s", conn.RemoteAddr())
}

func writeHandlerError(w io.Writer, handlerErr *HandlerError) {
	err := response.WriteStatusLine(w, handlerErr.StatusCode)
	if err != nil {
		log.Printf("ERROR: Cannot write status line: %v", err)
		return
	}

	errorHeaders := response.GetDefaultHeaders(len(handlerErr.Message))
	err = response.WriteHeaders(w, errorHeaders)
	if err != nil {
		log.Printf("ERROR: Cannot write headers: %v", err)
		return
	}

	_, err = w.Write([]byte(handlerErr.Message))
	if err != nil {
		log.Printf("ERROR: Cannot write body: %v", err)
		return
	}
}
