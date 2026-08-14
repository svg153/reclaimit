package cli

import (
	"strings"
	"testing"
)

func TestParseConfigScannerLimits(t *testing.T) {
	cfg, err := ParseConfig([]string{"analyze", "--max-depth", "3", "--workers", "4"})
	if err != nil {
		t.Fatalf("ParseConfig scanner limits: %v", err)
	}
	if cfg.MaxDepth != 3 {
		t.Fatalf("MaxDepth = %d, want 3", cfg.MaxDepth)
	}
	if cfg.Workers != 4 {
		t.Fatalf("Workers = %d, want 4", cfg.Workers)
	}
}

func TestParseConfigScannerLimitValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "negative max depth", args: []string{"--max-depth", "-1"}, want: "max-depth must be >= 0"},
		{name: "zero workers", args: []string{"--workers", "0"}, want: "workers must be >= 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseConfig(tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("ParseConfig(%v) error = %v, want %q", tt.args, err, tt.want)
			}
		})
	}
}
