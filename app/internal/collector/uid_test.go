package collector_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/indocker-app/app/internal/collector"
)

// transportFunc allows us to inject mock transport for testing. We define it
// here, so we can detect the tlsconfig and return nil for only this type.
type transportFunc func(*http.Request) (*http.Response, error)

func (tf transportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return tf(req)
}

func newMockClient(doer func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: transportFunc(doer),
	}
}

func TestDockerIDResolver_Resolve(t *testing.T) {
	t.Parallel()

	r, err := collector.NewDockerIDResolver(context.Background(), client.WithHTTPClient(
		newMockClient(func(req *http.Request) (*http.Response, error) {
			if strings.HasSuffix(req.URL.Path, "/info") {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(bytes.NewReader(
						[]byte(`{"ID": "7TRN:IPZB:QYBB:VPBQ:UMPP:KARE:6ZNR:XE6T:7EWV:PKF4:ZOJD:TPYS"}`),
					)),
				}, nil
			}

			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       http.NoBody,
			}, nil
		}),
	))
	assert.NoError(t, err)

	id, err := r.Resolve()

	assert.NoError(t, err)
	assert.Equal(t, "7TRN:IPZB:QYBB:VPBQ:UMPP:KARE:6ZNR:XE6T:7EWV:PKF4:ZOJD:TPYS", id)

	id2, err := r.Resolve()

	assert.NoError(t, err)
	assert.Equal(t, id, id2)
}
