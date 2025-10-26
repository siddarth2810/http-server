package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sid.tv/internal/request"
	"sid.tv/internal/response"
	"sid.tv/internal/server"
)

const port = 42069

func main() {

	s, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandleError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandleError{
				StatusCode: response.StatusBadRequest,
				Message:    "your problem lol\n"}

		case "/myproblem":
			return &server.HandleError{
				StatusCode: response.StatusInternalServerError,
				Message:    "whoops, my bad\n",
			}

		default:
			w.Write([]byte("All good, yay \n"))
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error starting the server: %s\n", err)
	}

	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server stopped gracefully")

}
