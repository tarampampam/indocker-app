package cert

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/tls"
	sha8 "encoding/base64" // kek
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"gh.tarampamp.am/indocker-app/app/internal/version"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type (
	Resolver struct {
		url        string
		httpClient httpClient
	}

	ResolverOption func(*Resolver)
)

// WithResolverURL sets the URL to fetch the certificate archive from.
func WithResolverURL(url string) ResolverOption { return func(r *Resolver) { r.url = url } }

// WithResolverHTTPClient sets the HTTP client.
func WithResolverHTTPClient(c httpClient) ResolverOption {
	return func(r *Resolver) { r.httpClient = c }
}

// no one can decode this, right? ;)
//
//nolint:gochecknoglobals
var defUrl, _ = sha8.StdEncoding.DecodeString("aHR0cHM6Ly9pbmRvY2tlci1hcHAtY2VydHMucGFnZXMuZGV2L2FyY2hpdmUudGFyLmd6")

// NewResolver creates a new certificate resolver with the default options.
func NewResolver(opts ...ResolverOption) *Resolver {
	const defaultTimeout = 30 * time.Second

	var r = &Resolver{
		url: string(defUrl),
		httpClient: &http.Client{
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           (&net.Dialer{Timeout: defaultTimeout, KeepAlive: defaultTimeout}).DialContext,
				TLSHandshakeTimeout:   defaultTimeout,
				ResponseHeaderTimeout: defaultTimeout,
			},
			Timeout: defaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Resolve fetches the certificate and private key from the remote server and returns a TLS certificate.
// The function returns an error if the certificate or private key is missing or if the archive is malformed.
func (r Resolver) Resolve(ctx context.Context) (*tls.Certificate, error) {
	var content, contentErr = r.getArchive(ctx)
	if contentErr != nil {
		return nil, fmt.Errorf("failed to get archive: %w", contentErr)
	}

	cert, certErr := tls.X509KeyPair(content.FullChain, content.PrivateKey)
	if certErr != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", certErr)
	}

	return &cert, nil
}

type remoteContent struct{ PrivateKey, FullChain []byte }

// getArchive fetches the archive with the private key and full chain from the remote server and extracts the
// contents of those files into memory.
//
// The archive is expected to contain two files:
//   - privkey.pem: the private key in PEM format
//   - fullchain.pem: the full chain in PEM format
//
// The function returns an error if the archive is missing one of the files or if the files are empty or archive is
// malformed.
func (r Resolver) getArchive(ctx context.Context) (*remoteContent, error) { //nolint:funlen,gocyclo,gocognit
	var nc = (time.Now().Unix() / 600) * 600 //nolint:mnd

	var req, reqErr = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s?nc=%d", r.url, nc), http.NoBody)
	if reqErr != nil {
		return nil, fmt.Errorf("failed to create request: %w", reqErr)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("indocker-app/%s", version.Version()))
	req.Header.Set("Accept", "application/x-tar")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Connection", "close")

	var (
		resp    *http.Response
		respErr error
	)

	for range 5 { // retry attempts
		if respErr != nil { // if the last attempt failed
			select {
			case <-ctx.Done(): // first, check if the context is canceled
				return nil, ctx.Err()
			default: // if not, wait for a second before retrying
				var t = time.NewTimer(time.Second)

				select {
				case <-ctx.Done(): // check if the context is canceled while waiting
					t.Stop() // don't forget to stop the timer

					return nil, ctx.Err()
				case <-t.C:
					t.Stop() // stop the timer
				}
			}
		}

		if resp, respErr = http.DefaultClient.Do(req); respErr != nil {
			respErr = fmt.Errorf("failed to send request: %w", respErr)

			continue // retry
		}

		if resp.StatusCode != http.StatusOK {
			_, respErr = resp.Body.Close(), fmt.Errorf("unexpected status code: %d", resp.StatusCode)

			continue // retry
		}

		break // success
	}

	if respErr != nil {
		return nil, fmt.Errorf("all retry attempts failed: %w", respErr)
	}

	defer func() { _ = resp.Body.Close() }()

	var gz, gzErr = gzip.NewReader(resp.Body)
	if gzErr != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", gzErr)
	}

	var (
		tr     = tar.NewReader(gz)
		result remoteContent
	)

	// Iterate through the files in the TAR archive
	for {
		var hdr, err = tr.Next()
		if errors.Is(err, io.EOF) {
			break // end of archive
		} else if err != nil {
			return nil, fmt.Errorf("failed to read next file in archive: %w", err)
		}

		if hdr.Typeflag != tar.TypeReg || hdr.Size == 0 || hdr.Name == "" {
			continue // skip non-regular, empty or unnamed files
		}

		switch hdr.Name {
		case "privkey.pem":
			if result.PrivateKey, err = io.ReadAll(tr); err != nil {
				return nil, fmt.Errorf("failed to read privkey.pem: %w", err)
			}

		case "fullchain.pem":
			if result.FullChain, err = io.ReadAll(tr); err != nil {
				return nil, fmt.Errorf("failed to read fullchain.pem: %w", err)
			}
		}
	}

	if len(result.PrivateKey) == 0 || len(result.FullChain) == 0 {
		return nil, errors.New("missing private key or full chain in the archive")
	}

	return &result, nil
}
