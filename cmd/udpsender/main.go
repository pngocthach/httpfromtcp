package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatalf("UDP resolve error: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Connect UDP error: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf(">")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Read from stdin error: %v", err)
		}

		_, err = conn.Write([]byte(input))
		if err != nil {
			log.Fatalf("Write to udp conn error: %v", err)
		}
	}
}
