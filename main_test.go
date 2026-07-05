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
