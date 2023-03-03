package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
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
