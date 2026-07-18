package cli

import (
	"strings"
	"testing"
)

// TestParseConfigDefaults validates all default values.
func TestParseConfigDefaults(t *testing.T) {
	cfg, err := ParseConfig([]string{})
	if err != nil {
		t.Fatalf("ParseConfig([]) returned error: %v", err)
	}
	if cfg.Command != "analyze" {
		t.Fatalf("expected command 'analyze', got %q", cfg.Command)
	}
	if cfg.Format != "plain" {
		t.Fatalf("expected format 'plain', got %q", cfg.Format)
	}
	if cfg.GroupMode != "repo" {
		t.Fatalf("expected groupMode 'repo', got %q", cfg.GroupMode)
	}
	if cfg.GroupDepth != 1 {
		t.Fatalf("expected groupDepth 1, got %d", cfg.GroupDepth)
	}
	if cfg.TopFiles != 20 {
		t.Fatalf("expected topFiles 20, got %d", cfg.TopFiles)
	}
	if cfg.TopGroups != 20 {
		t.Fatalf("expected topGroups 20, got %d", cfg.TopGroups)
	}
	if cfg.TopEntries != 15 {
		t.Fatalf("expected topEntries 15, got %d", cfg.TopEntries)
	}
	if cfg.MinCandidateSize != 1<<20 {
		t.Fatalf("expected minCandidateSize 1MB, got %d", cfg.MinCandidateSize)
	}
	if cfg.LogLevel != "warn" {
		t.Fatalf("expected logLevel 'warn', got %q", cfg.LogLevel)
	}
	if cfg.DryRun {
		t.Fatalf("expected dryRun false by default")
	}
	if cfg.Yes {
		t.Fatalf("expected yes false by default")
	}
}

// TestParseConfigCleanCommand validates clean command parsing.
func TestParseConfigCleanCommand(t *testing.T) {
	cfg, err := ParseConfig([]string{"clean", "--yes"})
	if err != nil {
		t.Fatalf("ParseConfig clean: %v", err)
	}
	if cfg.Command != "clean" {
		t.Fatalf("expected command 'clean', got %q", cfg.Command)
	}
	if !cfg.Yes {
		t.Fatalf("expected yes=true")
	}
}

// TestParseConfigDryRunFlag validates --dry-run flag parsing.
func TestParseConfigDryRunFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"clean", "--dry-run"})
	if err != nil {
		t.Fatalf("ParseConfig dry-run: %v", err)
	}
	if cfg.Command != "clean" {
		t.Fatalf("expected command 'clean', got %q", cfg.Command)
	}
	if !cfg.DryRun {
		t.Fatalf("expected dryRun=true")
	}
}

// TestParseConfigDryRunAndYes validates both flags together.
func TestParseConfigDryRunAndYes(t *testing.T) {
	cfg, err := ParseConfig([]string{"clean", "--dry-run", "--yes"})
	if err != nil {
		t.Fatalf("ParseConfig dry-run+yes: %v", err)
	}
	if !cfg.DryRun || !cfg.Yes {
		t.Fatalf("expected both dryRun and yes to be true")
	}
}

// TestParseConfigTUICommand validates tui command parsing.
func TestParseConfigTUICommand(t *testing.T) {
	cfg, err := ParseConfig([]string{"tui", "--format", "markdown"})
	if err != nil {
		t.Fatalf("ParseConfig tui: %v", err)
	}
	if cfg.Command != "tui" {
		t.Fatalf("expected command 'tui', got %q", cfg.Command)
	}
	if cfg.Format != "markdown" {
		t.Fatalf("expected format 'markdown', got %q", cfg.Format)
	}
}

// TestParseConfigHelpFlags validates help flag parsing.
func TestParseConfigHelpFlags(t *testing.T) {
	cfg, err := ParseConfig([]string{"--help"})
	if err != nil {
		t.Fatalf("ParseConfig --help: %v", err)
	}
	if cfg.Command != "help" {
		t.Fatalf("expected command 'help', got %q", cfg.Command)
	}

	cfg, err = ParseConfig([]string{"clean", "-h"})
	if err != nil {
		t.Fatalf("ParseConfig clean -h: %v", err)
	}
	if cfg.Command != "help" {
		t.Fatalf("expected command 'help', got %q", cfg.Command)
	}
}

// TestParseConfigVersionFlag validates version flag parsing.
func TestParseConfigVersionFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--version"})
	if err != nil {
		t.Fatalf("ParseConfig --version: %v", err)
	}
	if cfg.Command != "version" {
		t.Fatalf("expected command 'version', got %q", cfg.Command)
	}
}

// TestParseConfigInvalidFormat validates format validation.
func TestParseConfigInvalidFormat(t *testing.T) {
	_, err := ParseConfig([]string{"--format", "xml"})
	if err == nil {
		t.Fatalf("expected error for unsupported format 'xml'")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("expected 'unsupported format' error, got: %v", err)
	}
}

func TestParseConfigInvalidGroupMode(t *testing.T) {
	_, err := ParseConfig([]string{"--group-mode", "invalid"})
	if err == nil {
		t.Fatalf("expected error for unsupported group-mode")
	}
	if !strings.Contains(err.Error(), "unsupported group mode") {
		t.Fatalf("expected 'unsupported group mode' error, got: %v", err)
	}
}

// TestParseConfigInvalidGroupDepth validates group-depth >= 1.
func TestParseConfigInvalidGroupDepth(t *testing.T) {
	_, err := ParseConfig([]string{"--group-depth", "0"})
	if err == nil {
		t.Fatalf("expected error for group-depth 0")
	}
	if !strings.Contains(err.Error(), ">= 1") {
		t.Fatalf("expected '>= 1' error, got: %v", err)
	}
}

// TestParseConfigInvalidTopLimits validates top limits >= 1.
func TestParseConfigInvalidTopLimits(t *testing.T) {
	_, err := ParseConfig([]string{"--top-files", "0"})
	if err == nil {
		t.Fatalf("expected error for top-files 0")
	}

	_, err = ParseConfig([]string{"--top-groups", "-1"})
	if err == nil {
		t.Fatalf("expected error for top-groups -1")
	}
}

// TestParseConfigInvalidLogLevel validates log level validation.
func TestParseConfigInvalidLogLevel(t *testing.T) {
	_, err := ParseConfig([]string{"--log-level", "verbose"})
	if err == nil {
		t.Fatalf("expected error for invalid log level 'verbose'")
	}
	if !strings.Contains(err.Error(), "unsupported log level") {
		t.Fatalf("expected 'unsupported log level' error, got: %v", err)
	}
}

// TestParseConfigValidLogLevels validates all accepted log levels.
func TestParseConfigValidLogLevels(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error"} {
		cfg, err := ParseConfig([]string{"--log-level", level})
		if err != nil {
			t.Fatalf("log level %q: %v", level, err)
		}
		if cfg.LogLevel != level {
			t.Fatalf("log level %q: expected %q, got %q", level, level, cfg.LogLevel)
		}
	}
}

// TestParseConfigIncludeExcludeCategories validates category flags.
func TestParseConfigIncludeExcludeCategories(t *testing.T) {
	cfg, err := ParseConfig([]string{
		"--include-category", "node-modules",
		"--include-category", "python-venv",
		"--exclude-category", "__pycache__",
	})
	if err != nil {
		t.Fatalf("ParseConfig categories: %v", err)
	}
	if len(cfg.IncludeCategories) != 2 {
		t.Fatalf("expected 2 include categories, got %d", len(cfg.IncludeCategories))
	}
	if len(cfg.ExcludeCategories) != 1 {
		t.Fatalf("expected 1 exclude category, got %d", len(cfg.ExcludeCategories))
	}
}

// TestParseConfigExcludeGroupsAndPaths validates group and path exclusion flags.
func TestParseConfigExcludeGroupsAndPaths(t *testing.T) {
	cfg, err := ParseConfig([]string{
		"--exclude-group", "/tmp/foo",
		"--exclude-group", "/tmp/bar",
		"--exclude-path", "/tmp/foo/node_modules",
	})
	if err != nil {
		t.Fatalf("ParseConfig exclude: %v", err)
	}
	if len(cfg.ExcludeGroups) != 2 {
		t.Fatalf("expected 2 exclude groups, got %d", len(cfg.ExcludeGroups))
	}
	if len(cfg.ExcludePaths) != 1 {
		t.Fatalf("expected 1 exclude path, got %d", len(cfg.ExcludePaths))
	}
}

// TestParseConfigRootAbsPath validates root is converted to absolute path.
func TestParseConfigRootAbsPath(t *testing.T) {
	cfg, err := ParseConfig([]string{"--root", "."})
	if err != nil {
		t.Fatalf("ParseConfig root: %v", err)
	}
	if !filepathIsAbs(cfg.Root) {
		t.Fatalf("expected absolute root path, got %q", cfg.Root)
	}
}

// TestParseConfigOutFile validates --out flag.
func TestParseConfigOutFile(t *testing.T) {
	cfg, err := ParseConfig([]string{"--out", "/tmp/report.md"})
	if err != nil {
		t.Fatalf("ParseConfig out: %v", err)
	}
	if cfg.OutFile != "/tmp/report.md" {
		t.Fatalf("expected outFile '/tmp/report.md', got %q", cfg.OutFile)
	}
}

// TestParseConfigMinCandidateSize validates --min-candidate-size flag.
func TestParseConfigMinCandidateSize(t *testing.T) {
	cfg, err := ParseConfig([]string{"--min-candidate-size", "1024"})
	if err != nil {
		t.Fatalf("ParseConfig min-candidate-size: %v", err)
	}
	if cfg.MinCandidateSize != 1024 {
		t.Fatalf("expected minCandidateSize 1024, got %d", cfg.MinCandidateSize)
	}
}

// TestParseConfigHelpTopic validates help topic extraction.
func TestParseConfigHelpTopic(t *testing.T) {
	cfg, err := ParseConfig([]string{"help", "clean"})
	if err != nil {
		t.Fatalf("ParseConfig help clean: %v", err)
	}
	if cfg.HelpTopic != "clean" {
		t.Fatalf("expected helpTopic 'clean', got %q", cfg.HelpTopic)
	}
}

// filepathIsAbs is a local copy of filepath.IsAbs for testing.
func filepathIsAbs(path string) bool {
	return len(path) > 0 && path[0] == '/'
}

// TestParseConfigQuietFlag validates --quiet flag parsing.
func TestParseConfigQuietFlag(t *testing.T) {
	cfg, err := ParseConfig([]string{"--quiet"})
	if err != nil {
		t.Fatalf("ParseConfig --quiet: %v", err)
	}
	if !cfg.Quiet {
		t.Fatalf("expected quiet=true")
	}
	if cfg.LogLevel != "warn" {
		t.Fatalf("expected logLevel 'warn' by default (quiet override happens in main.go)")
	}
}

// TestParseConfigQuietOverridesLogLevel validates that --quiet + --log-level debug
// results in quiet being set (logLevel override to error happens in main.go).
func TestParseConfigQuietOverridesLogLevel(t *testing.T) {
	cfg, err := ParseConfig([]string{"--quiet", "--log-level", "debug"})
	if err != nil {
		t.Fatalf("ParseConfig quiet+debug: %v", err)
	}
	if !cfg.Quiet {
		t.Fatalf("expected quiet=true")
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("expected logLevel 'debug' (quiet override happens in main.go)")
	}
}
