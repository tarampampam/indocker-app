package cli

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"

	"gh.tarampamp.am/indocker-app/app/internal/cli/start"
	"gh.tarampamp.am/indocker-app/app/internal/logger"
	"gh.tarampamp.am/indocker-app/app/internal/version"
)

//go:generate go run app_generate.go

// NewApp creates new console application.
func NewApp() *cli.Command {
	var (
		logLevelFlag = cli.StringFlag{
			Name:     "log-level",
			Value:    logger.InfoLevel.String(),
			Usage:    "Logging level (" + strings.Join(logger.LevelStrings(), "/") + ")",
			Sources:  cli.EnvVars("LOG_LEVEL"),
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
			Validator: func(s string) error {
				if _, err := logger.ParseLevel(s); err != nil {
					return err
				}

				return nil
			},
		}

		logFormatFlag = cli.StringFlag{
			Name:     "log-format",
			Value:    logger.ConsoleFormat.String(),
			Usage:    "Logging format (" + strings.Join(logger.FormatStrings(), "/") + ")",
			Sources:  cli.EnvVars("LOG_FORMAT"),
			OnlyOnce: true,
			Config:   cli.StringConfig{TrimSpace: true},
			Validator: func(s string) error {
				if _, err := logger.ParseFormat(s); err != nil {
					return err
				}

				return nil
			},
		}
	)

	// create "default" logger (will be overwritten later with customized)
	var log, _ = logger.New(logger.InfoLevel, logger.ConsoleFormat) // error will never occur

	return &cli.Command{
		Usage: "indocker.app",
		Before: func(ctx context.Context, c *cli.Command) error {
			_ = log.Sync() // sync previous logger instance

			var (
				logLevel, _  = logger.ParseLevel(c.String(logLevelFlag.Name))   // error ignored because the flag validates itself
				logFormat, _ = logger.ParseFormat(c.String(logFormatFlag.Name)) // --//--
			)

			configured, err := logger.New(logLevel, logFormat) // create new logger instance
			if err != nil {
				return err
			}

			*log = *configured // replace "default" logger with customized

			return nil
		},
		Commands: []*cli.Command{
			start.NewCommand(log),
		},
		Version: fmt.Sprintf("%s (%s)", version.Version(), runtime.Version()),
		Flags: []cli.Flag{ // global flags
			&logLevelFlag,
			&logFormatFlag,
		},
	}
}
