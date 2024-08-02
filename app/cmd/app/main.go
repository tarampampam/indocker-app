package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gh.tarampamp.am/indocker-app/app/internal/cli"
)

// main CLI application entrypoint.
func main() {
	var ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := cli.NewApp().Run(ctx, os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())

		os.Exit(1)
	}
}
