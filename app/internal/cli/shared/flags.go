package shared

import (
	"fmt"
	"net"
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
		Usage:    "IP (v4 or v6) address to listen on (0.0.0.0 to bind to all interfaces)",
		Value:    "0.0.0.0",
		Sources:  cli.EnvVars("SERVER_ADDR", "LISTEN_ADDR"),
		OnlyOnce: true,
		Config:   cli.StringConfig{TrimSpace: true},
		Validator: func(ip string) error {
			if ip == "" {
				return fmt.Errorf("missing IP address")
			}

			if net.ParseIP(ip) == nil {
				return fmt.Errorf("wrong IP address [%s] for listening", ip)
			}

			return nil
		},
	}
	HttpPortFlag = cli.UintFlag{
		Name:      "http-port",
		Category:  httpCategory,
		Usage:     "HTTP server port",
		Value:     8080, // default HTTP port number
		Sources:   cli.EnvVars("HTTP_PORT"),
		OnlyOnce:  true,
		Validator: validateTCPPortNumber,
	}
	HttpsPortFlag = cli.UintFlag{
		Name:      "https-port",
		Category:  httpCategory,
		Usage:     "HTTPS server port",
		Value:     8443, // default HTTPS port number
		Sources:   cli.EnvVars("HTTPS_PORT"),
		OnlyOnce:  true,
		Validator: validateTCPPortNumber,
	}
	ReadTimeoutFlag = cli.DurationFlag{
		Name:      "read-timeout",
		Category:  httpCategory,
		Usage:     "maximum duration for reading the entire request, including the body (zero = no timeout)",
		Value:     time.Second * 60,
		Sources:   cli.EnvVars("HTTP_READ_TIMEOUT"),
		OnlyOnce:  true,
		Validator: validateDuration("read timeout", time.Millisecond, time.Hour),
	}
	WriteTimeoutFlag = cli.DurationFlag{
		Name:      "write-timeout",
		Category:  httpCategory,
		Usage:     "maximum duration before timing out writes of the response (zero = no timeout)",
		Value:     time.Second * 60,
		Sources:   cli.EnvVars("HTTP_WRITE_TIMEOUT"),
		OnlyOnce:  true,
		Validator: validateDuration("write timeout", time.Millisecond, time.Hour),
	}
	IdleTimeoutFlag = cli.DurationFlag{
		Name:      "idle-timeout",
		Category:  httpCategory,
		Usage:     "maximum amount of time to wait for the next request (keep-alive, zero = no timeout)",
		Value:     time.Second * 60,
		Sources:   cli.EnvVars("HTTP_IDLE_TIMEOUT"),
		OnlyOnce:  true,
		Validator: validateDuration("idle timeout", time.Millisecond, time.Hour),
	}
)

const tlsCategory = "TLS"

var (
	HttpsCertFileFlag = cli.StringFlag{
		Name:      "https-cert-file",
		Category:  tlsCategory,
		Usage:     "TLS certificate file path (if empty, the certificate will be automatically resolved)",
		Sources:   cli.EnvVars("HTTPS_CERT_FILE", "TLS_CERT_FILE"),
		OnlyOnce:  true,
		Config:    cli.StringConfig{TrimSpace: true},
		Validator: validateFilePath("certificate file", true),
	}
	HttpsKeyFileFlag = cli.StringFlag{
		Name:      "https-key-file",
		Category:  tlsCategory,
		Usage:     "TLS key file path (if empty, the key will be automatically resolved)",
		Sources:   cli.EnvVars("HTTPS_KEY_FILE", "TLS_KEY_FILE"),
		OnlyOnce:  true,
		Config:    cli.StringConfig{TrimSpace: true},
		Validator: validateFilePath("key file", true),
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
)

var (
	ShutdownTimeoutFlag = cli.DurationFlag{
		Name:      "shutdown-timeout",
		Usage:     "maximum duration for graceful shutdown",
		Value:     time.Second * 15,
		Sources:   cli.EnvVars("SHUTDOWN_TIMEOUT"),
		OnlyOnce:  true,
		Validator: validateDuration("shutdown timeout", time.Millisecond, time.Minute),
	}
)

// validateTCPPortNumber validates the given TCP port number.
func validateTCPPortNumber(port uint64) error {
	if port == 0 || port > 65535 {
		return fmt.Errorf("wrong TCP port number [%d]", port)
	}

	return nil
}

// validateDuration returns a validator for the given duration.
func validateDuration(name string, minValue, maxValue time.Duration) func(d time.Duration) error {
	return func(d time.Duration) error {
		switch {
		case d < 0:
			return fmt.Errorf("negative %s (%s)", name, d)
		case d < minValue:
			return fmt.Errorf("too small %s (%s)", name, d)
		case d > maxValue:
			return fmt.Errorf("too big %s (%s)", name, d)
		}

		return nil
	}
}

func validateFilePath(name string, isOptional bool) func(s string) error {
	return func(s string) error {
		if isOptional && s == "" {
			return nil
		}

		if s == "" {
			return fmt.Errorf("missing %s", name)
		}

		if stat, err := os.Stat(s); err != nil {
			return fmt.Errorf("failed to find %s (%s): %w", name, s, err)
		} else if stat.IsDir() {
			return fmt.Errorf("%s is a directory", name)
		}

		return nil
	}
}
