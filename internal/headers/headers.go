package headers

import (
	"bytes"
	"fmt"
)

var SEPERATOR = ":"

var ERROR_MALFORMED_HEADER_LINE = fmt.Errorf("malformed header line")

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}
func ParseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed field line")
	}

	name := parts[0]
	//fmt.Println("parts[0]: ", string(name))
	value := bytes.TrimSpace(parts[1])
	//fmt.Println("parts[1]: ", string(value))

	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", fmt.Errorf("malformed field name ")
	}

	return string(name), string(value), nil
}

func (h Headers) Parse(data []byte) (int, bool, error) {
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
			break
		}

		name, value, err := ParseHeader(data[read : read+idx])
		if err != nil {
			return read, false, err
		}

		//fmt.Println("parsing header2: ", string(data[:]))
		read = idx + len(rn)
		h[name] = value
	}
	return read, done, nil
}
