package reclaimit

import (
	"strings"
	"testing"
)

// TestParseConfigDefaults validates all default values.
func TestParseConfigDefaults(t *testing.T) {
	cfg, err := parseConfig([]string{})
	if err != nil {
		t.Fatalf("parseConfig([]) returned error: %v", err)
	}
	if cfg.command != "analyze" {
		t.Fatalf("expected command 'analyze', got %q", cfg.command)
	}
	if cfg.format != "plain" {
		t.Fatalf("expected format 'plain', got %q", cfg.format)
	}
	if cfg.groupMode != "repo" {
		t.Fatalf("expected groupMode 'repo', got %q", cfg.groupMode)
	}
	if cfg.groupDepth != 1 {
		t.Fatalf("expected groupDepth 1, got %d", cfg.groupDepth)
	}
	if cfg.topFiles != 20 {
		t.Fatalf("expected topFiles 20, got %d", cfg.topFiles)
	}
	if cfg.topGroups != 20 {
		t.Fatalf("expected topGroups 20, got %d", cfg.topGroups)
	}
	if cfg.topEntries != 15 {
		t.Fatalf("expected topEntries 15, got %d", cfg.topEntries)
	}
	if cfg.minCandidateSize != 1<<20 {
		t.Fatalf("expected minCandidateSize 1MB, got %d", cfg.minCandidateSize)
	}
	if cfg.logLevel != "warn" {
		t.Fatalf("expected logLevel 'warn', got %q", cfg.logLevel)
	}
	if cfg.dryRun {
		t.Fatalf("expected dryRun false by default")
	}
	if cfg.yes {
		t.Fatalf("expected yes false by default")
	}
}

// TestParseConfigCleanCommand validates clean command parsing.
func TestParseConfigCleanCommand(t *testing.T) {
	cfg, err := parseConfig([]string{"clean", "--yes"})
	if err != nil {
		t.Fatalf("parseConfig clean: %v", err)
	}
	if cfg.command != "clean" {
		t.Fatalf("expected command 'clean', got %q", cfg.command)
	}
	if !cfg.yes {
		t.Fatalf("expected yes=true")
	}
}

// TestParseConfigDryRunFlag validates --dry-run flag parsing.
func TestParseConfigDryRunFlag(t *testing.T) {
	cfg, err := parseConfig([]string{"clean", "--dry-run"})
	if err != nil {
		t.Fatalf("parseConfig dry-run: %v", err)
	}
	if cfg.command != "clean" {
		t.Fatalf("expected command 'clean', got %q", cfg.command)
	}
	if !cfg.dryRun {
		t.Fatalf("expected dryRun=true")
	}
}

// TestParseConfigDryRunAndYes validates both flags together.
func TestParseConfigDryRunAndYes(t *testing.T) {
	cfg, err := parseConfig([]string{"clean", "--dry-run", "--yes"})
	if err != nil {
		t.Fatalf("parseConfig dry-run+yes: %v", err)
	}
	if !cfg.dryRun || !cfg.yes {
		t.Fatalf("expected both dryRun and yes to be true")
	}
}

// TestParseConfigTUICommand validates tui command parsing.
func TestParseConfigTUICommand(t *testing.T) {
	cfg, err := parseConfig([]string{"tui", "--format", "markdown"})
	if err != nil {
		t.Fatalf("parseConfig tui: %v", err)
	}
	if cfg.command != "tui" {
		t.Fatalf("expected command 'tui', got %q", cfg.command)
	}
	if cfg.format != "markdown" {
		t.Fatalf("expected format 'markdown', got %q", cfg.format)
	}
}

// TestParseConfigHelpFlags validates help flag parsing.
func TestParseConfigHelpFlags(t *testing.T) {
	cfg, err := parseConfig([]string{"--help"})
	if err != nil {
		t.Fatalf("parseConfig --help: %v", err)
	}
	if cfg.command != "help" {
		t.Fatalf("expected command 'help', got %q", cfg.command)
	}

	cfg, err = parseConfig([]string{"clean", "-h"})
	if err != nil {
		t.Fatalf("parseConfig clean -h: %v", err)
	}
	if cfg.command != "help" {
		t.Fatalf("expected command 'help', got %q", cfg.command)
	}
}

// TestParseConfigVersionFlag validates version flag parsing.
func TestParseConfigVersionFlag(t *testing.T) {
	cfg, err := parseConfig([]string{"--version"})
	if err != nil {
		t.Fatalf("parseConfig --version: %v", err)
	}
	if cfg.command != "version" {
		t.Fatalf("expected command 'version', got %q", cfg.command)
	}
}

// TestParseConfigInvalidFormat validates format validation.
func TestParseConfigInvalidFormat(t *testing.T) {
	_, err := parseConfig([]string{"--format", "json"})
	if err == nil {
		t.Fatalf("expected error for unsupported format 'json'")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("expected 'unsupported format' error, got: %v", err)
	}
}

// TestParseConfigInvalidGroupMode validates group-mode validation.
func TestParseConfigInvalidGroupMode(t *testing.T) {
	_, err := parseConfig([]string{"--group-mode", "invalid"})
	if err == nil {
		t.Fatalf("expected error for unsupported group-mode")
	}
	if !strings.Contains(err.Error(), "unsupported group mode") {
		t.Fatalf("expected 'unsupported group mode' error, got: %v", err)
	}
}

// TestParseConfigInvalidGroupDepth validates group-depth >= 1.
func TestParseConfigInvalidGroupDepth(t *testing.T) {
	_, err := parseConfig([]string{"--group-depth", "0"})
	if err == nil {
		t.Fatalf("expected error for group-depth 0")
	}
	if !strings.Contains(err.Error(), ">= 1") {
		t.Fatalf("expected '>= 1' error, got: %v", err)
	}
}

// TestParseConfigInvalidTopLimits validates top limits >= 1.
func TestParseConfigInvalidTopLimits(t *testing.T) {
	_, err := parseConfig([]string{"--top-files", "0"})
	if err == nil {
		t.Fatalf("expected error for top-files 0")
	}

	_, err = parseConfig([]string{"--top-groups", "-1"})
	if err == nil {
		t.Fatalf("expected error for top-groups -1")
	}
}

// TestParseConfigInvalidLogLevel validates log level validation.
func TestParseConfigInvalidLogLevel(t *testing.T) {
	_, err := parseConfig([]string{"--log-level", "verbose"})
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
		cfg, err := parseConfig([]string{"--log-level", level})
		if err != nil {
			t.Fatalf("log level %q: %v", level, err)
		}
		if cfg.logLevel != level {
			t.Fatalf("log level %q: expected %q, got %q", level, level, cfg.logLevel)
		}
	}
}

// TestParseConfigIncludeExcludeCategories validates category flags.
func TestParseConfigIncludeExcludeCategories(t *testing.T) {
	cfg, err := parseConfig([]string{
		"--include-category", "node-modules",
		"--include-category", "python-venv",
		"--exclude-category", "__pycache__",
	})
	if err != nil {
		t.Fatalf("parseConfig categories: %v", err)
	}
	if len(cfg.includeCategories) != 2 {
		t.Fatalf("expected 2 include categories, got %d", len(cfg.includeCategories))
	}
	if len(cfg.excludeCategories) != 1 {
		t.Fatalf("expected 1 exclude category, got %d", len(cfg.excludeCategories))
	}
}

// TestParseConfigExcludeGroupsAndPaths validates group and path exclusion flags.
func TestParseConfigExcludeGroupsAndPaths(t *testing.T) {
	cfg, err := parseConfig([]string{
		"--exclude-group", "/tmp/foo",
		"--exclude-group", "/tmp/bar",
		"--exclude-path", "/tmp/foo/node_modules",
	})
	if err != nil {
		t.Fatalf("parseConfig exclude: %v", err)
	}
	if len(cfg.excludeGroups) != 2 {
		t.Fatalf("expected 2 exclude groups, got %d", len(cfg.excludeGroups))
	}
	if len(cfg.excludePaths) != 1 {
		t.Fatalf("expected 1 exclude path, got %d", len(cfg.excludePaths))
	}
}

// TestParseConfigRootAbsPath validates root is converted to absolute path.
func TestParseConfigRootAbsPath(t *testing.T) {
	cfg, err := parseConfig([]string{"--root", "."})
	if err != nil {
		t.Fatalf("parseConfig root: %v", err)
	}
	if !filepathIsAbs(cfg.root) {
		t.Fatalf("expected absolute root path, got %q", cfg.root)
	}
}

// TestParseConfigOutFile validates --out flag.
func TestParseConfigOutFile(t *testing.T) {
	cfg, err := parseConfig([]string{"--out", "/tmp/report.md"})
	if err != nil {
		t.Fatalf("parseConfig out: %v", err)
	}
	if cfg.outFile != "/tmp/report.md" {
		t.Fatalf("expected outFile '/tmp/report.md', got %q", cfg.outFile)
	}
}

// TestParseConfigMinCandidateSize validates --min-candidate-size flag.
func TestParseConfigMinCandidateSize(t *testing.T) {
	cfg, err := parseConfig([]string{"--min-candidate-size", "1024"})
	if err != nil {
		t.Fatalf("parseConfig min-candidate-size: %v", err)
	}
	if cfg.minCandidateSize != 1024 {
		t.Fatalf("expected minCandidateSize 1024, got %d", cfg.minCandidateSize)
	}
}

// TestParseConfigHelpTopic validates help topic extraction.
func TestParseConfigHelpTopic(t *testing.T) {
	cfg, err := parseConfig([]string{"help", "clean"})
	if err != nil {
		t.Fatalf("parseConfig help clean: %v", err)
	}
	if cfg.helpTopic != "clean" {
		t.Fatalf("expected helpTopic 'clean', got %q", cfg.helpTopic)
	}
}

// filepathIsAbs is a local copy of filepath.IsAbs for testing.
func filepathIsAbs(path string) bool {
	return len(path) > 0 && path[0] == '/'
}
