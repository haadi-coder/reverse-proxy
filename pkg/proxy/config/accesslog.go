package proxy

import (
	"fmt"
	"slices"

	"github.com/haadi-coder/reverse-proxy/pkg/accesslog"
)

type AccessLogConfig struct {
	Format accesslog.AccessLogFormat
}

func (c *AccessLogConfig) validate() error {
	if c != nil {
		if !isValidAccessLogFormat(c.Format) {
			return fmt.Errorf("invalid format: %s (must be common, json or combined)", c.Format)
		}
	}

	return nil
}

func isValidAccessLogFormat(format accesslog.AccessLogFormat) bool {
	formats := []accesslog.AccessLogFormat{
		accesslog.AccessLogCommon,
		accesslog.AccessLogJSON,
		accesslog.AccessLogCombined,
	}

	return slices.Contains(formats, format)
}
