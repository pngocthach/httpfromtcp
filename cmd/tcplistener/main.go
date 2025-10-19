package main

import (
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const port = 3000

const html400 = `<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>The request could not be processed.</p></body></html>`
const html500 = `<html><head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>An unexpected error occurred on the server.</p></body></html>`
const html200 = `<html><head><title>200 OK</title></head><body><h1>Success</h1><p>Request processed successfully.</p></body></html>`

func myHandler(w *response.Writer, req *request.Request) {
	log.Printf("Handling request for target: %s", req.RequestLine.RequestTarget)

	var statusCode response.StatusCode
	var bodyHTML string

	switch req.RequestLine.RequestTarget {
	case "/api/error":
		statusCode = response.StatusBadRequest
		bodyHTML = html400
	case "/api/internal":
		statusCode = response.StatusInternalServerError
		bodyHTML = html500
	default:
		statusCode = response.StatusOK
		bodyHTML = html200
	}

	err := w.WriteStatusLine(statusCode)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	h := headers.NewHeaders()
	h.Set("Content-Type", "text/html")
	h.Set("Content-Length", strconv.Itoa(len(bodyHTML)))
	h.Set("Connection", "close")

	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	_, err = w.WriteBody([]byte(bodyHTML))
	if err != nil {
		log.Printf("Error writing body: %v", err)
		return
	}

	log.Printf("Sent response %d for %s", statusCode, req.RequestLine.RequestTarget)
}

func main() {
	srv, err := server.Serve(port, myHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Printf("Server started on port %d", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Received shutdown signal, stopping server...")
}
