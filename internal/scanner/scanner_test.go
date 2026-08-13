package scanner

import (
	"os"

	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAnalyzeFindsCandidatesAndGroupsByRepo(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "project")
	mustMkdir(t, filepath.Join(repo, ".git"))
	mustMkdir(t, filepath.Join(repo, "node_modules", "pkg"))
	mustMkdir(t, filepath.Join(repo, ".venv", "bin"))
	mustMkdir(t, filepath.Join(repo, "src", "__pycache__"))
	mustWriteFile(t, filepath.Join(repo, "node_modules", "pkg", "bundle.js"), strings.Repeat("a", 3<<20))
	mustWriteFile(t, filepath.Join(repo, ".venv", "bin", "python"), strings.Repeat("b", 2<<20))
	mustWriteFile(t, filepath.Join(repo, "src", "__pycache__", "main.cpython-311.pyc"), strings.Repeat("c", 128))

	report, err := AnalyzeWithOptions("analyze", AnalyzeOptions{
		Root:             root,
		GroupMode:        "repo",
		GroupDepth:       1,
		TopFiles:         10,
		TopGroups:        10,
		TopEntries:       10,
		MinCandidateSize: 1,
	}, nil)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if len(report.Candidates) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(report.Candidates))
	}

	for _, candidate := range report.Candidates {
		if candidate.Group != repo {
			t.Fatalf("expected candidate group %s, got %s", repo, candidate.Group)
		}
	}
}

func TestExcludeGroupSkipsNestedCandidates(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "workspace")
	repo := filepath.Join(workspace, "repo-a")
	mustMkdir(t, filepath.Join(repo, ".git"))
	mustMkdir(t, filepath.Join(repo, "node_modules"))
	mustWriteFile(t, filepath.Join(repo, "node_modules", "dep.js"), strings.Repeat("x", 1024))

	report, err := AnalyzeWithOptions("analyze", AnalyzeOptions{
		Root:             workspace,
		GroupMode:        "repo",
		GroupDepth:       1,
		TopFiles:         10,
		TopGroups:        10,
		TopEntries:       10,
		MinCandidateSize: 1,
		ExcludeGroups:    []string{workspace},
	}, nil)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if len(report.Candidates) != 1 {
		t.Fatalf("expected one detected candidate, got %d", len(report.Candidates))
	}
	if len(report.SelectedCandidates) != 0 {
		t.Fatalf("expected excluded group to drop selected candidates, got %d", len(report.SelectedCandidates))
	}
}

func TestMatchHelpersAndSummaries(t *testing.T) {
	if _, ok := MatchDirectory("node_modules"); !ok {
		t.Fatalf("expected node_modules to match directory category")
	}
	if _, ok := MatchFile("/tmp/demo/main.pyc"); !ok {
		t.Fatalf("expected .pyc to match file category")
	}

	modified := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	candidates := []Candidate{
		{CategoryKey: "node-modules", Category: "node_modules", Group: "/tmp/a", Bytes: 10, ModifiedAt: modified},
		{CategoryKey: "node-modules", Category: "node_modules", Group: "/tmp/a", Bytes: 20, ModifiedAt: modified.Add(1 * time.Hour)},
		{CategoryKey: "python-venv", Category: ".venv", Group: "/tmp/b", Bytes: 30, ModifiedAt: modified.Add(2 * time.Hour)},
	}

	cat := SummarizeCategories(candidates)
	if len(cat) != 2 || cat[0].Bytes != 30 {
		t.Fatalf("unexpected category summaries: %#v", cat)
	}
	groups := SummarizeGroups(candidates, 10)
	if len(groups) != 2 || groups[0].ModifiedAt.IsZero() {
		t.Fatalf("unexpected group summaries: %#v", groups)
	}
}

func TestPrefixAndSetHelpers(t *testing.T) {
	if !HasPathPrefix("/tmp/root/repo", "/tmp/root") {
		t.Fatalf("expected prefix match")
	}
	if HasPathPrefix("/tmp/other", "/tmp/root") {
		t.Fatalf("did not expect unrelated prefix match")
	}
	set := ListToSet([]string{"a", "b"})
	if _, ok := set["a"]; !ok {
		t.Fatalf("expected set entry")
	}
	if IncludeCategory("node-modules", ListToSet([]string{"node-modules"}), nil) != true {
		t.Fatalf("expected included category to pass")
	}
	if IncludeCategory("node-modules", nil, ListToSet([]string{"node-modules"})) != false {
		t.Fatalf("expected excluded category to fail")
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
