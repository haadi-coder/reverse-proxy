package middleware

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/haadi-coder/reverse-proxy/internal/lib/logger"
	lru "github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/time/rate"
)

const (
	lruCapacity = 1000
	lruTTL      = 15 * time.Minute
)

// RatelimitConfig holds configuration for the rate limiting middleware.
type RatelimitConfig struct {
	// Requests is the number of requests allowed within the Window duration.
	Requests int

	// Window is the time duration in which Requests are allowed.
	Window time.Duration

	// Burst is the maximum burst size of requests allowed instantly.
	Burst int
}

type ratelimitMiddleware struct {
	cfg   *RatelimitConfig
	cache *lru.LRU[string, *rate.Limiter]
}

func RateLimit(cfg *RatelimitConfig) Middleware {
	cache := lru.NewLRU[string, *rate.Limiter](lruCapacity, nil, lruTTL)

	return &ratelimitMiddleware{
		cfg:   cfg,
		cache: cache,
	}
}

func (mw *ratelimitMiddleware) Type() Type {
	return TypeRateLimit
}

// RateLimit creates a middleware that enforces rate limiting per client IP address.
// It uses a token bucket algorithm via golang.org/x/time/rate.
//
// The rate is calculated as Requests per Window (e.g., 10 requests per minute).
// Burst allows short bursts of traffic beyond the sustained rate.
// Clients exceeding the limit receive HTTP 429 Too Many Requests.
func (mw *ratelimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := getIP(r)
		if err != nil {
			slog.Error("failed get ip", logger.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		limiter, ok := mw.cache.Get(ip)

		if !ok {
			limiter = rate.NewLimiter(
				rate.Every(mw.cfg.Window/time.Duration(mw.cfg.Requests)),
				mw.cfg.Burst,
			)
			mw.cache.Add(ip, limiter)
		}

		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getIP(r *http.Request) (string, error) {
	ips := r.Header.Get("X-Forwarded-For")
	splitIps := strings.Split(ips, ",")

	if len(splitIps) > 0 {
		netIP := net.ParseIP(splitIps[0])
		if netIP != nil {
			return netIP.String(), nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("invalid RemoteAddr: %w", err)
	}

	netIP := net.ParseIP(ip)
	if netIP != nil {
		ip := netIP.String()
		if ip == "::1" {
			return "127.0.0.1", nil
		}

		return ip, nil
	}

	return "", fmt.Errorf("failed to parse IP: %s", ip)
}
