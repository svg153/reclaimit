package main

func applySelection(report *Report, excludedGroups, excludedPaths []string) {
	report.SelectedCandidates = filterCandidates(report.Candidates, excludedGroups, excludedPaths)
	report.SelectedBytes = sumCandidateBytes(report.SelectedCandidates)
	report.SelectedCategorySummaries = summarizeCategories(report.SelectedCandidates)
	report.SelectedGroupSummaries = summarizeGroups(report.SelectedCandidates, len(report.GroupSummaries))
}

func filterCandidates(candidates []Candidate, excludedGroups, excludedPaths []string) []Candidate {
	if len(excludedGroups) == 0 && len(excludedPaths) == 0 {
		return append([]Candidate(nil), candidates...)
	}

	selected := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		if isGroupExcluded(candidate, excludedGroups) || isPathExcluded(candidate.Path, excludedPaths) {
			continue
		}
		selected = append(selected, candidate)
	}
	sortCandidates(selected)
	return selected
}

func isPathExcluded(path string, excludedPaths []string) bool {
	for _, excluded := range excludedPaths {
		if path == excluded {
			return true
		}
	}
	return false
}
