package accesslog

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/haadi-coder/reverse-proxy/internal/lib/logger"
)

// Middleware wraps an HTTP handler to log access information for each request.
// It creates a ResponseWriter that captures status code and content length,
// then logs the request details after the handler completes.
type Middleware struct {
	Logger *AccessLogger
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		rw := &ResponseWriter{
			ResponseWriter: w,
			ContentLength:  0,
			StatusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		if err := m.Logger.Log(r, rw, startTime); err != nil {
			slog.Error("failed to make accesslog: %w", logger.Error(err))
		}
	})
}
