package proxy

import (
	"crypto/tls"
	_ "embed"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/app/internal/docker"
	"gh.tarampamp.am/indocker-app/app/internal/httptools"
	"gh.tarampamp.am/indocker-app/app/internal/version"
)

type dockerRouter interface {
	RouteToContainerByHostname(hostname string) (*docker.Route, error)
	Routes() map[string]docker.Route
}

type Proxy struct {
	log    *zap.Logger
	router dockerRouter
}

func NewProxy(log *zap.Logger, router dockerRouter) *Proxy {
	return &Proxy{
		log:    log,
		router: router,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var host = httptools.TrimHostPortSuffix(r.Host) // foo.indocker.app -> foo

	if err := p.handle(w, r, host); err != nil {
		p.renderError(w, host, err)
	}
}

func (p *Proxy) handle(w http.ResponseWriter, r *http.Request, host string) error {
	route, err := p.router.RouteToContainerByHostname(host)
	if err != nil {
		return err
	}

	var requestURL = r.URL.RequestURI()

	var s strings.Builder
	s.Grow(len(route.Scheme) + 3 + len(route.IPAddr) + 5 + len(requestURL)) //nolint:wsl

	s.WriteString(route.Scheme)
	s.WriteString("://")
	s.WriteString(route.IPAddr)

	if route.Port > 0 {
		s.WriteRune(':')
		s.WriteString(strconv.FormatUint(uint64(route.Port), 10))
	}

	if !strings.HasPrefix(requestURL, "/") {
		s.WriteRune('/')
	}

	s.WriteString(requestURL)

	newUrl, err := url.Parse(s.String())
	if err != nil {
		return err
	}

	(&httputil.ReverseProxy{
		Director: func(pr *http.Request) { pr.URL = newUrl },
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec
			},
		},
		ErrorLog: zap.NewStdLog(p.log),
	}).ServeHTTP(w, r)

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
