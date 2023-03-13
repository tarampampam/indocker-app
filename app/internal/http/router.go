package http

import (
	"net/http"
	"strings"
)

type Router struct {
	prefix   string
	notFound http.Handler
	fallback http.Handler
	mw       []func(next http.Handler) http.Handler

	// no mutex, so it's thread safe only if you don't register routes after the server starts
	routes map[string]http.Handler
}

type RouterOption func(*Router)

// WithPrefix sets the prefix for the router.
func WithPrefix(p string) RouterOption { return func(r *Router) { r.prefix = p } }

// WithNotFound sets the handler for not found routes.
func WithNotFound(h http.Handler) RouterOption { return func(r *Router) { r.notFound = h } }

// WithFallback sets the handler for routes that don't match any registered route.
func WithFallback(h http.Handler) RouterOption { return func(r *Router) { r.fallback = h } }

// WithMiddleware adds middleware to the router.
func WithMiddleware(mw ...func(http.Handler) http.Handler) RouterOption {
	return func(r *Router) { r.mw = append(r.mw, mw...) }
}

// NewRouter creates a new Router.
func NewRouter(opts ...RouterOption) *Router {
	var r = Router{
		routes: make(map[string]http.Handler), // map[method+prefix+route]handler
	}

	for _, opt := range opts {
		opt(&r)
	}

	return &r
}

// Register registers a new route.
func (router *Router) Register(method, route string, h http.Handler) *Router {
	router.routes[strings.ToUpper(method)+router.prefix+route] = h

	return router
}

// ServeHTTP implements the http.Handler interface.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var urlPath = r.URL.Path

	if router.prefix == "" || strings.HasPrefix(urlPath, router.prefix) {
		var handler http.Handler

		if h, ok := router.routes[r.Method+router.prefix+urlPath]; ok && h != nil {
			handler = h
		} else if router.notFound != nil {
			handler = router.notFound
		} else {
			handler = http.NotFoundHandler()
		}

		// wrap the handler with the middleware
		for _, mw := range router.mw {
			handler = mw(handler)
		}

		handler.ServeHTTP(w, r)

		return
	}

	router.fallback.ServeHTTP(w, r)
}

var notFoundJSONBody = []byte(`{"error":"not found"}`) //nolint:gochecknoglobals

// NotFoundJSONHandler returns a simple request handler that returns a 404 Not Found JSON response.
func NotFoundJSONHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)

		_, _ = w.Write(notFoundJSONBody)
	})
}
