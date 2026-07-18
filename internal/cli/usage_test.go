package cli

import (
	"strings"
	"testing"
)

func TestUsageText_EmptyTopic(t *testing.T) {
	text := UsageText("")
	if !strings.Contains(text, "reclaimit scans a workspace") {
		t.Error("expected general help text for empty topic")
	}
	if !strings.Contains(text, "analyze") {
		t.Error("expected 'analyze' in general help")
	}
	if !strings.Contains(text, "clean") {
		t.Error("expected 'clean' in general help")
	}
	if !strings.Contains(text, "tui") {
		t.Error("expected 'tui' in general help")
	}
	if !strings.Contains(text, "help") {
		t.Error("expected 'help' in general help")
	}
	if !strings.Contains(text, "version") {
		t.Error("expected 'version' in general help")
	}
}

func TestUsageText_ValidTopic(t *testing.T) {
	text := UsageText("analyze")
	if !strings.Contains(text, "reclaimit analyze") {
		t.Error("expected command name in topic help")
	}
	if !strings.Contains(text, "Flags:") {
		t.Error("expected Flags section")
	}
}

func TestUsageText_TrimsWhitespace(t *testing.T) {
	text := UsageText("  analyze  ")
	if !strings.Contains(text, "reclaimit analyze") {
		t.Error("expected trimmed topic in output")
	}
}

func TestValidLogLevel_Debug(t *testing.T) {
	if !ValidLogLevel("debug") {
		t.Error("expected debug to be valid")
	}
}

func TestValidLogLevel_Info(t *testing.T) {
	if !ValidLogLevel("info") {
		t.Error("expected info to be valid")
	}
}

func TestValidLogLevel_Warn(t *testing.T) {
	if !ValidLogLevel("warn") {
		t.Error("expected warn to be valid")
	}
}

func TestValidLogLevel_Error(t *testing.T) {
	if !ValidLogLevel("error") {
		t.Error("expected error to be valid")
	}
}

func TestValidLogLevel_Invalid(t *testing.T) {
	if ValidLogLevel("verbose") {
		t.Error("expected verbose to be invalid")
	}
}

func TestValidLogLevel_CaseInsensitive(t *testing.T) {
	if !ValidLogLevel("debug") {
		t.Error("expected DEBUG (uppercase) to be valid")
	}
	if !ValidLogLevel("info") {
		t.Error("expected Info (mixed case) to be valid")
	}
}

func TestValidLogLevel_Empty(t *testing.T) {
	if ValidLogLevel("") {
		t.Error("expected empty to be invalid")
	}
}

func TestValidLogLevel_TrimsWhitespace(t *testing.T) {
	if !ValidLogLevel("  debug  ") {
		t.Error("expected trimmed debug to be valid")
	}
}
