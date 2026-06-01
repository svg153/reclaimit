package reclaimit

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

	report, err := Analyze(config{
		command:          "analyze",
		root:             root,
		format:           "plain",
		groupMode:        "repo",
		groupDepth:       1,
		topFiles:         10,
		topGroups:        10,
		topEntries:       10,
		minCandidateSize: 1,
	})
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

	report, err := Analyze(config{
		command:          "analyze",
		root:             root,
		format:           "plain",
		groupMode:        "repo",
		groupDepth:       1,
		topFiles:         10,
		topGroups:        10,
		topEntries:       10,
		minCandidateSize: 1,
		excludeGroups:    []string{workspace},
	})
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

func TestRenderMarkdown(t *testing.T) {
	report := Report{
		Root: "/tmp/demo",
		TopEntries: []PathSize{
			{Path: "/tmp/demo/cache", Bytes: 1024},
		},
		TopFiles: []PathSize{
			{Path: "/tmp/demo/file.bin", Bytes: 2048},
		},
		CategorySummaries: []CategorySummary{
			{CategoryKey: "node-modules", Description: "safe", Bytes: 4096, Count: 1},
		},
		GroupSummaries: []GroupSummary{
			{Group: "/tmp/demo/project", Bytes: 4096, Count: 1},
		},
		SelectedCandidates: []Candidate{
			{CategoryKey: "node-modules", Path: "/tmp/demo/project/node_modules", Bytes: 4096},
		},
		CandidateBytes: 4096,
	}

	out, err := RenderReport(report, "markdown")
	if err != nil {
		t.Fatalf("RenderReport returned error: %v", err)
	}
	if !strings.Contains(out, "| `node-modules` |") {
		t.Fatalf("expected markdown table, got %s", out)
	}
}

func TestMatchHelpersAndSummaries(t *testing.T) {
	if _, ok := matchDirectory("node_modules"); !ok {
		t.Fatalf("expected node_modules to match directory category")
	}
	if _, ok := matchFile("/tmp/demo/main.pyc"); !ok {
		t.Fatalf("expected .pyc to match file category")
	}

	modified := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	candidates := []Candidate{
		{CategoryKey: "node-modules", Category: "node_modules", Group: "/tmp/a", Bytes: 10, ModifiedAt: modified},
		{CategoryKey: "node-modules", Category: "node_modules", Group: "/tmp/a", Bytes: 20, ModifiedAt: modified.Add(1 * time.Hour)},
		{CategoryKey: "python-venv", Category: ".venv", Group: "/tmp/b", Bytes: 30, ModifiedAt: modified.Add(2 * time.Hour)},
	}

	cat := summarizeCategories(candidates)
	if len(cat) != 2 || cat[0].Bytes != 30 {
		t.Fatalf("unexpected category summaries: %#v", cat)
	}
	groups := summarizeGroups(candidates, 10)
	if len(groups) != 2 || groups[0].ModifiedAt.IsZero() {
		t.Fatalf("unexpected group summaries: %#v", groups)
	}
}

func TestPrefixAndSetHelpers(t *testing.T) {
	if !hasPathPrefix("/tmp/root/repo", "/tmp/root") {
		t.Fatalf("expected prefix match")
	}
	if hasPathPrefix("/tmp/other", "/tmp/root") {
		t.Fatalf("did not expect unrelated prefix match")
	}
	set := listToSet([]string{"a", "b"})
	if _, ok := set["a"]; !ok {
		t.Fatalf("expected set entry")
	}
	if includeCategory("node-modules", listToSet([]string{"node-modules"}), nil) != true {
		t.Fatalf("expected included category to pass")
	}
	if includeCategory("node-modules", nil, listToSet([]string{"node-modules"})) != false {
		t.Fatalf("expected excluded category to fail")
	}
}

func TestRenderPlainAndDetermineGroupHelpers(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	mustMkdir(t, filepath.Join(repo, ".git"))
	target := filepath.Join(repo, "node_modules")
	mustMkdir(t, target)

	if got := determineGroup(target, AnalyzeOptions{Root: root, GroupMode: "repo", GroupDepth: 1}, map[string]string{}); got != repo {
		t.Fatalf("expected repo group %q, got %q", repo, got)
	}
	if got := determineGroup(filepath.Join(root, "a", "b", "c"), AnalyzeOptions{Root: root, GroupMode: "depth", GroupDepth: 2}, map[string]string{}); got != filepath.Join(root, "a", "b") {
		t.Fatalf("unexpected depth group %q", got)
	}
	if got := ancestorGroup(root, root, 2); got != root {
		t.Fatalf("expected root ancestor group, got %q", got)
	}

	out, err := RenderReport(Report{
		Command:         "clean",
		Root:            root,
		FilesystemBytes: 10 << 20,
		TotalBytes:      5 << 20,
		AvailableBytes:  5 << 20,
		CandidateBytes:  3 << 20,
		SelectedBytes:   1 << 20,
		DeletedBytes:    1 << 20,
		Candidates: []Candidate{
			{CategoryKey: "node-modules", Path: target, Bytes: 3 << 20},
			{CategoryKey: "python-venv", Path: filepath.Join(repo, ".venv"), Bytes: 1 << 20},
		},
		SelectedCandidates: []Candidate{
			{CategoryKey: "python-venv", Path: filepath.Join(repo, ".venv"), Bytes: 1 << 20},
		},
		TopEntries: []PathSize{{Path: repo, Bytes: 5 << 20}},
		TopFiles:   []PathSize{{Path: filepath.Join(repo, "bundle.js"), Bytes: 2 << 20}},
		SelectedCategorySummaries: []CategorySummary{
			{CategoryKey: "python-venv", Description: "safe", Bytes: 1 << 20, Count: 1},
		},
		SelectedGroupSummaries: []GroupSummary{
			{Group: repo, Bytes: 1 << 20, Count: 1},
		},
	}, "plain")
	if err != nil {
		t.Fatalf("RenderReport returned error: %v", err)
	}
	for _, want := range []string{"deleted 1.0 MiB", "Cleanup by category (selected after exclusions)", "Largest files"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected plain report to contain %q, got %q", want, out)
		}
	}

	if _, err := RenderReport(Report{}, "json"); err == nil {
		t.Fatalf("expected unsupported format error")
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
