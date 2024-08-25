package start

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/app/internal/cert"
	"gh.tarampamp.am/indocker-app/app/internal/cli/shared"
	"gh.tarampamp.am/indocker-app/app/internal/cli/start/healthcheck"
	"gh.tarampamp.am/indocker-app/app/internal/docker"
	appHttp "gh.tarampamp.am/indocker-app/app/internal/http"
)

type (
	command struct {
		c *cli.Command

		options struct {
			addr string // IP (v4 or v6) address to listen on
			http struct {
				tcpPort uint16 // TCP port number for HTTP server
			}
			https struct {
				tcpPort uint16           // TCP port number for HTTPS server
				cert    *tls.Certificate // TLS certificate to use
			}
			timeouts struct {
				httpRead, httpWrite, httpIdle time.Duration // timeouts for HTTP(s) servers
				shutdown                      time.Duration // maximum amount of time to wait for the server to stop
			}
			docker struct {
				host string // Docker daemon host (e.g. "unix:///var/run/docker.sock")
			}
			frontend struct {
				distPath string // path to the frontend distribution to serve instead of the built-in one
			}
		}
	}
)

func NewCommand(log *zap.Logger) *cli.Command { //nolint:funlen
	var cmd command

	var (
		addrFlag              = shared.AddrFlag
		httpPortFlag          = shared.HttpPortFlag
		httpsPortFlag         = shared.HttpsPortFlag
		httpsCertFileFlag     = shared.HttpsCertFileFlag
		httpsKeyFileFlag      = shared.HttpsKeyFileFlag
		readTimeoutFlag       = shared.ReadTimeoutFlag
		writeTimeoutFlag      = shared.WriteTimeoutFlag
		idleTimeoutFlag       = shared.IdleTimeoutFlag
		shutdownTimeoutFlag   = shared.ShutdownTimeoutFlag
		dockerHostFlag        = shared.DockerHostFlag
		localFrontendPathFlag = shared.LocalFrontendPathFlag
	)

	cmd.c = &cli.Command{
		Name:    "start",
		Usage:   "Start HTTP/HTTPs servers",
		Aliases: []string{"server", "serve"},
		Action: func(ctx context.Context, c *cli.Command) error {
			var opt = &cmd.options

			// set options
			opt.addr = c.String(addrFlag.Name)
			opt.http.tcpPort = uint16(c.Uint(httpPortFlag.Name))   //nolint:gosec
			opt.https.tcpPort = uint16(c.Uint(httpsPortFlag.Name)) //nolint:gosec
			opt.timeouts.httpRead = c.Duration(readTimeoutFlag.Name)
			opt.timeouts.httpWrite = c.Duration(writeTimeoutFlag.Name)
			opt.timeouts.httpIdle = c.Duration(idleTimeoutFlag.Name)
			opt.timeouts.shutdown = c.Duration(shutdownTimeoutFlag.Name)
			opt.docker.host = c.String(dockerHostFlag.Name)
			opt.frontend.distPath = c.String(localFrontendPathFlag.Name)

			// if user provided both certificate and key files, use them
			if crt, key := c.String(httpsCertFileFlag.Name), c.String(httpsKeyFileFlag.Name); crt != "" && key != "" { //nolint:nestif,lll
				crtData, err := os.ReadFile(crt) // read certificate file
				if err != nil {
					return fmt.Errorf("failed to read certificate file: %w", err)
				}

				keyData, err := os.ReadFile(key) // read key file
				if err != nil {
					return fmt.Errorf("failed to read key file: %w", err)
				}

				if localCert, localCertErr := tls.X509KeyPair(crtData, keyData); localCertErr != nil {
					return fmt.Errorf("failed to load TLS certificate: %w", localCertErr)
				} else {
					opt.https.cert = &localCert // set the user-provided certificate
				}
			} else { // otherwise, try to get the certificate and key using our resolver
				log.Debug("Getting TLS certificate and key...")

				if remoteCert, remoteCertErr := cert.NewResolver().Resolve(ctx); remoteCertErr != nil {
					return fmt.Errorf("failed to get TLS certificate: %w", remoteCertErr)
				} else {
					log.Debug("TLS certificate and key are successfully loaded")

					opt.https.cert = remoteCert // set the resolved certificate
				}
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

	// create Docker client
	dc, dcClose, dcErr := cmd.makeDockerClient(ctx, log.Named("docker"))
	if dcErr != nil {
		return dcErr
	} else {
		defer dcClose()
	}

	// create Docker state watcher
	dockerState, stateClose, stateErr := cmd.makeDockerStateWatcher(ctx, log.Named("docker.state"), dc)
	if stateErr != nil {
		return stateErr
	} else {
		defer stateClose()
	}

	// create HTTP server
	var server = appHttp.NewServer(ctx, log,
		appHttp.WithReadTimeout(cmd.options.timeouts.httpRead),
		appHttp.WithWriteTimeout(cmd.options.timeouts.httpWrite),
		appHttp.WithIDLETimeout(cmd.options.timeouts.httpIdle),
	).Register(
		ctx,
		log,
		dockerState,
		cmd.options.frontend.distPath,
	)

	server.ShutdownTimeout = cmd.options.timeouts.shutdown // set shutdown timeout

	// open HTTP port
	httpLn, httpLnErr := net.Listen("tcp", fmt.Sprintf("%s:%d", cmd.options.addr, cmd.options.http.tcpPort))
	if httpLnErr != nil {
		return fmt.Errorf("HTTP port error (%s:%d): %w", cmd.options.addr, cmd.options.http.tcpPort, httpLnErr)
	}

	// open HTTPS port
	httpsLn, httpsLnErr := net.Listen("tcp", fmt.Sprintf("%s:%d", cmd.options.addr, cmd.options.https.tcpPort))
	if httpsLnErr != nil {
		return fmt.Errorf("HTTPS port error (%s:%d): %w", cmd.options.addr, cmd.options.https.tcpPort, httpsLnErr)
	}

	// try to determine if the app is running inside a Docker container
	var iAmInsideDocker = cmd.isInsideDocker()

	const monitorDomain = "monitor.indocker.app"

	// start HTTP server in separate goroutine
	go func() {
		defer func() { _ = httpLn.Close() }()

		const (
			defaultHttpPort = uint16(80)
			hstsNote        = "please note browsers open every domain in the .app zone using HTTPS due to the HSTS policy"
		)

		log.Info("HTTP server starting",
			zap.String("address", cmd.options.addr),
			zap.Uint16("port", cmd.options.http.tcpPort),
			zap.String("open", func() string {
				if iAmInsideDocker || cmd.options.http.tcpPort == defaultHttpPort {
					return fmt.Sprintf("http://%s (%s)", monitorDomain, hstsNote)
				}

				return fmt.Sprintf("http://%s:%d (%s)", monitorDomain, cmd.options.http.tcpPort, hstsNote)
			}()),
		)

		if err := server.StartHTTP(ctx, httpLn); err != nil {
			cancel() // cancel the context on error (this is critical for us)

			log.Error("Failed to start HTTP server", zap.Error(err))
		} else {
			log.Debug("HTTP server stopped")
		}
	}()

	// start HTTPS server in separate goroutine
	go func() {
		defer func() { _ = httpsLn.Close() }()

		const defaultHttpsPort uint16 = 443

		log.Info("HTTPS server starting",
			zap.String("address", cmd.options.addr),
			zap.Uint16("port", cmd.options.https.tcpPort),
			zap.String("open", func() string {
				if iAmInsideDocker || cmd.options.https.tcpPort == defaultHttpsPort {
					return fmt.Sprintf("https://%s", monitorDomain)
				}

				return fmt.Sprintf("https://%s:%d", monitorDomain, cmd.options.https.tcpPort)
			}()),
		)

		if err := server.StartHTTPs(ctx, httpsLn, *cmd.options.https.cert); err != nil {
			cancel() // cancel the context on error (this is critical for us)

			log.Error("Failed to start HTTPS server", zap.Error(err))
		} else {
			log.Debug("HTTPS server stopped")
		}
	}()

	// here, we are blocking until the context is canceled. this will occur when the user sends a signal to stop
	// the app by pressing Ctrl+C, terminating the process, or if the HTTP/HTTPS server fails to start
	<-ctx.Done()

	// if the context contains an error, and it's not a cancellation error, return it
	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func (cmd *command) makeDockerClient(ctx context.Context, log *zap.Logger) (*client.Client, func(), error) {
	dc, dcErr := client.NewClientWithOpts(client.WithHost(cmd.options.docker.host))
	if dcErr != nil {
		return nil, func() {}, fmt.Errorf("failed to create docker client: %w", dcErr)
	}

	if info, err := dc.Info(ctx); err != nil { // check connection to the Docker daemon
		return nil, func() {}, fmt.Errorf("failed to get docker info: %w", err)
	} else {
		log.Debug("Connected to the Docker daemon", zap.String("docker version", info.ServerVersion))
	}

	return dc, sync.OnceFunc(func() { _ = dc.Close(); log.Debug("Disconnected from the Docker daemon") }), nil
}

func (cmd *command) makeDockerStateWatcher(
	ctx context.Context,
	log *zap.Logger,
	dc *client.Client,
) (*docker.State, func(), error) {
	var state = docker.NewState(dc)

	if err := state.Update(ctx); err != nil { // initial update
		return nil, func() {}, fmt.Errorf("failed to update docker state: %w", err)
	}

	var (
		routesSub, closeRoutesSub = state.SubscribeForRoutingUpdates() // subscribe for routing updates
		stopAutoUpdate            = state.StartAutoUpdate(ctx)         // start auto-update

		logRoutes = func(msg string, routes map[string]map[string]url.URL) {
			var currentRoutes = make(map[string][]string, len(routes))

			// format routes map
			for domain, urls := range routes {
				for _, u := range urls {
					currentRoutes[domain] = append(currentRoutes[domain], u.String())
				}
			}

			log.Info(msg, zap.Any("routes", currentRoutes))
		}
	)

	// run a goroutine to log routing updates
	go func() {
		for {
			select {
			case routes, isOpened := <-routesSub:
				if !isOpened {
					return
				}

				logRoutes("Docker routing updated", routes)
			case <-ctx.Done():
				return
			}
		}
	}()

	if routes := state.AllContainerURLs(); len(routes) > 0 {
		logRoutes("Initial Docker routing", routes)
	}

	return state, sync.OnceFunc(func() { stopAutoUpdate(); closeRoutesSub() }), nil
}

func (*command) isInsideDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	return false
}
