package shared

import (
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/urfave/cli/v3"
)

const httpCategory = "HTTP"

var (
	AddrFlag = cli.StringFlag{
		Name:     "addr",
		Category: httpCategory,
		Usage:    "server address (hostname or port; 0.0.0.0 for all interfaces)",
		Value:    "0.0.0.0",
		Sources:  cli.EnvVars("SERVER_ADDR"),
		OnlyOnce: true,
		Config:   cli.StringConfig{TrimSpace: true},
	}
	HttpPortFlag = cli.UintFlag{
		Name:     "http-port",
		Category: httpCategory,
		Usage:    "HTTP server port",
		Value:    8080,
		Sources:  cli.EnvVars("HTTP_PORT"),
		OnlyOnce: true,
		Validator: func(port uint64) error {
			if port == 0 || port > 65535 {
				return fmt.Errorf("wrong TCP port number [%d]", port)
			}

			return nil
		},
	}
	HttpsPortFlag = cli.UintFlag{
		Name:     "https-port",
		Category: httpCategory,
		Usage:    "HTTPS server port",
		Value:    8443,
		Sources:  cli.EnvVars("HTTPS_PORT"),
		OnlyOnce: true,
		Validator: func(port uint64) error {
			if port == 0 || port > 65535 {
				return fmt.Errorf("wrong TCP port number [%d]", port)
			}

			return nil
		},
	}
	ReadTimeoutFlag = cli.DurationFlag{
		Name:     "read-timeout",
		Category: httpCategory,
		Usage:    "maximum duration for reading the entire request, including the body (zero = no timeout)",
		Value:    time.Second * 60,
		Sources:  cli.EnvVars("READ_TIMEOUT"),
		OnlyOnce: true,
		Validator: func(d time.Duration) error {
			const minValue, maxValue = time.Millisecond, time.Hour

			switch {
			case d < 0:
				return fmt.Errorf("negative read timeout (%s)", d)
			case d < minValue:
				return fmt.Errorf("too small read timeout (%s)", d)
			case d > maxValue:
				return fmt.Errorf("too big read timeout (%s)", d)
			}

			return nil
		},
	}
	WriteTimeoutFlag = cli.DurationFlag{
		Name:     "write-timeout",
		Category: httpCategory,
		Usage:    "maximum duration before timing out writes of the response (zero = no timeout)",
		Value:    time.Second * 60,
		Sources:  cli.EnvVars("WRITE_TIMEOUT"),
		OnlyOnce: true,
		Validator: func(d time.Duration) error {
			const minValue, maxValue = time.Millisecond, time.Hour

			switch {
			case d < 0:
				return fmt.Errorf("negative write timeout (%s)", d)
			case d < minValue:
				return fmt.Errorf("too small write timeout (%s)", d)
			case d > maxValue:
				return fmt.Errorf("too big write timeout (%s)", d)
			}

			return nil
		},
	}
	IdleTimeoutFlag = cli.DurationFlag{
		Name:     "idle-timeout",
		Category: httpCategory,
		Usage:    "maximum amount of time to wait for the next request (keep-alive, zero = no timeout)",
		Value:    time.Second * 60,
		Sources:  cli.EnvVars("IDLE_TIMEOUT"),
		OnlyOnce: true,
		Validator: func(d time.Duration) error {
			const minValue, maxValue = time.Millisecond, time.Hour

			switch {
			case d < 0:
				return fmt.Errorf("negative idle timeout (%s)", d)
			case d < minValue:
				return fmt.Errorf("too small idle timeout (%s)", d)
			case d > maxValue:
				return fmt.Errorf("too big idle timeout (%s)", d)
			}

			return nil
		},
	}
)

const tlsCategory = "TLS"

var (
	HttpsCertFileFlag = cli.StringFlag{
		Name:     "https-cert-file",
		Category: tlsCategory,
		Usage:    "TLS certificate file path (if empty, embedded certificate will be used)",
		Value:    "",
		Sources:  cli.EnvVars("HTTPS_CERT_FILE", "TLS_CERT_FILE"),
		OnlyOnce: true,
		Config:   cli.StringConfig{TrimSpace: true},
		Validator: func(s string) error {
			if s == "" {
				return nil // use embedded certificate
			}

			if _, err := os.Stat(s); err != nil {
				return fmt.Errorf("failed to find certificate file (%s): %w", s, err)
			}

			return nil
		},
	}
	HttpsKeyFileFlag = cli.StringFlag{
		Name:     "https-key-file",
		Category: tlsCategory,
		Usage:    "TLS key file path (if empty, embedded key will be used)",
		Value:    "",
		Sources:  cli.EnvVars("HTTPS_KEY_FILE", "TLS_KEY_FILE"),
		OnlyOnce: true,
		Config:   cli.StringConfig{TrimSpace: true},

		Validator: func(s string) error {
			if s == "" {
				return nil // use embedded key
			}

			if _, err := os.Stat(s); err != nil {
				return fmt.Errorf("failed to find key file (%s): %w", s, err)
			}

			return nil
		},
	}
)

const dockerCategory = "DOCKER"

var (
	DockerHostFlag = cli.StringFlag{
		Name:     "docker-socket",
		Category: dockerCategory,
		Usage:    "path to the docker socket (or docker host)",
		Value:    client.DefaultDockerHost,
		Sources:  cli.EnvVars("DOCKER_SOCKET", "DOCKER_HOST"),
		OnlyOnce: true,
		Config:   cli.StringConfig{TrimSpace: true},
		Validator: func(s string) error {
			if s == "" {
				return fmt.Errorf("missing docker socket path")
			}

			return nil
		},
	}
	DockerWatchIntervalFlag = cli.DurationFlag{
		Name:     "docker-watch-interval",
		Category: dockerCategory,
		Usage:    "how often to ask Docker for changes (minimum 100ms)",
		Value:    time.Second,
		Sources:  cli.EnvVars("DOCKER_WATCH_INTERVAL"),
		OnlyOnce: true,
		Validator: func(d time.Duration) error {
			const minValue, maxValue = time.Millisecond * 100, time.Minute

			switch {
			case d < 0:
				return fmt.Errorf("negative docker watch interval (%s)", d)
			case d < minValue:
				return fmt.Errorf("too small docker watch interval (%s)", d)
			case d > maxValue:
				return fmt.Errorf("too big docker watch interval (%s)", d)
			}

			return nil
		},
	}
)

var (
	ShutdownTimeoutFlag = cli.DurationFlag{
		Name:     "shutdown-timeout",
		Usage:    "maximum duration for graceful shutdown",
		Value:    time.Second * 15,
		Sources:  cli.EnvVars("SHUTDOWN_TIMEOUT"),
		OnlyOnce: true,
		Validator: func(d time.Duration) error {
			const minValue, maxValue = time.Millisecond, time.Minute

			switch {
			case d < 0:
				return fmt.Errorf("negative shutdown timeout (%s)", d)
			case d < minValue:
				return fmt.Errorf("too small shutdown timeout (%s)", d)
			case d > maxValue:
				return fmt.Errorf("too big shutdown timeout (%s)", d)
			}

			return nil
		},
	}
)

var (
	LocalFrontendPathFlag = cli.StringFlag{
		Name:     "local-frontend-path",
		Usage:    "path to the local frontend (if empty, embedded frontend will be used; useful for development)",
		OnlyOnce: true,
		Config:   cli.StringConfig{TrimSpace: true},
		Validator: func(s string) error {
			if s == "" {
				return nil // use embedded frontend
			}

			if stat, err := os.Stat(s); err != nil {
				return fmt.Errorf("failed to find local frontend path (%s): %w", s, err)
			} else if !stat.IsDir() {
				return fmt.Errorf("local frontend path is not a directory (%s)", s)
			}

			return nil
		},
	}
)
