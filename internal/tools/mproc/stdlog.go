package mproc

import (
	"encoding/json"
	"log/slog"
	"time"
)

type stdLogEntry struct {
	Time  time.Time `json:"time"`
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
}

func stdLogUnmarshal(data []byte) (slog.Record, error) {
	var entry stdLogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return slog.Record{}, err
	}

	var attrsMap map[string]any
	if err := json.Unmarshal(data, &attrsMap); err != nil {
		return slog.Record{}, err
	}

	level := slog.LevelInfo
	switch entry.Level {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	record := slog.Record{
		Time:    entry.Time,
		Level:   level,
		Message: entry.Msg,
	}

	delete(attrsMap, "time")
	delete(attrsMap, "level")
	delete(attrsMap, "msg")
	for k, v := range attrsMap {
		record.AddAttrs(slog.Any(k, v))
	}

	return record, nil
}
