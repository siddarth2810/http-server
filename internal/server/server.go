package server

import (
	"bytes"
	"fmt"
	"io"
	"net"

	"sid.tv/internal/request"
	"sid.tv/internal/response"
)

type HandleError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandleError

type Server struct {
	closed  bool
	handler Handler
}

// get req, add def headers
func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	headers := response.GetDefaultHeaders(0)
	request, err := request.RequestFromReader(conn)

	if err != nil {
		response.WriteStatusLine(conn, response.StatusBadRequest)
		response.WriteHeaders(conn, headers)
		return
	}

	//need a writer to write to
	writer := bytes.NewBuffer([]byte{})
	handleError := s.handler(writer, request)

	var body []byte = nil
	var status response.StatusCode = response.StatusOK

	if handleError != nil {
		status = handleError.StatusCode
		body = []byte(handleError.Message)
	} else {
		body = writer.Bytes()
	}

	//get body, reset content-len, write the body
	headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
	response.WriteStatusLine(conn, status)
	response.WriteHeaders(conn, headers)
	conn.Write(body)
}

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if s.closed {
			return
		}

		if err != nil {
			return
		}
		go runConnection(s, conn)
	}
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{
		closed:  false,
		handler: handler,
	}
	go runServer(server, listener)
	return server, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}
