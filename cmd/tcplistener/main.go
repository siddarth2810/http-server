package main

import (
	"fmt"
	"log"
	"net"

	"sid.tv/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err)
		}

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error: %s\n", err)
		}
		fmt.Printf("Request Line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
        fmt.Printf("Headers:\n")// TODO: Headers below wont get printed when this line is removed
		r.Headers.ForEach(func(n, v string) {
			fmt.Printf("- %s: %s\n", n, v)
		})

	}
}
