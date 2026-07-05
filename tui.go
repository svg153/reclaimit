package reclaimit

import internaltui "github.com/svg153/reclaimit/internal/tui"

type selectionSnapshot = internaltui.Selection

func RunTUI(report Report) (Report, selectionSnapshot, error) {
	snapshot, err := internaltui.Run(toTUIReport(report))
	if err != nil {
		return report, selectionSnapshot{}, err
	}
	if snapshot.Saved {
		applySelection(&report, snapshot.ExcludedGroups, snapshot.ExcludedPaths)
	}
	return report, snapshot, nil
}

func toTUIReport(report Report) internaltui.Report {
	candidates := make([]internaltui.Candidate, len(report.Candidates))
	for i, candidate := range report.Candidates {
		candidates[i] = internaltui.Candidate{
			CategoryKey: candidate.CategoryKey,
			Path:        candidate.Path,
			Bytes:       candidate.Bytes,
			Description: candidate.Description,
			ModifiedAt:  candidate.ModifiedAt,
			IsDir:       candidate.IsDir,
		}
	}
	return internaltui.Report{
		Root:       report.Root,
		Candidates: candidates,
	}
}
