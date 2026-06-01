package scanner

func ApplySelection(report *Report, excludedGroups, excludedPaths []string) {
	report.SelectedCandidates = FilterCandidates(report.Candidates, excludedGroups, excludedPaths)
	report.SelectedBytes = SumCandidateBytes(report.SelectedCandidates)
	report.SelectedCategorySummaries = SummarizeCategories(report.SelectedCandidates)
	report.SelectedGroupSummaries = SummarizeGroups(report.SelectedCandidates, len(report.GroupSummaries))
}

func FilterCandidates(candidates []Candidate, excludedGroups, excludedPaths []string) []Candidate {
	if len(excludedGroups) == 0 && len(excludedPaths) == 0 {
		return append([]Candidate(nil), candidates...)
	}

	selected := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		if IsGroupExcluded(candidate, excludedGroups) || IsPathExcluded(candidate.Path, excludedPaths) {
			continue
		}
		selected = append(selected, candidate)
	}
	SortCandidates(selected)
	return selected
}

func IsPathExcluded(path string, excludedPaths []string) bool {
	for _, excluded := range excludedPaths {
		if path == excluded {
			return true
		}
	}
	return false
}
