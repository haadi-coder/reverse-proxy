package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// BasicAuthConfig holds the configuration for the Basic Authentication middleware.
type BasicAuthConfig struct {
	Users map[string]string // Users is a map where keys are usernames and values are bcrypt-hashed passwords.
	Realm string            // Realm is the protection space for the authentication. If empty, the default realm will be used by the browser (often "Restricted").
}

type basicAuthMiddleware struct {
	cfg *BasicAuthConfig
}

func (mw *basicAuthMiddleware) Type() Type {
	return TypeBasicAuth
}

// BasicAuth returns a middleware that enforces HTTP Basic Authentication.
// It checks the Authorization header for valid credentials against the provided config.
// If authentication fails or is missing, it responds with a 401 Unauthorized status
// and a WWW-Authenticate header prompting the client to authenticate.
//
// The middleware uses constant-time comparison via bcrypt to prevent timing attacks,
// even for non-existent users.
func (mw *basicAuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const schema = "Basic"

		auth := r.Header.Get("Authorization")
		if auth == "" {
			authenticate(w, mw.cfg.Realm)
			return
		}

		if !strings.HasPrefix(auth, schema) {
			authenticate(w, mw.cfg.Realm)
			return
		}

		rawCredentials := auth[len(schema)+1:]

		credentials, err := base64.StdEncoding.DecodeString(rawCredentials)
		if err != nil {
			authenticate(w, mw.cfg.Realm)
			return
		}

		splitted := strings.SplitN(string(credentials), ":", 2)
		if len(splitted) != 2 {
			authenticate(w, mw.cfg.Realm)
			return
		}

		login := splitted[0]
		password := splitted[1]

		hash, ok := mw.cfg.Users[login]
		if !ok {
			// Perform a dummy bcrypt comparison to maintain constant-time behavior
			// and mitigate timing attacks.
			_ = bcrypt.CompareHashAndPassword(
				[]byte("$2a$10$dummy.hash.to.prevent.timing.attack"),
				[]byte(password),
			)

			authenticate(w, mw.cfg.Realm)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
			authenticate(w, mw.cfg.Realm)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func authenticate(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func BasicAuth(cfg *BasicAuthConfig) Middleware {
	return &basicAuthMiddleware{ cfg: cfg }
}
