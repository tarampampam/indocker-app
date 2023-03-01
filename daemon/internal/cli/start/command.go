package start

import (
	"context"
	"crypto/tls"
	stdErrors "errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/daemon/certs"
	"gh.tarampamp.am/indocker-app/daemon/internal/breaker"
	"gh.tarampamp.am/indocker-app/daemon/internal/env"
	appHttp "gh.tarampamp.am/indocker-app/daemon/internal/http"
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

		Docker struct {
			SocketPath string
		}
	}
)

func NewCommand(log *zap.Logger) *cli.Command {
	var cmd = command{}

	const (
		addrFlagName             = "addr"
		httpPortFlagName         = "http-port"
		httpsPortFlagName        = "https-port"
		httpsCertFileFlagName    = "https-cert-file"
		httpsKeyFileFlagName     = "https-key-file"
		readTimeoutFlagName      = "read-timeout"
		writeTimeoutFlagName     = "write-timeout"
		idleTimeoutFlagName      = "idle-timeout"
		shutdownTimeoutFlagName  = "shutdown-timeout"
		dockerSocketPathFlagName = "docker-socket"
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
			opt.Docker.SocketPath = c.String(dockerSocketPathFlagName)

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

			if stat, err := os.Stat(opt.Docker.SocketPath); err != nil {
				return errors.Wrapf(err, "failed to access the docker socket %s", opt.Docker.SocketPath)
			} else if stat.Mode().Type() != os.ModeSocket {
				return fmt.Errorf("%s is not a socket", opt.Docker.SocketPath)
			}

			return cmd.Run(c.Context, log, opt)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    addrFlagName,
				Usage:   "server address (hostname or port)",
				Value:   "0.0.0.0",
				EnvVars: []string{env.ServerAddress.String()},
			},
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
				Value:   time.Second * 60,
				EnvVars: []string{env.ReadTimeout.String()},
			},
			&cli.DurationFlag{
				Name:    writeTimeoutFlagName,
				Usage:   "maximum duration before timing out writes of the response (zero = no timeout)",
				Value:   time.Second * 60,
				EnvVars: []string{env.WriteTimeout.String()},
			},
			&cli.DurationFlag{
				Name:    idleTimeoutFlagName,
				Usage:   "maximum amount of time to wait for the next request (keep-alive, zero = no timeout)",
				Value:   time.Second * 60,
				EnvVars: []string{env.WriteTimeout.String()},
			},
			&cli.DurationFlag{
				Name:    shutdownTimeoutFlagName,
				Usage:   "maximum duration for graceful shutdown",
				Value:   time.Second * 15,
				EnvVars: []string{env.ShutdownTimeout.String()},
			},
			&cli.StringFlag{
				Name:    dockerSocketPathFlagName,
				Usage:   "path to the docker socket",
				Value:   "/var/run/docker.sock",
				EnvVars: []string{env.DockerSocketPath.String()},
			},
		},
	}

	return cmd.c
}

// Run current command.
func (cmd *command) Run(parentCtx context.Context, log *zap.Logger, opt options) error {
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
		appHttp.WithTLSConfig(&tls.Config{Certificates: []tls.Certificate{cert}}),
		appHttp.WithReadTimeout(opt.ReadTimeout),
		appHttp.WithWriteTimeout(opt.WriteTimeout),
		appHttp.WithIDLETimeout(opt.IDLETimeout),
	)

	startingErrCh := make(chan error, 1) // channel for server starting error
	defer close(startingErrCh)

	// start HTTP server in separate goroutine
	go func(errCh chan<- error) {
		log.Info("HTTP(S) servers starting",
			zap.String("address", opt.Addr),
			zap.Uint("http port", opt.HTTP.Port),
			zap.Uint("https port", opt.HTTPS.Port),
			zap.Duration("read timeout", opt.ReadTimeout),
			zap.Duration("write timeout", opt.WriteTimeout),
			zap.Duration("idle timeout", opt.IDLETimeout),
		)

		var startingError = server.Start("0.0.0.0", uint16(opt.HTTP.Port), uint16(opt.HTTPS.Port))
		if startingError != nil && !stdErrors.Is(startingError, http.ErrServerClosed) {
			errCh <- startingError
		}

		log.Info("HTTP(S) servers stopped")
	}(startingErrCh)

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
