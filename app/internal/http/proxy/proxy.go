package proxy

import (
	"crypto/tls"
	_ "embed"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/app/internal/docker"
	"gh.tarampamp.am/indocker-app/app/internal/version"
)

type dockerRouter interface {
	RouteToContainerByHostname(hostname string) (*docker.Route, error)
	Routes() map[string]docker.Route
}

type Proxy struct {
	log    *zap.Logger
	router dockerRouter
	client *http.Client
}

func NewProxy(log *zap.Logger, router dockerRouter, clientTimeout time.Duration) *Proxy {
	return &Proxy{
		log:    log,
		router: router,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec
				},
			},
			Timeout: clientTimeout,
		},
	}
}

var errInvalidHostRequested = errors.New("invalid host requested")

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var host string

	if hostPort := strings.Split(r.Host, ":"); len(hostPort) > 0 {
		host = hostPort[0]
	} else {
		p.renderError(w, "unknown", errInvalidHostRequested)

		return
	}

	const trimHostSuffix = ".indocker.app"
	if strings.HasSuffix(strings.ToLower(host), trimHostSuffix) {
		host = host[:len(host)-len(trimHostSuffix)]
	}

	if err := p.handle(w, r, host); err != nil {
		p.renderError(w, host, err)
	}
}

func (p *Proxy) handle(w http.ResponseWriter, r *http.Request, host string) error {
	newUrl, err := url.Parse(host + r.RequestURI)
	if err != nil {
		return err
	}

	// http: Request.RequestURI can't be set in client requests
	r.RequestURI = ""

	r.URL = newUrl

	resp, err := p.client.Do(r)
	if err != nil {
		return errors.Wrap(err, "failed to proxy request")
	}

	defer func() { _ = resp.Body.Close() }()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if _, err = io.Copy(w, resp.Body); err != nil {
		return err
	}

	return nil
}

var (
	//go:embed error.tpl.html
	errorTplHtml     string
	errorTemplate, _ = template.New("").Parse(errorTplHtml) //nolint:gochecknoglobals
)

func (p *Proxy) renderError(w http.ResponseWriter, host string, err error) {
	var (
		message = "Houston, we have a problem"
		code    = http.StatusInternalServerError
	)

	switch {
	case errors.Is(err, docker.ErrNoRegisteredRoutes):
		code, message = http.StatusUnprocessableEntity, "No registered routes"

	case errors.Is(err, docker.ErrNoRouteFound):
		code, message = http.StatusNotFound, "No route found"

	case errors.Is(err, errInvalidHostRequested):
		code, message = http.StatusBadRequest, "Invalid host requested"

	default:
		if err != nil {
			message = err.Error()
		}
	}

	w.WriteHeader(code)

	var (
		routes = p.router.Routes()
		hosts  = make([]string, 0, len(routes))
	)

	for h := range routes {
		hosts = append(hosts, h)
	}

	sort.Slice(hosts, func(i, j int) bool { return hosts[i] < hosts[j] })

	_ = errorTemplate.Execute(w, struct {
		Code            int
		Message         string
		Domain          string
		Version         string
		RegisteredHosts []string
	}{
		Code:            code,
		Message:         message,
		Domain:          host,
		Version:         version.Version(),
		RegisteredHosts: hosts,
	})
}
