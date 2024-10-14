package main

import (
	"fmt"
	"os"

	"go.uber.org/automaxprocs/maxprocs"

	"gh.tarampamp.am/indocker-app/daemon/internal/cli"
)

// set GOMAXPROCS to match Linux container CPU quota.
var _, _ = maxprocs.Set(maxprocs.Min(1), maxprocs.Logger(func(_ string, _ ...any) {}))

// exitFn is a function for application exiting.
var exitFn = os.Exit //nolint:gochecknoglobals

// main CLI application entrypoint.
func main() { exitFn(run()) }

// run this CLI application.
// Exit codes documentation: <https://tldp.org/LDP/abs/html/exitcodes.html>
func run() int {
	if err := (cli.NewApp()).Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())

		return 1
	}

	return 0
}
