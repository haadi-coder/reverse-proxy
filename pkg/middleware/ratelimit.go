package middleware

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type RatelimitConfig struct {
	Requests int
	Window   time.Duration
	Burst    int
}

func RateLimit(cfg *RatelimitConfig) Middleware {
	limiter := rate.NewLimiter(rate.Every(cfg.Window/time.Duration(cfg.Requests)), cfg.Burst)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
