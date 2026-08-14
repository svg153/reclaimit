package scanner

import (
	"errors"
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
	Path           string `json:"path"`
	QuarantinePath string `json:"quarantine_path,omitempty"`
	Status         string `json:"status"`
	Reason         string `json:"reason"`
	Bytes          int64  `json:"bytes"`
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
	info        os.FileInfo
}

type cleanOperations struct {
	lstat     func(string) (os.FileInfo, error)
	mkdirTemp func(string, string) (string, error)
	rename    func(string, string) error
	removeAll func(string) error
	remove    func(string) error
}

var defaultCleanOperations = cleanOperations{
	lstat:     os.Lstat,
	mkdirTemp: os.MkdirTemp,
	rename:    os.Rename,
	removeAll: os.RemoveAll,
	remove:    os.Remove,
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
	return cleanWithOperations(candidates, options, defaultCleanOperations)
}

func cleanWithOperations(candidates []Candidate, options CleanOptions, operations cleanOperations) (CleanResult, error) {
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
		actualBytes, info, issue := verifyCandidate(candidate)
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
		verified = append(verified, verifiedCandidate{candidate: candidate, actualBytes: actualBytes, info: info})
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
		deleted, issue := quarantineAndDelete(item, operations)
		if issue != nil {
			result.addIssue(item.candidate, *issue)
			logger.Warn("clean candidate outcome",
				"path", item.candidate.Path,
				"quarantine_path", issue.QuarantinePath,
				"status", issue.Status,
				"reason", issue.Reason,
			)
		}
		if !deleted {
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

// quarantineAndDelete atomically detaches a candidate from its public path by
// renaming it into a private sibling directory. Identity and contents are
// checked again after the rename; changed data is left recoverable instead of
// being recursively deleted.
func quarantineAndDelete(item verifiedCandidate, operations cleanOperations) (bool, *CleanIssue) {
	actualBytes, currentInfo, issue := verifyCandidate(item.candidate)
	if issue != nil {
		return false, issue
	}
	if !os.SameFile(item.info, currentInfo) {
		return false, &CleanIssue{
			Status: "skipped",
			Reason: "path identity changed since preflight",
			Bytes:  item.actualBytes,
		}
	}
	if actualBytes != item.actualBytes {
		return false, &CleanIssue{
			Status: "skipped",
			Reason: "path contents changed since preflight",
			Bytes:  item.actualBytes,
		}
	}

	parent := filepath.Dir(filepath.Clean(item.candidate.Path))
	quarantineDir, err := operations.mkdirTemp(parent, ".reclaimit-quarantine-")
	if err != nil {
		return false, &CleanIssue{Status: "failed", Reason: fmt.Sprintf("create quarantine: %v", err), Bytes: item.actualBytes}
	}
	quarantinePath := filepath.Join(quarantineDir, "candidate")
	if err := operations.rename(item.candidate.Path, quarantinePath); err != nil {
		reason := fmt.Sprintf("quarantine rename: %v", err)
		if cleanupErr := operations.remove(quarantineDir); cleanupErr != nil {
			reason += fmt.Sprintf("; empty quarantine cleanup: %v", cleanupErr)
		}
		return false, &CleanIssue{Status: "failed", Reason: reason, Bytes: item.actualBytes}
	}

	quarantinedInfo, err := operations.lstat(quarantinePath)
	if err != nil {
		return false, quarantinedIssue(quarantinePath, item.actualBytes, "inspect quarantined candidate", err)
	}
	if !os.SameFile(currentInfo, quarantinedInfo) {
		return false, &CleanIssue{
			Status:         "failed",
			Reason:         "candidate identity changed during quarantine; data was not deleted",
			Bytes:          item.actualBytes,
			QuarantinePath: quarantinePath,
		}
	}

	quarantinedCandidate := item.candidate
	quarantinedCandidate.Path = quarantinePath
	quarantinedBytes, _, changed := verifyCandidate(quarantinedCandidate)
	if changed != nil || quarantinedBytes != item.actualBytes {
		reason := "candidate contents changed during quarantine; data was not deleted"
		if changed != nil {
			reason += ": " + changed.Reason
		}
		return false, &CleanIssue{
			Status:         "failed",
			Reason:         reason,
			Bytes:          item.actualBytes,
			QuarantinePath: quarantinePath,
		}
	}

	if err := operations.removeAll(quarantinePath); err != nil {
		return false, quarantinedIssue(quarantinePath, item.actualBytes, "delete quarantined candidate", err)
	}
	if _, err := operations.lstat(quarantinePath); err == nil {
		return false, quarantinedIssue(quarantinePath, item.actualBytes, "verify quarantined deletion", errors.New("path still exists"))
	} else if !os.IsNotExist(err) {
		return false, quarantinedIssue(quarantinePath, item.actualBytes, "verify quarantined deletion", err)
	}

	if err := operations.remove(quarantineDir); err != nil {
		return true, &CleanIssue{
			Status:         "warning",
			Reason:         fmt.Sprintf("candidate deleted but empty quarantine cleanup failed: %v", err),
			Bytes:          item.actualBytes,
			QuarantinePath: quarantineDir,
		}
	}
	return true, nil
}

func quarantinedIssue(path string, bytes int64, action string, err error) *CleanIssue {
	return &CleanIssue{
		Status:         "failed",
		Reason:         fmt.Sprintf("%s: %v; data may remain recoverable at quarantine path", action, err),
		Bytes:          bytes,
		QuarantinePath: path,
	}
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

func verifyCandidate(candidate Candidate) (int64, os.FileInfo, *CleanIssue) {
	cleanedPath := filepath.Clean(candidate.Path)
	if candidate.Path == "" || cleanedPath == "." || filepath.Dir(cleanedPath) == cleanedPath {
		return 0, nil, &CleanIssue{Status: "skipped", Reason: "filesystem roots and empty paths are protected", Bytes: candidate.Bytes}
	}
	info, err := os.Lstat(candidate.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil, &CleanIssue{Status: "skipped", Reason: "path no longer exists", Bytes: candidate.Bytes}
		}
		return 0, nil, &CleanIssue{Status: "failed", Reason: fmt.Sprintf("stat: %v", err), Bytes: candidate.Bytes}
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return 0, nil, &CleanIssue{Status: "skipped", Reason: "symbolic links are not cleanup candidates", Bytes: candidate.Bytes}
	}
	if !info.IsDir() && !info.Mode().IsRegular() {
		return 0, nil, &CleanIssue{Status: "skipped", Reason: "special files are not cleanup candidates", Bytes: candidate.Bytes}
	}
	if info.IsDir() != candidate.IsDir {
		return 0, nil, &CleanIssue{
			Status: "skipped",
			Reason: "path type changed since scan",
			Bytes:  candidate.Bytes,
		}
	}

	actualBytes, latestModified, err := pathUsage(candidate.Path, info)
	if err != nil {
		return 0, nil, &CleanIssue{
			Status: "failed",
			Reason: fmt.Sprintf("measure: %v", err),
			Bytes:  candidate.Bytes,
		}
	}
	if actualBytes != candidate.Bytes {
		return 0, nil, &CleanIssue{
			Status: "skipped",
			Reason: fmt.Sprintf("size changed since scan: expected %d, found %d", candidate.Bytes, actualBytes),
			Bytes:  candidate.Bytes,
		}
	}
	if !candidate.ModifiedAt.IsZero() && !latestModified.Equal(candidate.ModifiedAt) {
		return 0, nil, &CleanIssue{
			Status: "skipped",
			Reason: "modified since scan",
			Bytes:  candidate.Bytes,
		}
	}
	return actualBytes, info, nil
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

// filesystemUsage is implemented per-OS in filesystem_*.go
