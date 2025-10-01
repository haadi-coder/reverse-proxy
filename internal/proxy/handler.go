package proxy

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"

	"github.com/haadi-coder/reverse-proxy/internal/lib/logger"
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

	backendURL := *route.Backend
	backendURL.Path = path.Join(backendURL.Path, r.URL.Path)
	backendURL.RawQuery = r.URL.RawQuery

	backendReq, err := http.NewRequest(r.Method, backendURL.String(), r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create backend request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	for k, vv := range r.Header {
		for _, v := range vv {
			backendReq.Header.Add(k, v)
		}
	}

	if route.PreserveHost {
		backendReq.Host = r.Host
	} else {
		backendReq.Host = route.Backend.Host
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	handler.ServeHTTP(w, r)
}
