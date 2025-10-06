package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/haadi-coder/filesize"
	"github.com/haadi-coder/reverse-proxy/internal/config"
	"github.com/haadi-coder/reverse-proxy/internal/lib/logger"
	"github.com/haadi-coder/reverse-proxy/internal/middleware"
	"github.com/haadi-coder/reverse-proxy/internal/proxy"
)

func main() {
	yamlCfg, err := config.Load()
	if err != nil {
		fmt.Print(err)
		return
	}

	cfg, err := MapConfig(yamlCfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	setupLogger(cfg)

	slog.Info("starting proxy", slog.String("addr", cfg.Server.Listen))

	p := proxy.New(cfg)

	gMiddlewares, err := buildMiddlewares(yamlCfg.Middlewares)
	if err != nil {
		slog.Error("failed to build global middlewares", logger.Error(err))
	}

	p.Use(gMiddlewares...)

	for host, route := range cfg.Routes {
		middlewares, err := buildMiddlewares(route.Middlewares)
		if err != nil {
			slog.Error("failed to build route middlewares",
				slog.String("host", host),
				slog.String("error", err.Error()))
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

	p.Run(context.Background())
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

func setupLogger(cfg *proxy.Config) {
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
