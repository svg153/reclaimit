package cli

import (
	"testing"
	"time"
)

func TestParseConfigOlderThan(t *testing.T) {
	cfg, err := ParseConfig([]string{"analyze", "--root", ".", "--older-than", "30d"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OlderThan != 30*24*time.Hour {
		t.Fatalf("older-than = %s, want %s", cfg.OlderThan, 30*24*time.Hour)
	}
}

func TestParseConfigRejectsInvalidOlderThan(t *testing.T) {
	for _, value := range []string{"0", "0d", "-1d", "wat"} {
		if _, err := ParseConfig([]string{"analyze", "--older-than", value}); err == nil {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}
