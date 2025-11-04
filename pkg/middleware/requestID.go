package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

// RequestIDConfig defines configuration for the RequestID middleware.
type RequestIDConfig struct {
	// HeaderName specifies the HTTP header used to store and propagate the request ID.
	// Common values include "X-Request-ID", "Request-Id", or "X-Correlation-ID".
	// If empty, defaults to "X-Request-ID" in typical usage.
	HeaderName string
}

// RequestID creates a middleware that ensures every request has a unique identifier.
//
// It checks for an existing request ID in the configured header (cfg.HeaderName).
// If present, it is reused. If absent, a new UUID (v4) is generated using github.com/google/uuid.
//
// The ID is:
//   - Set in the request header (for downstream handlers and context propagation)
//   - Set in the response header (for client and logging correlation)
//
// This enables end-to-end request tracing across services.
// The generated ID is a standard UUID v4 (e.g., "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8").
func RequestID(cfg *RequestIDConfig) *Middleware {
	handler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(cfg.HeaderName)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			r.Header.Set(cfg.HeaderName, requestID)
			w.Header().Set(cfg.HeaderName, requestID)

			next.ServeHTTP(w, r)
		})
	}

	return &Middleware{
		Type:    TypeRequestID,
		Handler: handler,
	}
}
