package proxy

import (
	"fmt"
)

type Config struct {
	Server    *ServerConfig
	Log       *LogConfig
	AccessLog *AccessLogConfig
}

func (c *Config) Validate() error {
	if err := c.Server.validate(); err != nil {
		return fmt.Errorf("failed to validate server config: %w", err)
	}

	if err := c.Log.validate(); err != nil {
		return fmt.Errorf("failed to validate log config: %w", err)
	}

	if err := c.AccessLog.validate(); err != nil {
		return fmt.Errorf("failed to validate access_log config: %w", err)
	}

	return nil
}
