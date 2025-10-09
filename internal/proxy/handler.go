package proxy

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path"
	"time"

	"github.com/haadi-coder/reverse-proxy/internal/lib/logger"
	"github.com/haadi-coder/reverse-proxy/internal/middleware"
)

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Host == "" {
		http.Error(w, "Missing Host header", http.StatusBadRequest)
		return
	}

	route, ok := p.router.Lookup(r.Host)
	if !ok {
		http.Error(w, "No route found for host", http.StatusNotFound)
		return
	}

	startTime := time.Now()

	backendURL := *route.Backend
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

	if route.PreserveHost {
		backendReq.Host = r.Host
	} else {
		backendReq.Host = route.Backend.Host
	}

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{Transport: route.Transport}
		resp, err := client.Do(backendReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to do request: %s", err.Error()), http.StatusBadGateway)
			return
		}

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

		if err := resp.Body.Close(); err != nil {
			slog.Error("Failed to close response body", logger.Error(err))
			return
		}
	})

	loggingResponseWriter := &AccessLogResponseWriter{ResponseWriter: w, contentLength: 0, statusCode: http.StatusOK}

	if p.cfg.AccessLog != nil {
		defer logAccess(*r, *loggingResponseWriter, p.cfg.AccessLog.Format, &startTime)
	}

	maxBodyMiddleware := middleware.MaxRequestBody(p.cfg.Server.MaxRequestBody)

	globalMiddlewares := append(p.gmiddlewares, maxBodyMiddleware, middleware.Recovery())

	allMiddlewares := append(globalMiddlewares, route.Middlewares...)

	handler := applyMiddlewares(baseHandler, allMiddlewares)
	handler.ServeHTTP(loggingResponseWriter, r)
}

func applyMiddlewares(h http.Handler, middlewares []middleware.Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
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
