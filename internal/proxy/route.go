package proxy

import (
	"net/http"
	"net/url"
	"time"
)

type Route struct {
	Backend      *url.URL
	Transport    *http.Transport
	PreserveHost bool
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

func NewRoute(backend *url.URL, opts ...RouteOption) *Route {
	route := &Route{
		Backend:      backend,
		PreserveHost: true,
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
