package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sid.tv/internal/request"
	"sid.tv/internal/response"
	"sid.tv/internal/server"
)

func respond400() []byte {
	return []byte(`
	<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
	`)
}

func respond500() []byte {
	return []byte(`
	<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
	`)
}

func respond200() []byte {
	return []byte(`
	<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
	`)
}

const port = 42069

func main() {

	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		headers := response.GetDefaultHeaders(0)
		body := respond200()
		status := response.StatusOK

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			body = respond400()
			status = response.StatusBadRequest

			headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
			headers.Replace("Content-type", "text/html")
			w.WriteHeaders(*headers)
			w.WriteBody(body)

		case "/myproblem":
			body = respond500()
			status = response.StatusInternalServerError

		default:
			headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
			w.WriteStatusLine(status)
			w.WriteHeaders(*headers)
			w.WriteBody(body)
		}
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
