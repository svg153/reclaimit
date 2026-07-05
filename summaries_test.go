package reclaimit

import (
	"testing"
	"time"
)

// TestSummarizeCategoriesWithDuplicates validates category aggregation.
func TestSummarizeCategoriesWithDuplicates(t *testing.T) {
	modified := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	candidates := []Candidate{
		{CategoryKey: "node-modules", Category: "node_modules", Bytes: 100, ModifiedAt: modified},
		{CategoryKey: "node-modules", Category: "node_modules", Bytes: 200, ModifiedAt: modified.Add(1 * time.Hour)},
		{CategoryKey: "python-venv", Category: ".venv", Bytes: 300, ModifiedAt: modified.Add(2 * time.Hour)},
		{CategoryKey: "python-venv", Category: ".venv", Bytes: 100, ModifiedAt: modified},
	}

	summaries := summarizeCategories(candidates)
	if len(summaries) != 2 {
		t.Fatalf("expected 2 category summaries, got %d", len(summaries))
	}

	// First should be python-venv (300+100=400 bytes, more than 100+200=300)
	if summaries[0].CategoryKey != "python-venv" {
		t.Fatalf("expected first category 'python-venv', got %q", summaries[0].CategoryKey)
	}
	if summaries[0].Bytes != 400 {
		t.Fatalf("expected python-venv bytes 400, got %d", summaries[0].Bytes)
	}
	if summaries[0].Count != 2 {
		t.Fatalf("expected python-venv count 2, got %d", summaries[0].Count)
	}

	if summaries[1].CategoryKey != "node-modules" {
		t.Fatalf("expected second category 'node-modules', got %q", summaries[1].CategoryKey)
	}
	if summaries[1].Bytes != 300 {
		t.Fatalf("expected node-modules bytes 300, got %d", summaries[1].Bytes)
	}
}

// TestSummarizeCategoriesEmpty validates empty input.
func TestSummarizeCategoriesEmpty(t *testing.T) {
	summaries := summarizeCategories(nil)
	if len(summaries) != 0 {
		t.Fatalf("expected 0 summaries for nil, got %d", len(summaries))
	}

	summaries = summarizeCategories([]Candidate{})
	if len(summaries) != 0 {
		t.Fatalf("expected 0 summaries for empty, got %d", len(summaries))
	}
}

// TestSummarizeGroupsWithDuplicates validates group aggregation.
func TestSummarizeGroupsWithDuplicates(t *testing.T) {
	modified := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	candidates := []Candidate{
		{Group: "/tmp/a", CategoryKey: "node-modules", Bytes: 100, ModifiedAt: modified},
		{Group: "/tmp/a", CategoryKey: "python-venv", Bytes: 200, ModifiedAt: modified.Add(1 * time.Hour)},
		{Group: "/tmp/b", CategoryKey: "node-modules", Bytes: 500, ModifiedAt: modified},
	}

	summaries := summarizeGroups(candidates, 10)
	if len(summaries) != 2 {
		t.Fatalf("expected 2 group summaries, got %d", len(summaries))
	}

	if summaries[0].Group != "/tmp/b" || summaries[0].Bytes != 500 {
		t.Fatalf("expected first group /tmp/b with 500 bytes, got %s/%d", summaries[0].Group, summaries[0].Bytes)
	}
}

// TestSummarizeGroupsLimit validates group limit.
func TestSummarizeGroupsLimit(t *testing.T) {
	candidates := []Candidate{
		{Group: "/tmp/a", Bytes: 100},
		{Group: "/tmp/b", Bytes: 200},
		{Group: "/tmp/c", Bytes: 300},
		{Group: "/tmp/d", Bytes: 400},
	}

	summaries := summarizeGroups(candidates, 2)
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries (limited), got %d", len(summaries))
	}
	if summaries[0].Group != "/tmp/d" {
		t.Fatalf("expected /tmp/d first, got %s", summaries[0].Group)
	}
}

// TestSummarizeGroupsEmpty validates empty input.
func TestSummarizeGroupsEmpty(t *testing.T) {
	summaries := summarizeGroups(nil, 10)
	if len(summaries) != 0 {
		t.Fatalf("expected 0 summaries for nil, got %d", len(summaries))
	}
}

// TestSortByBytesAndPath validates sorting order (desc by bytes, asc by path).
func TestSortByBytesAndPath(t *testing.T) {
	items := []PathSize{
		{Path: "/tmp/b", Bytes: 100},
		{Path: "/tmp/a", Bytes: 100}, // Same bytes, different path
		{Path: "/tmp/c", Bytes: 200},
	}
	sortPathSizes(items)
	if items[0].Path != "/tmp/c" || items[0].Bytes != 200 {
		t.Fatalf("expected /tmp/c first (200 bytes), got %s/%d", items[0].Path, items[0].Bytes)
	}
	if items[1].Path != "/tmp/a" {
		t.Fatalf("expected /tmp/a second (same bytes, alphabetically first), got %s", items[1].Path)
	}
	if items[2].Path != "/tmp/b" {
		t.Fatalf("expected /tmp/b third, got %s", items[2].Path)
	}
}

// TestSortCandidates validates candidate sorting.
func TestSortCandidates(t *testing.T) {
	candidates := []Candidate{
		{Path: "/tmp/b", Bytes: 100},
		{Path: "/tmp/a", Bytes: 200},
	}
	sortCandidates(candidates)
	if candidates[0].Bytes != 200 {
		t.Fatalf("expected 200 bytes first, got %d", candidates[0].Bytes)
	}
}

// TestPushTop validates pushTop behavior.
func TestPushTop(t *testing.T) {
	list := []PathSize{}
	list = pushTop(list, PathSize{Path: "/tmp/a", Bytes: 100}, 3)
	if len(list) != 1 || list[0].Bytes != 100 {
		t.Fatalf("expected 1 item with 100 bytes, got %#v", list)
	}

	list = pushTop(list, PathSize{Path: "/tmp/b", Bytes: 200}, 3)
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}

	// Exceed limit
	list = pushTop(list, PathSize{Path: "/tmp/c", Bytes: 300}, 2)
	if len(list) != 2 {
		t.Fatalf("expected 2 items (limit), got %d", len(list))
	}
	if list[0].Bytes != 300 {
		t.Fatalf("expected 300 first, got %d", list[0].Bytes)
	}
}

// TestAggregateCandidates validates the generic aggregation skeleton.
func TestAggregateCandidates(t *testing.T) {
	candidates := []Candidate{
		{Group: "a", Bytes: 100},
		{Group: "b", Bytes: 200},
		{Group: "a", Bytes: 50},
	}

	result := aggregateCandidates(
		candidates,
		func(c Candidate) string { return c.Group },
		func(c Candidate) GroupSummary { return GroupSummary{Group: c.Group} },
		func(s *GroupSummary, c Candidate) {
			s.Bytes += c.Bytes
			s.Count++
		},
		func(a, b GroupSummary) bool {
			return a.Bytes > b.Bytes
		},
	)

	if len(result) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(result))
	}
	if result[0].Group != "b" || result[0].Bytes != 200 {
		t.Fatalf("expected b/200 first, got %s/%d", result[0].Group, result[0].Bytes)
	}
	if result[1].Group != "a" || result[1].Bytes != 150 {
		t.Fatalf("expected a/150 second, got %s/%d", result[1].Group, result[1].Bytes)
	}
}

// TestAggregateCandidatesEmpty validates empty input.
func TestAggregateCandidatesEmpty(t *testing.T) {
	result := aggregateCandidates(
		nil,
		func(c Candidate) string { return c.Group },
		func(c Candidate) GroupSummary { return GroupSummary{} },
		func(s *GroupSummary, c Candidate) {},
		func(a, b GroupSummary) bool { return false },
	)
	if len(result) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result))
	}
}

// TestCandidateSortingWithSameBytes validates path tiebreaker.
func TestCandidateSortingWithSameBytes(t *testing.T) {
	candidates := []Candidate{
		{Path: "/tmp/z", Bytes: 100},
		{Path: "/tmp/a", Bytes: 100},
		{Path: "/tmp/m", Bytes: 100},
	}
	sortCandidates(candidates)
	if candidates[0].Path != "/tmp/a" {
		t.Fatalf("expected /tmp/a first (alphabetical tiebreaker), got %s", candidates[0].Path)
	}
	if candidates[2].Path != "/tmp/z" {
		t.Fatalf("expected /tmp/z last, got %s", candidates[2].Path)
	}
}

// TestGroupSummaryModifiedAt validates latest modification tracking.
func TestGroupSummaryModifiedAt(t *testing.T) {
	mod1 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	mod2 := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	mod3 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

	candidates := []Candidate{
		{Group: "/tmp/a", Bytes: 100, ModifiedAt: mod1},
		{Group: "/tmp/a", Bytes: 200, ModifiedAt: mod3}, // older
		{Group: "/tmp/a", Bytes: 300, ModifiedAt: mod2}, // newest
	}

	summaries := summarizeGroups(candidates, 10)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 group, got %d", len(summaries))
	}
	if summaries[0].ModifiedAt != mod2 {
		t.Fatalf("expected latest modified %v, got %v", mod2, summaries[0].ModifiedAt)
	}
}
