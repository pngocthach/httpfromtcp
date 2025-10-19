package server

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Handler func(w *response.Writer, req *request.Request)

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
		errWriter := response.NewWriter(conn)
		errWriter.WriteStatusLine(response.StatusBadRequest)
		h := headers.NewHeaders()
		h.Set("Connection", "close")
		h.Set("Content-Type", "text/plain")
		errWriter.WriteHeaders(h) // Ghi cả dòng trống
		errWriter.WriteBody([]byte(fmt.Sprintf("Bad Request: %v\n", err)))
		return
	}

	responseWriter := response.NewWriter(conn)
	s.handler(responseWriter, req)

	log.Printf("Sent response to %s", conn.RemoteAddr())
}
