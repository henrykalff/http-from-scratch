package http1

import (
	"bufio"
	"fmt"
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
		handler = http.DefaultServeMux
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		} else {
			HandleConnection(conn, handler)
		}
	}
}

func HandleConnection(conn net.Conn, mux http.Handler) error {
	defer conn.Close()
	req, err := ParseRequest(bufio.NewReader(conn))
	if err != nil {
		return err
	}
	log.Printf("%v\n", req)
	resp := NewResponse(conn)
	mux.ServeHTTP(resp, req)

	return nil

}
