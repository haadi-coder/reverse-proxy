package config

import (
	"github.com/haadi-coder/reverse-proxy/pkg/accesslog"
)

type AccessLogConfig struct {
	Format accesslog.AccessLogFormat `yaml:"format,omitempty"`
}

func (c *AccessLogConfig) applyDefaults() {
	if c != nil {
		if c.Format == "" {
			c.Format = "common"
		}
	}
}
