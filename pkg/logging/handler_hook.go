package logging

import (
	"context"
	"log/slog"
	"sync"
)

type HookHandlerOptions struct {
	Level     slog.Leveler
	WriteFunc func(ctx context.Context, record Record) error
}

var _ slog.Handler = (*HookHandler)(nil)

type HookHandler struct {
	mutex   *sync.Mutex
	base    slog.Handler
	parent  *HookHandler
	options *HookHandlerOptions
	group   string
	attrs   []slog.Attr
}

func NewHookHandler(handler slog.Handler, opts *HookHandlerOptions) *HookHandler {
	if opts == nil {
		opts = &HookHandlerOptions{}
	}

	h := &HookHandler{
		mutex:   &sync.Mutex{},
		base:    handler,
		options: opts,
	}

	if h.options.WriteFunc == nil {
		panic("the `options.WriteFunc` is nil")
	}

	if h.options.Level == nil {
		h.options.Level = slog.LevelInfo
	}

	return h
}

func (h *HookHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.options.Level.Level()
}

func (h *HookHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &HookHandler{
		parent:  h,
		mutex:   h.mutex,
		options: h.options,
		group:   name,
	}
}

func (h *HookHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	return &HookHandler{
		parent:  h,
		mutex:   h.mutex,
		options: h.options,
		attrs:   attrs,
	}
}

func (h *HookHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.group != "" {
		h.mutex.Lock()
		attrs := make([]any, 0, len(h.attrs)+r.NumAttrs())
		for _, a := range h.attrs {
			attrs = append(attrs, a)
		}
		h.mutex.Unlock()

		r.Attrs(func(a slog.Attr) bool {
			attrs = append(attrs, a)
			return true
		})

		r = slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
		r.AddAttrs(slog.Group(h.group, attrs...))
	} else if len(h.attrs) > 0 {
		r = r.Clone()

		h.mutex.Lock()
		r.AddAttrs(h.attrs...)
		h.mutex.Unlock()
	}

	if h.parent != nil {
		return h.parent.Handle(ctx, r)
	}

	if h.base != nil {
		h.base.Handle(ctx, r)
	}

	if err := h.writeRecord(ctx, Record{Record: r}); err != nil {
		return err
	}

	return nil
}

func (h *HookHandler) SetLevel(level slog.Level) {
	h.mutex.Lock()
	h.options.Level = level
	h.mutex.Unlock()
}

func (h *HookHandler) writeRecord(ctx context.Context, r Record) error {
	if h.parent != nil {
		return h.parent.writeRecord(ctx, r)
	}

	return h.options.WriteFunc(ctx, r)
}
