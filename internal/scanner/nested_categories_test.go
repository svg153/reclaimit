package scanner

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestMatchDirectoryPathSpecificCategories(t *testing.T) {
	root := filepath.Join(t.TempDir(), "home")
	for _, path := range []string{
		filepath.Join(root, ".cache", "pip"),
		filepath.Join(root, ".local", "pipx"),
	} {
		cat, ok := MatchDirectory(path)
		if !ok {
			t.Fatalf("MatchDirectory(%q) did not match", path)
		}
		if cat.Key != map[string]string{
			filepath.Join(root, ".cache", "pip"): "pip-cache",
			filepath.Join(root, ".local", "pipx"): "pipx-data",
		}[path] {
			t.Fatalf("MatchDirectory(%q) returned %q", path, cat.Key)
		}
	}
}

func TestAnalyzeFindsNestedPipDirectories(t *testing.T) {
	root := t.TempDir()
	pip := filepath.Join(root, ".cache", "pip")
	pipx := filepath.Join(root, ".local", "pipx")
	for _, dir := range []string{pip, pipx} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "payload"), []byte("cache"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	report, err := AnalyzeWithOptions("analyze", AnalyzeOptions{
		Root: root, GroupMode: "repo", GroupDepth: 1,
		TopFiles: 10, TopGroups: 10, TopEntries: 10, MinCandidateSize: 1,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{pip: "pip-cache", pipx: "pipx-data"}
	for path, key := range want {
		found := false
		for _, candidate := range report.Candidates {
			if candidate.Path == path && candidate.CategoryKey == key {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected %s as %s candidate, got %#v", path, key, report.Candidates)
		}
	}
}
