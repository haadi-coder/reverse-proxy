package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/haadi-coder/reverse-proxy/internals/middleware"
)

type RouteConfig struct {
	Backend               string                        `yaml:"backend"`
	PreserveHost          bool                          `yaml:"preserve_host"`
	DialTimeout           time.Duration                 `yaml:"dial_timeout"`
	ResponseHeaderTimeout time.Duration                 `yaml:"response_header_timeout"`
	IdleConnTimeout       time.Duration                 `yaml:"idle_conn_timeout"`
	MaxIdleConns          int                           `yaml:"max_idle_conns"`
	Middlewares           []middleware.MiddlewareConfig `yaml:"middlewares"`
}

func (c *RouteConfig) applyDefaults() {
	if c.DialTimeout == 0 {
		c.DialTimeout = 10 * time.Second
	}
	if c.ResponseHeaderTimeout == 0 {
		c.ResponseHeaderTimeout = 30 * time.Second
	}
	if c.IdleConnTimeout == 0 {
		c.IdleConnTimeout = 90 * time.Second
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 100
	}

	for i := range c.Middlewares {
		c.Middlewares[i].ApplyDefaults()
	}
}

func (c *RouteConfig) validate() error {
	if c.Backend == "" {
		return fmt.Errorf("backend is required")
	}
	if !isUrl(c.Backend) {
		return fmt.Errorf("invalid backend URL: %s", c.Backend)
	}

	if c.DialTimeout < 0 {
		return fmt.Errorf("dial_timeout can't be negative")
	}

	if c.ResponseHeaderTimeout < 0 {
		return fmt.Errorf("response_header_timeout can't be negative")
	}

	if c.IdleConnTimeout < 0 {
		return fmt.Errorf("idle_conn_timeout can't be negative")
	}

	if c.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns can't be negative")
	}

	for _, mw := range c.Middlewares {
		if err := mw.Validate(); err != nil {
			return fmt.Errorf("failed to validate middleware %s: %w", mw.Type, err)
		}
	}

	return nil
}

func isUrl(s string) bool {
	return (strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://"))
}
