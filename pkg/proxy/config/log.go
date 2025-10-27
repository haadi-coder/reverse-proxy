package proxy

import (
	"fmt"
	"slices"
)

type LogLevel string
type LogFormat string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"

	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

type LogConfig struct {
	Level  LogLevel
	Format LogFormat
}

func (c *LogConfig) validate() error {
	if !isValidLogLevel(c.Level) {
		return fmt.Errorf("invalid level: %s (must be debug, info, warn or error)", c.Level)
	}
	if !isValidLogFormat(c.Format) {
		return fmt.Errorf("invalid format: %s (must be text or json)", c.Format)
	}

	return nil
}

func isValidLogLevel(level LogLevel) bool {
	levels := []LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError}
	return slices.Contains(levels, level)
}

func isValidLogFormat(format LogFormat) bool {
	formats := []LogFormat{LogFormatJSON, LogFormatText}
	return slices.Contains(formats, format)
}
