package logger

import (
	"io"
	"log/slog"
	"strings"
)

var logLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

func ValidLogLevel(name string) bool {
	return logLevels[strings.ToLower(strings.TrimSpace(name))]
}

func NewLogger(level string, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	switch strings.ToLower(level) {
	case "debug":
		opts.Level = slog.LevelDebug
	case "warn":
		opts.Level = slog.LevelWarn
	case "error":
		opts.Level = slog.LevelError
	}
	return slog.New(slog.NewTextHandler(w, opts))
}
