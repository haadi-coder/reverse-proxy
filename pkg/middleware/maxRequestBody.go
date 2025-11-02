package middleware

import "net/http"

func MaxRequestBody(maxBytes int64) *Middleware {
	handler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes > 0 {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}

			next.ServeHTTP(w, r)
		})
	}

	return &Middleware{
		Type:    TypeMaxRequestBody,
		Handler: handler,
	}
}
