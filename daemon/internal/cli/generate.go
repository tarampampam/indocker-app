//go:build docs
// +build docs

package cli

import _ "gh.tarampamp.am/urfave-cli-docs/markdown"

// Run using `go generate -tags docs ./internal/cli`

// Generate CLI usage documentation and write it to the README.md file (between special tags).
//go:generate go run generate_readme.go
