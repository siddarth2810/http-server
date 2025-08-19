package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	s, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal("error", "error", err)
	}

	for {
		data := make([]byte, 8)
		n, err := s.Read(data)
		if err != nil {
			break
		}
		fmt.Printf("read: %s\n", string(data[:n]))
	}
}
