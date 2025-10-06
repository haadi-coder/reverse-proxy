package proxy

import (
	"time"

	"github.com/haadi-coder/reverse-proxy/internal/middleware"
)

type Config struct {
	Server      ServerConfig
	Log         LogConfig
	AccessLog   *AccessLogConfig
	Routes      map[string]*RouteConfig
	Middlewares []middleware.MiddlewareConfig
}

type ServerConfig struct {
	Listen          string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxHeaderBytes  int64
	MaxRequestBody  int64
}

type LogConfig struct {
	Level  string
	Format string
}

type AccessLogConfig struct {
	Format string
}


type RouteConfig struct {
	Backend               string
	PreserveHost          bool
	DialTimeout           time.Duration
	ResponseHeaderTimeout time.Duration
	IdleConnTimeout       time.Duration
	MaxIdleConns          int
	Middlewares           []middleware.MiddlewareConfig
}
