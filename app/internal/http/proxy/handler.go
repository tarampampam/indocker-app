package proxy

import (
	"crypto/tls"
	_ "embed"
	"errors"
	"html/template"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"strings"

	"go.uber.org/zap"
)

var (
	//go:embed error.tpl.html
	errorTplHtml string
	//go:embed 5xx.svg
	err5xxSvg string
	//go:embed 4xx.svg
	err4xxSvg string
)

type (
	dockerRouter interface {
		URLToContainerByHostname(hostname string) ([]url.URL, bool)
		AllContainerURLs() (routes map[string][]url.URL)
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
	if host, _, err := net.SplitHostPort(r.Host); err == nil {
		if urls, found := h.router.URLToContainerByHostname(host); found && len(urls) > 0 {
			// pick a random url in round-robin fashion
			var u = urls[rand.Intn(len(urls))] //nolint:gosec

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
			}).ServeHTTP(w, r)

			return
		}

		h.renderError(w, host, http.StatusNotFound, errors.New("container not found"))

		return
	}

	h.renderError(w, "", http.StatusUnprocessableEntity, errors.New("invalid host"))
}

var errorTemplate = func() *template.Template { //nolint:gochecknoglobals
	var s, err = template.New("").Parse(errorTplHtml)
	if err != nil {
		panic(err)
	}

	return s
}()

const (
	contentTypeHeader = "Content-Type"
	contentTypeHTML   = "text/html; charset=utf-8"
)

func (h *Handler) renderError(w http.ResponseWriter, host string, code int, err error) {
	var message string

	if err != nil {
		message = err.Error()
	} else {
		message = "Houston, we have a problem"
	}

	w.Header().Set(contentTypeHeader, contentTypeHTML)
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
