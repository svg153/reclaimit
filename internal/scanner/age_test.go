package scanner

import (
	"testing"
	"time"
)

func TestFilterCandidatesByAge(t *testing.T) {
	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	report := Report{Candidates: []Candidate{
		{Path: "/stale", ModifiedAt: now.Add(-31 * 24 * time.Hour)},
		{Path: "/boundary", ModifiedAt: now.Add(-30 * 24 * time.Hour)},
		{Path: "/active", ModifiedAt: now.Add(-29 * 24 * time.Hour)},
		{Path: "/unknown", ModifiedAt: time.Time{}},
	}}
	FilterCandidatesByAge(&report, 30*24*time.Hour, now)
	if len(report.Candidates) != 1 || report.Candidates[0].Path != "/stale" {
		t.Fatalf("unexpected candidates after age filter: %+v", report.Candidates)
	}
}

func TestFilterCandidatesByAgeNoOp(t *testing.T) {
	report := Report{Candidates: []Candidate{{Path: "/candidate"}}}
	FilterCandidatesByAge(&report, 0, time.Now())
	if len(report.Candidates) != 1 {
		t.Fatalf("zero age should not filter candidates: %+v", report.Candidates)
	}
}
