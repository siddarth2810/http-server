// cmd/udpsender/main.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("failed to resolve UDP address: %v", err)
	}

	// create a UDP connection
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("failed to dial UDP: %v", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("error closing connection: %v", closeErr)
		}
	}()

	// Step 3: set up a bufio.Reader for stdin
	reader := bufio.NewReader(os.Stdin)

	// Step 4: infinite loop
	for {
		fmt.Print(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading input: %v", err)
			break
		}

        //netcat command: nc -u -l 42069
		// write to UDP, so the conn works even after terminating netcat
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Printf("error writing to UDP: %v", err)
		}
	}
}

