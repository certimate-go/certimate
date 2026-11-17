package plugin

import (
	"context"
	"log/slog"
	"sync"
	"testing"
)

type captureHandler struct {
	mu      sync.Mutex
	records []capturedLifecycle
}

type capturedLifecycle struct {
	Message string
	Level   slog.Level
	Attrs   map[string]string
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := map[string]string{}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.String()
		return true
	})
	h.mu.Lock()
	h.records = append(h.records, capturedLifecycle{Message: r.Message, Level: r.Level, Attrs: attrs})
	h.mu.Unlock()
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

func TestLifecycle_DoesNotConsumeFrameCap(t *testing.T) {
	h := &captureHandler{}
	node := slog.New(h)
	run := runLog{node: node}
	sink := newForwardLogger(node, noopRedactor, 1)

	run.started("demo", 0)
	sink.Info("rpc-frame")
	sink.Info("rpc-frame-dropped")
	run.exited("demo", 0, exitStopped)

	started := requireLifecycle(t, h.records, "plugin process started")
	exited := requireLifecycle(t, h.records, "plugin process exited")
	if started.Level != slog.LevelInfo || exited.Level != slog.LevelInfo {
		t.Fatalf("levels: start=%v exit=%v", started.Level, exited.Level)
	}
	if exited.Attrs["cause"] != "stopped" {
		t.Fatalf("cause: %+v", exited.Attrs)
	}
	if !hasLifecycle(h.records, "rpc-frame") || hasLifecycle(h.records, "rpc-frame-dropped") {
		t.Fatalf("frame cap: %+v", h.records)
	}
}

func TestLifecycle_NilLoggerNoop(t *testing.T) {
	run := runLog{}
	run.started("demo", 1)
	run.exited("demo", 1, exitDied)
}

func TestLifecycle_DiedIsWarn(t *testing.T) {
	h := &captureHandler{}
	run := runLog{node: slog.New(h)}
	run.started("demo", 9)
	run.exited("demo", 9, exitDied)
	exited := requireLifecycle(t, h.records, "plugin process exited")
	if exited.Level != slog.LevelWarn || exited.Attrs["cause"] != "died" || exited.Attrs["pid"] != "9" {
		t.Fatalf("exit: %+v", exited)
	}
}

func hasLifecycle(records []capturedLifecycle, msg string) bool {
	for _, r := range records {
		if r.Message == msg {
			return true
		}
	}
	return false
}

func requireLifecycle(t *testing.T, records []capturedLifecycle, msg string) capturedLifecycle {
	t.Helper()
	var found []capturedLifecycle
	for _, r := range records {
		if r.Message == msg {
			found = append(found, r)
		}
	}
	if len(found) != 1 {
		t.Fatalf("message %q count = %d, records %+v", msg, len(found), records)
	}
	return found[0]
}
