package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type AccessLogJSONHandler struct {
	w io.Writer
}

func (h *AccessLogJSONHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *AccessLogJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs := make(map[string]string)

	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.String()
		return true
	})

	enc := json.NewEncoder(h.w)
	return enc.Encode(attrs)
}

func (h *AccessLogJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *AccessLogJSONHandler) WithGroup(name string) slog.Handler {
	return h
}

func logAccess(format string, r http.Request, startTime *time.Time) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	if format == "json" {
		logger := slog.New(&AccessLogJSONHandler{w: os.Stdout})
		logger.Info("",
			slog.String("time", startTime.Format(time.RFC3339)),
			slog.String("method", r.Method),
			slog.String("path", r.RequestURI),
			slog.String("status", "200"),
			slog.String("path", r.RequestURI),
			slog.String("protocol", r.Proto),
			slog.String("size", "1234"),
			slog.String("ip", host),
		)
		return
	}

	username, _, ok := r.BasicAuth()
	if !ok {
		username = "-"
	}

	logLine := []string{host, username, startTime.Format("[02/Jan/2006:15:04:05 -0700]"), r.Method, r.RequestURI, r.Proto, "200", "123"}

	if format == "combined" {
		referer := r.Referer()
		if referer == "" {
			referer = "-"
		}

		userAgent := r.UserAgent()
		if userAgent == "" {
			userAgent = "-"
		}

		combined := fmt.Sprintf(`"%s" "%s"`, referer, userAgent)
		logLine = append(logLine, combined)
	}

	_, _ = fmt.Fprintln(os.Stdout, strings.Join(logLine, " "))
}
