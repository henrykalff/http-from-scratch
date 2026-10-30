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
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

	wrappedHandler := otelhttp.NewHandler(handler, "http1server-request")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			continue
		}
		go func(c net.Conn) {
			_ = HandleConnection(conn, wrappedHandler)
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

		metrics.RequestsTotal.Inc() // Metrics does not work currently in testcases.
		start := time.Now()
		shouldEnd, err := handleRequest(conn, br, mux)
		metrics.RequestProcessingDuration.UpdateDuration(start)
		if err != nil {
			return err
		}
		if shouldEnd {
			return nil
		}
	}
}

func handleRequest(conn net.Conn, br *bufio.Reader, mux http.Handler) (bool, error) {
	req, err := ParseRequest(br)

	if err != nil {
		if errors.Is(err, ErrUnsupportedVersion) {
			fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", http.StatusHTTPVersionNotSupported, http.StatusText(http.StatusHTTPVersionNotSupported))
			return true, nil
		}

		status := http.StatusBadRequest
		fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", status, http.StatusText(status))
		return true, nil
	}

	resp := NewResponse(req, req.Close)
	mux.ServeHTTP(resp, req)

	if req.Body != nil {
		_, _ = io.Copy(io.Discard, req.Body)
		_ = req.Body.Close()
	}

	if err := resp.flush(conn); err != nil {
		return true, err
	}
	if req.Close {
		return true, nil
	}
	return false, nil
}
