package config

import (
	"fmt"
	"slices"
)

type AccessLogFormat string

const (
	AccessLogCommon   AccessLogFormat = "common"
	AccessLogJSON     AccessLogFormat = "json"
	AccessLogCombined AccessLogFormat = "combined"
)

type AccessLogConfig struct {
	Format AccessLogFormat `yaml:"format,omitempty"`
}

func (c *AccessLogConfig) applyDefaults() {
	if c != nil {
		if c.Format == "" {
			c.Format = "common"
		}
	}
}

func (c *AccessLogConfig) validate() error {
	if c != nil {
		if !isValidAccessLogFormat(c.Format) {
			return fmt.Errorf("invalid format: %s (must be common, json or combined)", c.Format)
		}
	}

	return nil
}

func isValidAccessLogFormat(format AccessLogFormat) bool {
	formats := []AccessLogFormat{AccessLogCommon, AccessLogJSON, AccessLogCombined}
	return slices.Contains(formats, format)
}
