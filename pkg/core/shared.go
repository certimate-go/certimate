package core

import (
	"log/slog"
)

type LoggerSetter interface {
	SetLogger(logger *slog.Logger)
}
