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
	"github.com/haadi-coder/reverse-proxy/pkg/middleware"
	"github.com/haadi-coder/reverse-proxy/pkg/proxy"
	proxyCfg "github.com/haadi-coder/reverse-proxy/pkg/proxy/config"
)

const revision = "unknown"

func main() {
	flags := parseFlags()

	if flags.version {
		fmt.Printf("Reverse Proxy: %s\n", revision)
		os.Exit(0)
	}

	yamlCfg, err := config.Load(flags.configPath)
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

type flags struct {
	version    bool
	configPath string
}

func parseFlags() flags {
	var f flags

	flag.BoolVar(&f.version, "version", false, "Print appplication version (long)")
	flag.BoolVar(&f.version, "v", false, "Print application version (short)")

	flag.Parse()

	f.configPath = flag.Arg(0)

	return f
}

func run(yamlCfg *config.Config) error {
	cfg, err := mapConfig(yamlCfg)
	if err != nil {
		return fmt.Errorf("failed to map yaml config to proxy one: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("failed to validate proxy config: %w", err)
	}

	slog.Info("starting proxy", slog.String("addr", cfg.Server.Listen))

	p := proxy.New(cfg)

	gMiddlewares, err := buildMiddlewares(yamlCfg.Middlewares)
	if err != nil {
		return fmt.Errorf("failed to build global middlewares: %w", err)
	}

	p.Use(gMiddlewares...)

	for host, route := range yamlCfg.Routes {
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

func mapConfig(cliCfg *config.Config) (*proxyCfg.Config, error) {
	headersBytes, err := filesize.Parse(cliCfg.Server.MaxHeaderBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse max_header_bytes: %w", err)
	}

	requestBodyBytes, err := filesize.Parse(cliCfg.Server.MaxRequestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse max_request_body: %w", err)
	}

	cfg := &proxyCfg.Config{
		Server: &proxyCfg.ServerConfig{
			Listen:          cliCfg.Server.Listen,
			ReadTimeout:     cliCfg.Server.ReadTimeout,
			WriteTimeout:    cliCfg.Server.WriteTimeout,
			IdleTimeout:     cliCfg.Server.IdleTimeout,
			ShutdownTimeout: cliCfg.Server.ShutdownTimeout,
			MaxHeaderBytes:  headersBytes,
			MaxRequestBody:  requestBodyBytes,
		},
		Log: &proxyCfg.LogConfig{
			Level:  cliCfg.Log.Level,
			Format: cliCfg.Log.Format,
		},
	}

	if cliCfg.AccessLog != nil {
		cfg.AccessLog = &proxyCfg.AccessLogConfig{
			Format: cliCfg.AccessLog.Format,
		}
	}

	return cfg, nil
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

func buildMiddlewares(mwConfigs []config.MiddlewareConfig) ([]*middleware.Middleware, error) {
	middlewares := make([]*middleware.Middleware, 0, len(mwConfigs))

	for _, mwCfg := range mwConfigs {
		mw, err := mwCfg.Build()
		if err != nil {
			return nil, fmt.Errorf("failed to build middleware %s: %w", mwCfg.Type, err)
		}

		middlewares = append(middlewares, mw)
	}

	return middlewares, nil
}
