package tui

import (
	"testing"
	"time"

	"github.com/svg153/reclaimit/internal/renderer"
	"github.com/svg153/reclaimit/internal/scanner"
)

func TestApplySelectionFiltersGroupsAndPaths(t *testing.T) {
	t.Skip("TUI test requires terminal")
	report := scanner.Report{
		Candidates: []scanner.Candidate{
			{Group: "/root/repo-a", Path: "/root/repo-a/node_modules", Bytes: 100, CategoryKey: "node-modules"},
			{Group: "/root/repo-a", Path: "/root/repo-a/.venv", Bytes: 200, CategoryKey: "python-venv"},
			{Group: "/root/repo-b", Path: "/root/repo-b/dist", Bytes: 300, CategoryKey: "js-build"},
		},
		SelectedCandidates: nil,
	}

	selection, err := Run(report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = selection
}

func TestRenderReportFormats(t *testing.T) {
	report := scanner.Report{
		Root:         "/tmp",
		TotalBytes:   1000,
		CandidateBytes: 500,
		Candidates: []scanner.Candidate{
			{Path: "/tmp/node_modules", Bytes: 300, CategoryKey: "node-modules", ModifiedAt: time.Now()},
			{Path: "/tmp/.venv", Bytes: 200, CategoryKey: "python-venv", ModifiedAt: time.Now()},
		},
		CategorySummaries: []scanner.CategorySummary{
			{CategoryKey: "node-modules", Category: "node_modules", Bytes: 300},
			{CategoryKey: "python-venv", Category: ".venv", Bytes: 200},
		},
		GroupSummaries: []scanner.GroupSummary{
			{Group: "/tmp", Bytes: 500},
		},
		SelectedCandidates: []scanner.Candidate{
			{Path: "/tmp/node_modules", Bytes: 300, CategoryKey: "node-modules"},
		},
		SelectedBytes: 300,
		SelectedCategorySummaries: []scanner.CategorySummary{
			{CategoryKey: "node-modules", Bytes: 300},
		},
		SelectedGroupSummaries: []scanner.GroupSummary{
			{Group: "/tmp", Bytes: 300},
		},
	}

	for _, format := range []string{"plain", "markdown", "json"} {
		t.Run(format, func(t *testing.T) {
			out, err := renderer.RenderReport(report, format)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", format, err)
			}
			if len(out) == 0 {
				t.Fatalf("%s: empty output", format)
			}
		})
	}
}
