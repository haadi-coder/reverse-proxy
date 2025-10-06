package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/haadi-coder/filesize"
	"github.com/haadi-coder/reverse-proxy/internal/config"
	"github.com/haadi-coder/reverse-proxy/internal/lib/logger"
	"github.com/haadi-coder/reverse-proxy/internal/middleware"
	"github.com/haadi-coder/reverse-proxy/internal/proxy"
)

const revision = "unknown"

func main() {
	var version bool
	flag.BoolVar(&version, "version", false, "Print appplication version (long)")
	flag.BoolVar(&version, "v", false, "Print application version (short)")

	flag.Parse()

	if version {
		fmt.Printf("Reverse Proxy: %s\n", revision)
		os.Exit(0)
	}

	arg := flag.Arg(0)
	yamlCfg, err := config.Load(arg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load yaml config: %v\n", err)
		os.Exit(1)
	}

	setupLogger(yamlCfg)

	if err := run(yamlCfg); err != nil {
		slog.Error("failed to start reverse-proxy", logger.Error(err))
		os.Exit(1)
	}
}

func run(yamlCfg *config.Config) error {
	cfg, err := MapConfig(yamlCfg)
	if err != nil {
		return fmt.Errorf("failed to map yaml config to proxy one: %w", err)
	}

	slog.Info("starting proxy", slog.String("addr", cfg.Server.Listen))

	p := proxy.New(cfg)

	gMiddlewares, err := buildMiddlewares(yamlCfg.Middlewares)
	if err != nil {
		return fmt.Errorf("failed to build global middlewares: %w", err)
	}

	p.Use(gMiddlewares...)

	for host, route := range cfg.Routes {
		middlewares, err := buildMiddlewares(route.Middlewares)
		if err != nil {
			return fmt.Errorf("failed to build route %s middlewares: %w", host, err)
		}

		p.Route(
			host,
			route.Backend,
			proxy.WithPreserveHost(route.PreserveHost),
			proxy.WithIdleConnTimeout(route.IdleConnTimeout),
			proxy.WithResponseHeaderTimeout(route.ResponseHeaderTimeout),
			proxy.WithMaxIdleConns(route.MaxIdleConns),
			proxy.WithDialTimeout(route.DialTimeout),
			proxy.WithMiddlewares(middlewares...),
		)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := p.Run(ctx); err != nil {
		return fmt.Errorf("failed to start proxy server: %w", err)
	}

	return nil
}

func MapConfig(cliCfg *config.Config) (*proxy.Config, error) {
	headersBytes, err := filesize.Parse(cliCfg.Server.MaxHeaderBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse max_header_bytes: %w", err)
	}

	requestBodyBytes, err := filesize.Parse(cliCfg.Server.MaxRequestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse max_request_body: %w", err)
	}

	proxyCfg := &proxy.Config{
		Server: proxy.ServerConfig{
			Listen:          cliCfg.Server.Listen,
			ReadTimeout:     cliCfg.Server.ReadTimeout,
			WriteTimeout:    cliCfg.Server.WriteTimeout,
			IdleTimeout:     cliCfg.Server.IdleTimeout,
			ShutdownTimeout: cliCfg.Server.ShutdownTimeout,
			MaxHeaderBytes:  headersBytes,
			MaxRequestBody:  requestBodyBytes,
		},
		Log: proxy.LogConfig{
			Level:  string(cliCfg.Log.Level),
			Format: string(cliCfg.Log.Format),
		},
		Routes: make(map[string]*proxy.RouteConfig),
	}

	if cliCfg.AccessLog != nil {
		proxyCfg.AccessLog = &proxy.AccessLogConfig{
			Format: string(cliCfg.AccessLog.Format),
		}
	}

	for host, cliRoute := range cliCfg.Routes {
		route := &proxy.RouteConfig{
			Backend:               cliRoute.Backend,
			PreserveHost:          cliRoute.PreserveHost,
			DialTimeout:           cliRoute.DialTimeout,
			ResponseHeaderTimeout: cliRoute.ResponseHeaderTimeout,
			IdleConnTimeout:       cliRoute.IdleConnTimeout,
			MaxIdleConns:          cliRoute.MaxIdleConns,
			Middlewares:           cliRoute.Middlewares,
		}

		proxyCfg.Routes[host] = route
	}

	return proxyCfg, nil
}

func setupLogger(cfg *config.Config) {
	var handler slog.Handler
	var level slog.Level

	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	if cfg.Log.Format == "json" {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(handler))
}

func buildMiddlewares(mwConfigs []middleware.MiddlewareConfig) ([]middleware.Middleware, error) {
	middlewares := make([]middleware.Middleware, 0, len(mwConfigs))

	for _, mwCfg := range mwConfigs {
		mw, err := mwCfg.Build()
		if err != nil {
			return nil, fmt.Errorf("failed to build middleware %s: %w", mwCfg.Type, err)
		}

		middlewares = append(middlewares, mw)
	}

	return middlewares, nil
}
