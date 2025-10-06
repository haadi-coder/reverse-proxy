package proxy

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/haadi-coder/reverse-proxy/internal/middleware"
)

type Route struct {
	Backend      *url.URL
	Transport    *http.Transport
	PreserveHost bool
	Middlewares  []middleware.Middleware
}

type RouteOption func(r *Route)

func WithPreserveHost(preserve bool) RouteOption {
	return func(r *Route) {
		r.PreserveHost = preserve
	}
}

func WithIdleConnTimeout(timeout time.Duration) RouteOption {
	return func(r *Route) {
		r.Transport.IdleConnTimeout = timeout
	}
}

func WithResponseHeaderTimeout(timeout time.Duration) RouteOption {
	return func(r *Route) {
		r.Transport.ResponseHeaderTimeout = timeout
	}
}

func WithMaxIdleConns(max int) RouteOption {
	return func(r *Route) {
		r.Transport.MaxIdleConns = max
	}
}

func WithDialTimeout(timeout time.Duration) RouteOption {
	return func(r *Route) {
		r.Transport.DialContext = (&net.Dialer{Timeout: timeout}).DialContext
	}
}

func WithMiddlewares(middlewares ...middleware.Middleware) RouteOption {
	return func(r *Route) {
		r.Middlewares = append(r.Middlewares, middlewares...)
	}
}

func NewRoute(backend *url.URL, opts ...RouteOption) *Route {
	route := &Route{
		Backend:      backend,
		PreserveHost: true,
		Middlewares:  []middleware.Middleware{},
		Transport: &http.Transport{
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
