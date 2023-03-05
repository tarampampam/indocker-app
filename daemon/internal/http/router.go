package http

import (
	"net/http"
	"strings"

	"gh.tarampamp.am/indocker-app/daemon/internal/http/middleware"
)

type Router struct {
	routes   map[string]http.Handler
	prefix   string
	fallback http.Handler
	notFound http.Handler
}

func NewRouter(prefix string, fallback http.Handler) *Router {
	var r = Router{
		routes:   make(map[string]http.Handler), // map[method+prefix+route]handler
		prefix:   prefix,
		fallback: fallback,
	}

	r.notFound = r.defaultNotFoundHandler()

	return &r
}

func (router *Router) Register(method, route string, handler http.Handler) {
	router.routes[method+router.prefix+route] = handler
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if url := r.URL.Path; strings.HasPrefix(url, router.prefix) {
		if handler, ok := router.routes[r.Method+url]; ok {
			middleware.Cors(handler).ServeHTTP(w, r)
		} else {
			router.notFound.ServeHTTP(w, r)
		}
	} else {
		router.fallback.ServeHTTP(w, r)
	}
}

func (*Router) defaultNotFoundHandler() http.Handler {
	var body = []byte(`{"error":"not found"}`)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		_, _ = w.Write(body)
	})
}
