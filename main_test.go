package reclaimit

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/svg153/reclaimit/internal/cli"
	"github.com/svg153/reclaimit/internal/tui"
)

func mustMkdirRootTest(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func TestRun_Version(t *testing.T) {
	old := Version
	Version = "1.2.3"
	defer func() { Version = old }()

	var buf bytes.Buffer
	code := Run([]string{"version"}, &buf, &buf)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(buf.String(), "1.2.3") {
		t.Errorf("expected version output, got %q", buf.String())
	}
}

func TestRun_Help(t *testing.T) {
	var buf bytes.Buffer
	code := Run([]string{"help"}, &buf, &buf)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(buf.String(), "reclaimit") {
		t.Errorf("expected help output, got %q", buf.String())
	}
}

func TestRun_HelpTopic(t *testing.T) {
	var buf bytes.Buffer
	code := Run([]string{"help", "analyze"}, &buf, &buf)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

func TestRun_InvalidCommand(t *testing.T) {
	var buf bytes.Buffer
	code := Run([]string{"analyze", "--format", "xml"}, &buf, &buf)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(buf.String(), "unsupported format") {
		t.Errorf("expected format error, got %q", buf.String())
	}
}

func TestRun_QuietMode(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".DS_Store"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	var buf bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--quiet"}, &buf, &buf)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestToScannerOpts(t *testing.T) {
	opts := toScannerOpts(cli.Options{
		Root:              "/tmp/test",
		GroupMode:         "repo",
		GroupDepth:        2,
		TopFiles:          10,
		TopGroups:         5,
		TopEntries:        20,
		MinCandidateSize:  100,
		IncludeCategories: []string{"node-modules"},
		ExcludeCategories: []string{"python-venv"},
		ExcludeGroups:     []string{"test"},
		ExcludePaths:      []string{"/tmp/ignored"},
	})

	if opts.Root != "/tmp/test" {
		t.Errorf("Root = %q, want %q", opts.Root, "/tmp/test")
	}
	if opts.GroupMode != "repo" {
		t.Errorf("GroupMode = %q, want %q", opts.GroupMode, "repo")
	}
	if opts.GroupDepth != 2 {
		t.Errorf("GroupDepth = %d, want 2", opts.GroupDepth)
	}
	if len(opts.IncludeCategories) != 1 || opts.IncludeCategories[0] != "node-modules" {
		t.Errorf("IncludeCategories = %v, want [node-modules]", opts.IncludeCategories)
	}
	if len(opts.ExcludeCategories) != 1 || opts.ExcludeCategories[0] != "python-venv" {
		t.Errorf("ExcludeCategories = %v, want [python-venv]", opts.ExcludeCategories)
	}
	if len(opts.ExcludeGroups) != 1 || opts.ExcludeGroups[0] != "test" {
		t.Errorf("ExcludeGroups = %v, want [test]", opts.ExcludeGroups)
	}
}

func TestWriteSelection(t *testing.T) {
	var buf bytes.Buffer
	selection := tui.Selection{
		ExcludedGroups: []string{"group1"},
		ExcludedPaths:  []string{"/path1"},
	}
	cfg := cli.Options{Root: "/tmp/test"}
	err := writeSelection(&buf, cfg, selection)
	if err != nil {
		t.Fatalf("writeSelection returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "Selection") {
		t.Error("expected 'Selection' in output")
	}
	if !strings.Contains(buf.String(), "group1") {
		t.Error("expected group1 in output")
	}
	if !strings.Contains(buf.String(), "/path1") {
		t.Error("expected /path1 in output")
	}
}

func TestWriteOutput_ToFile(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "output.txt")
	var buf bytes.Buffer

	code := writeOutput(&buf, &buf, outFile, "hello world")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("file content = %q, want %q", string(content), "hello world")
	}
}

func TestWriteOutput_ToStdout(t *testing.T) {
	var buf bytes.Buffer
	code := writeOutput(&buf, &buf, "", "stdout output")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if buf.String() != "stdout output" {
		t.Errorf("stdout = %q, want %q", buf.String(), "stdout output")
	}
}

func TestWriteOutput_FileErrorUsesStderrAndFails(t *testing.T) {
	var stdout, stderr bytes.Buffer
	outFile := filepath.Join(t.TempDir(), "missing", "report.txt")
	code := writeOutput(&stdout, &stderr, outFile, "report")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected stdout to remain empty, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "error: writing report:") {
		t.Fatalf("expected write error on stderr, got %q", stderr.String())
	}
}

func TestHumanizeBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TiB"},
		{1536, "1.5 KiB"},
		{1048576 * 100, "100.0 MiB"},
	}
	for _, tt := range tests {
		result := humanizeBytes(tt.input)
		if result != tt.expected {
			t.Errorf("humanizeBytes(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExitf(t *testing.T) {
	var buf bytes.Buffer
	code := exitf(&buf, "error: %s", "test")
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(buf.String(), "error: test") {
		t.Errorf("output = %q, want 'error: test'", buf.String())
	}
}

func TestWriteString(t *testing.T) {
	var buf bytes.Buffer
	err := writeString(&buf, "hello")
	if err != nil {
		t.Fatalf("writeString returned error: %v", err)
	}
	if buf.String() != "hello" {
		t.Errorf("output = %q, want %q", buf.String(), "hello")
	}
}

func TestWritef(t *testing.T) {
	var buf bytes.Buffer
	err := writef(&buf, "%d %s", 42, "test")
	if err != nil {
		t.Fatalf("writef returned error: %v", err)
	}
	if buf.String() != "42 test" {
		t.Errorf("output = %q, want %q", buf.String(), "42 test")
	}
}

func TestRun_AnalyzeInvalidFormat(t *testing.T) {
	var stderr bytes.Buffer
	code := Run([]string{"analyze", "--format", "xml", "--root", t.TempDir()}, &bytes.Buffer{}, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unsupported format") {
		t.Fatalf("expected error about unsupported format, got %q", stderr.String())
	}
}

func TestRun_AnalyzeMissingRoot(t *testing.T) {
	// analyze without -root: the flag package doesn't error on missing required flags
	// it just runs with zero value (empty string)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze"}, &stdout, &stderr)
	// Should succeed (exit 0) but produce empty/minimal output
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%q", code, stderr.String())
	}
}

func TestRun_AnalyzeQuietMode(t *testing.T) {
	var stdout, stderr bytes.Buffer
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))
	code := Run([]string{"analyze", "--root", root, "--quiet", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%q", code, stderr.String())
	}
	// Quiet mode suppresses verbose output but still produces the report
	if stderr.String() != "" {
		t.Fatalf("quiet mode should not write to stderr: %q", stderr.String())
	}
}

func TestRun_AnalyzeWithSelection(t *testing.T) {
	root := t.TempDir()
	// Create a node_modules directory to be detected as a candidate
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "node_modules") {
		t.Fatalf("expected node_modules in output, got: %s", stdout.String())
	}
}

func TestRun_AnalyzeWithIgnoreFile(t *testing.T) {
	root := t.TempDir()
	// Create a .reclaimitignore file
	ignoreFile := filepath.Join(root, ".reclaimitignore")
	if err := os.WriteFile(ignoreFile, []byte("node_modules\n"), 0o644); err != nil {
		t.Fatalf("write ignore file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--ignore-file", ignoreFile, "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithSelectionFile(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithInvalidLogLevel(t *testing.T) {
	var stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", t.TempDir(), "--log-level", "invalid"}, &bytes.Buffer{}, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unsupported log level") {
		t.Fatalf("expected error about unsupported log level, got %q", stderr.String())
	}
}

func TestRun_AnalyzeWithJSONOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", t.TempDir(), "--format", "json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if !json.Valid(stdout.Bytes()) {
		t.Fatalf("expected valid JSON, got %q", stdout.String())
	}
}

func TestRun_AnalyzeWithMarkdownOutput(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--format", "markdown", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "# Disk usage report") {
		t.Fatalf("expected markdown header in output, got: %s", stdout.String())
	}
}

func TestRun_AnalyzeWithOutputFile(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	outputFile := filepath.Join(root, "output.txt")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--format", "plain", "--out", outputFile, "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(data), "Disk usage report") {
		t.Fatalf("expected output file to contain report, got: %s", string(data))
	}
}

func TestRun_AnalyzeWithMinCandidateSize(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithExcludePath(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--exclude-path", filepath.Join(root, "node_modules"), "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithIncludeCategory(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--include-category", "cache", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_CleanHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"clean", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "reclaimit clean") {
		t.Fatalf("expected clean help text, got: %s", stdout.String())
	}
}

func TestRun_AnalyzeWithGroupModeDepth(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--group-mode", "depth", "--group-depth", "2", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithGroupModeRepo(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--group-mode", "repo", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithTopEntries(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--top-entries", "5", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithTopFiles(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--top-files", "10", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithTopGroups(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--top-groups", "3", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithExcludeGroup(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--exclude-group", filepath.Join(root, "node_modules"), "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithExcludeCategory(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--exclude-category", "cache", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithMultipleCategories(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--include-category", "cache", "--include-category", "build", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_VersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "reclaimit") {
		t.Fatalf("expected version output, got: %s", stdout.String())
	}
}

func TestRun_HelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "reclaimit") {
		t.Fatalf("expected help output, got: %s", stdout.String())
	}
}

func TestRun_AnalyzeWithLogLevelWarn(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--log-level", "warn", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithLogLevelError(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--log-level", "error", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_CleanDryRun(t *testing.T) {
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"clean", "--root", root, "--dry-run", "--yes", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "DRY RUN") {
		t.Fatalf("expected DRY RUN output, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Disk usage report") {
		t.Fatalf("expected post-clean report, got: %s", stdout.String())
	}
}

func TestRun_CleanWithoutYesOrDryRun(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"clean", "--root", t.TempDir()}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "clean requires --yes or --dry-run") {
		t.Fatalf("expected error about --yes/--dry-run, got %q", stderr.String())
	}
}

func TestRun_TUIHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"tui", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "reclaimit tui") {
		t.Fatalf("expected tui help, got: %s", stdout.String())
	}
}

func TestRun_AnalyzeWithYesFlag(t *testing.T) {
	// --yes is a clean flag but should not cause issues when analyzing
	root := t.TempDir()
	mustMkdirRootTest(t, filepath.Join(root, "node_modules"))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--yes", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

type failNthWriter struct {
	writes int
	failAt int
}

func (writer *failNthWriter) Write(p []byte) (int, error) {
	if writer.writes == writer.failAt {
		return 0, errors.New("write failed")
	}
	writer.writes++
	return len(p), nil
}

func TestRunReportsWriterAndRootErrors(t *testing.T) {
	for name, args := range map[string][]string{
		"help":    {"help"},
		"version": {"version"},
	} {
		t.Run(name, func(t *testing.T) {
			var stderr bytes.Buffer
			if code := Run(args, failingWriter{}, &stderr); code != 1 {
				t.Fatalf("code = %d, stderr=%q", code, stderr.String())
			}
		})
	}
	var stderr bytes.Buffer
	if code := Run([]string{"analyze", "--root", filepath.Join(t.TempDir(), "missing")}, io.Discard, &stderr); code != 1 {
		t.Fatalf("missing root code = %d", code)
	}
}

func TestRunTUIWithInjectedSelection(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "node_modules")
	mustMkdirRootTest(t, target)
	original := runTUI
	t.Cleanup(func() { runTUI = original })
	runTUI = func(report tui.Report) (tui.Selection, error) {
		if len(report.Candidates) != 1 {
			t.Fatalf("expected one TUI candidate, got %#v", report.Candidates)
		}
		return tui.Selection{ExcludedPaths: []string{target}, Saved: true}, nil
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"tui", "--root", root, "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Reproduce this selection") || !strings.Contains(stdout.String(), target) {
		t.Fatalf("selection output missing: %s", stdout.String())
	}
}

func TestRunTUIError(t *testing.T) {
	original := runTUI
	t.Cleanup(func() { runTUI = original })
	runTUI = func(tui.Report) (tui.Selection, error) {
		return tui.Selection{}, errors.New("terminal unavailable")
	}
	var stderr bytes.Buffer
	if code := Run([]string{"tui", "--root", t.TempDir()}, io.Discard, &stderr); code != 1 {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(stderr.String(), "terminal unavailable") {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
}

func TestRunCleanDeletesCandidate(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "node_modules")
	mustMkdirRootTest(t, target)
	if err := os.WriteFile(filepath.Join(target, "dep.js"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"clean", "--root", root, "--yes", "--min-candidate-size", "1"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("candidate still exists: %v", err)
	}
	if !strings.Contains(stdout.String(), "[CLEAN] Deleted 1 B") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestWriteSelectionAndOutputPropagateWriterErrors(t *testing.T) {
	if err := writeSelection(failingWriter{}, cli.Options{Root: "/tmp"}, tui.Selection{}); err == nil {
		t.Fatal("writeSelection should fail")
	}
	if code := writeOutput(failingWriter{}, io.Discard, "", "report"); code != 1 {
		t.Fatalf("writeOutput code = %d", code)
	}
}

func TestWriteSelectionPropagatesEveryWriteFailure(t *testing.T) {
	selection := tui.Selection{
		ExcludedGroups: []string{"/tmp/group"},
		ExcludedPaths:  []string{"/tmp/path"},
	}
	for failAt := 0; failAt < 6; failAt++ {
		writer := &failNthWriter{failAt: failAt}
		if err := writeSelection(writer, cli.Options{Root: "/tmp"}, selection); err == nil {
			t.Fatalf("expected failure at write %d", failAt)
		}
	}
}

func TestRunCleanPropagatesPreviewAndSummaryWriteErrors(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "node_modules")
	mustMkdirRootTest(t, target)
	if err := os.WriteFile(filepath.Join(target, "dep.js"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, dryRun := range []bool{true, false} {
		args := []string{"clean", "--root", root, "--min-candidate-size", "1"}
		if dryRun {
			args = append(args, "--dry-run")
		} else {
			args = append(args, "--yes")
		}
		writer := &failNthWriter{failAt: 1}
		if code := Run(args, writer, io.Discard); code != 1 {
			t.Fatalf("dryRun=%v code=%d", dryRun, code)
		}
		if !dryRun {
			// Recreate the candidate removed before the summary write failed.
			mustMkdirRootTest(t, target)
			if err := os.WriteFile(filepath.Join(target, "dep.js"), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
}
