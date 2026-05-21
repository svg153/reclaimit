package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestUsageTextTopics(t *testing.T) {
	cases := []struct {
		topic string
		want  string
	}{
		{"", "Analyze disk usage"},
		{"analyze", "reclaimit analyze"},
		{"clean", "reclaimit clean"},
		{"tui", "reclaimit tui"},
	}

	for _, tc := range cases {
		t.Run(tc.topic, func(t *testing.T) {
			if got := usageText(tc.topic); got == "" || !strings.Contains(got, tc.want) {
				t.Fatalf("usageText(%q) missing %q", tc.topic, tc.want)
			}
		})
	}
}

func TestParseConfigVersionAndHelp(t *testing.T) {
	cfg, err := parseConfig([]string{"--version"})
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}
	if cfg.command != "version" {
		t.Fatalf("expected version command, got %q", cfg.command)
	}

	cfg, err = parseConfig([]string{"tui", "--help"})
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}
	if cfg.command != "help" || cfg.helpTopic != "tui" {
		t.Fatalf("expected help/tui, got command=%q topic=%q", cfg.command, cfg.helpTopic)
	}
}

func TestRenderDeletionPreview(t *testing.T) {
	preview := renderDeletionPreview([]Candidate{
		{Path: "/tmp/demo/.venv", Bytes: 1024},
		{Path: "/tmp/demo/node_modules", Bytes: 2048},
	})
	for _, want := range []string{"Deleting 2 candidates", "/tmp/demo/.venv", "/tmp/demo/node_modules"} {
		if !strings.Contains(preview, want) {
			t.Fatalf("expected preview to contain %q, got %q", want, preview)
		}
	}
}

func TestStringListAndParseConfigValidation(t *testing.T) {
	var list stringList
	if err := list.Set("node-modules"); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if got := list.String(); got != "node-modules" {
		t.Fatalf("unexpected string list value %q", got)
	}
	if err := list.Set(""); err == nil {
		t.Fatalf("expected empty string to fail")
	}

	root := t.TempDir()
	cfg, err := parseConfig([]string{"analyze", "--root", root, "--exclude-group", root, "--exclude-path", filepath.Join(root, ".venv")})
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}
	if cfg.root != root || cfg.excludeGroups[0] != root || cfg.excludePaths[0] != filepath.Join(root, ".venv") {
		t.Fatalf("parseConfig did not normalize paths: %#v", cfg)
	}

	for _, args := range [][]string{
		{"analyze", "--format", "json"},
		{"analyze", "--group-mode", "invalid"},
		{"analyze", "--group-depth", "0"},
		{"analyze", "--top-files", "0"},
	} {
		if _, err := parseConfig(args); err == nil {
			t.Fatalf("expected parseConfig to fail for args %#v", args)
		}
	}
}
