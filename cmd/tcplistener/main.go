package main

import (
	"fmt"
	"log"
	"net"

	"httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			break
		}
		fmt.Printf("Connection success from %s\n", conn.RemoteAddr())

		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error reading request from conn", err)
			break
		}
		fmt.Printf(
			"Request line: \n - Method: %s\n - Target: %s\n - Version: %s\n",
			request.RequestLine.Method,
			request.RequestLine.RequestTarget,
			request.RequestLine.HttpVersion)

		fmt.Printf("Connection close from %s\n", conn.RemoteAddr())
	}
}
