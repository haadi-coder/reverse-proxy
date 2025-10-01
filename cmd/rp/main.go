package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/haadi-coder/filesize"
	"github.com/haadi-coder/reverse-proxy/internal/config"
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

	p := proxy.New(cfg)

	for host, route := range cfg.Routes {

		p.Route(
			host,
			route.Backend,
			proxy.WithPreserveHost(route.PreserveHost),
			proxy.WithIdleConnTimeout(route.IdleConnTimeout),
			proxy.WithResponseHeaderTimeout(route.ResponseHeaderTimeout),
			proxy.WithMaxIdleConns(route.MaxIdleConns),
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
