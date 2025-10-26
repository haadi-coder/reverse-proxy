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
	"strconv"
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

type AccessLogResponseWriter struct {
	http.ResponseWriter
	contentLength int
	statusCode    int
}

func (rw *AccessLogResponseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	rw.contentLength += n

	return n, err
}

func (rw *AccessLogResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func logAccess(req http.Request, resp AccessLogResponseWriter, format string, startTime *time.Time) {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}

	if format == "json" {
		logger := slog.New(&AccessLogJSONHandler{w: os.Stdout})
		logger.Info("",
			slog.String("time", startTime.Format(time.RFC3339)),
			slog.String("method", req.Method),
			slog.String("path", req.RequestURI),
			slog.Int("status", resp.statusCode),
			slog.String("path", req.RequestURI),
			slog.String("protocol", req.Proto),
			slog.Int64("size", int64(resp.contentLength)),
			slog.String("ip", host),
		)
		return
	}

	username, _, ok := req.BasicAuth()
	if !ok {
		username = "-"
	}

	logLine := []string{
		host,
		username,
		startTime.Format("[02/Jan/2006:15:04:05 -0700]"),
		req.Method, req.RequestURI,
		req.Proto,
		strconv.Itoa(resp.statusCode),
		strconv.Itoa(int(resp.contentLength)),
	}

	if format == "combined" {
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
}
