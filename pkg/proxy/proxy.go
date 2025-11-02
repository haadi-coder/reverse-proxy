package proxy

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/haadi-coder/reverse-proxy/internal/lib/logger"
	"github.com/haadi-coder/reverse-proxy/pkg/middleware"
	proxyCfg "github.com/haadi-coder/reverse-proxy/pkg/proxy/config"
)

type Proxy struct {
	cfg         *proxyCfg.Config
	server      *http.Server
	router      *Router
	middlewares []*middleware.Middleware
}

func New(cfg *proxyCfg.Config) *Proxy {
	p := &Proxy{
		cfg: cfg,
		server: &http.Server{
			Addr:           cfg.Server.Listen,
			ReadTimeout:    cfg.Server.ReadTimeout,
			WriteTimeout:   cfg.Server.WriteTimeout,
			IdleTimeout:    cfg.Server.IdleTimeout,
			MaxHeaderBytes: int(cfg.Server.MaxHeaderBytes),
		},
		router: &Router{
			exact:     make(map[string]*route),
			wildcards: make(map[string]*route),
		},
		middlewares: make([]*middleware.Middleware, 0),
	}

	p.server.Handler = http.HandlerFunc(p.serveHTTP)

	return p
}

func (p *Proxy) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Host == "" {
		http.Error(w, "Missing Host header", http.StatusBadRequest)
		return
	}

	route, ok := p.router.lookup(r.Host)
	if !ok {
		http.Error(w, "No route found for host", http.StatusNotFound)
		return
	}

	route.handle(w, r, p.cfg, p.middlewares)
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

	route := newRoute(backendURL, opts...)
	p.router.add(host, route)

	slog.Info("route registered", slog.String("host", host), slog.String("backend", backend))
}

func (p *Proxy) Use(middlewares ...*middleware.Middleware) {
	p.middlewares = append(p.middlewares, middlewares...)
}
