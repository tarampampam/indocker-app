package healthcheck

import (
	"context"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/indocker-app/app/internal/cli/shared"
)

// NewCommand creates `healthcheck` command.
func NewCommand() *cli.Command {
	var (
		httpPortFlag  = shared.HttpPortFlag
		httpsPortFlag = shared.HttpsPortFlag
	)

	return &cli.Command{
		Name:    "healthcheck",
		Aliases: []string{"hc", "health", "check"},
		Usage:   "Health checker for the HTTP(S) servers. Use case - docker healthcheck",
		Action: func(ctx context.Context, c *cli.Command) error {
			return NewHealthChecker().Check(ctx, uint(c.Uint(httpPortFlag.Name)), uint(c.Uint(httpsPortFlag.Name)))
		},
		Flags: []cli.Flag{
			&httpPortFlag,
			&httpsPortFlag,
		},
	}
}
