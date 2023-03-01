package cli

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"

	"gh.tarampamp.am/indocker-app/daemon/internal/cli/start"
	"gh.tarampamp.am/indocker-app/daemon/internal/env"
	"gh.tarampamp.am/indocker-app/daemon/internal/logger"
	"gh.tarampamp.am/indocker-app/daemon/internal/version"
)

// NewApp creates new console application.
func NewApp() *cli.App {
	const (
		logLevelFlagName  = "log-level"
		logFormatFlagName = "log-format"

		defaultLogLevel  = logger.InfoLevel
		defaultLogFormat = logger.ConsoleFormat
	)

	// create "default" logger (will be overwritten later with customized)
	var log, _ = logger.New(defaultLogLevel, defaultLogFormat) // error will never occurs

	return &cli.App{
		Usage: "indocker.app daemon",
		Before: func(c *cli.Context) (err error) {
			_ = log.Sync() // sync previous logger instance

			var logLevel, logFormat = defaultLogLevel, defaultLogFormat //nolint:ineffassign

			// parse logging level
			if logLevel, err = logger.ParseLevel(c.String(logLevelFlagName)); err != nil {
				return err
			}

			// parse logging format
			if logFormat, err = logger.ParseFormat(c.String(logFormatFlagName)); err != nil {
				return err
			}

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
			&cli.StringFlag{
				Name:    logLevelFlagName,
				Value:   defaultLogLevel.String(),
				Usage:   "logging level (`" + strings.Join(logger.LevelStrings(), "/") + "`)",
				EnvVars: []string{env.LogLevel.String()},
			},
			&cli.StringFlag{
				Name:    logFormatFlagName,
				Value:   defaultLogFormat.String(),
				Usage:   "logging format (`" + strings.Join(logger.FormatStrings(), "/") + "`)",
				EnvVars: []string{env.LogFormat.String()},
			},
		},
	}
}
