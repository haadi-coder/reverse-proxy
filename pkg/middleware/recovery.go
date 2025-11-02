package middleware

import (
	"log/slog"
	"net/http"
)

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
