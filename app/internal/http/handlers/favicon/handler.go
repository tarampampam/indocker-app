package favicon

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"image"
	"image/png"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"

	_ "gh.tarampamp.am/indocker-app/app/internal/http/handlers/favicon/ico" // register the ICO format
	"gh.tarampamp.am/indocker-app/app/internal/http/openapi"
)

type (
	httpClient interface {
		Do(req *http.Request) (*http.Response, error)
	}

	Handler struct {
		client httpClient

		cache        cache
		purgeCacheAt time.Time
	}
)

const (
	cacheTTL       = time.Hour // purge the cache every hour
	handlerTimeout = 10 * time.Second
)

// New creates a handler to fetch the remote favicon using different approaches. First, it tries to fetch the
// favicon.ico from the base URL. If it fails, it tries to fetch the list of favicons from the HTML page and
// downloads the first one.
// The fetched favicons are cached in memory for future requests (every hour the cache is purged).
func New(client ...httpClient) *Handler {
	var handler = Handler{
		cache:        newCache(),
		purgeCacheAt: time.Now().Add(cacheTTL),
	}

	if len(client) > 0 {
		handler.client = client[0]
	} else {
		const requestTimeout = 5 * time.Second

		handler.client = &http.Client{
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           (&net.Dialer{Timeout: requestTimeout, KeepAlive: requestTimeout}).DialContext,
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
				TLSHandshakeTimeout:   requestTimeout,
				ResponseHeaderTimeout: requestTimeout,
			},
			Timeout: requestTimeout,
		}
	}

	return &handler
}

func (h *Handler) Handle(ctx context.Context, w http.ResponseWriter, params openapi.GetFaviconParams) error {
	var baseUrl, baseUrlErr = h.normalizeBaseUrl(params.BaseUrl)
	if baseUrlErr != nil {
		return baseUrlErr
	}

	if h.purgeCacheAt.Before(time.Now()) { // purge the cache if it is time
		h.cache.Clear()
		h.purgeCacheAt = time.Now().Add(cacheTTL)
	} else if img, hit := h.cache.Get(baseUrl); hit { // otherwise, try to get the favicon from the cache
		return h.writeImage(w, img)
	}

	ctx, cancel := context.WithTimeout(ctx, handlerTimeout) // set the timeout for the handler
	defer cancel()

	var lastErr error

	// try to fetch the favicon.ico from the base URL
	if icon, dlErr := h.getRemoteImage(ctx, fmt.Sprintf("%s/favicon.ico", baseUrl)); dlErr == nil {
		h.cache.Put(baseUrl, icon)

		return h.writeImage(w, icon) // favicon.ico
	} else {
		lastErr = dlErr // store the last error
	}

	// if the favicon.ico is not found (or failed to download), try to fetch the list of favicons from the HTML page
	if iconUrls, err := h.getFaviconsList(ctx, baseUrl); err == nil {
		// try to download the first found favicon
		for _, uri := range iconUrls {
			if icon, dlErr := h.getRemoteImage(ctx, uri); dlErr == nil {
				h.cache.Put(baseUrl, icon)

				return h.writeImage(w, icon) // first found favicon from the HTML page
			} else {
				lastErr = errors.Join(lastErr, dlErr)
			}
		}
	} else {
		lastErr = errors.Join(lastErr, err)
	}

	if lastErr != nil {
		// the errors.Join function concatenates the errors with a newline separator, which is not allowed in the header
		w.Header().Set("X-Error", strings.ReplaceAll(lastErr.Error(), "\n", "; "))
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

// normalizeBaseUrl normalizes the base URL. It adds the scheme if it is missing and removes the trailing slash.
// If the scheme is missing, it uses HTTPS by default. If the URL is invalid, it returns an error.
func (*Handler) normalizeBaseUrl(s string) (string, error) {
	// if the params.BaseUrl does not have a scheme, add one
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = fmt.Sprintf("https://%s", s) // use https by default
	}

	var u, uErr = url.Parse(s)
	if uErr != nil {
		return "", fmt.Errorf("failed to parse base url: %w", uErr)
	}

	// remove the trailing slash
	return strings.TrimRight(u.String(), "/"), nil
}

// writeImage encodes the image to PNG and writes it to the response. It sets the Content-Type and Content-Length
// headers. Also, it sets the Cache-Control header to public with a max-age of 1 hour.
// If the encoding fails, it returns an error.
func (*Handler) writeImage(w http.ResponseWriter, img image.Image) error {
	var buf = new(bytes.Buffer)

	if png.Encode(buf, img) != nil {
		return fmt.Errorf("failed to encode image to PNG")
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.Header().Set("Cache-Control", "public, max-age=3600") // 1 hour

	if _, writeErr := w.Write(buf.Bytes()); writeErr != nil {
		return fmt.Errorf("failed to write image to response: %w", writeErr)
	}

	return nil
}

// getRemoteImage fetches the image from the remote server and decodes it.
func (h *Handler) getRemoteImage(ctx context.Context, uri string) (image.Image, error) {
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, uri, http.NoBody)
	if reqErr != nil {
		return nil, reqErr
	}

	resp, respErr := h.client.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf("failed to fetch favicon image (%s): %w", uri, respErr)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code (%s): %d", uri, resp.StatusCode)
	}

	icon, _, decodeErr := image.Decode(resp.Body)
	if decodeErr != nil {
		return nil, fmt.Errorf("failed to decode favicon image (%s): %w", uri, decodeErr)
	}

	return icon, nil
}

// getFaviconsList fetches the HTML page and extracts the list of favicon URLs. The URLs are normalized to be absolute.
// If the URL is invalid, it is skipped.
func (h *Handler) getFaviconsList(ctx context.Context, baseUrl string) ([]string, error) { //nolint:gocyclo
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, baseUrl, http.NoBody)
	if reqErr != nil {
		return nil, reqErr
	}

	resp, respErr := h.client.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf("failed to fetch favicons list (%s): %w", baseUrl, respErr)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code (%s): %d", baseUrl, resp.StatusCode)
	}

	// parse the HTML and extract the favicons
	doc, docErr := html.Parse(resp.Body)
	if docErr != nil {
		return nil, fmt.Errorf("failed to parse HTML (%s): %w", baseUrl, docErr)
	}

	var (
		favicons []string
		visit    func(*html.Node)
	)

	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "link" { //nolint:nestif
			var rel, href string

			for _, attr := range n.Attr {
				if attr.Key == "rel" {
					rel = strings.ToLower(attr.Val)
				} else if attr.Key == "href" {
					href = attr.Val
				}
			}

			if rel == "icon" || rel == "apple-touch-icon" {
				// append base URL if the href is a relative path
				if !strings.HasPrefix(href, "http://") && !strings.HasPrefix(href, "https://") {
					href = fmt.Sprintf("%s/%s", baseUrl, strings.TrimLeft(href, "./"))
				}

				// make sure the URL is valid
				if _, urlErr := url.Parse(href); urlErr == nil {
					favicons = append(favicons, href)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c) // recursive call
		}
	}

	visit(doc)

	return favicons, nil
}
