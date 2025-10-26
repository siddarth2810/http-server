package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"sid.tv/internal/headers"
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

// pad with two zeros for hex
func toStr(bytes []byte) string {
	out := ""
	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

func main() {

	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := respond200()
		status := response.StatusOK

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			body = respond400()
			status = response.StatusBadRequest

			h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
			h.Replace("Content-type", "text/html")
			w.WriteHeaders(*h)
			w.WriteBody(body)

		case "/myproblem":
			body = respond500()
			status = response.StatusInternalServerError

		case "/httpbin/":
			target := req.RequestLine.RequestTarget
			res, err := http.Get("http://httpbin.org/" + target[len("/httpbin/"):])
			if err != nil {
				body = respond500()
				status = response.StatusInternalServerError
			} else {
				w.WriteStatusLine(response.StatusOK)
				h.Delete("Content-Length")
				h.Set("Transfer-Encoding", "chunked")
				h.Replace("Content-Type", "text/plain")
				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")
				w.WriteHeaders(*h)

				fullbody := []byte{}
				for {
					data := make([]byte, 32)
					n, err := res.Body.Read(data)
					if err != nil {
						break
					}
					fullbody = append(fullbody, data[:n]...)
					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}
				w.WriteBody([]byte("0\r\n"))
				tailers := headers.NewHeaders()
				out := sha256.Sum256(fullbody)
				tailers.Set("X-Content-SHA256", toStr(out[:]))
				tailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullbody)))
				w.WriteHeaders(*tailers)

				return
			}
		case "/video":
			f, _ := os.ReadFile("assets/vim.mp4")
			h.Replace("Content-Type", "video/mp4")
			h.Replace("Content-Length", fmt.Sprintf("%d", len(f)))

			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(*h)
			w.WriteBody(f)

		default:
			h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
			w.WriteStatusLine(status)
			w.WriteHeaders(*h)
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
