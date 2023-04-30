package fileserver

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

var fallback404 = []byte("<!doctype html><html><body><h1>Error 404</h1><h2>Not found</h2></body></html>") //nolint:lll,gochecknoglobals

func NewHandler(root http.FileSystem) http.Handler { //nolint:funlen
	var (
		fileServer = http.FileServer(root)
		index      []byte
	)

	if f, err := root.Open("index.html"); err == nil {
		index, _ = io.ReadAll(f)
		_ = f.Close()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var filePath = path.Clean(r.URL.Path)

		f, fErr := root.Open(filePath)
		if os.IsNotExist(fErr) { //nolint:nestif // requested file not found
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusNotFound)

				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			if len(index) > 0 { // always return index.html if exists (required for SPA)
				_, _ = w.Write(index)

				return
			}

			_, _ = w.Write(fallback404)

			return
		} else if fErr == nil { // file exists
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusOK)

				return
			}

			const gzFileExt = ".gz" // note: dot at the beginning is required

			if gzFile, err := root.Open(filePath + gzFileExt); err == nil { // and we have a gzipped version
				_ = gzFile.Close()

				var contentType = mime.TypeByExtension(filepath.Ext(filePath)) // first try to detect by file extension

				if contentType == "" { // if failed, try to detect by content
					var buf = make([]byte, 32) //nolint:gomnd // 32 bytes are enough for "first bytes" checking

					if _, err = io.ReadFull(f, buf); err == nil { // read first bytes to detect content type of original file
						contentType = http.DetectContentType(buf) // if failed, try to detect by content
					}
				}

				w.Header().Set("Content-Encoding", "gzip") // set content encoding to gzip
				w.Header().Del("Content-Length")

				if contentType != "" {
					w.Header().Set("Content-Type", contentType) // set content type of original file
				}

				r.URL.Path += gzFileExt // force to serve gzipped version using http.FileServer below
			}
		}

		if f != nil { // looks like unneeded, but so looks better
			_ = f.Close()
		}

		fileServer.ServeHTTP(w, r)
	})
}
