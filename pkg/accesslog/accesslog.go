package accesslog

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type AccessLogFormat string

const (
	AccessLogCommon   AccessLogFormat = "common"
	AccessLogJSON     AccessLogFormat = "json"
	AccessLogCombined AccessLogFormat = "combined"
)

type AccessLogConfig struct {
	Format    AccessLogFormat
	StartTime *time.Time
}

func Log(req http.Request, resp AccessLogResponseWriter, cfg *AccessLogConfig) error {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}

	if cfg.Format == AccessLogJSON {
		logger := slog.New(&JSONHandler{w: os.Stdout})
		logger.Info("",
			slog.String("time", cfg.StartTime.Format(time.RFC3339)),
			slog.String("method", req.Method),
			slog.String("path", req.RequestURI),
			slog.Int("status", resp.StatusCode),
			slog.String("path", req.RequestURI),
			slog.String("protocol", req.Proto),
			slog.Int64("size", int64(resp.ContentLength)),
			slog.String("ip", host),
		)
		return nil
	}

	username, _, ok := req.BasicAuth()
	if !ok {
		username = "-"
	}

	logLine := []string{
		host,
		username,
		cfg.StartTime.Format("[02/Jan/2006:15:04:05 -0700]"),
		req.Method, req.RequestURI,
		req.Proto,
		strconv.Itoa(resp.StatusCode),
		strconv.Itoa(int(resp.ContentLength)),
	}

	if cfg.Format == AccessLogCombined {
		referer := req.Referer()
		if referer == "" {
			referer = "-"
		}

		userAgent := req.UserAgent()
		if userAgent == "" {
			userAgent = "-"
		}

		combined := fmt.Sprintf(`"%s" "%s"`, referer, userAgent)
		logLine = append(logLine, combined)
	}

	_, _ = fmt.Fprintln(os.Stdout, strings.Join(logLine, " "))
	return nil
}
