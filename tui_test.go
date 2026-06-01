package reclaimit

import (
	"strings"
	"testing"
)

func TestApplySelectionFiltersGroupsAndPaths(t *testing.T) {
	report := Report{
		Candidates: []Candidate{
			{Group: "/root/repo-a", Path: "/root/repo-a/node_modules", Bytes: 100, CategoryKey: "node-modules"},
			{Group: "/root/repo-a", Path: "/root/repo-a/.venv", Bytes: 200, CategoryKey: "python-venv"},
			{Group: "/root/repo-b", Path: "/root/repo-b/node_modules", Bytes: 300, CategoryKey: "node-modules"},
		},
		GroupSummaries: []GroupSummary{{Group: "/root/repo-a"}, {Group: "/root/repo-b"}},
	}

	applySelection(&report, []string{"/root/repo-b"}, []string{"/root/repo-a/.venv"})

	if got := len(report.SelectedCandidates); got != 1 {
		t.Fatalf("expected 1 selected candidate, got %d", got)
	}
	if report.SelectedCandidates[0].Path != "/root/repo-a/node_modules" {
		t.Fatalf("unexpected selected path %q", report.SelectedCandidates[0].Path)
	}
	if report.SelectedBytes != 100 {
		t.Fatalf("expected selected bytes 100, got %d", report.SelectedBytes)
	}
}

func TestUsageTextIncludesBuiltBinaryExamples(t *testing.T) {
	text := usageText("")
	if !strings.Contains(text, "./bin/reclaimit tui") {
		t.Fatalf("expected main usage to mention built binary")
	}
}

func TestRenderMarkdownIncludesVisualBlocks(t *testing.T) {
	report := Report{
		Root:            "/tmp/demo",
		FilesystemBytes: 10 << 30,
		TotalBytes:      4 << 30,
		AvailableBytes:  2 << 30,
		CandidateBytes:  1 << 30,
		Candidates: []Candidate{
			{CategoryKey: "node-modules", Group: "/tmp/demo/repo", Path: "/tmp/demo/repo/node_modules", Bytes: 1 << 20},
		},
		SelectedCandidates: []Candidate{
			{CategoryKey: "node-modules", Group: "/tmp/demo/repo", Path: "/tmp/demo/repo/node_modules", Bytes: 1 << 20},
		},
		CategorySummaries: []CategorySummary{
			{CategoryKey: "node-modules", Description: "safe", Bytes: 1 << 20, Count: 1},
		},
		GroupSummaries: []GroupSummary{
			{Group: "/tmp/demo/repo", Bytes: 1 << 20, Count: 1},
		},
		SelectedCategorySummaries: []CategorySummary{
			{CategoryKey: "node-modules", Description: "safe", Bytes: 1 << 20, Count: 1},
		},
		SelectedGroupSummaries: []GroupSummary{
			{Group: "/tmp/demo/repo", Bytes: 1 << 20, Count: 1},
		},
	}

	out, err := RenderReport(report, "markdown")
	if err != nil {
		t.Fatalf("RenderReport returned error: %v", err)
	}

	for _, marker := range []string{"```mermaid", "```plantuml", "<details>", "Executive summary"} {
		if !strings.Contains(out, marker) {
			t.Fatalf("expected markdown output to contain %q", marker)
		}
	}
}
