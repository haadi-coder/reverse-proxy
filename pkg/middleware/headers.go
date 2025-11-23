package middleware

import "net/http"

// HeadersConfig defines rules for modifying HTTP headers on incoming requests
// and outgoing responses.
type HeadersConfig struct {
	// Request specifies header modifications to apply to the incoming HTTP request
	// before it is passed to the next handler in the chain.
	Request *HeaderRules

	// Response specifies header modifications to apply to the outgoing HTTP response.
	Response *HeaderRules
}

// HeaderRules defines a set of operations (add, set, remove) to perform on HTTP headers.
type HeaderRules struct {
	// Add appends headers without removing existing ones.
	// If a header with the same name already exists, the new value is added to it.
	// Example: Add["X-Frame-Options"] = "DENY"
	Add map[string]string

	// Set overwrites headers completely. Any existing values for the header are replaced.
	// Example: Set["Content-Security-Policy"] = "default-src 'self'"
	Set map[string]string

	// Remove lists header names to be deleted from the request or response.
	// Example: Remove = []string{"Server", "X-Powered-By"}
	Remove []string
}

type headersMiddleware struct {
	cfg *HeadersConfig
}

func (mw *headersMiddleware) Type() Type {
	return TypeHeaders
}

// Headers returns a middleware that applies custom header rules to HTTP requests
// and/or responses according to the provided configuration.
//
//   - "Add" appends header values (preserving existing ones).
//   - "Set" replaces header values entirely.
//   - "Remove" deletes specified headers.
//
// Request headers are modified before the next handler is called.
// Response headers are modified early (before the next handler runs), so subsequent
// middleware or the final handler may override them.
func (mw *headersMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mw.cfg.Request != nil {
			applyHeaders(r.Header, mw.cfg.Request)
		}

		if mw.cfg.Response != nil {
			applyHeaders(w.Header(), mw.cfg.Response)
		}

		next.ServeHTTP(w, r)
	})
}

func applyHeaders(h http.Header, cfg *HeaderRules) {
	for k, v := range cfg.Add {
		h.Add(k, v)
	}

	for k, v := range cfg.Set {
		h.Set(k, v)
	}

	for _, v := range cfg.Remove {
		h.Del(v)
	}
}

func Headers(cfg *HeadersConfig) Middleware {
	return &headersMiddleware{
		cfg: cfg,
	}
}
