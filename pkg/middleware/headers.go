package middleware

import "net/http"

type HeadersConfig struct {
	Request  *HeaderRules
	Response *HeaderRules
}
type HeaderRules struct {
	Add    map[string]string
	Set    map[string]string
	Remove []string
}

func Headers(cfg *HeadersConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Request != nil {
				applyHeaders(r.Header, cfg.Request)
			}

			if cfg.Response != nil {
				applyHeaders(w.Header(), cfg.Response)
			}

			next.ServeHTTP(w, r)
		})
	}
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
