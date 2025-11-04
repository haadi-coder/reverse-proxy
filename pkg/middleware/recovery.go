package middleware

import (
	"log/slog"
	"net/http"
)

// Recovery creates a middleware that recovers from panics anywhere in the request
// handling chain and converts them into HTTP 500 Internal Server Error responses.
//
// It logs the recovered panic value using slog for debugging and monitoring.
//
// This middleware should be placed early in the chain to catch panics from
// downstream handlers or other middleware.
func Recovery() *Middleware {
	handler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					slog.Error("panic recovered", slog.Any("error", err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}

	return &Middleware{
		Type:    TypeRecovery,
		Handler: handler,
	}
}
