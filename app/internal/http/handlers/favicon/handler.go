package favicon

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"strings"
	"time"

	"gh.tarampamp.am/indocker-app/app/internal/docker"
	_ "gh.tarampamp.am/indocker-app/app/internal/http/handlers/favicon/ico" // register the ICO format
)

type (
	Handler struct {
		faviconResolver interface {
			Resolve(ctx context.Context, baseUrl string, timeout time.Duration) (image.Image, error)
		}
		cache interface {
			Get(string) (image.Image, bool)
			Put(string, image.Image)
			Clear()
		}
		containerResolver docker.RoutingURLResolver
		handlerTimeout    time.Duration
	}
)

// New creates a new favicon handler with the given cache TTL for the favicon images. The cache is purged every TTL.
// To stop the cache purging, cancel the given context.
// The handlerTimeout is used for the whole operation (including the HTTP requests and image encoding).
func New(ctx context.Context, r docker.RoutingURLResolver, cacheTTL, handlerTimeout time.Duration) *Handler {
	var (
		handler = Handler{
			faviconResolver:   NewResolver(),
			cache:             newCache(),
			containerResolver: r,
			handlerTimeout:    handlerTimeout,
		}
		ticker = time.NewTicker(cacheTTL)
	)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				handler.cache.Clear()
			case <-ctx.Done():
				return
			}
		}
	}()

	return &handler
}

func (h *Handler) Handle(ctx context.Context, w http.ResponseWriter, hostname string) error {
	urls, routeFound := h.containerResolver.URLToContainerByHostname(hostname)
	if !routeFound || len(urls) == 0 {
		w.WriteHeader(http.StatusNotFound)

		return nil
	}

	var (
		startedAt = time.Now()
		baseUrl   string
	)

	// pick a random url in round-robin fashion
	for _, u := range urls {
		baseUrl = strings.TrimRight(u.String(), "/") // remove the trailing slash

		break
	}

	favicon, faviconErr := h.faviconResolver.Resolve(ctx, baseUrl, h.handlerTimeout)
	if faviconErr == nil {
		h.cache.Put(baseUrl, favicon)

		var buf = new(bytes.Buffer)

		if err := png.Encode(buf, favicon); err != nil {
			return fmt.Errorf("failed to encode image to PNG: %w", err)
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
		w.Header().Set("Cache-Control", "public, max-age=3600") // 1 hour

		if _, err := w.Write(buf.Bytes()); err != nil {
			return fmt.Errorf("failed to write image to response: %w", err)
		}

		return nil
	} else {
		var msg = faviconErr.Error()

		msg = strings.ReplaceAll(msg, "\n", "; ")
		msg = strings.ReplaceAll(msg, "\"", "'")

		w.Header().Set("Server-Timing", fmt.Sprintf("error;desc=\"%s\";dur=%d", msg, time.Since(startedAt).Milliseconds()))
	}

	w.Header().Set("Content-Length", "0")
	w.WriteHeader(http.StatusNoContent)

	return nil
}
