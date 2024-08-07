package frontend

import (
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

//go:embed error404.html
var error404html []byte

const (
	contentTypeHeader = "Content-Type"
	contentTypeHTML   = "text/html; charset=utf-8"
)

// New returns a new middleware that serves files from the given file system.
//
// If the requested file does not exist, it will return index.html if it exists (otherwise html-formatted 404 error).
func New(root fs.FS, skipper func(*http.Request) bool) func(http.Handler) http.Handler { //nolint:funlen
	var fileServer = http.FileServerFS(root)

	const indexFileName = "index.html"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipper != nil && skipper(r) {
				next.ServeHTTP(w, r)

				return
			}

			var filePath = strings.TrimLeft(path.Clean(r.URL.Path), "/")

			if filePath == "" {
				filePath = indexFileName
			}

			fd, fErr := root.Open(filePath)
			switch { //nolint:wsl
			case os.IsNotExist(fErr): // if requested file does not exist
				index, indexErr := root.Open(indexFileName)
				if indexErr == nil { // always return index.html, if it exists (required for SPA to work)
					defer func() { _ = index.Close() }()

					if r.Method == http.MethodHead {
						w.WriteHeader(http.StatusOK)

						return
					}

					w.Header().Set(contentTypeHeader, contentTypeHTML)
					w.WriteHeader(http.StatusOK)
					_, _ = io.Copy(w, index)

					return
				}

				w.Header().Set(contentTypeHeader, contentTypeHTML)
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(error404html)

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
