package handlers

import (
	"crypto/tls"
	_ "embed"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type router interface {
	FindRoute(domain string) (string, error)
}

type Proxy struct {
	router router
	client *http.Client
}

func NewProxy(router router) *Proxy {
	return &Proxy{
		router: router,
		client: &http.Client{ // TODO timeout
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

func (c *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := c.handle(w, r); err != nil {
		c.error(w, http.StatusInternalServerError, err)
	}
}

func (c *Proxy) handle(w http.ResponseWriter, r *http.Request) error {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		return err
	}

	route, err := c.router.FindRoute(host)
	if err != nil {
		return errors.Wrap(err, "route")
	}

	newUrl, err := url.Parse(route + r.RequestURI)
	if err != nil {
		return err
	}

	// http: Request.RequestURI can't be set in client requests
	r.RequestURI = ""

	r.URL = newUrl

	resp, err := c.client.Do(r)
	if err != nil {
		return err
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

//go:embed error.html
var errorTemplate string

func (c *Proxy) error(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)

	var message string

	if err != nil {
		message = err.Error()
	} else {
		message = http.StatusText(code)
	}

	content := strings.ReplaceAll(errorTemplate, "{{ code }}", strconv.Itoa(code))
	content = strings.ReplaceAll(content, "{{ message }}", message)

	_, _ = w.Write([]byte(content))
}
