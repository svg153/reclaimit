package reclaimit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/svg153/reclaimit/internal/cli"
	"github.com/svg153/reclaimit/internal/tui"
)

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
	os.MkdirAll(filepath.Join(root, ".git"), 0o755)
	os.WriteFile(filepath.Join(root, ".DS_Store"), []byte("x"), 0o644)

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

	code := writeOutput(&buf, outFile, "hello world")
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
	code := writeOutput(&buf, "", "stdout output")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if buf.String() != "stdout output" {
		t.Errorf("stdout = %q, want %q", buf.String(), "stdout output")
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
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)
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
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

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
	os.WriteFile(ignoreFile, []byte("node_modules\n"), 0o644)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--ignore-file", ignoreFile, "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithSelectionFile(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

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
	// json is not a supported format — this should fail
	var stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", t.TempDir(), "--format", "json"}, &bytes.Buffer{}, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unsupported format") {
		t.Fatalf("expected error about unsupported format, got %q", stderr.String())
	}
}

func TestRun_AnalyzeWithMarkdownOutput(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

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
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

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
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithExcludePath(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--exclude-path", filepath.Join(root, "node_modules"), "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithIncludeCategory(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

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
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--group-mode", "depth", "--group-depth", "2", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithGroupModeRepo(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--group-mode", "repo", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithTopEntries(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--top-entries", "5", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithTopFiles(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--top-files", "10", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithTopGroups(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--top-groups", "3", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithExcludeGroup(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--exclude-group", filepath.Join(root, "node_modules"), "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithExcludeCategory(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--exclude-category", "cache", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithMultipleCategories(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

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
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--log-level", "warn", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}

func TestRun_AnalyzeWithLogLevelError(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--log-level", "error", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}


func TestRun_CleanDryRun(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"clean", "--root", root, "--dry-run", "--yes", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "DRY RUN") {
		t.Fatalf("expected DRY RUN output, got: %s", stdout.String())
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
	os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--yes", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", code, stderr.String())
	}
}
