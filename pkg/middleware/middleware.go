package middleware

import (
	"net/http"
)

// Type represents a unique identifier for a middleware kind.
type Type string

const (
	TypeRateLimit       Type = "rate_limit"       // Enforces request rate limits.
	TypeBasicAuth       Type = "basic_auth"       // Provides HTTP Basic Authentication.
	TypeCORS            Type = "cors"             // Handles Cross-Origin Resource Sharing.
	TypeHeaders         Type = "headers"          // Modifies request and response headers.
	TypeCompress        Type = "compress"         // Compresses response bodies with gzip.
	TypeRequestID       Type = "request_id"       // Injects a unique ID into each request for tracing.
	TypeSecurityHeaders Type = "security_headers" // Adds common HTTP security headers.
)

// Middleware represents a configured middleware instance.
// It combines a handler function with a type identifier for introspection and management.
type Middleware interface {
	Type() Type                        // Type identifies the kind of middleware (e.g., "cors", "basic_auth").
	Handler(http.Handler) http.Handler // Handler contains the actual middleware logic that wraps the next handler in the chain.
}
