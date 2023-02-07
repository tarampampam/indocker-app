package middleware

import (
	"context"
	"net/http"
)

// Config the plugins configuration.
type Config struct {
	// ...
}

// CreateConfig creates the default plugins configuration.
func CreateConfig() *Config {
	return &Config{
		// ...
	}
}

// Plugin a plugins.
type Plugin struct {
	next http.Handler
	name string
	// ...
}

// New created a new plugins.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	// ...
	return &Plugin{
		next: next,
		name: name,
	}, nil
}

func (e *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/local-middleware" {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("it works!"))

		return
	}

	e.next.ServeHTTP(rw, req)
}
