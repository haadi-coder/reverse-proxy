package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type BasicAuthConfig struct {
	Users map[string]string
	Realm string
}

func BasicAuth(cfg *BasicAuthConfig) *Middleware {
	handler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const schema = "Basic"

			auth := r.Header.Get("Authorization")
			if auth == "" {
				authenticate(w, cfg.Realm)
				return
			}

			if !strings.HasPrefix(auth, schema) {
				authenticate(w, cfg.Realm)
				return
			}

			rawCredentials := auth[len(schema)+1:]

			credentials, err := base64.StdEncoding.DecodeString(rawCredentials)
			if err != nil {
				authenticate(w, cfg.Realm)
				return
			}

			splitted := strings.Split(string(credentials), ":")

			login := splitted[0]
			password := splitted[1]

			hash, ok := cfg.Users[login]
			if !ok {
				authenticate(w, cfg.Realm)
				return
			}

			if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
				authenticate(w, cfg.Realm)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	return &Middleware{
		Type:    TypeBasicAuth,
		Handler: handler,
	}
}

func authenticate(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%s", realm))
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
