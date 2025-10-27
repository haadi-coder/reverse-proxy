package config

import (
	"fmt"

	"github.com/haadi-coder/reverse-proxy/pkg/middleware"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server      *ServerConfig           `yaml:"server"`
	Log         *LogConfig              `yaml:"log"`
	AccessLog   *AccessLogConfig        `yaml:"access_log,omitempty"`
	Routes      map[string]*RouteConfig `yaml:"routes"`
	Middlewares []MiddlewareConfig      `yaml:"middlewares,omitempty"`
}

func Load(path string) (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server == nil {
		c.Server = &ServerConfig{}
	}
	c.Server.applyDefaults()

	if c.Log == nil {
		c.Log = new(LogConfig)
	}
	c.Log.applyDefaults()

	if c.AccessLog != nil {
		c.AccessLog.applyDefaults()
	}

	for host := range c.Routes {
		c.Routes[host].applyDefaults()
	}

	for i := range c.Middlewares {
		c.Middlewares[i].ApplyDefaults()
	}
}

func (c *Config) validate() error {
	if len(c.Routes) == 0 {
		return fmt.Errorf("failed to validate routes. There must be at least one route")
	}

	for host, route := range c.Routes {
		if err := route.validate(); err != nil {
			return fmt.Errorf("failed to validate route for %s: %w", host, err)
		}
	}

	mwTypes := make(map[middleware.Type]bool)
	for _, mw := range c.Middlewares {
		if mwTypes[mw.Type] {
			return fmt.Errorf("duplicate middleware: %s", mw.Type)
		}

		mwTypes[mw.Type] = true

		if err := mw.Validate(); err != nil {
			return fmt.Errorf("failed to validate middleware %s: %w", mw.Type, err)
		}
	}

	return nil
}
