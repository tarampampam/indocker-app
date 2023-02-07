package middleware

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
    UrlPrefix:  "/local-middleware",
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

const (
  routeListContainers = "/containers/list"
)

// ServeHTTP implements http.Handler.
func (p *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
  if strings.HasPrefix(req.URL.Path, p.config.UrlPrefix) {
    // establish a connection to the docker socket
    conn, cErr := (&net.Dialer{}).DialContext(p.ctx, "unix", p.config.SocketPath)
    if cErr != nil {
      p.fail(rw, cErr)

      return
    }

    defer func() { _ = conn.Close() }()

    // the simplest (and, probably, fastest) router in the world :D
    switch path := req.URL.Path; path {
    case p.config.UrlPrefix + routeListContainers:
      // get a list of containers
      containersList, err := p.containersList(p.ctx, conn)
      if err != nil {
        p.fail(rw, err)

        return
      }

      // prepare a response
      var response = make([]struct {
        Names []string `json:"names"`
      }, len(containersList))

      // fill it
      for i, c := range containersList {
        response[i].Names = c.Names
      }

      // and send
      p.json(rw, http.StatusOK, response)

      return
    }

    // if we are here, then the route was not found
    p.json(rw, http.StatusNotFound, map[string]string{"error": "no route found"})

    return
  }

  p.next.ServeHTTP(rw, req)
}

// json writes a json response.
func (p *Plugin) json(rw http.ResponseWriter, status int, v interface{}) {
  rw.Header().Set("Content-Type", "application/json; charset=utf-8")

  rw.WriteHeader(status)

  if err := json.NewEncoder(rw).Encode(v); err != nil {
    p.fail(rw, err)
  }
}

// containersList returns a list of containers.
func (p *Plugin) containersList(ctx context.Context, conn net.Conn) ([]Container, error) {
  var containers []Container

  if err := p.docker(ctx, conn, http.MethodGet, "containers/json", http.NoBody, &containers); err != nil {
    return nil, err
  }

  return containers, nil
}

// docker executes a request against the docker socket.
func (p *Plugin) docker(ctx context.Context, conn net.Conn, method, path string, body io.Reader, out interface{}) error {
  // create a new http client
  c := &http.Client{
    Transport: &http.Transport{
      DialContext: func(_ context.Context, _, _ string) (net.Conn, error) { return conn, nil },
    },
  }

  // create a new request
  req, err := http.NewRequestWithContext(ctx, method, "http://docker/"+path, body)
  if err != nil {
    return err
  }

  // execute the request
  resp, err := c.Do(req)
  if err != nil {
    return err
  }

  // decode the response
  return json.NewDecoder(resp.Body).Decode(&out)
}

// fail writes an error message to the response writer.
func (p *Plugin) fail(rw http.ResponseWriter, err error) {
  rw.WriteHeader(http.StatusInternalServerError)

  var msg string

  if err != nil {
    msg = err.Error()
  } else {
    msg = "internal error"
  }

  _, _ = rw.Write([]byte(msg))
}
