package middleware

import (
	"net/http"
)

type Type string

const (
	TypeRateLimit       Type = "rate_limit"
	TypeBasicAuth       Type = "basic_auth"
	TypeCORS            Type = "cors"
	TypeHeaders         Type = "headers"
	TypeCompress        Type = "compress"
	TypeRequestID       Type = "request_id"
	TypeSecurityHeaders Type = "security_headers"
	TypeMaxRequestBody  Type = "max_request_body"
	TypeRecovery        Type = "recovery"
)

type Handler func(http.Handler) http.Handler

type Middleware struct {
	Type    Type
	Handler Handler
}
