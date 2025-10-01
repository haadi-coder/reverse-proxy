package proxy

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/haadi-coder/reverse-proxy/internal/lib/logger"
)

type Proxy struct {
	cfg    *Config
	server *http.Server
	router *Router
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
	}

	proxy.server.Handler = proxy

	return proxy
}

func (p *Proxy) Run(ctx context.Context) error {
	slog.Info("Server started", slog.String("Address", p.server.Addr))

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
}
