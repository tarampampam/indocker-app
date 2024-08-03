package web

import (
	"embed"
	"io/fs"
)

// Generate mock distributive files, if needed.
//go:generate go run generate_dist.go

//go:embed dist
var content embed.FS

// Dist returns the content of the "dist" directory, which contains the built frontend.
func Dist() (data fs.FS) { data, _ = fs.Sub(content, "dist"); return } //nolint:nlreturn
