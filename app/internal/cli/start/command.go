package start

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/client"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/app/certs"
	"gh.tarampamp.am/indocker-app/app/internal/cli/shared/flags"
	"gh.tarampamp.am/indocker-app/app/internal/cli/start/healthcheck"
	"gh.tarampamp.am/indocker-app/app/internal/collector"
	"gh.tarampamp.am/indocker-app/app/internal/docker"
	appHttp "gh.tarampamp.am/indocker-app/app/internal/http"
	"gh.tarampamp.am/indocker-app/app/internal/version"
)

type (
	command struct {
		c *cli.Command
	}

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
		Dashboard struct {
			Domain               string
			DontUseEmbeddedFront bool
		}
		DontSendUsageStats bool
	}
)

func NewCommand(log *zap.Logger) *cli.Command { //nolint:funlen
	var cmd = command{}

	var (
		addrFlag                   = flags.AddrFlag
		httpPortFlag               = flags.HttpPortFlag
		httpsPortFlag              = flags.HttpsPortFlag
		httpsCertFileFlag          = flags.HttpsCertFileFlag
		httpsKeyFileFlag           = flags.HttpsKeyFileFlag
		readTimeoutFlag            = flags.ReadTimeoutFlag
		writeTimeoutFlag           = flags.WriteTimeoutFlag
		idleTimeoutFlag            = flags.IdleTimeoutFlag
		shutdownTimeoutFlag        = flags.ShutdownTimeoutFlag
		dockerHostFlag             = flags.DockerHostFlag
		dockerWatchIntervalFlag    = flags.DockerWatchIntervalFlag
		dashboardDomainFlag        = flags.DashboardDomainFlag
		dontUseEmbeddedFrontFlag   = flags.DontUseEmbeddedFrontFlag
		dontSendAnonymousUsageFlag = flags.DontSendAnonymousUsageFlag
	)

	cmd.c = &cli.Command{
		Name:    "start",
		Usage:   "Start HTTP/HTTPs servers",
		Aliases: []string{"server", "serve"},
		Action: func(ctx context.Context, c *cli.Command) error {
			var opt options

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
			opt.Dashboard.Domain = c.String(dashboardDomainFlag.Name)
			opt.Dashboard.DontUseEmbeddedFront = c.Bool(dontUseEmbeddedFrontFlag.Name)
			opt.DontSendUsageStats = c.Bool(dontSendAnonymousUsageFlag.Name)

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

			return cmd.Run(ctx, log, opt)
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
			&dashboardDomainFlag,
			&dontUseEmbeddedFrontFlag,
			&dontSendAnonymousUsageFlag,
		},
		Commands: []*cli.Command{
			healthcheck.NewCommand(),
		},
	}

	return cmd.c
}

// Run current command.
func (cmd *command) Run(parentCtx context.Context, log *zap.Logger, opt options) error { //nolint:funlen,gocyclo
	var ctx, cancel = context.WithCancel(parentCtx)

	defer cancel()

	dc, dcErr := client.NewClientWithOpts(client.WithHost(opt.Docker.Host))
	if dcErr != nil {
		return fmt.Errorf("failed to create docker client: %w", dcErr)
	}

	defer func() { _ = dc.Close() }()

	// load certificate
	cert, certErr := tls.X509KeyPair(opt.HTTPS.CertFile, opt.HTTPS.KeyFile)
	if certErr != nil {
		return fmt.Errorf("failed to load certificate: %w", certErr)
	}

	// create HTTP server
	var server = appHttp.NewServer(ctx, log,
		&tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
		appHttp.WithReadTimeout(opt.ReadTimeout),
		appHttp.WithWriteTimeout(opt.WriteTimeout),
		appHttp.WithIDLETimeout(opt.IDLETimeout),
	)

	// prepare dependencies for the http server and register routes
	var dockerWatcher = docker.NewContainersWatch(opt.Docker.WatchInterval, dc)

	go func() { // start a docker containers watcher in a separate goroutine
		if err := dockerWatcher.Watch(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Error("Failed to watch docker containers", zap.Error(err))

			cancel() // this is a critical error for us
		}
	}()

	var dockerRouter = docker.NewContainersRoute()

	go func() { // start a docker routes watching in a separate goroutine
		if err := dockerRouter.Watch(ctx, dockerWatcher); err != nil && !errors.Is(err, context.Canceled) {
			log.Error("Failed to start watching for the docker router", zap.Error(err))

			cancel() // this is a critical error for us
		}
	}()

	var dockerStateWatcher = docker.NewContainerStateWatch(dc)

	go func() { // start a docker containers state watcher in a separate goroutine
		if err := dockerStateWatcher.Watch(ctx, dockerWatcher); err != nil && !errors.Is(err, context.Canceled) {
			log.Error("Failed to watch docker containers state updates", zap.Error(err))
		}
	}()

	// register all routes
	if err := server.Register(ctx,
		dockerRouter,
		dockerStateWatcher,
		opt.Dashboard.Domain,
		opt.Dashboard.DontUseEmbeddedFront,
	); err != nil {
		return err
	}

	startingErrCh := make(chan error, 1) // channel for server starting error
	defer close(startingErrCh)

	// start HTTP server in separate goroutine
	go func(errCh chan<- error) {
		if ctx.Err() != nil { // check if the context is already canceled
			errCh <- ctx.Err()

			return
		}

		httpListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", opt.Addr, opt.HTTP.Port))
		if err != nil {
			errCh <- fmt.Errorf("failed to listen on HTTP port (%s:%d): %w", opt.Addr, opt.HTTP.Port, err)

			return
		}

		httpsListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", opt.Addr, opt.HTTPS.Port))
		if err != nil {
			errCh <- fmt.Errorf("failed to listen on HTTPS port (%s:%d): %w", opt.Addr, opt.HTTPS.Port, err)

			return
		}

		log.Info("HTTP(S) servers starting",
			zap.String("address", opt.Addr),
			zap.Uint("http_port", opt.HTTP.Port),
			zap.Uint("https_port", opt.HTTPS.Port),
			zap.Duration("read_timeout", opt.ReadTimeout),
			zap.Duration("write_timeout", opt.WriteTimeout),
			zap.Duration("idle_timeout", opt.IDLETimeout),
			zap.String("dashboard_domain", opt.Dashboard.Domain),
		)

		var startingError = server.Start(httpListener, httpsListener)
		if startingError != nil && !errors.Is(startingError, http.ErrServerClosed) {
			errCh <- startingError
		}

		log.Info("HTTP(S) servers stopped")
	}(startingErrCh)

	if opt.DontSendUsageStats {
		log.Warn("Anonymous usage statistics sending is disabled",
			zap.String("please", "leave it enabled, it helps us to improve the project"),
		)
	} else {
		collect := cmd.runStatsCollector(ctx, log, dockerRouter, dc)
		defer collect.Stop()
	}

	// and wait for..
	select {
	case err := <-startingErrCh: // ..server starting error
		return err

	case <-ctx.Done(): // ..or context cancellation
		// create context for server graceful shutdown
		ctxShutdown, ctxCancelShutdown := context.WithTimeout(context.Background(), opt.ShutdownTimeout)
		defer ctxCancelShutdown()

		// stop the server using created context above
		if err := server.Stop(ctxShutdown); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *command) runStatsCollector(
	ctx context.Context, log *zap.Logger, dr *docker.ContainersRoute, dc *client.Client,
) collector.Collector {
	const (
		initDelay, interval = 5 * time.Second, 30 * time.Minute
		mixPanelProjectID   = "e39a1eb7c7732fef947e07c4caf6a844"
	)

	// create docker ID resolver (ID from this resolver will be used to generate a unique user ID hash)
	resolver := collector.NewDockerIDResolver(ctx, dc)

	// create a new collector
	var collect = collector.NewCollector(ctx, log, initDelay, interval,
		collector.NewMixPanelSender(mixPanelProjectID, version.Version()),
		resolver,
	)

	// schedule initial event
	collect.Schedule(collector.Event{Name: "app_run"})

	go func() { // schedule heartbeat event sending every 14 minutes
		const hbInterval = 14 * time.Minute

		var t = time.NewTicker(hbInterval)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				collect.Schedule(collector.Event{
					Name: "app_heartbeat",
					Properties: map[string]string{
						// routes count is used to understanding how actively the project is used
						"routes_count": strconv.Itoa(dr.RoutesCount()),
					}},
				)

			case <-ctx.Done():
				return
			}
		}
	}()

	return collect
}
