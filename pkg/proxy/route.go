package proxy

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/haadi-coder/reverse-proxy/pkg/middleware"
)

type Route struct {
	backend      *url.URL
	transport    *http.Transport
	preserveHost bool
	middlewares  []middleware.Middleware
}

type RouteOption func(r *Route)

func WithPreserveHost(preserve bool) RouteOption {
	return func(r *Route) {
		r.preserveHost = preserve
	}
}

func WithIdleConnTimeout(timeout time.Duration) RouteOption {
	return func(r *Route) {
		r.transport.IdleConnTimeout = timeout
	}
}

func WithResponseHeaderTimeout(timeout time.Duration) RouteOption {
	return func(r *Route) {
		r.transport.ResponseHeaderTimeout = timeout
	}
}

func WithMaxIdleConns(max int) RouteOption {
	return func(r *Route) {
		r.transport.MaxIdleConns = max
	}
}

func WithDialTimeout(timeout time.Duration) RouteOption {
	return func(r *Route) {
		r.transport.DialContext = (&net.Dialer{Timeout: timeout}).DialContext
	}
}

func WithMiddlewares(middlewares ...middleware.Middleware) RouteOption {
	return func(r *Route) {
		r.middlewares = append(r.middlewares, middlewares...)
	}
}

func NewRoute(backend *url.URL, opts ...RouteOption) *Route {
	route := &Route{
		backend:      backend,
		preserveHost: true,
		middlewares:  []middleware.Middleware{},
		transport: &http.Transport{
			ResponseHeaderTimeout: 30 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			MaxIdleConns:          100,
		},
	}

	for _, opt := range opts {
		opt(route)
	}

	return route
}
