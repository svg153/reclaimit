package scanner

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type CleanOptions struct {
	DryRun bool
	Logger *slog.Logger
}

type CleanIssue struct {
	Path   string
	Status string
	Reason string
	Bytes  int64
}

type CleanResult struct {
	Candidates        int
	ExpectedBytes     int64
	VerifiedBytes     int64
	DeletedBytes      int64
	DeletedCandidates int
	SkippedBytes      int64
	SkippedCandidates int
	FailedBytes       int64
	FailedCandidates  int
	Issues            []CleanIssue
}

type verifiedCandidate struct {
	candidate   Candidate
	actualBytes int64
}

// Clean removes verified candidates and keeps processing after an individual
// candidate fails. It preserves the historical API; callers that need the
// per-candidate audit trail should use CleanWithOptions.
func Clean(candidates []Candidate) (int64, error) {
	result, err := CleanWithOptions(candidates, CleanOptions{})
	if err != nil {
		return result.DeletedBytes, err
	}
	if result.FailedCandidates > 0 {
		return result.DeletedBytes, fmt.Errorf("clean failed for %d candidate(s)", result.FailedCandidates)
	}
	return result.DeletedBytes, nil
}

// DryRun validates candidates without removing anything. Missing or changed
// candidates are skipped, while filesystem failures are returned.
func DryRun(candidates []Candidate) (int64, error) {
	result, err := CleanWithOptions(candidates, CleanOptions{DryRun: true})
	if err != nil {
		return result.VerifiedBytes, err
	}
	if result.FailedCandidates > 0 {
		return result.VerifiedBytes, fmt.Errorf("dry-run failed for %d candidate(s)", result.FailedCandidates)
	}
	return result.VerifiedBytes, nil
}

// CleanWithOptions performs a full preflight before deletion, records every
// candidate outcome, and verifies that successful deletions removed the path.
// DryRun uses the same preflight and reporting path without deleting anything.
func CleanWithOptions(candidates []Candidate, options CleanOptions) (CleanResult, error) {
	logger := options.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	normalized := collapseNestedCandidates(candidates)
	result := CleanResult{
		Candidates:    len(normalized),
		ExpectedBytes: sumCandidateBytes(normalized),
	}
	verified := make([]verifiedCandidate, 0, len(normalized))

	// Preflight every candidate before deleting any candidate.
	for _, candidate := range normalized {
		actualBytes, issue := verifyCandidate(candidate)
		if issue != nil {
			result.addIssue(candidate, *issue)
			logger.Warn("clean candidate skipped",
				"path", candidate.Path,
				"status", issue.Status,
				"reason", issue.Reason,
			)
			continue
		}
		result.VerifiedBytes += actualBytes
		verified = append(verified, verifiedCandidate{candidate: candidate, actualBytes: actualBytes})
		logger.Info("clean candidate verified",
			"path", candidate.Path,
			"bytes", actualBytes,
		)
	}

	if options.DryRun {
		logger.Info("clean dry-run completed",
			"candidates", result.Candidates,
			"verified", len(verified),
			"verified_bytes", result.VerifiedBytes,
			"skipped", result.SkippedCandidates,
			"failed", result.FailedCandidates,
		)
		return result, nil
	}

	for _, item := range verified {
		if err := os.RemoveAll(item.candidate.Path); err != nil {
			result.addIssue(item.candidate, CleanIssue{
				Status: "failed",
				Reason: err.Error(),
				Bytes: item.actualBytes,
			})
			logger.Warn("clean candidate failed",
				"path", item.candidate.Path,
				"status", "failed",
				"reason", err,
			)
			continue
		}
		if _, err := os.Lstat(item.candidate.Path); err == nil {
			result.addIssue(item.candidate, CleanIssue{
				Status: "failed",
				Reason: "path still exists after deletion",
				Bytes: item.actualBytes,
			})
			logger.Warn("clean candidate failed",
				"path", item.candidate.Path,
				"status", "failed",
				"reason", "path still exists after deletion",
			)
			continue
		} else if !os.IsNotExist(err) {
			result.addIssue(item.candidate, CleanIssue{
				Status: "failed",
				Reason: fmt.Sprintf("post-delete verification: %v", err),
				Bytes: item.actualBytes,
			})
			logger.Warn("clean candidate failed",
				"path", item.candidate.Path,
				"status", "failed",
				"reason", err,
			)
			continue
		}

		result.DeletedCandidates++
		result.DeletedBytes += item.actualBytes
		logger.Info("clean candidate deleted",
			"path", item.candidate.Path,
			"bytes", item.actualBytes,
		)
	}

	logger.Info("clean completed",
		"candidates", result.Candidates,
		"deleted", result.DeletedCandidates,
		"deleted_bytes", result.DeletedBytes,
		"skipped", result.SkippedCandidates,
		"failed", result.FailedCandidates,
	)
	return result, nil
}

func (result *CleanResult) addIssue(candidate Candidate, issue CleanIssue) {
	if issue.Bytes == 0 {
		issue.Bytes = candidate.Bytes
	}
	issue.Path = candidate.Path
	result.Issues = append(result.Issues, issue)
	switch issue.Status {
	case "skipped":
		result.SkippedCandidates++
		result.SkippedBytes += issue.Bytes
	case "failed":
		result.FailedCandidates++
		result.FailedBytes += issue.Bytes
	}
}

func verifyCandidate(candidate Candidate) (int64, *CleanIssue) {
	info, err := os.Lstat(candidate.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, &CleanIssue{Status: "skipped", Reason: "path no longer exists", Bytes: candidate.Bytes}
		}
		return 0, &CleanIssue{Status: "failed", Reason: fmt.Sprintf("stat: %v", err), Bytes: candidate.Bytes}
	}
	if info.IsDir() != candidate.IsDir {
		return 0, &CleanIssue{
			Status: "skipped",
			Reason: "path type changed since scan",
			Bytes: candidate.Bytes,
		}
	}

	actualBytes, latestModified, err := pathUsage(candidate.Path, info)
	if err != nil {
		return 0, &CleanIssue{
			Status: "failed",
			Reason: fmt.Sprintf("measure: %v", err),
			Bytes: candidate.Bytes,
		}
	}
	if actualBytes != candidate.Bytes {
		return 0, &CleanIssue{
			Status: "skipped",
			Reason: fmt.Sprintf("size changed since scan: expected %d, found %d", candidate.Bytes, actualBytes),
			Bytes: candidate.Bytes,
		}
	}
	if !candidate.ModifiedAt.IsZero() && !latestModified.Equal(candidate.ModifiedAt) {
		return 0, &CleanIssue{
			Status: "skipped",
			Reason: "modified since scan",
			Bytes: candidate.Bytes,
		}
	}
	return actualBytes, nil
}

func pathUsage(path string, info os.FileInfo) (int64, time.Time, error) {
	if !info.IsDir() {
		if !info.Mode().IsRegular() {
			return 0, time.Time{}, nil
		}
		return info.Size(), info.ModTime(), nil
	}

	total := int64(0)
	latest := info.ModTime()
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, time.Time{}, err
	}
	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		childInfo, err := os.Lstat(childPath)
		if err != nil {
			return 0, time.Time{}, err
		}
		if childInfo.Mode()&os.ModeSymlink != 0 {
			continue
		}
		childBytes, childModified, err := pathUsage(childPath, childInfo)
		if err != nil {
			return 0, time.Time{}, err
		}
		total += childBytes
		if childModified.After(latest) {
			latest = childModified
		}
	}
	return total, latest, nil
}

func collapseNestedCandidates(candidates []Candidate) []Candidate {
	normalized := append([]Candidate(nil), candidates...)
	sort.SliceStable(normalized, func(i, j int) bool {
		iDepth := pathDepth(normalized[i].Path)
		jDepth := pathDepth(normalized[j].Path)
		if iDepth != jDepth {
			return iDepth < jDepth
		}
		return normalized[i].Path < normalized[j].Path
	})

	selected := make([]Candidate, 0, len(normalized))
	for _, candidate := range normalized {
		if candidate.Path == "" {
			selected = append(selected, candidate)
			continue
		}
		nested := false
		for _, parent := range selected {
			if isDescendant(candidate.Path, parent.Path) {
				nested = true
				break
			}
		}
		if !nested {
			selected = append(selected, candidate)
		}
	}
	return selected
}

func isDescendant(path, parent string) bool {
	relative, err := filepath.Rel(filepath.Clean(parent), filepath.Clean(path))
	if err != nil || relative == "." || relative == ".." {
		return false
	}
	return !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

func pathDepth(path string) int {
	cleaned := filepath.Clean(path)
	if cleaned == "." || cleaned == string(filepath.Separator) {
		return 0
	}
	return strings.Count(cleaned, string(filepath.Separator))
}

func sumCandidateBytes(candidates []Candidate) int64 {
	var total int64
	for _, candidate := range candidates {
		total += candidate.Bytes
	}
	return total
}

// filesystemUsage is implemented per-OS in filesystem_*.go
