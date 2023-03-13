package fileserver

import (
	"io"
	"net/http"
	"os"
	"path"
)

var fallback404 = []byte("<!doctype html><html><body><h1>Error 404</h1><h2>Not found</h2></body></html>") //nolint:lll,gochecknoglobals

func NewHandler(root http.FileSystem) http.Handler {
	var (
		fileServer       = http.FileServer(root)
		errorPageContent []byte
	)

	if f, err := root.Open("404.html"); err == nil {
		errorPageContent, _ = io.ReadAll(f)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := root.Open(path.Clean(r.URL.Path))
		if os.IsNotExist(err) {
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusNotFound)

				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)

			if len(errorPageContent) > 0 {
				_, _ = w.Write(errorPageContent)

				return
			}

			_, _ = w.Write(fallback404)

			return
		}

		if err != nil { // looks like unneeded, but so looks better
			_ = f.Close()
		}

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)

			return
		}

		fileServer.ServeHTTP(w, r)
	})
}
