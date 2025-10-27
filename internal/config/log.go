package config

import proxy "github.com/haadi-coder/reverse-proxy/pkg/proxy/config"

type LogConfig struct {
	Level  proxy.LogLevel  `yaml:"level"`
	Format proxy.LogFormat `yaml:"format"`
}

func (c *LogConfig) applyDefaults() {
	if c.Level == "" {
		c.Level = "info"
	}
	if c.Format == "" {
		c.Format = "text"
	}
}
