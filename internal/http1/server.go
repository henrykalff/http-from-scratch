package http1

import (
	"bufio"
	"errors"
	"fmt"
	"httpServer/internal/metrics"
	"io"
	"log"
	"net"
	"net/http"
)

func ListenAndServe(addr string, handler http.Handler) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	if handler == nil {
		// handler = http.DefaultServeMux
		handler = DefaultRouter
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			continue
		}
		go func(c net.Conn) {
			_ = HandleConnection(conn, handler)
		}(conn)
	}
}

func HandleConnection(conn net.Conn, mux http.Handler) error {
	defer conn.Close()

	br := bufio.NewReader(conn)

	for {
		// Clean disconnect before reading the next request
		if _, err := br.Peek(1); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}

		metrics.RequestsTotal.Inc()
		req, err := ParseRequest(br)

		if err != nil {
			if errors.Is(err, ErrUnsupportedVersion) {
				fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", http.StatusHTTPVersionNotSupported, http.StatusText(http.StatusHTTPVersionNotSupported))
				return nil
			}

			status := http.StatusBadRequest
			fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", status, http.StatusText(status))
			return nil
		}

		resp := NewResponse(req, req.Close)
		mux.ServeHTTP(resp, req)

		if req.Body != nil {
			_, _ = io.Copy(io.Discard, req.Body)
			_ = req.Body.Close()
		}

		if err := resp.flush(conn); err != nil {
			return err
		}
		log.Println("Is closed set:", req.Close)
		if req.Close {
			return nil
		}
	}
}
