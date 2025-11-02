package middleware

import "net/http"

type SecurityHeadersConfig struct {
	ContentTypeOptions string
	FrameOptions       string
	XSSProtection      string
	ReferrerPolicy     string
	PermissionsPolicy  string
}

func SecurityHeaders(cfg *SecurityHeadersConfig) *Middleware {
	handler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.ContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", cfg.ContentTypeOptions)
			}

			if cfg.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", cfg.FrameOptions)
			}

			if cfg.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", cfg.XSSProtection)
			}

			if cfg.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
			}

			if cfg.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", cfg.PermissionsPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}

	return &Middleware{
		Type:    TypeSecurityHeaders,
		Handler: handler,
	}
}
