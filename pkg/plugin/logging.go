package plugin

import (
	"context"
	"io"
	"log"
	"log/slog"

	gohclog "github.com/hashicorp/go-hclog"
)

type slogHclogAdapter struct {
	logger  *slog.Logger
	level   gohclog.Level
	name    string
	implied []any
}

func NewHclogLogger(base *slog.Logger) gohclog.Logger {
	if base == nil {
		base = slog.Default()
	}
	return &slogHclogAdapter{logger: base, level: gohclog.Info}
}

func (a *slogHclogAdapter) Log(level gohclog.Level, msg string, args ...any) {
	a.logAt(level, msg, args)
}

func (a *slogHclogAdapter) Trace(msg string, args ...any) { a.logAt(gohclog.Trace, msg, args) }
func (a *slogHclogAdapter) Debug(msg string, args ...any) { a.logAt(gohclog.Debug, msg, args) }
func (a *slogHclogAdapter) Info(msg string, args ...any)  { a.logAt(gohclog.Info, msg, args) }
func (a *slogHclogAdapter) Warn(msg string, args ...any)  { a.logAt(gohclog.Warn, msg, args) }
func (a *slogHclogAdapter) Error(msg string, args ...any) { a.logAt(gohclog.Error, msg, args) }

func (a *slogHclogAdapter) logAt(level gohclog.Level, msg string, args []any) {
	if level < a.level {
		return
	}
	attrs := a.implied
	if len(args) > 0 {
		attrs = append(append([]any{}, a.implied...), args...)
	}
	ctx := context.Background()
	switch {
	case level >= gohclog.Error:
		a.logger.Log(ctx, slog.LevelError, msg, attrs...)
	case level >= gohclog.Warn:
		a.logger.Log(ctx, slog.LevelWarn, msg, attrs...)
	case level >= gohclog.Info:
		a.logger.Log(ctx, slog.LevelInfo, msg, attrs...)
	default:
		a.logger.Log(ctx, slog.LevelDebug, msg, attrs...)
	}
}

func (a *slogHclogAdapter) IsTrace() bool { return a.level <= gohclog.Trace }
func (a *slogHclogAdapter) IsDebug() bool { return a.level <= gohclog.Debug }
func (a *slogHclogAdapter) IsInfo() bool  { return a.level <= gohclog.Info }
func (a *slogHclogAdapter) IsWarn() bool  { return a.level <= gohclog.Warn }
func (a *slogHclogAdapter) IsError() bool { return a.level <= gohclog.Error }

func (a *slogHclogAdapter) ImpliedArgs() []any {
	out := make([]any, len(a.implied))
	copy(out, a.implied)
	return out
}

func (a *slogHclogAdapter) With(args ...any) gohclog.Logger {
	clone := *a
	clone.implied = append(append([]any{}, a.implied...), args...)
	clone.logger = a.logger.With(args...)
	return &clone
}

func (a *slogHclogAdapter) Name() string { return a.name }

func (a *slogHclogAdapter) Named(name string) gohclog.Logger {
	clone := *a
	if a.name == "" {
		clone.name = name
	} else {
		clone.name = a.name + "." + name
	}
	clone.logger = a.logger.With(slog.String("plugin", clone.name))
	return &clone
}

func (a *slogHclogAdapter) ResetNamed(name string) gohclog.Logger {
	clone := *a
	clone.name = name
	clone.logger = a.logger.With(slog.String("plugin", name))
	return &clone
}

func (a *slogHclogAdapter) SetLevel(level gohclog.Level) { a.level = level }
func (a *slogHclogAdapter) GetLevel() gohclog.Level      { return a.level }

func (a *slogHclogAdapter) StandardLogger(_ *gohclog.StandardLoggerOptions) *log.Logger {
	return slog.NewLogLogger(a.logger.Handler(), slog.LevelInfo)
}

func (a *slogHclogAdapter) StandardWriter(_ *gohclog.StandardLoggerOptions) io.Writer {
	return io.Discard
}
