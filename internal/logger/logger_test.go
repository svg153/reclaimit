package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestValidLogLevel_AllLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	for _, l := range levels {
		if !ValidLogLevel(l) {
			t.Errorf("ValidLogLevel(%q) = false, want true", l)
		}
	}
}

func TestValidLogLevel_CaseInsensitive(t *testing.T) {
	if !ValidLogLevel("DEBUG") {
		t.Error("ValidLogLevel(\"DEBUG\") = false, want true")
	}
	if !ValidLogLevel("Info") {
		t.Error("ValidLogLevel(\"Info\") = false, want true")
	}
	if !ValidLogLevel(" WARN ") {
		t.Error("ValidLogLevel(\" WARN \") = false, want true")
	}
}

func TestValidLogLevel_Invalid(t *testing.T) {
	if ValidLogLevel("trace") {
		t.Error("ValidLogLevel(\"trace\") = true, want false")
	}
	if ValidLogLevel("") {
		t.Error("ValidLogLevel(\"\") = true, want false")
	}
	if ValidLogLevel("invalid") {
		t.Error("ValidLogLevel(\"invalid\") = true, want false")
	}
}

func TestNewLogger_DefaultLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("", &buf)
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}
	// Default should be Info level - debug messages should be filtered
	logger.Debug("debug msg")
	if buf.Len() != 0 {
		t.Error("unexpected output from Debug at default level")
	}
	logger.Info("info msg")
	if buf.Len() == 0 {
		t.Error("expected output from Info at default level")
	}
}

func TestNewLogger_DebugLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", &buf)
	logger.Debug("debug msg")
	if !strings.Contains(buf.String(), "debug msg") {
		t.Error("expected debug output")
	}
}

func TestNewLogger_WarnLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("warn", &buf)
	logger.Info("info msg")
	if buf.Len() != 0 {
		t.Error("unexpected info output at warn level")
	}
	logger.Warn("warn msg")
	if !strings.Contains(buf.String(), "warn msg") {
		t.Error("expected warn output")
	}
}

func TestNewLogger_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("error", &buf)
	logger.Info("info msg")
	logger.Warn("warn msg")
	if buf.Len() != 0 {
		t.Error("unexpected output at error level")
	}
	logger.Error("error msg")
	if !strings.Contains(buf.String(), "error msg") {
		t.Error("expected error output")
	}
}

func TestNewLogger_ReturnsSlogLogger(t *testing.T) {
	logger := NewLogger("info", nil)
	// Should be a valid slog.Logger - if it panics, type is wrong
	var _ *slog.Logger = logger
}
