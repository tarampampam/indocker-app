package cert

import (
	"net/http"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Resolver struct{}

//func (Resolver) Resolve() (*tls.Certificate, error) {
//
//}
//
//func (Resolver) downloadArchive(ctx context.Context) (io.ReadCloser, error) {
//	var req, reqErr = http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com/archive.zip", nil)
//}
