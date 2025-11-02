package middleware

import (
	"bytes"
	"compress/gzip"
	"log/slog"
	"net/http"
	"slices"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gw      *gzip.Writer
	buf     *bytes.Buffer
	size    int
	level   int
	minSize int
}

func (gw *gzipResponseWriter) Write(b []byte) (int, error) {
	if gw.size == 0 {
		gw.buf = &bytes.Buffer{}
		gw.gw, _ = gzip.NewWriterLevel(gw.buf, gw.level)
		gw.Header().Set("Content-Encoding", "gzip")
		gw.Header().Set("Vary", "Accept-Encoding")
	}

	gw.size += len(b)
	if gw.size > gw.minSize {
		return gw.gw.Write(b)
	}
	return gw.buf.Write(b)
}

func (gw *gzipResponseWriter) Close() error {
	if gw.gw == nil {
		if gw.buf != nil {
			_, _ = gw.ResponseWriter.Write(gw.buf.Bytes())
		}
		return nil
	}

	if err := gw.gw.Close(); err != nil {
		return err
	}

	_, err := gw.ResponseWriter.Write(gw.buf.Bytes())
	return err
}

type CompressConfig struct {
	MinSize int
	Level   int
	Types   []string
}

func Compress(cfg *CompressConfig) *Middleware {
	handler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acceptEncoding := r.Header.Get("Accept-Encoding")
			if !strings.Contains(acceptEncoding, "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			if !slices.Contains(cfg.Types, r.Header.Get("Content-Type")) {
				next.ServeHTTP(w, r)
				return
			}

			gw := &gzipResponseWriter{
				ResponseWriter: w,
				level:          cfg.Level,
				minSize:        cfg.MinSize,
			}
			next.ServeHTTP(gw, r)

			if err := gw.Close(); err != nil {
				slog.Error("failed to close gzip writer", slog.Any("error", err))
			}
		})
	}

	return &Middleware{
		Type:    TypeCompress,
		Handler: handler,
	}
}
