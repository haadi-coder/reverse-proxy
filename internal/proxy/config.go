package proxy

import (
	"time"
)

type Config struct {
	Server    ServerConfig
	Log       LogConfig
	AccessLog *AccessLogConfig
	Routes    map[string]*RouteConfig
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
}
