//go:build ignore
// +build ignore

package main

import (
	"os"

	"gh.tarampamp.am/urfave-cli-docs/markdown"

	"gh.tarampamp.am/indocker-app/app/internal/cli"
)

func main() {
	var app = cli.NewApp()

	// generate markdown documentation for the app
	docs, err := markdown.Render(app)
	if err != nil {
		panic(err)
	}

	const readmeFilePath = "../../readme.md"

	// read readme file
	readme, err := os.ReadFile(readmeFilePath)
	if err != nil {
		panic(err)
	}

	const start, end = "<!--GENERATED:CLI_DOCS-->", "<!--/GENERATED:CLI_DOCS-->"

	// replace the documentation section in the readme file
	updated, err := markdown.ReplaceBetween(start, end, string(readme), docs)
	if err != nil {
		panic(err)
	}

	// write the updated readme file
	if err = os.WriteFile(readmeFilePath, []byte(updated), 0664); err != nil {
		panic(err)
	}
}
