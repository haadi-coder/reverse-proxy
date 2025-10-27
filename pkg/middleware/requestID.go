package middleware

import (
	"net/http"

	"github.com/haadi-coder/passgen"
)

type RequestIDConfig struct {
	HeaderName string
}

func RequestID(cfg *RequestIDConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			requestID := r.Header.Get(cfg.HeaderName)
			if requestID == "" {
				requestID, _ = passgen.Generate()
			}

			r.Header.Set(cfg.HeaderName, requestID)
			w.Header().Set(cfg.HeaderName, requestID)

			next.ServeHTTP(w, r)
		})
	}
}
