package cli

import (
	"os"
	"testing"
)

func TestParseConfig_Defaults(t *testing.T) {
	cfg, err := ParseConfig(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "analyze" {
		t.Errorf("expected command analyze, got %s", cfg.Command)
	}
	if cfg.Format != "plain" {
		t.Errorf("expected format plain, got %s", cfg.Format)
	}
	if cfg.GroupMode != "repo" {
		t.Errorf("expected group-mode repo, got %s", cfg.GroupMode)
	}
	if cfg.GroupDepth != 1 {
		t.Errorf("expected group-depth 1, got %d", cfg.GroupDepth)
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("expected log-level warn, got %s", cfg.LogLevel)
	}
}

func TestParseConfig_AnalyzeCommand(t *testing.T) {
	cfg, err := ParseConfig([]string{"analyze"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "analyze" {
		t.Errorf("expected command analyze, got %s", cfg.Command)
	}
}

func TestParseConfig_CleanCommand(t *testing.T) {
	cfg, err := ParseConfig([]string{"clean"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "clean" {
		t.Errorf("expected command clean, got %s", cfg.Command)
	}
}

func TestParseConfig_TUICommand(t *testing.T) {
	cfg, err := ParseConfig([]string{"tui"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "tui" {
		t.Errorf("expected command tui, got %s", cfg.Command)
	}
}

func TestParseConfig_HelpCommand(t *testing.T) {
	cfg, err := ParseConfig([]string{"help"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "help" {
		t.Errorf("expected command help, got %s", cfg.Command)
	}
}

func TestParseConfig_HelpTopic(t *testing.T) {
	cfg, err := ParseConfig([]string{"help", "analyze"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "help" {
		t.Errorf("expected command help, got %s", cfg.Command)
	}
	if cfg.HelpTopic != "analyze" {
		t.Errorf("expected help-topic analyze, got %s", cfg.HelpTopic)
	}
}

func TestParseConfig_HelpFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--help"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "help" {
		t.Errorf("expected command help, got %s", cfg.Command)
	}
}

func TestParseConfig_VersionCommand(t *testing.T) {
	cfg, err := ParseConfig([]string{"--version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "version" {
		t.Errorf("expected command version, got %s", cfg.Command)
	}
}

func TestParseConfig_RootFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--root", "/tmp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Root != "/tmp" {
		t.Errorf("expected root /tmp, got %s", cfg.Root)
	}
}

func TestParseConfig_FormatFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--format", "markdown"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Format != "markdown" {
		t.Errorf("expected format markdown, got %s", cfg.Format)
	}
}

func TestParseConfig_InvalidFormat(t *testing.T) {
	_, err := ParseConfig([]string{"--format", "json"})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestParseConfig_GroupModeFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--group-mode", "depth"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GroupMode != "depth" {
		t.Errorf("expected group-mode depth, got %s", cfg.GroupMode)
	}
}

func TestParseConfig_InvalidGroupMode(t *testing.T) {
	_, err := ParseConfig([]string{"--group-mode", "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid group-mode")
	}
}

func TestParseConfig_GroupDepthFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--group-depth", "3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GroupDepth != 3 {
		t.Errorf("expected group-depth 3, got %d", cfg.GroupDepth)
	}
}

func TestParseConfig_InvalidGroupDepth(t *testing.T) {
	_, err := ParseConfig([]string{"--group-depth", "0"})
	if err == nil {
		t.Fatal("expected error for group-depth 0")
	}
}

func TestParseConfig_TopFilesFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--top-files", "50"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TopFiles != 50 {
		t.Errorf("expected top-files 50, got %d", cfg.TopFiles)
	}
}

func TestParseConfig_TopGroupsFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--top-groups", "10"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TopGroups != 10 {
		t.Errorf("expected top-groups 10, got %d", cfg.TopGroups)
	}
}

func TestParseConfig_TopEntriesFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--top-entries", "5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TopEntries != 5 {
		t.Errorf("expected top-entries 5, got %d", cfg.TopEntries)
	}
}

func TestParseConfig_MinCandidateSizeFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--min-candidate-size", "1048576"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MinCandidateSize != 1048576 {
		t.Errorf("expected min-candidate-size 1048576, got %d", cfg.MinCandidateSize)
	}
}

func TestParseConfig_OutFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--out", "report.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutFile != "report.md" {
		t.Errorf("expected out report.md, got %s", cfg.OutFile)
	}
}

func TestParseConfig_YesFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--yes"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Yes {
		t.Error("expected yes to be true")
	}
}

func TestParseConfig_LogLevelFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--log-level", "debug"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log-level debug, got %s", cfg.LogLevel)
	}
}

func TestParseConfig_InvalidLogLevel(t *testing.T) {
	_, err := ParseConfig([]string{"--log-level", "verbose"})
	if err == nil {
		t.Fatal("expected error for invalid log-level")
	}
}

func TestParseConfig_IncludeCategoryFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--include-category", "node-modules"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.IncludeCategories) != 1 || cfg.IncludeCategories[0] != "node-modules" {
		t.Errorf("expected include node-modules, got %v", cfg.IncludeCategories)
	}
}

func TestParseConfig_ExcludeCategoryFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--exclude-category", "python-venv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.ExcludeCategories) != 1 || cfg.ExcludeCategories[0] != "python-venv" {
		t.Errorf("expected exclude python-venv, got %v", cfg.ExcludeCategories)
	}
}

func TestParseConfig_ExcludeGroupFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--exclude-group", "/tmp/ignored"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.ExcludeGroups) != 1 {
		t.Errorf("expected 1 exclude-group, got %d", len(cfg.ExcludeGroups))
	}
}

func TestParseConfig_ExcludePathFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--exclude-path", "/tmp/specific"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.ExcludePaths) != 1 {
		t.Errorf("expected 1 exclude-path, got %d", len(cfg.ExcludePaths))
	}
}

func TestParseConfig_MultipleRepeatableFlags(t *testing.T) {
	cfg, err := ParseConfig([]string{
		"--include-category", "node-modules",
		"--include-category", "python-venv",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.IncludeCategories) != 2 {
		t.Errorf("expected 2 include-categories, got %d", len(cfg.IncludeCategories))
	}
}

func TestParseConfig_AbsoluteRoot(t *testing.T) {
	orig, _ := os.Getwd()
	os.Chdir("/tmp")
	defer os.Chdir(orig)

	cfg, err := ParseConfig([]string{"--root", "."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Root != "/tmp" {
		t.Errorf("expected root /tmp (absolute), got %s", cfg.Root)
	}
}

func TestParseConfig_TopLimitsZero(t *testing.T) {
	_, err := ParseConfig([]string{"--top-files", "0"})
	if err == nil {
		t.Fatal("expected error for top-files 0")
	}
}

func TestParseConfig_Complex(t *testing.T) {
	cfg, err := ParseConfig([]string{
		"analyze",
		"--root", "/home",
		"--format", "markdown",
		"--group-mode", "depth",
		"--group-depth", "2",
		"--top-files", "10",
		"--top-groups", "5",
		"--top-entries", "3",
		"--min-candidate-size", "524288",
		"--out", "report.md",
		"--include-category", "node-modules",
		"--exclude-category", "python-venv",
		"--exclude-group", "/home/user/.local",
		"--exclude-path", "/home/user/.cache",
		"--yes",
		"--log-level", "info",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "analyze" {
		t.Errorf("expected command analyze, got %s", cfg.Command)
	}
	if cfg.Format != "markdown" {
		t.Errorf("expected format markdown, got %s", cfg.Format)
	}
	if cfg.GroupMode != "depth" {
		t.Errorf("expected group-mode depth, got %s", cfg.GroupMode)
	}
	if cfg.GroupDepth != 2 {
		t.Errorf("expected group-depth 2, got %d", cfg.GroupDepth)
	}
	if cfg.TopFiles != 10 {
		t.Errorf("expected top-files 10, got %d", cfg.TopFiles)
	}
	if cfg.TopGroups != 5 {
		t.Errorf("expected top-groups 5, got %d", cfg.TopGroups)
	}
	if cfg.TopEntries != 3 {
		t.Errorf("expected top-entries 3, got %d", cfg.TopEntries)
	}
	if cfg.MinCandidateSize != 524288 {
		t.Errorf("expected min-candidate-size 524288, got %d", cfg.MinCandidateSize)
	}
	if cfg.OutFile != "report.md" {
		t.Errorf("expected out report.md, got %s", cfg.OutFile)
	}
	if !cfg.Yes {
		t.Error("expected yes to be true")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected log-level info, got %s", cfg.LogLevel)
	}
}
