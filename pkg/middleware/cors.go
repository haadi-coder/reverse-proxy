package middleware

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

// CORSConfig defines the configuration for Cross-Origin Resource Sharing (CORS) middleware.
// It controls which origins, methods, and headers are permitted in cross-origin requests.
type CORSConfig struct {
	// AllowedOrigins is a list of trusted origins (e.g., "https://example.com").
	// Note: Wildcards (e.g., "*") are not supported when AllowCredentials is true.
	AllowedOrigins []string

	// AllowedMethods specifies the HTTP methods allowed in cross-origin requests
	// (e.g., "GET", "POST", "PUT"). Used in response to preflight (OPTIONS) requests.
	AllowedMethods []string

	// AllowedHeaders lists the HTTP headers that can be used during the actual request.
	// This is returned in the "Access-Control-Allow-Headers" header during preflight.
	AllowedHeaders []string

	// ExposedHeaders defines which headers are safe for the client to read from the response.
	// These are sent in the "Access-Control-Expose-Headers" header.
	ExposedHeaders []string

	// AllowCredentials indicates whether the browser should include credentials
	// (cookies, HTTP authentication, etc.) in cross-origin requests.
	// If true, "Access-Control-Allow-Origin" cannot be "*", and origins must be explicitly listed.
	AllowCredentials bool

	// MaxAge defines how long (in seconds) the results of a preflight request
	// can be cached by the browser. Set to 0 to disable caching.
	MaxAge time.Duration
}

type corsMiddleware struct {
	cfg *CORSConfig
}

func (mw *corsMiddleware) Type() Type {
	return TypeCORS
}

// CORS returns a middleware that implements Cross-Origin Resource Sharing (CORS) protection.
// It inspects the "Origin" header of incoming requests and enforces the configured policy.
//
// For non-preflight requests, it adds appropriate CORS response headers if the origin is allowed.
// For preflight (OPTIONS) requests, it responds with CORS headers and a 204 No Content status.
//
// If the origin is not in AllowedOrigins, the middleware responds with 403 Forbidden.
func (mw *corsMiddleware) Handler(next http.Handler) http.Handler {
	exposedHeaders := strings.Join(mw.cfg.ExposedHeaders, ",")
	allowedMethods := strings.Join(mw.cfg.AllowedMethods, ",")
	allowedHeaders := strings.Join(mw.cfg.AllowedHeaders, ",")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		if !slices.Contains(mw.cfg.AllowedOrigins, origin) {
			http.Error(w, "CORS policy: Origin not allowed", http.StatusForbidden)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)

		if mw.cfg.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if exposedHeaders != "" {
			w.Header().Set("Access-Control-Expose-Headers", exposedHeaders)
		}

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(int(mw.cfg.MaxAge.Seconds())))

			if allowedHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CORS(cfg *CORSConfig) Middleware {
	return &corsMiddleware{
		cfg: cfg,
	}
}
