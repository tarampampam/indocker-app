package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// A Handler responds to an HTTP API request.
//
// It should return an error WITHOUT calling WriteHeader on http.ResponseWriter if the request cannot be handled.
type Handler interface {
	Handle(http.ResponseWriter, *http.Request) error
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions as API handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a Handler that calls f.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

var _ Handler = (*HandlerFunc)(nil) // verify interface implementation

// Handle calls f(w, r).
func (f HandlerFunc) Handle(w http.ResponseWriter, r *http.Request) error { return f(w, r) }

// Router is a super-simple (and probably fast) API router.
type Router struct {
	prefix   string // should be not empty to make fallback work
	fallback http.Handler

	// no mutex, so it's thread safe only if you don't register routes after the server starts
	routes map[string]Handler
}

// NewRouter creates a new Router.
func NewRouter(prefix string, fallback http.Handler) *Router {
	return &Router{
		prefix:   prefix,
		fallback: fallback,
		routes:   make(map[string]Handler), // map[method+prefix+route]handler
	}
}

// Register registers a new route.
func (router *Router) Register(method, route string, h Handler) *Router {
	router.routes[strings.ToUpper(method)+router.prefix+route] = h

	return router
}

// ServeHTTP implements the http.Handler interface.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, router.prefix) {
		if h, ok := router.routes[r.Method+r.URL.Path]; ok && h != nil {
			router.addCorsHeaders(w, r)

			if err := h.Handle(w, r); err != nil {
				router.error(w, err)
			}

			return
		}

		router.notFound(w, r)

		return
	}

	if router.fallback != nil {
		router.fallback.ServeHTTP(w, r)

		return
	}

	router.error(w, errors.New("fallback handler is not set"))
}

// addCorsHeaders adds the CORS headers to the response.
func (router *Router) addCorsHeaders(w http.ResponseWriter, r *http.Request) {
	var scheme = "http"

	if r.TLS != nil {
		scheme = "https"
	}

	w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("%s://%s", scheme, r.Host))
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}

func (router *Router) error(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)

	if err == nil { // kinda fuse
		err = errors.New("unknown error")
	}

	_ = json.NewEncoder(w).Encode(struct { //nolint:errchkjson
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}

var notFoundJSONBody = []byte(`{"error":"not found"}`) //nolint:gochecknoglobals

// notFound is the default handler for not found routes.
func (router *Router) notFound(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	_, _ = w.Write(notFoundJSONBody)
}
