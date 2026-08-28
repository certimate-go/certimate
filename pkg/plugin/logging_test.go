package plugin

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	gohclog "github.com/hashicorp/go-hclog"
)

func newCaptureLogger() (*slog.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return logger, buf
}

func TestHclogLogger_AdaptsToSlog(t *testing.T) {
	logger, buf := newCaptureLogger()
	adapter := NewHclogLogger(logger)

	adapter.Info("hello", "provider", "demo")
	adapter.Warn("careful", "code", 42)
	adapter.Error("broken", "err", "x")

	out := buf.String()
	for _, want := range []string{"hello", "careful", "broken"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in slog output, got %s", want, out)
		}
	}
	if !strings.Contains(out, `"provider":"demo"`) {
		t.Fatalf("expected kv attrs in slog output, got %s", out)
	}
}

func TestHclogLogger_LevelGuards(t *testing.T) {
	logger, buf := newCaptureLogger()
	adapter := NewHclogLogger(logger).(*slogHclogAdapter)
	adapter.level = gohclog.Warn

	adapter.Info("should-be-suppressed")
	adapter.Warn("should-pass")

	out := buf.String()
	if strings.Contains(out, "should-be-suppressed") {
		t.Fatalf("info leaked at warn level: %s", out)
	}
	if !strings.Contains(out, "should-pass") {
		t.Fatalf("warn missing: %s", out)
	}
}

func TestHclogLogger_NamedReturnsLogger(t *testing.T) {
	logger, _ := newCaptureLogger()
	adapter := NewHclogLogger(logger)
	child := adapter.Named("webhook-deployer")
	if child == nil {
		t.Fatal("Named returned nil")
	}
	if child.Name() != "webhook-deployer" {
		t.Fatalf("unexpected name %q", child.Name())
	}
	withChild := adapter.With("k", "v")
	if withChild == nil {
		t.Fatal("With returned nil")
	}
}
