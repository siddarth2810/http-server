package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

func main() {
	s, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal("error", "error", err)
	}

	str := ""
	for {
		data := make([]byte, 8)
		n, err := s.Read(data)
		if err != nil {
			break
		}
		data = data[:n]
		if i := bytes.IndexByte(data, '\n'); i != -1 {
			str += string(data[:i])
			data = data[i+1:]
			fmt.Printf("read: %s\n", str)
			str = ""
		}
		str += string(data)
	}
}
