package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/svg153/reclaimit/internal/scanner"
	"github.com/svg153/reclaimit/internal/tui"
)

func TestLoadIgnoreFile(t *testing.T) {
	tmpDir := t.TempDir()
	ignoreFile := filepath.Join(tmpDir, ".reclaimitignore")
	os.WriteFile(ignoreFile, []byte("node_modules\n*.log\n\n# comment\n\n"), 0o644)

	patterns, err := loadIgnoreFile(ignoreFile)
	if err != nil {
		t.Fatalf("loadIgnoreFile: %v", err)
	}
	if len(patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
	if patterns[0] != "node_modules" {
		t.Errorf("first pattern = %q, want %q", patterns[0], "node_modules")
	}
	if patterns[1] != "*.log" {
		t.Errorf("second pattern = %q, want %q", patterns[1], "*.log")
	}
}

func TestLoadIgnoreFile_NonExistent(t *testing.T) {
	_, err := loadIgnoreFile("/nonexistent/.reclaimitignore")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadIgnoreFile_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	ignoreFile := filepath.Join(tmpDir, ".reclaimitignore")
	os.WriteFile(ignoreFile, []byte(""), 0o644)

	patterns, err := loadIgnoreFile(ignoreFile)
	if err != nil {
		t.Fatalf("loadIgnoreFile: %v", err)
	}
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns, got %d", len(patterns))
	}
}

func TestRenderDeletionPreview_Empty(t *testing.T) {
	result := RenderDeletionPreview([]scanner.Candidate{})
	if !strings.Contains(result, "No cleanup candidates") {
		t.Errorf("expected no candidates message, got %q", result)
	}
}

func TestRenderDeletionPreview_Single(t *testing.T) {
	candidates := []scanner.Candidate{
		{Path: "/tmp/test.log", Bytes: 1024, IsDir: false},
	}
	result := RenderDeletionPreview(candidates)
	if !strings.Contains(result, "/tmp/test.log") {
		t.Errorf("expected path in output, got %q", result)
	}
	if !strings.Contains(result, "1.0 KiB") {
		t.Errorf("expected size in output, got %q", result)
	}
}

func TestRenderDeletionPreview_Multiple(t *testing.T) {
	candidates := []scanner.Candidate{
		{Path: "/tmp/a.log", Bytes: 512, IsDir: false},
		{Path: "/tmp/b.log", Bytes: 2048, IsDir: false},
	}
	result := RenderDeletionPreview(candidates)
	if !strings.Contains(result, "a.log") {
		t.Error("expected a.log in output")
	}
	if !strings.Contains(result, "b.log") {
		t.Error("expected b.log in output")
	}
	if !strings.Contains(strings.ToLower(result), "total") {
		t.Error("expected total in output")
	}
}

func TestRenderDeletionPreview_Directories(t *testing.T) {
	candidates := []scanner.Candidate{
		{Path: "/tmp/node_modules", Bytes: 1048576, IsDir: true},
	}
	result := RenderDeletionPreview(candidates)
	if !strings.Contains(result, "node_modules") {
		t.Error("expected node_modules in output")
	}
	if !strings.Contains(result, "1.0 MiB") {
		t.Error("expected MiB in output")
	}
}

func TestHumanizeBytes_RunCommands(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{1024, "1.0 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
	}
	for _, tt := range tests {
		result := humanizeBytes(tt.input)
		if result != tt.expected {
			t.Errorf("humanizeBytes(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestWriteSelection_NoExclusions(t *testing.T) {
	var buf bytes.Buffer
	selection := tui.Selection{}
	cfg := Options{Root: "/tmp/test"}
	err := writeSelection(&buf, cfg, selection)
	if err != nil {
		t.Fatalf("writeSelection returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "Selection") {
		t.Error("expected 'Selection' header")
	}
	if strings.Contains(buf.String(), "Reproduce") {
		t.Error("should not contain Reproduce header when no exclusions")
	}
}


func TestRenderDeletionPreview_NoCandidates(t *testing.T) {
	// Edge case: empty slice should produce a message
	result := RenderDeletionPreview([]scanner.Candidate{})
	if !strings.Contains(result, "No cleanup candidates") {
		t.Errorf("expected no candidates message, got %q", result)
	}
}

func TestRenderDeletionPreview_BigNumber(t *testing.T) {
	candidates := []scanner.Candidate{
		{Path: "/tmp/huge", Bytes: 5368709120, IsDir: true}, // 5 GiB
	}
	result := RenderDeletionPreview(candidates)
	if !strings.Contains(result, "5.0 GiB") {
		t.Error("expected 5.0 GiB in output")
	}
}

func TestParseConfig_GroupMode(t *testing.T) {
	tests := []struct {
		args     []string
		wantErr  bool
		wantMode string
	}{
		{[]string{"--group-mode", "repo"}, false, "repo"},
		{[]string{"--group-mode", "depth"}, false, "depth"},
		{[]string{"--group-mode", "invalid"}, true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.wantMode, func(t *testing.T) {
			cfg, err := ParseConfig(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error for invalid group-mode")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cfg.GroupMode != tt.wantMode {
					t.Errorf("GroupMode = %q, want %q", cfg.GroupMode, tt.wantMode)
				}
			}
		})
	}
}

func TestParseConfig_WithIgnoreFile(t *testing.T) {
	tmpDir := t.TempDir()
	ignoreFile := filepath.Join(tmpDir, ".reclaimitignore")
	os.WriteFile(ignoreFile, []byte("node_modules\n"), 0o644)

	cfg, err := ParseConfig([]string{"--root", tmpDir, "--ignore-file", ignoreFile})
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if cfg.IgnoreFile != ignoreFile {
		t.Errorf("IgnoreFile = %q, want %q", cfg.IgnoreFile, ignoreFile)
	}
}

func TestParseConfig_AllCategories(t *testing.T) {
	tmpDir := t.TempDir()
	cfg, err := ParseConfig([]string{
		"--root", tmpDir,
		"--include-category", "node-modules",
		"--include-category", "python-venv",
		"--exclude-category", "bun-cache",
		"--exclude-category", "ds-store",
	})
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if len(cfg.IncludeCategories) != 2 {
		t.Errorf("IncludeCategories = %v, want 2 items", cfg.IncludeCategories)
	}
	if len(cfg.ExcludeCategories) != 2 {
		t.Errorf("ExcludeCategories = %v, want 2 items", cfg.ExcludeCategories)
	}
}
