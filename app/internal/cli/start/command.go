package start

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/app/certs"
	"gh.tarampamp.am/indocker-app/app/internal/cli/shared"
	"gh.tarampamp.am/indocker-app/app/internal/cli/start/healthcheck"
	"gh.tarampamp.am/indocker-app/app/internal/docker"
	appHttp "gh.tarampamp.am/indocker-app/app/internal/http"
)

type (
	command struct {
		c *cli.Command

		options struct {
			Addr string
			HTTP struct {
				Port uint
			}
			HTTPS struct {
				Port              uint
				CertFile, KeyFile []byte // contents of the certificate and key files
			}
			ReadTimeout     time.Duration
			WriteTimeout    time.Duration
			IDLETimeout     time.Duration
			ShutdownTimeout time.Duration
			Docker          struct {
				Host          string
				WatchInterval time.Duration
			}
			LocalFrontendPath string
		}
	}
)

func NewCommand(log *zap.Logger) *cli.Command { //nolint:funlen
	var cmd = command{}

	var (
		addrFlag                = shared.AddrFlag
		httpPortFlag            = shared.HttpPortFlag
		httpsPortFlag           = shared.HttpsPortFlag
		httpsCertFileFlag       = shared.HttpsCertFileFlag
		httpsKeyFileFlag        = shared.HttpsKeyFileFlag
		readTimeoutFlag         = shared.ReadTimeoutFlag
		writeTimeoutFlag        = shared.WriteTimeoutFlag
		idleTimeoutFlag         = shared.IdleTimeoutFlag
		shutdownTimeoutFlag     = shared.ShutdownTimeoutFlag
		dockerHostFlag          = shared.DockerHostFlag
		dockerWatchIntervalFlag = shared.DockerWatchIntervalFlag
		localFrontendPathFlag   = shared.LocalFrontendPathFlag
	)

	cmd.c = &cli.Command{
		Name:    "start",
		Usage:   "Start HTTP/HTTPs servers",
		Aliases: []string{"server", "serve"},
		Action: func(ctx context.Context, c *cli.Command) error {
			var opt = &cmd.options

			opt.Addr = c.String(addrFlag.Name)
			opt.HTTP.Port = uint(c.Uint(httpPortFlag.Name))
			opt.HTTPS.Port = uint(c.Uint(httpsPortFlag.Name))
			httpsCertFilePath := c.String(httpsCertFileFlag.Name)
			httpsKeyFilePath := c.String(httpsKeyFileFlag.Name)
			opt.ReadTimeout = c.Duration(readTimeoutFlag.Name)
			opt.WriteTimeout = c.Duration(writeTimeoutFlag.Name)
			opt.IDLETimeout = c.Duration(idleTimeoutFlag.Name)
			opt.ShutdownTimeout = c.Duration(shutdownTimeoutFlag.Name)
			opt.Docker.Host = c.String(dockerHostFlag.Name)
			opt.Docker.WatchInterval = c.Duration(dockerWatchIntervalFlag.Name)
			opt.LocalFrontendPath = c.String(localFrontendPathFlag.Name)

			if httpsCertFilePath == "" {
				opt.HTTPS.CertFile = certs.FullChain()
			} else {
				data, err := os.ReadFile(httpsCertFilePath)
				if err != nil {
					return fmt.Errorf("failed to read certificate file: %w", err)
				}

				opt.HTTPS.CertFile = data
			}

			if httpsKeyFilePath == "" {
				opt.HTTPS.KeyFile = certs.PrivateKey()
			} else {
				data, err := os.ReadFile(httpsKeyFilePath)
				if err != nil {
					return fmt.Errorf("failed to read key file: %w", err)
				}

				opt.HTTPS.KeyFile = data
			}

			if opt.Docker.WatchInterval < time.Millisecond*100 {
				return fmt.Errorf("too small docker watch interval (%s)", opt.Docker.WatchInterval)
			}

			return cmd.Run(ctx, log)
		},
		Flags: []cli.Flag{
			&addrFlag,
			&httpPortFlag,
			&httpsPortFlag,
			&httpsCertFileFlag,
			&httpsKeyFileFlag,
			&readTimeoutFlag,
			&writeTimeoutFlag,
			&idleTimeoutFlag,
			&shutdownTimeoutFlag,
			&dockerHostFlag,
			&dockerWatchIntervalFlag,
			&localFrontendPathFlag,
		},
		Commands: []*cli.Command{
			healthcheck.NewCommand(),
		},
	}

	return cmd.c
}

// Run current command.
func (cmd *command) Run(parentCtx context.Context, log *zap.Logger) error { //nolint:funlen
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	dc, dcErr := client.NewClientWithOpts(client.WithHost(cmd.options.Docker.Host))
	if dcErr != nil {
		return fmt.Errorf("failed to create docker client: %w", dcErr)
	}

	defer func() { _ = dc.Close() }()

	var dockerState = docker.NewState(dc)

	if err := dockerState.Update(ctx); err != nil { // initial update
		return fmt.Errorf("failed to update docker state: %w", err)
	}

	var stopAutoUpdate = dockerState.StartAutoUpdate(ctx) // start auto-update
	defer stopAutoUpdate()

	// create HTTP server
	var server = appHttp.NewServer(ctx, log,
		appHttp.WithReadTimeout(cmd.options.ReadTimeout),
		appHttp.WithWriteTimeout(cmd.options.WriteTimeout),
		appHttp.WithIDLETimeout(cmd.options.IDLETimeout),
	)

	server.Register(ctx, log, dockerState, cmd.options.LocalFrontendPath)

	httpLn, httpLnErr := net.Listen("tcp", fmt.Sprintf("%s:%d", cmd.options.Addr, cmd.options.HTTP.Port))
	if httpLnErr != nil {
		return fmt.Errorf("failed to listen on HTTP port (%s:%d): %w", cmd.options.Addr, cmd.options.HTTP.Port, httpLnErr)
	}

	httpsLn, httpsLnErr := net.Listen("tcp", fmt.Sprintf("%s:%d", cmd.options.Addr, cmd.options.HTTPS.Port))
	if httpsLnErr != nil {
		return fmt.Errorf("failed to listen on HTTPS port (%s:%d): %w", cmd.options.Addr, cmd.options.HTTPS.Port, httpsLnErr)
	}

	go func() {
		log.Info("HTTP servers starting",
			zap.String("address", cmd.options.Addr),
			zap.Uint("port", cmd.options.HTTP.Port),
		)

		if err := server.StartHTTP(ctx, httpLn); err != nil {
			cancel()

			log.Error("Failed to start HTTP server", zap.Error(err))
		}
	}()

	go func() {
		log.Info("HTTPS servers starting",
			zap.String("address", cmd.options.Addr),
			zap.Uint("port", cmd.options.HTTPS.Port),
		)

		if err := server.StartHTTPs(ctx, httpsLn, cmd.options.HTTPS.CertFile, cmd.options.HTTPS.KeyFile); err != nil {
			cancel()

			log.Error("Failed to start HTTPS server", zap.Error(err))
		}
	}()

	<-ctx.Done()

	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
