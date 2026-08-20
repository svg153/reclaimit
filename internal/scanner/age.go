package scanner

import "time"

// FilterCandidatesByAge keeps only candidates strictly older than age.
// Candidates without a modification timestamp are excluded because their age
// cannot be verified safely. The report summaries are rebuilt by the caller.
func FilterCandidatesByAge(report *Report, age time.Duration, now time.Time) {
	if report == nil || age <= 0 {
		return
	}
	cutoff := now.UTC().Add(-age)
	filtered := make([]Candidate, 0, len(report.Candidates))
	for _, candidate := range report.Candidates {
		if candidate.ModifiedAt.IsZero() {
			continue
		}
		if candidate.ModifiedAt.UTC().Before(cutoff) {
			filtered = append(filtered, candidate)
		}
	}
	report.Candidates = filtered
	report.SelectedCandidates = nil
	report.SelectedBytes = 0
	report.SelectedCategorySummaries = nil
	report.SelectedGroupSummaries = nil
}
