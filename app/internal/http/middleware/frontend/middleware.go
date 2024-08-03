package frontend

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeHTML   = "text/html; charset=utf-8"
)

// New returns a new middleware that serves files from the given file system.
//
// If the requested file does not exist, it will return index.html if it exists (otherwise html-formatted 404 error).
func New(root fs.FS, skipper func(*http.Request) bool) func(http.Handler) http.Handler { //nolint:funlen
	var (
		fileServer = http.FileServerFS(root)
		index      []byte

		fallback404 = []byte("<!doctype html><html><body><h1>Error 404</h1><h2>Not found</h2></body></html>")
	)

	if f, err := root.Open("index.html"); err == nil {
		index, _ = io.ReadAll(f)
		_ = f.Close()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipper != nil && skipper(r) {
				next.ServeHTTP(w, r)

				return
			}

			var filePath = strings.TrimLeft(path.Clean(r.URL.Path), "/")

			if filePath == "" {
				filePath = "index.html"
			}

			fd, fErr := root.Open(filePath)
			switch { //nolint:wsl
			case os.IsNotExist(fErr): // if requested file does not exist
				if len(index) > 0 { // always return index.html, if it exists (required for SPA to work)
					if r.Method == http.MethodHead {
						w.WriteHeader(http.StatusOK)

						return
					}

					w.Header().Set(contentTypeHeader, contentTypeHTML)
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(index)

					return
				}

				w.Header().Set(contentTypeHeader, contentTypeHTML)
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(fallback404)

				return
			case fErr != nil: // some other error
				if r.Method == http.MethodHead {
					w.WriteHeader(http.StatusInternalServerError)

					return
				}

				http.Error(w, fmt.Errorf("failed to open file %s: %w", filePath, fErr).Error(), http.StatusInternalServerError)

				return
			}

			defer func() { _ = fd.Close() }()

			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusOK)

				return
			}

			fileServer.ServeHTTP(w, r)
		})
	}
}
