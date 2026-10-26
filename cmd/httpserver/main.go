package main

import (
	"fmt"
	"httpServer/internal/http1"
	"log"
	"net/http"
)

const addr = ":8080"

func main() {
	http.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {

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
		log.Println("Before Write")
		log.Println(w.Write(data))

		// fmt.Fprintf(w, "hello")
	})
	if err := http1.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
