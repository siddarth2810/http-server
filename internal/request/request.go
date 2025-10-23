package request

import (
	"bytes"
	"fmt"
	"io"
	//"log/slog"
	"strconv"

	"sid.tv/internal/headers"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        string
	state       parserState
}

type parserState string

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http version")
var ERROR_Request_In_Error_State = fmt.Errorf("request in error state")
var SEPERATOR = []byte("\r\n")

const (
	StateInit    parserState = "init"
	StateDone    parserState = "done"
	StateError   parserState = "error"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
)

// looks up header map and get the value
func getInt(h *headers.Headers, name string, defaultValue int) int {
	valueStr, exists := h.Get(name)
	if !exists {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

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

// TODO: chunked enoding
func (r *Request) hasBody() bool {
	length := getInt(r.Headers, "content-length", 0)
	return length > 0
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

dance:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break dance
		}
		switch r.state {
		case StateError:
			return 0, ERROR_Request_In_Error_State

		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break dance
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders
			return read, nil

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return n, err
			}
			read += n

			//log.Println("headers parsed is: ", string(currentData[:n]))
			//log.Println("what is done ? ", done)

			if n == 0 {
				break dance
			}

			if done {
				if r.hasBody() {
					r.state = StateBody
					//return read, nil //go to body
				} else {
					r.state = StateDone
				}
			}

			//log.Println("go to body ? ", done)

		case StateBody:
			contentLength := getInt(r.Headers, "content-length", 0)
			if contentLength == 0 {
				panic("chunked not implemented")
			}

			remaining := min(contentLength-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == contentLength {
				r.state = StateDone
			}

		case StateDone:
			break dance
		default:
			panic("Oh yea we are bad programmers")
		}
	}
	return read, nil
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
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
		//slog.Info("RequestFromReader", "state", request.state)
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
		//log.Printf("→ Before copy: readToIdx=%d, readN=%d, buf[:readToIdx]=%q\n", readToIdx, readN, buf[:readToIdx])
		copy(buf, buf[readN:readToIdx])
		readToIdx -= readN
		//log.Printf("→ After  copy: readToIdx=%d, buf[:readToIdx]=%q\n", readToIdx, buf[:readToIdx])

	}
	//log.Printf("done")
	return request, nil
}
