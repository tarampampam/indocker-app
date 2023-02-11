package docker_info

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
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
	ctx    context.Context
	next   http.Handler
	name   string
	config *Config
	client *Client
}

// New creates a new plugins.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if stat, err := os.Stat(config.SocketPath); err != nil {
		return nil, errors.New(name + ": " + err.Error())
	} else if stat.Mode().Type() != os.ModeSocket {
		return nil, errors.New(name + ": is not a socket")
	}

	client, err := NewClient(config.SocketPath)
	if err != nil {
		return nil, err
	}

	return &Plugin{
		ctx:    ctx,
		next:   next,
		name:   name,
		config: config,
		client: client,
	}, nil
}

var errRouteNotFound = errors.New("route not found")

// ServeHTTP implements http.Handler.
func (p *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(req.URL.Path, p.config.UrlPrefix) {
		var routeErr = errRouteNotFound

		switch route := req.URL.Path[len(p.config.UrlPrefix):]; route { // the simplest router in the world :D
		case "/ping":
			routeErr = p.Ping(rw, req)

		case "/version":
			routeErr = p.DockerVersion(rw, req)

		case "/containers/list":
			routeErr = p.DockerContainersList(rw, req)

		case "/inspect": // ?id=<container-hash>
			routeErr = p.DockerInspect(rw, req)

		case "/stats": // ?id=<container-hash>
			routeErr = p.DockerContainerStats(rw, req)

		case "/logs": // ?id=<container-hash>
			routeErr = p.DockerContainerLogs(rw, req)
		}

		if routeErr != nil {
			p.fail(rw, routeErr)
		}

		return
	}

	p.next.ServeHTTP(rw, req)
}

func (p *Plugin) Ping(rw http.ResponseWriter, _ *http.Request) error {
	p.jsonb(rw, http.StatusOK, []byte(`"OK"`))

	return nil
}

// DockerVersion returns the docker version information.
func (p *Plugin) DockerVersion(rw http.ResponseWriter, _ *http.Request) error {
	var list, code, err = p.client.Version(p.ctx)
	if err != nil {
		return err
	}

	p.jsonb(rw, code, list)

	return nil
}

// DockerContainersList returns a list with all containers.
func (p *Plugin) DockerContainersList(rw http.ResponseWriter, _ *http.Request) error {
	var list, code, err = p.client.ContainersList(p.ctx)
	if err != nil {
		return err
	}

	p.jsonb(rw, code, list)

	return nil
}

// DockerInspect returns the container inspect information.
func (p *Plugin) DockerInspect(rw http.ResponseWriter, req *http.Request) error {
	var data, code, err = p.client.ContainerInspect(p.ctx, req.URL.Query().Get("id"))
	if err != nil {
		return err
	}

	p.jsonb(rw, code, data)

	return nil
}

// DockerContainerStats returns the container stats information.
func (p *Plugin) DockerContainerStats(rw http.ResponseWriter, req *http.Request) error {
	var data, code, err = p.client.ContainerStats(p.ctx, req.URL.Query().Get("id"))
	if err != nil {
		return err
	}

	p.jsonb(rw, code, data)

	return nil
}

// DockerContainerLogs returns the base64-encoded container logs.
func (p *Plugin) DockerContainerLogs(rw http.ResponseWriter, req *http.Request) error {
	var data, code, err = p.client.ContainerLogs(p.ctx, req.URL.Query().Get("id"))
	if err != nil {
		return err
	}

	var (
		lines  = bytes.FieldsFunc(data, func(r rune) bool { return r == '\n' || r == '\r' })
		result = make([]string, len(lines))
	)

	for i, line := range lines {
		result[i] = b64.StdEncoding.EncodeToString(line)
	}

	p.json(rw, code, result)

	return err
}

func (*Plugin) setJSONHeaders(rw http.ResponseWriter) {
	const name, value = "Content-Type", "application/json; charset=utf-8"

	rw.Header().Set(name, value)
}

// json writes a json response.
func (p *Plugin) json(rw http.ResponseWriter, status int, v any) {
	p.setJSONHeaders(rw)

	rw.WriteHeader(status)

	if err := json.NewEncoder(rw).Encode(v); err != nil {
		p.fail(rw, err)
	}
}

// jsonb writes a json response as binary data.
func (p *Plugin) jsonb(rw http.ResponseWriter, status int, v []byte) {
	p.setJSONHeaders(rw)

	rw.WriteHeader(status)

	if _, err := rw.Write(v); err != nil {
		p.fail(rw, err)
	}
}

// fail writes an error message to the response writer.
func (p *Plugin) fail(rw http.ResponseWriter, err error) {
	p.setJSONHeaders(rw)

	rw.WriteHeader(http.StatusInternalServerError)

	var msg string

	if err != nil {
		msg = err.Error()
	} else {
		msg = "internal error"
	}

	_, _ = rw.Write([]byte(`{"error":"` + msg + `"}`))
}
