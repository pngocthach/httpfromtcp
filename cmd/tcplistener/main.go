package main

import (
	"httpfromtcp/internal/request"  // Thêm import request
	"httpfromtcp/internal/response" // Thêm import response
	"httpfromtcp/internal/server"
	"io" // Thêm import io
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func myHandler(w io.Writer, req *request.Request) *server.HandlerError {
	log.Printf("Handling request for target: %s", req.RequestLine.RequestTarget)

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    "Your problem is not my problem\n",
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: response.StatusInternalServerError,
			Message:    "Woopsie, my bad\n",
		}
	default:
		_, err := w.Write([]byte("All good, frfr\n"))
		if err != nil {
			log.Printf("Error writing response body: %v", err)
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    "Failed to write response body\n",
			}
		}
		return nil
	}
}

func main() {
	server, err := server.Serve(port, myHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
