package core

import (
	"log/slog"
)

type WithLogger interface {
	SetLogger(logger *slog.Logger)
}
