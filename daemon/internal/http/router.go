package http

import (
	"net/http"
	"strings"

	"gh.tarampamp.am/indocker-app/daemon/internal/http/middleware"
)

type Router struct {
	routes   map[string]http.Handler
	prefix   string
	origins  []string
	fallback http.Handler
	notFound http.Handler
}

func NewRouter(prefix string, fallback http.Handler) *Router {
	var r = Router{
		routes:   make(map[string]http.Handler), // map[method+prefix+route]handler
		prefix:   prefix,
		origins:  []string{"indocker.app", "frontend.indocker.app" /* for local development */},
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
		if router.isInAllowedOrigins(r) {
			if handler, ok := router.routes[r.Method+url]; ok {
				middleware.Cors(handler).ServeHTTP(w, r)
			} else {
				router.notFound.ServeHTTP(w, r)
			}

			return
		}
	}

	router.fallback.ServeHTTP(w, r)
}

func (router *Router) isInAllowedOrigins(r *http.Request) bool {
	// r.Host can be "localhost:8080" or "localhost"
	if hostPort := strings.Split(r.Host, ":"); len(hostPort) > 0 {
		for _, origin := range router.origins {
			if hostPort[0] == origin {
				return true
			}
		}
	}

	return false
}

func (*Router) defaultNotFoundHandler() http.Handler {
	var body = []byte(`{"error":"not found"}`)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		_, _ = w.Write(body)
	})
}
