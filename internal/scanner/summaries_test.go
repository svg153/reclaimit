package scanner

import (
	"testing"
	"time"
)

func TestApplySelection_NoExclusions(t *testing.T) {
	report := &Report{
		Candidates: []Candidate{
			{Path: "/a", Bytes: 100, CategoryKey: "node-modules"},
			{Path: "/b", Bytes: 200, CategoryKey: "python-venv"},
		},
	}
	ApplySelection(report, nil, nil)
	if len(report.SelectedCandidates) != 2 {
		t.Errorf("expected 2 selected, got %d", len(report.SelectedCandidates))
	}
	if report.SelectedBytes != 300 {
		t.Errorf("expected 300 bytes, got %d", report.SelectedBytes)
	}
}

func TestApplySelection_ExcludePath(t *testing.T) {
	report := &Report{
		Candidates: []Candidate{
			{Path: "/a", Bytes: 100, CategoryKey: "node-modules"},
			{Path: "/b", Bytes: 200, CategoryKey: "python-venv"},
		},
	}
	ApplySelection(report, nil, []string{"/a"})
	if len(report.SelectedCandidates) != 1 {
		t.Errorf("expected 1 selected, got %d", len(report.SelectedCandidates))
	}
	if report.SelectedCandidates[0].Path != "/b" {
		t.Errorf("expected /b, got %s", report.SelectedCandidates[0].Path)
	}
}

func TestApplySelection_ExcludeGroup(t *testing.T) {
	report := &Report{
		Candidates: []Candidate{
			{Path: "/a/node_modules", Group: "/a", Bytes: 100, CategoryKey: "node-modules"},
			{Path: "/b/node_modules", Group: "/b", Bytes: 200, CategoryKey: "node-modules"},
		},
	}
	ApplySelection(report, []string{"/a"}, nil)
	if len(report.SelectedCandidates) != 1 {
		t.Errorf("expected 1 selected, got %d", len(report.SelectedCandidates))
	}
}

func TestFilterCandidates_NoExclusions(t *testing.T) {
	candidates := []Candidate{
		{Path: "/a", Bytes: 100},
		{Path: "/b", Bytes: 200},
	}
	result := FilterCandidates(candidates, nil, nil)
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}

func TestFilterCandidates_EmptyInput(t *testing.T) {
	result := FilterCandidates(nil, []string{"/a"}, []string{"/b"})
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

func TestIsPathExcluded_Match(t *testing.T) {
	if !IsPathExcluded("/a/b/c", []string{"/a/b/c"}) {
		t.Error("expected true for exact match")
	}
}

func TestIsPathExcluded_NoMatch(t *testing.T) {
	if IsPathExcluded("/a/b/d", []string{"/a/b/c"}) {
		t.Error("expected false for different path")
	}
}

func TestIsPathExcluded_MultiplePaths(t *testing.T) {
	if !IsPathExcluded("/x", []string{"/a", "/x", "/z"}) {
		t.Error("expected true for middle path")
	}
	if IsPathExcluded("/y", []string{"/a", "/x", "/z"}) {
		t.Error("expected false for non-matching path")
	}
}

func TestSummarizeCategories_SingleCategory(t *testing.T) {
	candidates := []Candidate{
		{CategoryKey: "node-modules", Bytes: 100, Category: "node_modules"},
		{CategoryKey: "node-modules", Bytes: 200, Category: "node_modules"},
	}
	summaries := SummarizeCategories(candidates)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].Bytes != 300 {
		t.Errorf("expected 300 bytes, got %d", summaries[0].Bytes)
	}
	if summaries[0].Count != 2 {
		t.Errorf("expected count 2, got %d", summaries[0].Count)
	}
}

func TestSummarizeCategories_MultipleCategories(t *testing.T) {
	candidates := []Candidate{
		{CategoryKey: "node-modules", Bytes: 100},
		{CategoryKey: "python-venv", Bytes: 200},
	}
	summaries := SummarizeCategories(candidates)
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	// Should be sorted by bytes descending
	if summaries[0].CategoryKey != "python-venv" {
		t.Errorf("expected python-venv first, got %s", summaries[0].CategoryKey)
	}
}

func TestSummarizeCategories_SortByCategoryOnTie(t *testing.T) {
	candidates := []Candidate{
		{CategoryKey: "zzz", Category: "ZZZ", Bytes: 100},
		{CategoryKey: "aaa", Category: "AAA", Bytes: 100},
	}
	summaries := SummarizeCategories(candidates)
	// Both have 100 bytes, sort by display name alphabetically ascending
	if summaries[0].Category != "AAA" {
		t.Errorf("expected AAA first on tie, got %s", summaries[0].Category)
	}
}

func TestSummarizeGroups(t *testing.T) {
	candidates := []Candidate{
		{Group: "/repo-a", Bytes: 100, ModifiedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Group: "/repo-a", Bytes: 200, ModifiedAt: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
		{Group: "/repo-b", Bytes: 300, ModifiedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	summaries := SummarizeGroups(candidates, 10)
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	// /repo-a has 300 bytes (100+200), /repo-b has 300 bytes
	// Tie on bytes, so sorted by group name ascending -> /repo-a first
	if summaries[0].Group != "/repo-a" {
		t.Errorf("expected /repo-a first (tie-break by name), got %s", summaries[0].Group)
	}
}

func TestSummarizeGroups_Limit(t *testing.T) {
	candidates := []Candidate{
		{Group: "/a", Bytes: 100},
		{Group: "/b", Bytes: 200},
		{Group: "/c", Bytes: 300},
	}
	summaries := SummarizeGroups(candidates, 2)
	if len(summaries) != 2 {
		t.Errorf("expected 2 summaries (limited), got %d", len(summaries))
	}
}

func TestIsGroupExcluded_PrefixMatch(t *testing.T) {
	cand := Candidate{Group: "/repo/a", Path: "/repo/a/file"}
	if !IsGroupExcluded(cand, []string{"/repo"}) {
		t.Error("expected true for group prefix match")
	}
}

func TestIsGroupExcluded_NoMatch(t *testing.T) {
	cand := Candidate{Group: "/repo/a", Path: "/repo/a/file"}
	if IsGroupExcluded(cand, []string{"/other"}) {
		t.Error("expected false for non-matching group")
	}
}

func TestHasPathPrefix_Equal(t *testing.T) {
	if !HasPathPrefix("/a/b", "/a/b") {
		t.Error("expected true for equal paths")
	}
}

func TestHasPathPrefix_SubPath(t *testing.T) {
	if !HasPathPrefix("/a/b/c", "/a/b") {
		t.Error("expected true for sub-path")
	}
}

func TestHasPathPrefix_NoPrefix(t *testing.T) {
	if HasPathPrefix("/a/c", "/a/b") {
		t.Error("expected false for non-prefix")
	}
}

func TestHasPathPrefix_ParentOfPrefix(t *testing.T) {
	if HasPathPrefix("/a", "/a/b") {
		t.Error("expected false for parent of prefix")
	}
}

func TestSumBytes(t *testing.T) {
	items := []PathSize{{Bytes: 100}, {Bytes: 200}, {Bytes: 300}}
	if got := SumBytes(items); got != 600 {
		t.Errorf("expected 600, got %d", got)
	}
}

func TestSumBytes_Empty(t *testing.T) {
	if got := SumBytes(nil); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestSumCandidateBytes(t *testing.T) {
	items := []Candidate{{Bytes: 100}, {Bytes: 200}}
	if got := SumCandidateBytes(items); got != 300 {
		t.Errorf("expected 300, got %d", got)
	}
}

func TestListToSet(t *testing.T) {
	s := ListToSet([]string{"a", "b", "c"})
	if len(s) != 3 {
		t.Errorf("expected set size 3, got %d", len(s))
	}
	if _, ok := s["b"]; !ok {
		t.Error("expected b in set")
	}
}

func TestListToSet_Empty(t *testing.T) {
	s := ListToSet(nil)
	if len(s) != 0 {
		t.Errorf("expected empty set, got size %d", len(s))
	}
}

func TestPushTop_BelowLimit(t *testing.T) {
	items := []PathSize{{Bytes: 100}}
	result := PushTop(items, PathSize{Bytes: 200}, 5)
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}

func TestPushTop_AboveLimit(t *testing.T) {
	items := []PathSize{{Bytes: 100}, {Bytes: 200}, {Bytes: 300}}
	result := PushTop(items, PathSize{Bytes: 400}, 2)
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
	if result[0].Bytes != 400 {
		t.Errorf("expected 400 first, got %d", result[0].Bytes)
	}
}

func TestSortPathSizes(t *testing.T) {
	items := []PathSize{{Bytes: 300}, {Bytes: 100}, {Bytes: 200}}
	SortPathSizes(items)
	if items[0].Bytes != 300 {
		t.Errorf("expected 300 first, got %d", items[0].Bytes)
	}
}

func TestSortCandidates(t *testing.T) {
	items := []Candidate{{Bytes: 300}, {Bytes: 100}, {Bytes: 200}}
	SortCandidates(items)
	if items[0].Bytes != 300 {
		t.Errorf("expected 300 first, got %d", items[0].Bytes)
	}
}
