package http1

import (
	"fmt"
	"net"
	"net/http"
)

type Response struct {
	conn        net.Conn
	StatusCode  int
	bytes       int
	header      http.Header
	wroteHeader bool
}

func NewResponse(conn net.Conn) *Response {
	return &Response{
		conn:   conn,
		header: make(http.Header),
	}
}

func (r *Response) Header() http.Header {
	return r.header
}

func (r *Response) Write(data []byte) (n int, err error) {
	if !r.wroteHeader {
		r.WriteHeader(200)
	}

	n, err = r.conn.Write(data)
	r.bytes += n
	return n, err
}

func (r *Response) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}

	r.StatusCode = statusCode
	r.wroteHeader = true

	statusText := http.StatusText(statusCode)

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, statusText)

	r.conn.Write([]byte(statusLine))

	for key, values := range r.header {
		for _, val := range values {
			line := fmt.Sprintf("%s: %s\r\n", key, val)
			r.conn.Write([]byte(line))
		}
	}

	r.conn.Write([]byte("\r\n"))
}
