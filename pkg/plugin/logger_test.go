package plugin

import (
	"log/slog"
	"testing"

	"github.com/certimate-go/certimate/pkg/plugin/proto"
)

type fakeSender struct {
	frames []*proto.DeployFrame
	err    error
}

func (s *fakeSender) Send(f *proto.DeployFrame) error {
	if s.err != nil {
		return s.err
	}
	s.frames = append(s.frames, f)
	return nil
}

func TestParseLogLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"DEBUG":   slog.LevelDebug,
		"info":    slog.LevelInfo,
		"Warn":    slog.LevelWarn,
		"ERROR":   slog.LevelError,
		"":        slog.LevelInfo,
		"warning": slog.LevelWarn,
		"8":       slog.Level(8),
		"bogus":   slog.LevelInfo,
	}
	for in, want := range cases {
		if got := parseLogLevel(in); got != want {
			t.Errorf("parseLogLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestStreamLoggerEmitsFramesAndRespectsLevel(t *testing.T) {
	s := &fakeSender{}
	logger := NewStreamLogger(s, slog.LevelWarn)
	logger.Debug("d")
	logger.Info("i")
	logger.Warn("w", slog.String("k", "v"))
	logger.Error("e")

	if len(s.frames) != 2 {
		t.Fatalf("expected 2 frames (warn+error), got %d", len(s.frames))
	}
	w := s.frames[0].GetLog()
	if w.Message != "w" || w.Level != int32(slog.LevelWarn) {
		t.Fatalf("warn frame mismatch: %+v", w)
	}
	if w.Data["k"] != "v" {
		t.Fatalf("attr not flattened: %+v", w.Data)
	}
	if w.TimestampMilli == 0 {
		t.Fatal("timestamp not set")
	}
	if s.frames[1].GetLog().Level != int32(slog.LevelError) {
		t.Fatalf("error level mismatch: %+v", s.frames[1].GetLog())
	}
}

func TestStreamLoggerNoAttrsNilMap(t *testing.T) {
	s := &fakeSender{}
	NewStreamLogger(s, slog.LevelInfo).Info("plain")
	if len(s.frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(s.frames))
	}
	if s.frames[0].GetLog().Data != nil {
		t.Fatalf("expected nil data for no-attr record, got %+v", s.frames[0].GetLog().Data)
	}
}
