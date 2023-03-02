package healthcheck

import (
	"github.com/urfave/cli/v2"

	"gh.tarampamp.am/indocker-app/daemon/internal/env"
)

// NewCommand creates `healthcheck` command.
func NewCommand() *cli.Command {
	const (
		httpPortFlagName  = "http-port"
		httpsPortFlagName = "https-port"
	)

	return &cli.Command{
		Name:    "healthcheck",
		Aliases: []string{"chk", "health", "check"},
		Usage:   "Health checker for the HTTP(S) servers. Use case - docker healthcheck",
		Action: func(c *cli.Context) error {
			return NewHealthChecker(c.Context).Check(c.Uint(httpPortFlagName), c.Uint(httpsPortFlagName))
		},
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:    httpPortFlagName,
				Usage:   "HTTP server port",
				Value:   8080,
				EnvVars: []string{env.HTTPPort.String()},
			},
			&cli.UintFlag{
				Name:    httpsPortFlagName,
				Usage:   "HTTPS server port",
				Value:   8443,
				EnvVars: []string{env.HTTPSPort.String()},
			},
		},
	}
}
