package proxy

import (
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"strings"

	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/app/internal/docker"
)

type (
	dockerRouter interface {
		docker.RoutingURLResolver
		docker.AllContainerURLsResolver
	}

	Handler struct {
		router     dockerRouter
		log        *zap.Logger
		appVersion string
	}
)

var _ http.Handler = (*Handler)(nil) // verify interface implementation

func New(log *zap.Logger, router dockerRouter, appVersion string) *Handler {
	return &Handler{log: log, router: router, appVersion: appVersion}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var host = r.Host

	if strings.ContainsRune(host, ':') { // remove the port from the host
		if splitHost, _, err := net.SplitHostPort(host); err != nil {
			http.Error(w, fmt.Sprintf("invalid host: %s", host), http.StatusBadRequest)

			return
		} else {
			host = splitHost
		}
	}

	if urls, found := h.router.URLToContainerByHostname(host); found && len(urls) > 0 {
		var u url.URL

		// pick a random url in round-robin fashion
		for _, u = range urls {
			break
		}

		(&httputil.ReverseProxy{
			Director: func(pr *http.Request) {
				var clone = r.Clone(r.Context())

				clone.URL.Scheme = u.Scheme // set target scheme
				clone.URL.Host = u.Host     // set target host
				clone.Host = u.Host         // --//--

				*pr = *clone // swap the request
			},
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, //nolint:gosec
			ErrorLog:  zap.NewStdLog(h.log),
			ModifyResponse: func(resp *http.Response) error {
				resp.Header.Set("X-Indocker-Downstream-Url", u.String())

				return nil
			},
		}).ServeHTTP(w, r) //nolint:gosec

		return
	}

	h.renderErrorNice(w, host, http.StatusNotFound, errors.New("container not found"))
}

var (
	//go:embed error.tpl.html
	errorTplHtml string
	//go:embed 5xx.svg
	err5xxSvg string
	//go:embed 4xx.svg
	err4xxSvg string
)

var errorTemplate = func() *template.Template { //nolint:gochecknoglobals
	var s, err = template.New("").Parse(errorTplHtml)
	if err != nil {
		panic(err)
	}

	return s
}()

func (h *Handler) renderErrorNice(w http.ResponseWriter, host string, code int, err error) {
	var message string

	if err != nil {
		message = err.Error()
	} else {
		message = "Houston, we have a problem"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)

	var (
		allRoutes  = h.router.AllContainerURLs()       // get all routes
		allDomains = make([]string, 0, len(allRoutes)) // prepare slice for all domains
	)

	// convert map keys to slice
	for routeHost := range allRoutes {
		allDomains = append(allDomains, routeHost)
	}

	slices.SortFunc(allDomains, strings.Compare) // sort hosts

	if execErr := errorTemplate.Execute(w, struct {
		Code                     int
		Message, Domain, Version string
		RegisteredHosts          []string
		Err4xxSvg, Err5xxSvg     template.HTML
	}{
		Code:            code,
		Message:         message,
		Domain:          host,
		Version:         h.appVersion,
		RegisteredHosts: allDomains,
		Err4xxSvg:       template.HTML(err4xxSvg), //nolint:gosec
		Err5xxSvg:       template.HTML(err5xxSvg), //nolint:gosec
	}); execErr != nil {
		h.log.Error("failed to render error template", zap.Error(execErr))
	}
}
