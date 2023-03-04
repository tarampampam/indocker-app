// Package env contains all about environment variables, that can be used by current application.
package env

import "os"

type envVariable string

const (
	LogLevel  envVariable = "LOG_LEVEL"  // logging level
	LogFormat envVariable = "LOG_FORMAT" // logging format (json|console)

	ServerAddress       envVariable = "SERVER_ADDR"           // server address (hostname or port)
	HTTPPort            envVariable = "HTTP_PORT"             // HTTP server port
	HTTPSPort           envVariable = "HTTPS_PORT"            // HTTPS server port
	HTTPSCertFile       envVariable = "HTTPS_CERT_FILE"       // HTTPS certificate file path
	HTTPSKeyFile        envVariable = "HTTPS_KEY_FILE"        // HTTPS certificate key file path
	ReadTimeout         envVariable = "READ_TIMEOUT"          // Read timeout
	WriteTimeout        envVariable = "WRITE_TIMEOUT"         // Write timeout
	ShutdownTimeout     envVariable = "SHUTDOWN_TIMEOUT"      // Shutdown timeout
	DockerHost          envVariable = "DOCKER_HOST"           // Docker host (or socket path)
	DockerWatchInterval envVariable = "DOCKER_WATCH_INTERVAL" // Docker watch interval
)

// String returns environment variable name in the string representation.
func (e envVariable) String() string { return string(e) }

// Lookup retrieves the value of the environment variable. If the variable is present in the environment the value
// (which may be empty) is returned and the boolean is true. Otherwise, the returned value will be empty and the
// boolean will be false.
func (e envVariable) Lookup() (string, bool) { return os.LookupEnv(string(e)) }
