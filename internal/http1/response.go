package http1

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Response struct {
	request    *http.Request
	header     http.Header
	statusCode int
	buf        bytes.Buffer
	closeConn  bool
}

func NewResponse(req *http.Request, closeConn bool) *Response {
	return &Response{
		request:    req,
		header:     make(http.Header),
		statusCode: http.StatusOK,
		closeConn:  closeConn,
	}
}

func (r *Response) Header() http.Header {
	return r.header
}

func (r *Response) Write(data []byte) (int, error) {
	return r.buf.Write(data)
}

func (r *Response) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

// finish writes the framed response to the network connection.
func (r *Response) finish(w io.Writer) error {
	body := r.buf.Bytes()
	if r.request.Method == http.MethodHead {
		body = nil
	}

	// Set Content-Length if not already specified
	if r.header.Get("Content-Length") == "" && r.header.Get("Transfer-Encoding") == "" {
		r.header.Set("Content-Length", strconv.Itoa(len(body)))
	}

	if r.closeConn {
		r.header.Set("Connection", "close")
	}

	proto := r.request.Proto
	if proto == "" {
		proto = "HTTP/1.1"
	}

	// Write status line
	if _, err := fmt.Fprintf(w, "%s %d %s\r\n", proto, r.statusCode, http.StatusText(r.statusCode)); err != nil {
		return err
	}

	// Write headers
	if err := r.header.Write(w); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "\r\n"); err != nil {
		return err
	}

	// Write body
	if len(body) > 0 {
		if _, err := w.Write(body); err != nil {
			return err
		}
	}
	return nil
}
