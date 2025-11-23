package proxy

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/haadi-coder/reverse-proxy/internal/lib/logger"
	"github.com/haadi-coder/reverse-proxy/pkg/accesslog"
	"github.com/haadi-coder/reverse-proxy/pkg/middleware"
	proxyCfg "github.com/haadi-coder/reverse-proxy/pkg/proxy/config"
)

type route struct {
	backend      *url.URL
	transport    *http.Transport
	preserveHost bool
	middlewares  []middleware.Middleware
}

type RouteOption func(r *route)

func WithPreserveHost(preserve bool) RouteOption {
	return func(r *route) {
		r.preserveHost = preserve
	}
}

func WithIdleConnTimeout(timeout time.Duration) RouteOption {
	return func(r *route) {
		r.transport.IdleConnTimeout = timeout
	}
}

func WithResponseHeaderTimeout(timeout time.Duration) RouteOption {
	return func(r *route) {
		r.transport.ResponseHeaderTimeout = timeout
	}
}

func WithMaxIdleConns(max int) RouteOption {
	return func(r *route) {
		r.transport.MaxIdleConns = max
	}
}

func WithDialTimeout(timeout time.Duration) RouteOption {
	return func(r *route) {
		r.transport.DialContext = (&net.Dialer{Timeout: timeout}).DialContext
	}
}

func WithMiddlewares(middlewares ...middleware.Middleware) RouteOption {
	return func(r *route) {
		r.middlewares = append(r.middlewares, middlewares...)
	}
}

func newRoute(backend *url.URL, opts ...RouteOption) *route {
	route := &route{
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

func (rt *route) handle(w http.ResponseWriter, r *http.Request, cfg *proxyCfg.Config, globalMws []middleware.Middleware, accessLogger *accesslog.AccessLogger) {
	startTime := time.Now()

	backendURL := *rt.backend
	backendURL.Path = path.Join(backendURL.Path, r.URL.Path)
	backendURL.RawQuery = r.URL.RawQuery

	backendReq, err := http.NewRequestWithContext(r.Context(), r.Method, backendURL.String(), r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create backend request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	for k, vv := range r.Header {
		for _, v := range vv {
			backendReq.Header.Add(k, v)
		}
	}

	addForwardedHeaders(backendReq, r)

	if rt.preserveHost {
		backendReq.Host = r.Host
	} else {
		backendReq.Host = rt.backend.Host
	}

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{Transport: rt.transport}
		resp, err := client.Do(backendReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to do request: %s", err.Error()), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			slog.Error("Failed to copy response body", logger.Error(err))
		}
	})

	loggingWriter := &accesslog.ResponseWriter{ResponseWriter: w, ContentLength: 0, StatusCode: http.StatusOK}

	defer func() {
		err := accessLogger.Log(r, loggingWriter, startTime)

		if err != nil {
			slog.Error("failed to make accesslog: %w", logger.Error(err))
		}
	}()

	maxBodyMw := middleware.MaxRequestBody(cfg.Server.MaxRequestBody)
	recoveryMw := middleware.Recovery()

	mergedMws := mergeMiddlewares(globalMws, rt.middlewares)

	allMws := make([]middleware.Middleware, 0, len(mergedMws)+2)
	allMws = append(allMws, mergedMws...)
	allMws = append(allMws, maxBodyMw, recoveryMw)

	handler := applyMiddlewares(baseHandler, allMws)
	handler.ServeHTTP(loggingWriter, r)
}

func mergeMiddlewares(global, route []middleware.Middleware) []middleware.Middleware {
	if len(route) == 0 {
		return global
	}

	routeByType := make(map[middleware.Type]middleware.Middleware)
	for _, mw := range route {
		routeByType[mw.Type()] = mw
	}

	result := make([]middleware.Middleware, 0, len(global)+len(route))

	for _, gmw := range global {
		if routeMw, exists := routeByType[gmw.Type()]; exists {
			result = append(result, routeMw)
			delete(routeByType, gmw.Type())
		} else {
			result = append(result, gmw)
		}
	}

	for _, mw := range route {
		if _, exists := routeByType[mw.Type()]; exists {
			result = append(result, mw)
		}
	}

	return result
}

func applyMiddlewares(h http.Handler, middlewares []middleware.Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i].Handler(h)
	}

	return h
}

func addForwardedHeaders(backendReq *http.Request, originalReq *http.Request) {
	clientIP := getClientIP(originalReq)

	forwarded := backendReq.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		clientIP = fmt.Sprintf("%s,%s", forwarded, clientIP)
	}

	backendReq.Header.Set("X-Forwarded-For", clientIP)
	backendReq.Header.Set("X-Forwarded-Host", originalReq.Host)
	backendReq.Header.Set("X-Forwarded-Proto", "http")
}

func getClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	return host
}
