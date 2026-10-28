package main

import (
	"fmt"
	"httpServer/internal/http1"
	"httpServer/internal/metrics"
	"io"
	"log"
	"net/http"
)

const addr = ":8080"

func main() {
	metrics.InitMetrics()

	http1.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {

		data := []byte(`<html>
		<head>
		<title>Hello World</title>
		</head>
		<body>
		Hello World
		</body>
		</html>`)
		w.Header().Set("Server", "nginx/1.15.8")
		w.Header().Set("Date", "Fri, 09 Apr 2021 13:34:03 GMT")
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.Write(data)
	})
	http1.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len("hello")))
		fmt.Fprintf(w, "hello")

	})
	http1.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Body)
		_, _ = io.Copy(w, r.Body)
	})
	http1.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WriteMetrics(w)
	})
	if err := http1.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}

}
