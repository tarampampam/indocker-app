package start

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/app/certs"
	"gh.tarampamp.am/indocker-app/app/internal/breaker"
	"gh.tarampamp.am/indocker-app/app/internal/cli/start/healthcheck"
	"gh.tarampamp.am/indocker-app/app/internal/collector"
	"gh.tarampamp.am/indocker-app/app/internal/docker"
	"gh.tarampamp.am/indocker-app/app/internal/env"
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
		DontSendUsageStats bool
	}
)

func NewCommand(log *zap.Logger) *cli.Command { //nolint:funlen
	var cmd = command{}

	const (
		addrFlagName                   = "addr"
		httpPortFlagName               = "http-port"
		httpsPortFlagName              = "https-port"
		httpsCertFileFlagName          = "https-cert-file"
		httpsKeyFileFlagName           = "https-key-file"
		readTimeoutFlagName            = "read-timeout"
		writeTimeoutFlagName           = "write-timeout"
		idleTimeoutFlagName            = "idle-timeout"
		shutdownTimeoutFlagName        = "shutdown-timeout"
		dockerHostFlagName             = "docker-socket"
		dockerWatchIntervalFlagName    = "docker-watch-interval"
		dontSendAnonymousUsageFlagName = "dont-send-anonymous-usage"
	)

	cmd.c = &cli.Command{
		Name:    "start",
		Usage:   "Start HTTP/HTTPs servers",
		Aliases: []string{"server", "serve"},
		Action: func(c *cli.Context) error {
			var opt options

			opt.Addr = c.String(addrFlagName)
			opt.HTTP.Port = c.Uint(httpPortFlagName)
			opt.HTTPS.Port = c.Uint(httpsPortFlagName)
			httpsCertFilePath := c.String(httpsCertFileFlagName)
			httpsKeyFilePath := c.String(httpsKeyFileFlagName)
			opt.ReadTimeout = c.Duration(readTimeoutFlagName)
			opt.WriteTimeout = c.Duration(writeTimeoutFlagName)
			opt.IDLETimeout = c.Duration(idleTimeoutFlagName)
			opt.ShutdownTimeout = c.Duration(shutdownTimeoutFlagName)
			opt.Docker.Host = c.String(dockerHostFlagName)
			opt.Docker.WatchInterval = c.Duration(dockerWatchIntervalFlagName)
			opt.DontSendUsageStats = c.Bool(dontSendAnonymousUsageFlagName)

			if opt.HTTP.Port == 0 || opt.HTTP.Port > 65535 {
				return fmt.Errorf("wrong HTTP port number (%d)", opt.HTTP.Port)
			}

			if opt.HTTPS.Port == 0 || opt.HTTPS.Port > 65535 {
				return fmt.Errorf("wrong HTTP port number (%d)", opt.HTTPS.Port)
			}

			if httpsCertFilePath == "" {
				opt.HTTPS.CertFile = certs.FullChain()
			} else {
				data, err := os.ReadFile(httpsCertFilePath)
				if err != nil {
					return errors.Wrap(err, "failed to read certificate file")
				}

				opt.HTTPS.CertFile = data
			}

			if httpsKeyFilePath == "" {
				opt.HTTPS.KeyFile = certs.PrivateKey()
			} else {
				data, err := os.ReadFile(httpsKeyFilePath)
				if err != nil {
					return errors.Wrap(err, "failed to read key file")
				}

				opt.HTTPS.KeyFile = data
			}

			if opt.Docker.WatchInterval < time.Millisecond*100 {
				return fmt.Errorf("too small docker watch interval (%s)", opt.Docker.WatchInterval)
			}

			return cmd.Run(c.Context, log, opt)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    addrFlagName,
				Usage:   "server address (hostname or port; 0.0.0.0 for all interfaces)",
				Value:   "0.0.0.0",
				EnvVars: []string{env.ServerAddress.String()},
			},
			&cli.UintFlag{
				Name:    httpPortFlagName,
				Usage:   "HTTP server port",
				Value:   8080, //nolint:gomnd
				EnvVars: []string{env.HTTPPort.String()},
			},
			&cli.UintFlag{
				Name:    httpsPortFlagName,
				Usage:   "HTTPS server port",
				Value:   8443, //nolint:gomnd
				EnvVars: []string{env.HTTPSPort.String()},
			},
			&cli.StringFlag{
				Name:    httpsCertFileFlagName,
				Usage:   "TLS certificate file path (if empty, embedded certificate will be used)",
				Value:   "",
				EnvVars: []string{env.HTTPSCertFile.String()},
			},
			&cli.StringFlag{
				Name:    httpsKeyFileFlagName,
				Usage:   "TLS key file path (if empty, embedded key will be used)",
				Value:   "",
				EnvVars: []string{env.HTTPSKeyFile.String()},
			},
			&cli.DurationFlag{
				Name:    readTimeoutFlagName,
				Usage:   "maximum duration for reading the entire request, including the body (zero = no timeout)",
				Value:   time.Second * 60, //nolint:gomnd
				EnvVars: []string{env.ReadTimeout.String()},
			},
			&cli.DurationFlag{
				Name:    writeTimeoutFlagName,
				Usage:   "maximum duration before timing out writes of the response (zero = no timeout)",
				Value:   time.Second * 60, //nolint:gomnd
				EnvVars: []string{env.WriteTimeout.String()},
			},
			&cli.DurationFlag{
				Name:    idleTimeoutFlagName,
				Usage:   "maximum amount of time to wait for the next request (keep-alive, zero = no timeout)",
				Value:   time.Second * 60, //nolint:gomnd
				EnvVars: []string{env.WriteTimeout.String()},
			},
			&cli.DurationFlag{
				Name:    shutdownTimeoutFlagName,
				Usage:   "maximum duration for graceful shutdown",
				Value:   time.Second * 15, //nolint:gomnd
				EnvVars: []string{env.ShutdownTimeout.String()},
			},
			&cli.StringFlag{
				Name:    dockerHostFlagName,
				Usage:   "docker host (or path to the docker socket)",
				Value:   client.DefaultDockerHost,
				EnvVars: []string{env.DockerHost.String()},
			},
			&cli.DurationFlag{
				Name:    dockerWatchIntervalFlagName,
				Usage:   "how often to ask Docker for changes (minimum 100ms)",
				Value:   time.Second,
				EnvVars: []string{env.DockerWatchInterval.String()},
			},
			&cli.BoolFlag{
				Name:  dontSendAnonymousUsageFlagName,
				Usage: "Don't send anonymous usage statistics (please, leave it enabled, it helps us to improve the project)",
			},
		},
		Subcommands: []*cli.Command{
			healthcheck.NewCommand(),
		},
	}

	return cmd.c
}

// Run current command.
func (cmd *command) Run(parentCtx context.Context, log *zap.Logger, opt options) error { //nolint:funlen,gocognit,gocyclo,lll
	var (
		ctx, cancel = context.WithCancel(parentCtx) // main context creation
		oss         = breaker.NewOSSignals(ctx)     // OS signals listener
	)

	// subscribe for system signals
	oss.Subscribe(func(sig os.Signal) {
		log.Warn("Stopping by OS signal..", zap.String("signal", sig.String()))

		cancel()
	})

	defer func() {
		cancel()   // call the cancellation function after all
		oss.Stop() // stop system signals listening
	}()

	// load certificate
	cert, certErr := tls.X509KeyPair(opt.HTTPS.CertFile, opt.HTTPS.KeyFile)
	if certErr != nil {
		return errors.Wrap(certErr, "failed to load certificate")
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
	var dockerWatcher, dwErr = docker.NewContainersWatch(opt.Docker.WatchInterval, client.WithHost(opt.Docker.Host))
	if dwErr != nil {
		return errors.Wrap(dwErr, "failed to create docker watcher")
	}

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

	var dockerStateWatcher, swErr = docker.NewContainerStateWatch(client.WithHost(opt.Docker.Host))
	if swErr != nil {
		return errors.Wrap(swErr, "failed to create docker state watcher")
	}

	go func() { // start a docker containers state watcher in a separate goroutine
		if err := dockerStateWatcher.Watch(ctx, dockerWatcher); err != nil && !errors.Is(err, context.Canceled) {
			log.Error("Failed to watch docker containers state updates", zap.Error(err))
		}
	}()

	// register all routes
	if err := server.Register(ctx, dockerRouter, dockerStateWatcher); err != nil {
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
			errCh <- errors.Wrapf(err, "failed to listen on HTTP port (%s:%d)", opt.Addr, opt.HTTP.Port)

			return
		}

		httpsListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", opt.Addr, opt.HTTPS.Port))
		if err != nil {
			errCh <- errors.Wrapf(err, "failed to listen on HTTPS port (%s:%d)", opt.Addr, opt.HTTPS.Port)

			return
		}

		log.Info("HTTP(S) servers starting",
			zap.String("address", opt.Addr),
			zap.Uint("http port", opt.HTTP.Port),
			zap.Uint("https port", opt.HTTPS.Port),
			zap.Duration("read timeout", opt.ReadTimeout),
			zap.Duration("write timeout", opt.WriteTimeout),
			zap.Duration("idle timeout", opt.IDLETimeout),
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
		collect := cmd.runStatsCollector(ctx, log, dockerRouter, opt.Docker.Host)
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
	ctx context.Context, log *zap.Logger, dr *docker.ContainersRoute, dockerHost string,
) collector.Collector {
	const (
		initDelay, interval = 5 * time.Second, 30 * time.Minute
		mixPanelProjectID   = "e39a1eb7c7732fef947e07c4caf6a844"
	)

	// create docker ID resolver (ID from this resolver will be used to generate a unique user ID hash)
	resolver, rErr := collector.NewDockerIDResolver(ctx, client.WithHost(dockerHost))
	if rErr != nil {
		log.Debug("Failed to create docker ID resolver", zap.Error(rErr))

		// return noop collector if we failed above
		return &collector.NoopCollector{}
	}

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
