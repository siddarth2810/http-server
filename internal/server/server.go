package server

import (
	"fmt"
	"io"
	"net"

	"sid.tv/internal/request"
	"sid.tv/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	closed  bool
	handler Handler
}

// get req, add def headers
func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	request, err := request.RequestFromReader(conn)

	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(*response.GetDefaultHeaders(0))
		return
	}

	s.handler(responseWriter, request)
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
