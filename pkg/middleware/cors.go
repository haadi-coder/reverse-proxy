package middleware

import (
	"net/http"
	"slices"
	"strings"
	"time"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

func CORS(cfg *CORSConfig) Middleware {
	exposedHeaders := strings.Join(cfg.ExposedHeaders, ",")
	allowedMethods := strings.Join(cfg.AllowedMethods, ",")
	allowedHeaders := strings.Join(cfg.AllowedHeaders, ",")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" && !slices.Contains(cfg.AllowedOrigins, origin) {
				http.Error(w, "CORS policy: Origin not allowed", http.StatusForbidden)
				return
			}

			if origin != "" && slices.Contains(cfg.AllowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)

				if cfg.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				if exposedHeaders != "" {
					w.Header().Set("Access-Control-Expose-Headers", exposedHeaders)
				}
			}

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
				w.Header().Set("Access-Control-Max-Age", cfg.MaxAge.String())

				if allowedHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
				}

				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
