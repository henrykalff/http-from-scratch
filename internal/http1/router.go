package http1

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type handlerFunc func(http.ResponseWriter, *http.Request)

func (f handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}

var isRouterSet = false
var DefaultRouter = &defaultRouter
var defaultRouter Router

func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	Handle(pattern, handlerFunc(handler))
}

func Handle(pattern string, handler Handler) {
	if !isRouterSet {
		isRouterSet = true
		defaultRouter = Router{routes: make(map[string]Handler)}
	}
	defaultRouter.register(pattern, handler)
}

type Router struct {
	mu     sync.Mutex
	routes map[string]Handler
}

func (r *Router) register(pattern string, handler Handler) {
	r.setRoute(pattern, handler)
}

func (r *Router) setRoute(pattern string, handler Handler) {
	r.mu.Lock()
	r.routes[pattern] = handler
	r.mu.Unlock()
}

func (r *Router) findRoute(req *http.Request) Handler {
	for route := range r.routes {
		log.Println("Route: ", route)
	}
	path := req.URL.Path

	r.mu.Lock()
	h := r.routes[path]
	r.mu.Unlock()
	return h
}

func (r *Router) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	h := r.findRoute(req)
	if h == nil {
		status := 404
		fmt.Fprintf(resp, "HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", status, http.StatusText(status))
		return
	}
	h.ServeHTTP(resp, req)
}
