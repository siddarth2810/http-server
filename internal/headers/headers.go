package headers

import (
	"bytes"
	"fmt"
	"strings"
)

func isToken(str []byte) bool {
	for _, ch := range str {
		found := false

		if ch >= 'A' && ch <= 'Z' ||
			ch >= 'a' && ch <= 'z' ||
			ch >= '0' && ch <= '9' {
			found = true
		}
		switch ch {
		case '#', '!', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}
	return true
}

var SEPERATOR = ":"

var ERROR_MALFORMED_HEADER_LINE = fmt.Errorf("malformed header line")

type Headers struct {
	headers map[string]string
}

func ParseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed field line")
	}

	name := parts[0]

	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", fmt.Errorf("malformed field name ")
	}

	//fmt.Println("parts[0]: ", string(name))
	value := bytes.TrimSpace(parts[1])
	//fmt.Println("parts[1]: ", string(value))

	return string(name), string(value), nil
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(name string) (string, bool) {
    str, ok := h.headers[strings.ToLower(name)]
    return str, ok
}

func (h *Headers) Replace(name, value string) {
	name = strings.ToLower(name)
		h.headers[name] = value
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)
	if v, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.headers[name] = value
	}
}


func (h *Headers) ForEach(cb func (n, v string)) {
    for n, v := range h.headers {
        cb(n,v)
    }
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	rn := "\r\n"
	done := false
	for {
		//fmt.Println("parsing header", string(data[read:]))
		idx := bytes.Index(data[read:], []byte(rn))

		if idx == -1 {
			break
		}

		if idx == 0 {
			done = true
			read += len(rn) //add the second crlf
			break
		}

		name, value, err := ParseHeader(data[read : read+idx])
		if err != nil {
			return read, false, err
		}

		if !isToken([]byte(name)) {
			return 0, false, fmt.Errorf("malformed header name")
		}
		h.Set(name, value)

		//fmt.Println("parsing header2: ", string(data[:]))
		read += idx + len(rn)
	}
	return read, done, nil
}

