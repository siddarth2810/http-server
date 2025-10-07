package main

import (
	"fmt"
)

type counter struct {
	n int
}

func (c *counter) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte('A' + (c.n % 26))
		c.n += 1
	}
	return len(p), nil
}

func main() {
	r := &counter{}
	buf := make([]byte, 10)
	for range 3 {
		n, _ := r.Read(buf)
		fmt.Printf("%q (read %d bytes) \n", buf[:n], n)
	}
}
