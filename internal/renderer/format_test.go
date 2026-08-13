package renderer

import (
	"strings"
	"testing"

	"github.com/svg153/reclaimit/internal/scanner"
)

func TestRenderPlain(t *testing.T) {
	report := scanner.Report{
		Root:   "/tmp/test",
		TotalBytes: 1048676,
		FreeBytes:  738858448896,
		Candidates: []scanner.Candidate{
			{Path: "/tmp/test/node_modules", Bytes: 1048576, IsDir: true, Category: "node-modules"},
			{Path: "/tmp/test/.DS_Store", Bytes: 100, IsDir: false, Category: "ds-store"},
		},
		CandidateBytes: 1048676,
		CategorySummaries: []scanner.CategorySummary{
			{Category: "node-modules", Bytes: 1048576, Count: 1, Description: "JS dependencies"},
			{Category: "ds-store", Bytes: 100, Count: 1, Description: "macOS storage"},
		},
		GroupSummaries: []scanner.GroupSummary{
			{Group: "repo1", Bytes: 1048576, Count: 1},
		},
		TopFiles: []scanner.PathSize{
			{Path: "/tmp/test/node_modules", Bytes: 1048576},
		},
		TopEntries: []scanner.PathSize{
			{Path: "/tmp/test/node_modules", Bytes: 1048576},
			{Path: "/tmp/test/.DS_Store", Bytes: 100},
		},
	}

	output, err := RenderReport(report, "plain")
	if err != nil {
		t.Fatalf("RenderReport: %v", err)
	}

	if !strings.Contains(output, "Disk usage report") {
		t.Error("expected 'Disk usage report' header")
	}
	if !strings.Contains(output, "node_modules") {
		t.Error("expected node_modules in output")
	}
	if !strings.Contains(output, "1.0 MiB") {
		t.Error("expected 1.0 MiB in output")
	}
}

func TestRenderPlain_NoCandidates(t *testing.T) {
	report := scanner.Report{Root: "/tmp/empty", CandidateBytes: 0, CategorySummaries: []scanner.CategorySummary{}, GroupSummaries: []scanner.GroupSummary{}}
	output, err := RenderReport(report, "plain")
	if err != nil {
		t.Fatalf("RenderReport: %v", err)
	}

	if !strings.Contains(output, "0 candidates") {
		t.Error("expected 0 candidates message")
	}
}

func TestRenderMarkdown(t *testing.T) {
	report := scanner.Report{
		Root:       "/tmp/test",
		TotalBytes: 1048676,
		Candidates: []scanner.Candidate{
			{Path: "/tmp/test/node_modules", Bytes: 1048576, IsDir: true, Category: "node-modules"},
		},
		CandidateBytes: 1048576,
		GroupSummaries: []scanner.GroupSummary{
			{Group: "repo1", Bytes: 1048576, Count: 1},
		},
		TopFiles: []scanner.PathSize{
			{Path: "/tmp/test/node_modules", Bytes: 1048576},
		},
	}
	output, err := RenderReport(report, "markdown")
	if err != nil {
		t.Fatalf("RenderReport: %v", err)
	}

	if !strings.Contains(output, "# Disk usage report") {
		t.Error("expected markdown header")
	}
	if !strings.Contains(output, "node_modules") {
		t.Error("expected node_modules in output")
	}
	if !strings.Contains(output, "|") {
		t.Error("expected table format in markdown")
	}
}

func TestRenderJSON(t *testing.T) {
	report := scanner.Report{
		Root:       "/tmp/test",
		TotalBytes: 1048576,
		Candidates: []scanner.Candidate{
			{Path: "/tmp/test/node_modules", Bytes: 1048576, IsDir: true, Category: "node-modules"},
		},
		CandidateBytes: 1048576,
	}
	output, err := RenderReport(report, "json")
	if err != nil {
		t.Fatalf("RenderReport: %v", err)
	}

	if !strings.Contains(output, `"root"`) {
		t.Error("expected root field in JSON")
	}
	if !strings.Contains(output, "node_modules") {
		t.Error("expected node_modules in JSON")
	}
}

func TestRenderJSON_Empty(t *testing.T) {
	report := scanner.Report{Root: "/tmp/empty", Candidates: []scanner.Candidate{}, CandidateBytes: 0}
	output, err := RenderReport(report, "json")
	if err != nil {
		t.Fatalf("RenderReport: %v", err)
	}

	if !strings.Contains(output, `"root"`) {
		t.Error("expected root field in empty JSON")
	}
	if !strings.Contains(output, `"candidates"`) {
		t.Error("expected candidates field in JSON")
	}
}

func TestEscapePlant(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello\nworld", "hello\nworld"},
		{"hello\tworld", "hello\tworld"},
		{"", ""},
	}
	for _, tt := range tests {
		result := escapePlant(tt.input)
		if result != tt.expected {
			t.Errorf("escapePlant(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestLimitCandidates(t *testing.T) {
	candidates := []scanner.Candidate{
		{Path: "/tmp/a", Bytes: 100},
		{Path: "/tmp/b", Bytes: 200},
		{Path: "/tmp/c", Bytes: 300},
	}
	limited := limitCandidates(candidates, 2)
	if len(limited) != 2 {
		t.Errorf("expected 2 candidates, got %d", len(limited))
	}
	if limited[0].Path != "/tmp/a" {
		t.Errorf("first = %q, want %q", limited[0].Path, "/tmp/a")
	}
}

func TestLimitCandidates_FewerThanLimit(t *testing.T) {
	candidates := []scanner.Candidate{
		{Path: "/tmp/a", Bytes: 100},
	}
	limited := limitCandidates(candidates, 5)
	if len(limited) != 1 {
		t.Errorf("expected 1 candidate, got %d", len(limited))
	}
}

func TestLimitCandidates_Empty(t *testing.T) {
	limited := limitCandidates([]scanner.Candidate{}, 5)
	if len(limited) != 0 {
		t.Errorf("expected 0 candidates, got %d", len(limited))
	}
}

func TestHumanizeBytes_Renderer(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
		{1099511627776, "1.0 TiB"},
	}
	for _, tt := range tests {
		result := humanizeBytes(tt.input)
		if result != tt.expected {
			t.Errorf("humanizeBytes(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestRenderMarkdownSummary(t *testing.T) {
	report := scanner.Report{CandidateBytes: 1048576}
	result := renderMarkdownSummary(report, false)
	if !strings.Contains(result, "Executive summary") {
		t.Error("expected Executive summary header")
	}
	if !strings.Contains(result, "1.0 MiB") {
		t.Error("expected 1.0 MiB in output")
	}
	if !strings.Contains(result, "|") {
		t.Error("expected table format")
	}
}
