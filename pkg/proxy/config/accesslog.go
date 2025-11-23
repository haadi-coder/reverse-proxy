package proxy

import (
	"fmt"
	"slices"

	"github.com/haadi-coder/reverse-proxy/pkg/accesslog"
)

type AccessLogConfig struct {
	Format accesslog.Format
}

func (c *AccessLogConfig) validate() error {
	if c != nil {
		if !isValidAccessLogFormat(c.Format) {
			return fmt.Errorf("invalid format: %s (must be common, json or combined)", c.Format)
		}
	}

	return nil
}

func isValidAccessLogFormat(format accesslog.Format) bool {
	formats := []accesslog.Format{
		accesslog.CommonFormat,
		accesslog.JSONFormat,
		accesslog.CombinedFormat,
	}

	return slices.Contains(formats, format)
}
