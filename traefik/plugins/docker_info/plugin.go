package docker_info

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

// Config the plugins configuration.
type Config struct {
	UrlPrefix  string `json:"urlPrefix" yaml:"urlPrefix"`
	SocketPath string `json:"socketPath" yaml:"socketPath"`
}

// CreateConfig creates the default plugins configuration.
func CreateConfig() *Config {
	return &Config{
		UrlPrefix:  "/docker-info",
		SocketPath: "/var/run/docker.sock",
	}
}

// Plugin a plugins.
type Plugin struct {
	next   http.Handler
	name   string
	config *Config

	ctx context.Context
}

// New creates a new plugins.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if stat, err := os.Stat(config.SocketPath); err != nil {
		return nil, errors.New(name + ": " + err.Error())
	} else if stat.Mode().Type() != os.ModeSocket {
		return nil, errors.New(name + ": is not a socket")
	}

	return &Plugin{
		config: config,
		next:   next,
		name:   name,
		ctx:    ctx,
	}, nil
}

var errRouteNotFound = errors.New("route not found")

// ServeHTTP implements http.Handler.
func (p *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(req.URL.Path, p.config.UrlPrefix) {
		var routeErr = errRouteNotFound

		switch route := req.URL.Path[len(p.config.UrlPrefix):]; route { // the simplest router in the world :D
		case "/containers/list":
			routeErr = p.DockerContainersList(rw, req)
		}

		if routeErr != nil {
			p.fail(rw, routeErr)
		}

		return
	}

	p.next.ServeHTTP(rw, req)
}

// DockerContainersList returns a list with all containers.
func (p *Plugin) DockerContainersList(rw http.ResponseWriter, _ *http.Request) error {
	var containers []Container
	if err := p.requestDocker(p.ctx, http.MethodGet, "containers/json", http.NoBody, &containers); err != nil {
		return err
	}

	type container struct {
		Names []string `json:"names"`
	}

	// prepare a response
	var resp = make([]container, len(containers))

	// fill it
	for i, c := range containers {
		resp[i].Names = c.Names
	}

	p.json(rw, http.StatusOK, resp)

	return nil
}

// dockerClient creates a new docker client.
func (p *Plugin) dockerClient() (*http.Client, func(), error) {
	var conn, cErr = (&net.Dialer{}).DialContext(p.ctx, "unix", p.config.SocketPath)
	if cErr != nil {
		return nil, nil, cErr
	}

	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) { return conn, nil },
		},
	}, func() { _ = conn.Close() }, nil
}

// requestDocker executes a request against the docker socket.
func (p *Plugin) requestDocker(ctx context.Context, method, path string, body io.Reader, out any) error {
	c, end, err := p.dockerClient()
	if err != nil {
		return err
	}

	defer end()

	// create a new request
	req, err := http.NewRequestWithContext(ctx, method, "http://docker/"+path, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// execute the request
	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	// decode the response
	return json.NewDecoder(resp.Body).Decode(&out)
}

// json writes a json response.
func (p *Plugin) json(rw http.ResponseWriter, status int, v any) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	rw.WriteHeader(status)

	if err := json.NewEncoder(rw).Encode(v); err != nil {
		p.fail(rw, err)
	}
}

// fail writes an error message to the response writer.
func (p *Plugin) fail(rw http.ResponseWriter, err error) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	rw.WriteHeader(http.StatusInternalServerError)

	var msg string

	if err != nil {
		msg = err.Error()
	} else {
		msg = "internal error"
	}

	_, _ = rw.Write([]byte(`{"error": "` + msg + `"}`))
}
