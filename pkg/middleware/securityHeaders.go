package middleware

import "net/http"

// SecurityHeadersConfig defines configuration for security-related HTTP response headers.
//
// Each field corresponds to a standard security header. Empty values are ignored,
// allowing selective application of headers.
type SecurityHeadersConfig struct {
	// ContentTypeOptions sets the X-Content-Type-Options header.
	// Recommended: "nosniff" — prevents MIME type sniffing.
	ContentTypeOptions string

	// FrameOptions sets the X-Frame-Options header.
	// Recommended: "DENY" or "SAMEORIGIN" — mitigates clickjacking.
	FrameOptions string

	// XSSProtection sets the X-XSS-Protection header.
	// Recommended: "0" (disable) or omit entirely.
	XSSProtection string

	// ReferrerPolicy sets the Referrer-Policy header.
	// Recommended: "strict-origin-when-cross-origin" or "no-referrer".
	ReferrerPolicy string

	// PermissionsPolicy sets the Permissions-Policy header (formerly Feature-Policy).
	// Example: "geolocation=(), microphone=(), camera=()".
	// Disallows dangerous browser features.
	PermissionsPolicy string
}

type securityHeadersMiddleware struct {
	cfg *SecurityHeadersConfig
}

func SecurityHeaders(cfg *SecurityHeadersConfig) Middleware {
	return &securityHeadersMiddleware{
		cfg: cfg,
	}
}

func (mw *securityHeadersMiddleware) Type() Type {
	return TypeSecurityHeaders
}

// SecurityHeaders creates a middleware that adds security-focused HTTP response headers.
//
// It applies only the headers with non-empty values from the config,
// allowing fine-grained control and safe defaults.
func (mw *securityHeadersMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mw.cfg.ContentTypeOptions != "" {
			w.Header().Set("X-Content-Type-Options", mw.cfg.ContentTypeOptions)
		}

		if mw.cfg.FrameOptions != "" {
			w.Header().Set("X-Frame-Options", mw.cfg.FrameOptions)
		}

		if mw.cfg.XSSProtection != "" {
			w.Header().Set("X-XSS-Protection", mw.cfg.XSSProtection)
		}

		if mw.cfg.ReferrerPolicy != "" {
			w.Header().Set("Referrer-Policy", mw.cfg.ReferrerPolicy)
		}

		if mw.cfg.PermissionsPolicy != "" {
			w.Header().Set("Permissions-Policy", mw.cfg.PermissionsPolicy)
		}

		next.ServeHTTP(w, r)
	})
}
