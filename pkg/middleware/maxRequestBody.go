package middleware

import "net/http"

// MaxRequestBody returns a middleware that limits the maximum size of the incoming request body.
// If maxBytes is greater than zero, the request body will be wrapped with http.MaxBytesReader,
// which enforces the specified byte limit. Requests exceeding this limit will result in
// an http.StatusRequestEntityTooLarge (413) error sent automatically by the HTTP server.
//
// Setting maxBytes to 0 or a negative value disables the limit.
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
