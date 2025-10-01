package config

import (
	"fmt"
	"time"
)

type ServerConfig struct {
	Listen          string        `yaml:"listen"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	MaxHeaderBytes  string        `yaml:"max_header_bytes"`
	MaxRequestBody  string        `yaml:"max_request_body"`
}

func (c *ServerConfig) applyDefaults() {
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 15 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 15 * time.Second
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 60 * time.Second
	}

	if c.MaxHeaderBytes == "" {
		c.MaxHeaderBytes = "1MB"
	}
	if c.MaxRequestBody == "" {
		c.MaxRequestBody = "10MB"
	}
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 30 * time.Second
	}
}

func (c *ServerConfig) validate() error {
	if c.Listen == "" {
		return fmt.Errorf("listen is required")
	}
	if c.ReadTimeout < 0 {
		return fmt.Errorf("read_timeout can't be negative")
	}
	if c.WriteTimeout < 0 {
		return fmt.Errorf("write_timeout can't be negative")
	}
	if c.IdleTimeout < 0 {
		return fmt.Errorf("idle_timeout can't be negative")
	}
	if c.ShutdownTimeout < 0 {
		return fmt.Errorf("shutdown_timeout can't be negative")
	}

	return nil
}
