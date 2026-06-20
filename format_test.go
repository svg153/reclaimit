package reclaimit

import (
	"strings"
	"testing"
	"time"
)

// TestApplySelection validates applySelection filters candidates correctly.
func TestApplySelection(t *testing.T) {
	candidates := []Candidate{
		{Path: "/tmp/a/node_modules", CategoryKey: "node-modules", Bytes: 100, Group: "/tmp/a"},
		{Path: "/tmp/b/node_modules", CategoryKey: "node-modules", Bytes: 200, Group: "/tmp/b"},
		{Path: "/tmp/a/.venv", CategoryKey: "python-venv", Bytes: 300, Group: "/tmp/a"},
	}

	report := Report{
		Candidates:       candidates,
		CandidateBytes:   600,
		SelectedCandidates: candidates,
		SelectedBytes:    600,
	}

	applySelection(&report, []string{"/tmp/a"}, nil)

	if len(report.SelectedCandidates) != 1 {
		t.Fatalf("expected 1 selected candidate, got %d", len(report.SelectedCandidates))
	}
	if report.SelectedCandidates[0].Path != "/tmp/b/node_modules" {
		t.Fatalf("expected /tmp/b/node_modules, got %s", report.SelectedCandidates[0].Path)
	}
	if report.SelectedBytes != 200 {
		t.Fatalf("expected selected bytes 200, got %d", report.SelectedBytes)
	}
}

// TestApplySelectionNoExclusions validates no-op when no exclusions.
func TestApplySelectionNoExclusions(t *testing.T) {
	candidates := []Candidate{
		{Path: "/tmp/a/node_modules", Bytes: 100},
		{Path: "/tmp/a/.venv", Bytes: 200},
	}

	report := Report{
		Candidates:       candidates,
		CandidateBytes:   300,
		SelectedCandidates: candidates,
		SelectedBytes:    300,
	}

	applySelection(&report, nil, nil)

	if len(report.SelectedCandidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(report.SelectedCandidates))
	}
	if report.SelectedBytes != 300 {
		t.Fatalf("expected 300 bytes, got %d", report.SelectedBytes)
	}
}

// TestFilterCandidatesPathExclusion validates path-based exclusion.
func TestFilterCandidatesPathExclusion(t *testing.T) {
	candidates := []Candidate{
		{Path: "/tmp/a/node_modules", Bytes: 100},
		{Path: "/tmp/b/node_modules", Bytes: 200},
	}

	filtered := filterCandidates(candidates, nil, []string{"/tmp/a/node_modules"})

	if len(filtered) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(filtered))
	}
	if filtered[0].Path != "/tmp/b/node_modules" {
		t.Fatalf("expected /tmp/b/node_modules, got %s", filtered[0].Path)
	}
}

// TestFilterCandidatesGroupExclusion validates group-based exclusion.
func TestFilterCandidatesGroupExclusion(t *testing.T) {
	candidates := []Candidate{
		{Path: "/tmp/a/node_modules", Group: "/tmp/a", Bytes: 100},
		{Path: "/tmp/b/node_modules", Group: "/tmp/b", Bytes: 200},
	}

	filtered := filterCandidates(candidates, []string{"/tmp/a"}, nil)

	if len(filtered) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(filtered))
	}
	if filtered[0].Path != "/tmp/b/node_modules" {
		t.Fatalf("expected /tmp/b/node_modules, got %s", filtered[0].Path)
	}
}

// TestIsPathExcluded validates exact path matching.
func TestIsPathExcluded(t *testing.T) {
	if !isPathExcluded("/tmp/a/node_modules", []string{"/tmp/a/node_modules"}) {
		t.Fatalf("expected exact match to be excluded")
	}
	if isPathExcluded("/tmp/a/node_modules", []string{"/tmp/b/node_modules"}) {
		t.Fatalf("expected different path not to be excluded")
	}
	if isPathExcluded("/tmp/a/node_modules", nil) {
		t.Fatalf("expected nil exclusions to not exclude")
	}
}

// TestRenderPlainWithClean validates plain report shows deleted bytes.
func TestRenderPlainWithClean(t *testing.T) {
	report := Report{
		Root:       "/tmp",
		TotalBytes: 10 << 20,
		FilesystemBytes: 20 << 20,
		AvailableBytes: 10 << 20,
		Command:    "clean",
		DeletedBytes: 1 << 20,
		SelectedCandidates: []Candidate{
			{CategoryKey: "node-modules", Path: "/tmp/node_modules", Bytes: 1 << 20},
		},
		SelectedCategorySummaries: []CategorySummary{
			{CategoryKey: "node-modules", Bytes: 1 << 20, Count: 1, Description: "safe"},
		},
		SelectedGroupSummaries: []GroupSummary{
			{Group: "/tmp", Bytes: 1 << 20, Count: 1},
		},
	}

	out := renderPlain(report)
	if !strings.Contains(out, "deleted 1.0 MiB") {
		t.Fatalf("expected plain report to contain 'deleted 1.0 MiB', got: %s", out)
	}
}

// TestRenderPlainWithSelection validates selection mode labels.
func TestRenderPlainWithSelection(t *testing.T) {
	report := Report{
		Root:       "/tmp",
		TotalBytes: 10 << 20,
		FilesystemBytes: 20 << 20,
		AvailableBytes: 10 << 20,
		Candidates: []Candidate{
			{CategoryKey: "node-modules", Path: "/tmp/a/node_modules", Bytes: 1000},
			{CategoryKey: "python-venv", Path: "/tmp/b/.venv", Bytes: 2000},
		},
		SelectedCandidates: []Candidate{
			{CategoryKey: "python-venv", Path: "/tmp/b/.venv", Bytes: 2000},
		},
		SelectedBytes: 2000,
		CandidateBytes: 3000,
		SelectedCategorySummaries: []CategorySummary{
			{CategoryKey: "python-venv", Bytes: 2000, Count: 1, Description: "safe"},
		},
		SelectedGroupSummaries: []GroupSummary{
			{Group: "/tmp/b", Bytes: 2000, Count: 1},
		},
	}

	out := renderPlain(report)
	if !strings.Contains(out, "selected after exclusions") {
		t.Fatalf("expected 'selected after exclusions' in plain report, got: %s", out)
	}
}

// TestRenderMarkdownStructure validates markdown report has expected sections.
func TestRenderMarkdownStructure(t *testing.T) {
	report := Report{
		Root:       "/tmp/demo",
		TotalBytes: 10 << 20,
		FilesystemBytes: 20 << 20,
		AvailableBytes: 10 << 20,
		Candidates: []Candidate{
			{CategoryKey: "node-modules", Path: "/tmp/demo/node_modules", Bytes: 1024},
		},
		SelectedCandidates: []Candidate{
			{CategoryKey: "node-modules", Path: "/tmp/demo/node_modules", Bytes: 1024},
		},
		CandidateBytes: 1024,
		SelectedBytes:  1024,
		TopEntries:     []PathSize{{Path: "/tmp/demo", Bytes: 10 << 20}},
		TopFiles:       []PathSize{{Path: "/tmp/demo/bundle.js", Bytes: 5 << 20}},
		CategorySummaries: []CategorySummary{
			{CategoryKey: "node-modules", Bytes: 1024, Count: 1, Description: "safe"},
		},
		GroupSummaries: []GroupSummary{
			{Group: "/tmp/demo", Bytes: 1024, Count: 1},
		},
	}

	out := renderMarkdown(report)
	checks := []string{
		"# Disk usage report",
		"Filesystem usage:",
		"Cleanup by category",
		"Cleanup by group",
		"Top cleanup candidates",
		"Generated by reclaimit",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Fatalf("expected markdown to contain %q, got: %s", want, out)
		}
	}
}

// TestRenderMarkdownWithDeleted validates markdown shows deleted bytes.
func TestRenderMarkdownWithDeleted(t *testing.T) {
	report := Report{
		Root:       "/tmp",
		TotalBytes: 10 << 20,
		FilesystemBytes: 20 << 20,
		AvailableBytes: 10 << 20,
		Command:    "clean",
		DeletedBytes: 1 << 20,
		Candidates: []Candidate{},
	}

	out := renderMarkdown(report)
	if !strings.Contains(out, "- **Deleted:** 1.0 MiB") {
		t.Fatalf("expected markdown to contain deleted bytes, got: %s", out)
	}
}

// TestCandidateRows validates candidate row formatting.
func TestCandidateRows(t *testing.T) {
	modified := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	candidates := []Candidate{
		{CategoryKey: "node-modules", Path: "/tmp/a/node_modules", Bytes: 1024, ModifiedAt: modified, IsDir: true},
	}
	rows := candidateRows(candidates, 10)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !strings.Contains(rows[0], "1.0 KiB") {
		t.Fatalf("expected row to contain '1.0 KiB', got: %s", rows[0])
	}
	if !strings.Contains(rows[0], "dir") {
		t.Fatalf("expected row to contain 'dir', got: %s", rows[0])
	}
}

// TestGroupRows validates group row formatting.
func TestGroupRows(t *testing.T) {
	modified := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	groups := []GroupSummary{
		{Group: "/tmp/project", Bytes: 1024, Count: 3, ModifiedAt: modified},
	}
	rows := groupRows(groups)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !strings.Contains(rows[0], "1.0 KiB") {
		t.Fatalf("expected row to contain '1.0 KiB', got: %s", rows[0])
	}
}

// TestPathSizeRows validates path size row formatting.
func TestPathSizeRows(t *testing.T) {
	items := []PathSize{
		{Path: "/tmp/cache", Bytes: 2048},
	}
	rows := pathSizeRows(items)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !strings.Contains(rows[0], "2.0 KiB") {
		t.Fatalf("expected row to contain '2.0 KiB', got: %s", rows[0])
	}
}

// TestCategoryRows validates category row formatting.
func TestCategoryRows(t *testing.T) {
	summaries := []CategorySummary{
		{CategoryKey: "node-modules", Bytes: 4096, Count: 2, Description: "safe"},
	}
	rows := categoryRows(summaries)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !strings.Contains(rows[0], "4.0 KiB") {
		t.Fatalf("expected row to contain '4.0 KiB', got: %s", rows[0])
	}
}

// TestRenderMarkdownSummary validates the executive summary section.
func TestRenderMarkdownSummary(t *testing.T) {
	report := Report{
		Root:            "/tmp",
		FilesystemBytes: 100 << 20,
		TotalBytes:      50 << 20,
		AvailableBytes:  50 << 20,
		CandidateBytes:  10 << 20,
		Candidates:      []Candidate{{Bytes: 10 << 20}},
		SelectedCandidates: []Candidate{{Bytes: 5 << 20}},
		SelectedBytes:   5 << 20,
	}

	out := renderMarkdownSummary(report, true)
	if !strings.Contains(out, "Filesystem total") {
		t.Fatalf("expected 'Filesystem total' in summary, got: %s", out)
	}
	if !strings.Contains(out, "Cleanup candidates") {
		t.Fatalf("expected 'Cleanup candidates' in summary, got: %s", out)
	}
	if !strings.Contains(out, "Selected reclaimable") {
		t.Fatalf("expected 'Selected reclaimable' in summary, got: %s", out)
	}
}

// TestCandidateKind validates kind detection.
func TestCandidateKind(t *testing.T) {
	dir := Candidate{IsDir: true}
	file := Candidate{IsDir: false}

	if candidateKind(dir) != "dir" {
		t.Fatalf("expected 'dir', got %s", candidateKind(dir))
	}
	if candidateKind(file) != "file" {
		t.Fatalf("expected 'file', got %s", candidateKind(file))
	}
}

// TestHumanizeTimestamp validates timestamp formatting.
func TestHumanizeTimestamp(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if got := humanizeTimestamp(t1); got != "2026-01-01 12:00" {
		t.Fatalf("expected '2026-01-01 12:00', got %s", got)
	}

	if got := humanizeTimestamp(time.Time{}); got != "-" {
		t.Fatalf("expected '-', got %s", got)
	}
}

// TestLimitCandidates validates limiting behavior.
func TestLimitCandidates(t *testing.T) {
	items := []Candidate{
		{Bytes: 1}, {Bytes: 2}, {Bytes: 3}, {Bytes: 4}, {Bytes: 5},
	}
	limited := limitCandidates(items, 3)
	if len(limited) != 3 {
		t.Fatalf("expected 3 items, got %d", len(limited))
	}
	if limited[2].Bytes != 3 {
		t.Fatalf("expected first 3 items, got %d", limited[2].Bytes)
	}

	// No truncation needed
	limited = limitCandidates(items, 10)
	if len(limited) != 5 {
		t.Fatalf("expected 5 items (no truncation), got %d", len(limited))
	}
}

// TestEscapeMarkdownCell validates pipe escaping.
func TestEscapeMarkdownCell(t *testing.T) {
	if got := escapeMarkdownCell("a|b"); got != "a\\|b" {
		t.Fatalf("expected 'a\\\\|b', got %s", got)
	}
	if got := escapeMarkdownCell("no pipes"); got != "no pipes" {
		t.Fatalf("expected 'no pipes', got %s", got)
	}
}

// TestRenderMarkdownDetails validates details section.
func TestRenderMarkdownDetails(t *testing.T) {
	out := renderMarkdownDetails("Title", "Header\n", []string{"row1", "row2"})
	if !strings.Contains(out, "<summary>Title</summary>") {
		t.Fatalf("expected summary tag, got: %s", out)
	}
	if !strings.Contains(out, "row1") {
		t.Fatalf("expected row1, got: %s", out)
	}
	if !strings.Contains(out, "row2") {
		t.Fatalf("expected row2, got: %s", out)
	}
}

// TestRenderMarkdownCategoryPie validates mermaid pie chart.
func TestRenderMarkdownCategoryPie(t *testing.T) {
	summaries := []CategorySummary{
		{CategoryKey: "node-modules", Bytes: 1024},
		{CategoryKey: "python-venv", Bytes: 2048},
	}
	out := renderMarkdownCategoryPie(summaries)
	if !strings.Contains(out, "mermaid") {
		t.Fatalf("expected mermaid, got: %s", out)
	}
	if !strings.Contains(out, "node-modules") {
		t.Fatalf("expected node-modules, got: %s", out)
	}
}

// TestRenderMarkdownTopGroupsChart validates mermaid chart.
func TestRenderMarkdownTopGroupsChart(t *testing.T) {
	groups := []GroupSummary{
		{Group: "/tmp/a", Bytes: 1024},
		{Group: "/tmp/b", Bytes: 2048},
	}
	out := renderMarkdownTopGroupsChart(groups)
	if !strings.Contains(out, "xychart-beta") {
		t.Fatalf("expected xychart-beta, got: %s", out)
	}
}

// TestRenderPlantUMLOverview validates plantuml output.
func TestRenderPlantUMLOverview(t *testing.T) {
	groups := []GroupSummary{
		{Group: "/tmp/a", Bytes: 1024},
	}
	out := renderPlantUMLOverview(groups)
	if !strings.Contains(out, "@startmindmap") {
		t.Fatalf("expected @startmindmap, got: %s", out)
	}
}

// TestEscapePlant validates plantuml escaping.
func TestEscapePlant(t *testing.T) {
	if got := escapePlant(`test"quote`); got != `test\"quote` {
		t.Fatalf("expected 'test\\\"quote', got %s", got)
	}
}

// TestBytesToMiB validates byte conversion.
func TestBytesToMiB(t *testing.T) {
	if got := bytesToMiB(1024 * 1024); got != 1.0 {
		t.Fatalf("expected 1.0 MiB, got %f", got)
	}
}

// TestBytesToGiB validates byte conversion.
func TestBytesToGiB(t *testing.T) {
	if got := bytesToGiB(1024 * 1024 * 1024); got != 1.0 {
		t.Fatalf("expected 1.0 GiB, got %f", got)
	}
}

// TestRenderPlainNoSelection validates plain report when no exclusions applied.
func TestRenderPlainNoSelection(t *testing.T) {
	report := Report{
		Root:       "/tmp",
		TotalBytes: 10 << 20,
		FilesystemBytes: 20 << 20,
		AvailableBytes: 10 << 20,
		Candidates: []Candidate{
			{CategoryKey: "node-modules", Path: "/tmp/node_modules", Bytes: 1000},
		},
		SelectedCandidates: []Candidate{
			{CategoryKey: "node-modules", Path: "/tmp/node_modules", Bytes: 1000},
		},
		CandidateBytes: 1000,
		SelectedBytes:  1000,
		CategorySummaries: []CategorySummary{
			{CategoryKey: "node-modules", Bytes: 1000, Count: 1, Description: "safe"},
		},
		GroupSummaries: []GroupSummary{
			{Group: "/tmp", Bytes: 1000, Count: 1},
		},
	}

	out := renderPlain(report)
	// When no selection, should NOT say "selected after exclusions"
	if strings.Contains(out, "selected after exclusions") {
		t.Fatalf("expected no 'selected after exclusions' when no exclusions applied, got: %s", out)
	}
}
