package favicon

import (
	"context"
	"crypto/tls"
	"fmt"
	"image"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"

	_ "gh.tarampamp.am/indocker-app/app/internal/http/handlers/favicon/ico" // register the ICO format
)

type (
	httpClient interface {
		Do(req *http.Request) (*http.Response, error)
	}

	// Resolver fetches the favicon from the given base URL.
	Resolver struct{ httpClient httpClient }

	// ResolverOption allows to set options for the resolver.
	ResolverOption func(*Resolver)
)

// WithHTTPClient sets the HTTP client to be used by the resolver.
func WithHTTPClient(c httpClient) ResolverOption { return func(r *Resolver) { r.httpClient = c } }

// NewResolver creates a new favicon resolver with the given options. If no HTTP client is provided, a default one is
// created with a 5-second timeout.
func NewResolver(opts ...ResolverOption) *Resolver {
	var r Resolver

	for _, opt := range opts {
		opt(&r)
	}

	if r.httpClient == nil { // set default HTTP client
		const requestTimeout = 5 * time.Second

		r.httpClient = &http.Client{
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

	return &r
}

// Resolve fetches the favicon from the given base URL. It tries to fetch the favicon.ico first. If it fails, it tries
// to fetch the list of favicons from the HTML page and downloads the first one.
// The timeout is used for the whole operation (including the HTTP requests and image decoding).
func (r *Resolver) Resolve(ctx context.Context, baseUrl string, timeout time.Duration) (image.Image, error) {
	// create a context with a timeout for the whole operation
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// try to fetch the favicon.ico from the base URL
	favicon, faviconErr := r.downloadImage(ctx, fmt.Sprintf("%s/favicon.ico", baseUrl))
	if faviconErr == nil {
		return favicon, nil
	}

	// if the favicon.ico is not found (or failed to download), try to fetch the list of favicons from the HTML page
	faviconsList, listErr := r.getFaviconsLinksList(ctx, baseUrl)
	if listErr != nil {
		return nil, fmt.Errorf("%w; failed to fetch the list of favicons from the HTML: %w", faviconErr, listErr)
	} else if len(faviconsList) == 0 {
		return nil, fmt.Errorf("%w; no links to favicons found in the HTML", faviconErr)
	}

	for _, uri := range faviconsList {
		if icon, dlErr := r.downloadImage(ctx, uri); dlErr == nil {
			return icon, nil
		}
	}

	return nil, fmt.Errorf("%w; failed to download any of the favicons", faviconErr)
}

// downloadImage downloads the image from the given URI. It returns the image or an error if the download fails,
// the HTTP request fails, or the image cannot be decoded (unsupported format or corrupted data).
func (r *Resolver) downloadImage(ctx context.Context, uri string) (image.Image, error) {
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, uri, http.NoBody)
	if reqErr != nil {
		return nil, reqErr
	}

	resp, respErr := r.httpClient.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf("failed to fetch the image (%s): %w", uri, respErr)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code (%d) while fetching the image (%s)", resp.StatusCode, uri)
	}

	icon, _, decodeErr := image.Decode(resp.Body)
	if decodeErr != nil {
		return nil, fmt.Errorf("failed to decode the image (%s): %w", uri, decodeErr)
	}

	return icon, nil
}

// getFaviconsLinksList fetches the index page from the base URL and extracts the list of favicon URLs from the HTML
// document. The URLs are normalized to be absolute. The function returns the list of favicon URLs or an error if the
// request fails, the HTML parsing fails.
func (r *Resolver) getFaviconsLinksList(ctx context.Context, baseUrl string) ([]string, error) {
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, baseUrl, http.NoBody)
	if reqErr != nil {
		return nil, reqErr
	}

	// fetch the index page (usually the home page, index.html or something similar)
	resp, respErr := r.httpClient.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf("failed to fetch the index page (%s): %w", baseUrl, respErr)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code (%d) while fetching the index page (%s)", resp.StatusCode, baseUrl)
	}

	// try to parse the HTML
	doc, docErr := html.Parse(resp.Body)
	if docErr != nil {
		return nil, fmt.Errorf("failed to parse HTML (%s): %w", baseUrl, docErr)
	}

	var (
		links    = r.extractFaviconLinks(doc)
		filtered = make([]string, 0, len(links))
	)

	// filter out invalid URLs
	for _, link := range links {
		// append base URL if the href is a relative path
		if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
			link = fmt.Sprintf("%s/%s", baseUrl, strings.TrimLeft(link, "./"))
		}

		// validate the URL
		if _, err := url.Parse(link); err == nil {
			filtered = append(filtered, link)
		}
	}

	return filtered, nil
}

// extractFaviconLinks extracts the list of favicon URLs from the HTML document (from the
// `<link rel="..." href="..."/>` HTML tags). The returned URLs are not normalized and may be relative.
func (r *Resolver) extractFaviconLinks(doc *html.Node) (favicons []string) {
	var visit func(*html.Node)

	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "link" { // <link rel="..." href="..." />
			var rel, href string

			for _, attr := range n.Attr {
				if attr.Key == "rel" { // rel="..."
					rel = strings.ToLower(attr.Val)
				} else if attr.Key == "href" {
					href = attr.Val // href="..."
				}
			}

			// <link rel="icon" href="..."/>
			// <link rel="apple-touch-icon" href="..."/>
			if rel == "icon" || rel == "apple-touch-icon" {
				favicons = append(favicons, href)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c) // recursive call
		}
	}

	visit(doc)

	return
}
