package request

import (
	"bytes"
	"fmt"
	"io"
	"log"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
	state       parserState
}

type parserState string

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http version")
var ERROR_Request_In_Error_State = fmt.Errorf("request in error state")
var SEPERATOR = []byte("\r\n")

const (
	StateInit  parserState = "init"
	StateDone  parserState = "done"
	StateError parserState = "error"
)

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPERATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPERATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil

}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		switch r.state {
		case StateError:
			return 0, ERROR_Request_In_Error_State

		case StateInit:
			rl, n, err := parseRequestLine(data)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateDone
			return read, nil

		case StateDone:
			break outer
		}
	}
	return read, nil
}

func newRequest() *Request {
	return &Request{
		state: StateInit,
	}
}

func (r *Request) done() bool {
	return r.state == StateDone && r.state != StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	buf := make([]byte, 8)
	readToIdx := 0
	//log.Printf("start")

	for !request.done() {
		n, err := reader.Read(buf[readToIdx:])
		if err != nil {
			return nil, err
		}
		readToIdx += n
		//grow the size when its full
		if readToIdx == len(buf) {
			tmp := make([]byte, len(buf)*2)
			copy(tmp, buf)
			buf = tmp
		}
		readN, err := request.parse(buf[:readToIdx])
		if err != nil {
			return nil, err
		}
		//	log.Printf("→ Before copy: readToIdx=%d, readN=%d, buf[:readToIdx]=%q\n", readToIdx, readN, buf[:readToIdx])
		copy(buf, buf[readN:readToIdx])
		readToIdx -= readN
		//	log.Printf("→ After  copy: readToIdx=%d, buf[:readToIdx]=%q\n", readToIdx, buf[:readToIdx])

	}
	// log.Printf("done")
	return request, nil
}
