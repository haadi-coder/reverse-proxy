package accesslog

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
)

type JSONHandler struct {
	w io.Writer
}

func (h *JSONHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *JSONHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs := make(map[string]string)

	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.String()
		return true
	})

	enc := json.NewEncoder(h.w)
	return enc.Encode(attrs)
}

func (h *JSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *JSONHandler) WithGroup(name string) slog.Handler {
	return h
}
