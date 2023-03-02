package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

// HealthChecker is a heals checker.
type HealthChecker struct {
	ctx        context.Context
	httpClient httpClient
}

const defaultHTTPClientTimeout = time.Second * 3

// NewHealthChecker creates heals checker.
func NewHealthChecker(ctx context.Context, client ...httpClient) *HealthChecker {
	var c httpClient

	if len(client) == 1 {
		c = client[0]
	} else {
		c = &http.Client{Timeout: defaultHTTPClientTimeout} // default
	}

	return &HealthChecker{ctx: ctx, httpClient: c}
}

// Check application using liveness probe.
func (c *HealthChecker) Check(httpPort, httpsPort uint) error {
	var eg, egCtx = errgroup.WithContext(c.ctx)

	for _, _uri := range []string{
		fmt.Sprintf("http://127.0.0.1:%d/healthz", httpPort),
		fmt.Sprintf("https://127.0.0.1:%d/healthz", httpsPort),
	} {
		uri := _uri

		eg.Go(func() error {
			req, err := http.NewRequestWithContext(egCtx, http.MethodGet, uri, http.NoBody)
			if err != nil {
				return err
			}

			req.Header.Set("User-Agent", "HealthChecker/internal")

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return err
			}

			_ = resp.Body.Close()

			if code := resp.StatusCode; code != http.StatusOK {
				return fmt.Errorf("wrong status code [%d] from live endpoint (%s)", code, uri)
			}

			return nil
		})
	}

	return eg.Wait()
}
