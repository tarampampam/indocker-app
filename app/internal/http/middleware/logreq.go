package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"gh.tarampamp.am/indocker-app/app/internal/httptools"
)

func LogReq(log *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			start = time.Now()
			obs   = &observer{ResponseWriter: w}
		)

		next.ServeHTTP(obs, r)

		var level, message = zap.InfoLevel, "Request successfully processed"

		switch {
		case obs.metrics.status >= http.StatusInternalServerError: // >= 500
			level, message = zap.ErrorLevel, "Server error"

		case obs.metrics.status >= http.StatusBadRequest: // >= 400
			level, message = zap.WarnLevel, "Client error"

		case obs.metrics.status >= http.StatusMultipleChoices: // >= 300
			message = "Redirection"
		}

		if ce := log.Check(level, message); ce != nil {
			var fields = []zap.Field{
				zap.Int("status", obs.metrics.status),
				zap.String("domain", httptools.TrimHostPortSuffix(r.Host)),
				zap.String("uri", r.URL.String()),
				zap.String("method", r.Method),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("useragent", r.UserAgent()),
				zap.Duration("duration", time.Since(start)),
				zap.Int("size", obs.metrics.size),
			}

			if id := w.Header().Get("X-Request-Id"); id != "" {
				fields = append(fields, zap.String("request_id", id))
			}

			ce.Write(fields...)
		}
	})
}

type observer struct {
	http.ResponseWriter // compose original http.ResponseWriter

	metrics struct {
		status int
		size   int
	}
}

var ( // verify interface implementations
	_ http.ResponseWriter = (*observer)(nil)
	_ http.Flusher        = (*observer)(nil)
	_ http.Hijacker       = (*observer)(nil)
	_ http.Pusher         = (*observer)(nil)
)

func (o *observer) Write(b []byte) (int, error) {
	size, err := o.ResponseWriter.Write(b) // write response using original http.ResponseWriter
	o.metrics.size += size                 // capture size

	return size, err
}

func (o *observer) WriteHeader(statusCode int) {
	o.ResponseWriter.WriteHeader(statusCode) // write status code using original http.ResponseWriter
	o.metrics.status = statusCode            // capture status code
}

func (o *observer) Flush() {
	if f, ok := o.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (o *observer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if j, ok := o.ResponseWriter.(http.Hijacker); ok {
		return j.Hijack()
	}

	return nil, nil, errors.New("observer does not implement http.Hijacker")
}

func (o *observer) Push(target string, opts *http.PushOptions) error {
	if p, ok := o.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}

	return errors.New("observer does not implement http.Pusher")
}