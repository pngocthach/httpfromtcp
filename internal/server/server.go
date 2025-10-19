package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
}

func Serve(port int) (*Server, error) {
	addr := ":" + strconv.Itoa(port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot listen on port %d: %w", port, err)
	}

	server := &Server{
		listener: listener,
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

	statusLine := "HTTP/1.1 200 OK\r\n"
	contentType := "Content-Type: text/plain\r\n"
	body := "Hello World!"
	contentLength := fmt.Sprintf("Content-Length: %d\r\n", len(body))

	response := statusLine +
		contentType +
		contentLength +
		"\r\n" +
		body

	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("ERROR: Cannot send response to %s: %v", conn.RemoteAddr(), err)
		return
	}
	log.Printf("Sent response to %s", conn.RemoteAddr())
}
