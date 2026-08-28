package plugin

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/certimate-go/certimate/pkg/plugin/proto"
)

type frameSender interface {
	Send(*proto.DeployFrame) error
}

type streamLogHandler struct {
	mu     sync.Mutex
	sender frameSender
	level  slog.Level
}

func NewStreamLogger(sender frameSender, level slog.Level) *slog.Logger {
	return slog.New(&streamLogHandler{sender: sender, level: level})
}

func (h *streamLogHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *streamLogHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sender.Send(&proto.DeployFrame{
		Frame: &proto.DeployFrame_Log{Log: &proto.LogEntry{
			TimestampMilli: r.Time.UnixMilli(),
			Level:          int32(r.Level),
			Message:        r.Message,
			Data:           recordToMap(r),
		}},
	})
}

func (h *streamLogHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *streamLogHandler) WithGroup(_ string) slog.Handler      { return h }

func recordToMap(r slog.Record) map[string]string {
	if r.NumAttrs() == 0 {
		return nil
	}
	out := make(map[string]string, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		flattenAttr(out, "", a)
		return true
	})
	return out
}

func flattenAttr(out map[string]string, prefix string, a slog.Attr) {
	a.Value = a.Value.Resolve()
	if a.Value.Kind() == slog.KindGroup {
		gp := a.Key
		if prefix != "" {
			gp = prefix + "." + a.Key
		}
		for _, sub := range a.Value.Group() {
			flattenAttr(out, gp, sub)
		}
		return
	}
	k := a.Key
	if prefix != "" {
		k = prefix + "." + a.Key
	}
	out[k] = a.Value.String()
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info", "":
		return slog.LevelInfo
	}
	if n, err := strconv.Atoi(s); err == nil {
		return slog.Level(n)
	}
	return slog.LevelInfo
}
