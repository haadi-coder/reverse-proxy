package proxy

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/haadi-coder/reverse-proxy/internal/lib/logger"
	"github.com/haadi-coder/reverse-proxy/internal/middleware"
)

type Proxy struct {
	cfg          *Config
	server       *http.Server
	router       *Router
	gmiddlewares []middleware.Middleware
}

func New(cfg *Config) *Proxy {
	proxy := &Proxy{
		cfg: cfg,
		server: &http.Server{
			Addr:           cfg.Server.Listen,
			ReadTimeout:    cfg.Server.ReadTimeout,
			WriteTimeout:   cfg.Server.WriteTimeout,
			IdleTimeout:    cfg.Server.IdleTimeout,
			MaxHeaderBytes: int(cfg.Server.MaxHeaderBytes),
		},
		router: &Router{
			exact:     make(map[string]*Route),
			wildcards: make(map[string]*Route),
		},
		gmiddlewares: make([]middleware.Middleware, 0),
	}

	proxy.server.Handler = proxy

	return proxy
}

func (p *Proxy) Run(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), p.cfg.Server.ShutdownTimeout)
		defer cancel()
		return p.server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

func (p *Proxy) Route(host string, backend string, opts ...RouteOption) {
	backendURL, err := url.Parse(backend)
	if err != nil {
		slog.Error("failed to parse backend url", logger.Error(err))
	}

	route := NewRoute(backendURL, opts...)
	p.router.Add(host, route)

	slog.Info("route registered", slog.String("host", host), slog.String("backend", backend))
}

func (p *Proxy) Use(middlewares ...middleware.Middleware) {
	p.gmiddlewares = append(p.gmiddlewares, middlewares...)
}
