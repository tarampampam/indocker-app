package indocker

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config the plugins configuration.
type Config struct {
	UrlPrefix  string `json:"urlPrefix" yaml:"urlPrefix"`
	SocketPath string `json:"socketPath" yaml:"socketPath"`
}

// CreateConfig creates the default plugins configuration.
func CreateConfig() *Config {
	return &Config{ // defaults
		UrlPrefix:  "/indocker",
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
	docker *Docker
}

// New creates a new plugins.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if stat, err := os.Stat(config.SocketPath); err != nil {
		return nil, errors.New(name + ": " + err.Error())
	} else if stat.Mode().Type() != os.ModeSocket {
		return nil, errors.New(name + ": is not a socket")
	}

	docker := NewDocker(config.SocketPath)

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
		docker: docker,
	}, nil
}

var errRouteNotFound = errors.New("route not found")

// ServeHTTP implements http.Handler.
func (p *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(req.URL.Path, p.config.UrlPrefix) {
		var routeErr = errRouteNotFound

		switch route := req.URL.Path[len(p.config.UrlPrefix):]; route { // the simplest router in the world :D
		case "/stream-docker-state":
			routeErr = p.StreamDockerState(rw, req)

		case "/ping":
			routeErr = p.Ping(rw, req)

		case "/version":
			routeErr = p.DockerVersion(rw, req)

		case "/containers/list":
			routeErr = p.DockerContainersList(rw, req)

		case "/Inspect": // ?id=<container-hash>
			routeErr = p.DockerInspect(rw, req)

		case "/stats": // ?id=<container-hash>
			routeErr = p.DockerContainerStats(rw, req)

		case "/logs": // ?id=<container-hash>&tail=<lines>
			routeErr = p.DockerContainerLogs(rw, req)
		}

		if routeErr != nil {
			p.fail(rw, routeErr)
		}

		return
	}

	p.next.ServeHTTP(rw, req)
}

// StreamDockerState streams the docker state.
// Docs: <https://html.spec.whatwg.org/multipage/server-sent-events.html#event-stream-interpretation>
func (p *Plugin) StreamDockerState(rw http.ResponseWriter, r *http.Request) error {
	flusher, isFlusher := rw.(http.Flusher)
	if !isFlusher {
		return errors.New("streaming unsupported")
	}

	var t = time.NewTicker(1 * time.Second)
	defer t.Stop()

	var buf bytes.Buffer // reuse buffer to reduce allocations

	type payload struct {
		Foo string `json:"foo"`
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	rw.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		buf.WriteString("data: ")

		if j, err := json.Marshal(payload{Foo: "bar"}); err != nil {
			return err
		} else {
			buf.Write(j)
		}

		buf.WriteRune('\n')

		if _, err := buf.WriteTo(rw); err != nil { // writing automatically resets the buffer
			return err
		}

		flusher.Flush()

		select {
		case <-p.ctx.Done():
			return p.ctx.Err()

		case <-r.Context().Done(): // received browser disconnection
			return nil

		case <-t.C:
		}
	}
}

func (p *Plugin) Ping(rw http.ResponseWriter, _ *http.Request) error {
	var _, code, err = p.client.Ping(p.ctx)
	if err != nil {
		return err
	}

	if code != http.StatusOK {
		return errors.New("docker ping failed")
	}

	p.jsonb(rw, code, []byte(`"OK"`))

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

// DockerInspect returns the container Inspect information.
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
	var tail uint64 = 0

	if t := req.URL.Query().Get("tail"); t != "" {
		v, err := strconv.ParseUint(t, 10, 32)
		if err != nil {
			return err
		}

		tail = v
	}

	var data, code, err = p.client.ContainerLogs(p.ctx, req.URL.Query().Get("id"), tail)
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
