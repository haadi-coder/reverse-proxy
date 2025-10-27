package proxy

import (
	"fmt"
	"time"
)

type ServerConfig struct {
	Listen          string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxHeaderBytes  int64
	MaxRequestBody  int64
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
