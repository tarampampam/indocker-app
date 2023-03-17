package web

import (
	"embed"
	"io/fs"
)

// Generate mock distributive files, if needed.
//go:generate go run dist.go

//go:embed dist
var content embed.FS

// Content returns the embedded web content.
func Content() fs.FS {
	data, _ := fs.Sub(content, "dist")

	return data
}
