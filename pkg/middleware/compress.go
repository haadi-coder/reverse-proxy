package middleware

import (
	"bytes"
	"compress/gzip"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/haadi-coder/reverse-proxy/internal/lib/logger"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gw             *gzip.Writer
	buf            *bytes.Buffer
	size           int
	level          int
	minSize        int
	contentType    string
	shouldCompress bool
	allowedTypes   []string
}

func (c *gzipResponseWriter) WriteHeader(code int) {
	c.contentType = c.Header().Get("Content-Type")

	contentTypeBase := strings.Split(c.contentType, ";")[0]
	c.shouldCompress = slices.Contains(c.allowedTypes, contentTypeBase)

	if c.shouldCompress {
		c.Header().Set("Content-Encoding", "gzip")
		c.Header().Set("Vary", "Accept-Encoding")
	}

	c.ResponseWriter.WriteHeader(code)
}

func (gw *gzipResponseWriter) Write(b []byte) (int, error) {
	if !gw.shouldCompress {
		return gw.ResponseWriter.Write(b)
	}

	if gw.size == 0 {
		gw.buf = &bytes.Buffer{}

		var err error
		gw.gw, err = gzip.NewWriterLevel(gw.buf, gw.level)
		if err != nil {
			slog.Error("failed to set gzip level", logger.Error(err))
			return gw.ResponseWriter.Write(b)
		}

	}

	gw.size += len(b)
	if gw.size > gw.minSize {
		return gw.gw.Write(b)
	}

	return gw.buf.Write(b)
}

func (gw *gzipResponseWriter) Close() error {
	if !gw.shouldCompress {
		return nil
	}

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

// CompressConfig defines the configuration for gzip compression middleware.
type CompressConfig struct {
	// MinSize is the minimum response body size (in bytes) required to enable compression. Responses smaller than this will be sent uncompressed.
	MinSize int

	// Level specifies the gzip compression level. Acceptable values are constants from
	// the compress/gzip or compress/flate packages, such as:
	//   - gzip.DefaultCompression (-1)
	//   - gzip.NoCompression (0)
	//   - gzip.BestSpeed (1)
	//   - gzip.BestCompression (9)
	// Invalid values will be automatically replaced with gzip.DefaultCompression.
	Level int

	// Types is a list of allowed MIME types (e.g., "text/html", "application/json")
	// that are eligible for compression. Only responses with a matching Content-Type
	// will be compressed.
	Types []string
}

// Compress returns a middleware that compresses HTTP responses using gzip,
// provided the client supports it (via the Accept-Encoding header), the response
// meets the minimum size requirement, and its Content-Type is in the allowed list.
//
// The middleware uses a lazy-write strategy: small responses are buffered and only
// compressed if they exceed MinSize. This avoids overhead for tiny payloads.
func Compress(cfg *CompressConfig) *Middleware {
	cfg.validateLevel()

	handler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acceptEncoding := r.Header.Get("Accept-Encoding")
			if !strings.Contains(acceptEncoding, "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			gw := &gzipResponseWriter{
				ResponseWriter: w,
				level:          cfg.Level,
				minSize:        cfg.MinSize,
				allowedTypes:   cfg.Types,
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

func (c *CompressConfig) validateLevel() {
	levels := []int{gzip.NoCompression, gzip.BestSpeed, gzip.BestCompression, gzip.DefaultCompression}

	if !slices.Contains(levels, c.Level) {
		slog.Warn("invalid gzip compression level, falling back to DefaultCompression", slog.Int("provided_level", c.Level))
		c.Level = gzip.DefaultCompression
	}
}
