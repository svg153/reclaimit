package reclaimit

import (
	"io"
	"log/slog"
	"strings"
)

// logLevels maps the user-facing flag values to slog severities.
var logLevels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

// validLogLevel reports whether name is an accepted --log-level value.
func validLogLevel(name string) bool {
	_, ok := logLevels[strings.ToLower(name)]
	return ok
}

// newLogger builds a text logger that writes structured diagnostics to w at the
// requested level. Reports always go to stdout, so logs stay on stderr and never
// corrupt machine-readable output.
func newLogger(level string, w io.Writer) *slog.Logger {
	lvl, ok := logLevels[strings.ToLower(level)]
	if !ok {
		lvl = slog.LevelWarn
	}
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: lvl}))
}

// log returns the configured logger or a discarding logger so callers never have
// to nil-check. Tests construct config literals without a logger and stay silent.
func (cfg config) log() *slog.Logger {
	if cfg.logger != nil {
		return cfg.logger
	}
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
