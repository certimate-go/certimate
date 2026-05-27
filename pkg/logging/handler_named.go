package logging

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*NamedHandler)(nil)

type NamedHandler struct {
	handler   slog.Handler
	namespace string
}

func NewNamedHandler(handler slog.Handler, namespace string) *NamedHandler {
	h := &NamedHandler{
		handler:   handler,
		namespace: namespace,
	}

	return h
}

func (h *NamedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *NamedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &NamedHandler{
		namespace: h.namespace,
		handler:   h.handler.WithAttrs(attrs),
	}
}

func (h *NamedHandler) WithGroup(name string) slog.Handler {
	return &NamedHandler{
		namespace: h.namespace,
		handler:   h.handler.WithGroup(name),
	}
}

func (h *NamedHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.namespace != "" {
		r.Message = h.namespace + ": " + r.Message
	}
	return h.handler.Handle(ctx, r)
}
