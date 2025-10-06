package middleware

import (
	"net/http"
)

type MiddlewareType string

const (
	TypeRateLimit       MiddlewareType = "rate_limit"
	TypeBasicAuth       MiddlewareType = "basic_auth"
	TypeCORS            MiddlewareType = "cors"
	TypeHeaders         MiddlewareType = "headers"
	TypeCompress        MiddlewareType = "compress"
	TypeRequestID       MiddlewareType = "request_id"
	TypeSecurityHeaders MiddlewareType = "security_headers"
)

type Middleware func(http.Handler) http.Handler
