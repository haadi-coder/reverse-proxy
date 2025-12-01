package proxy

import (
	"log/slog"
	"net/http"
)

// internalMiddleware represents proxy-specific middleware that doesn't belong
// in the generic middleware package.
type internalMiddleware interface {
	Handler(next http.Handler) http.Handler
}

type maxRequestBodyMiddleware struct {
	maxBytes int64
}

func (m *maxRequestBodyMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.maxBytes > 0 {
			r.Body = http.MaxBytesReader(w, r.Body, m.maxBytes)
		}
		next.ServeHTTP(w, r)
	})
}

type recoveryMiddleware struct{}

func (m *recoveryMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", slog.Any("error", err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			next.ServeHTTP(w, r)
		}()
	})
}

func applyInternalMiddlewares(h http.Handler, middlewares []internalMiddleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i].Handler(h)
	}
	return h
}
