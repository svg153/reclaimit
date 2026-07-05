package reclaimit

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func Analyze(cfg config) (Report, error) {
	return AnalyzeWithOptions(cfg.command, analyzeOptionsFromConfig(cfg), cfg.log())
}

func analyzeOptionsFromConfig(cfg config) AnalyzeOptions {
	return AnalyzeOptions{
		Root:              cfg.root,
		GroupMode:         cfg.groupMode,
		GroupDepth:        cfg.groupDepth,
		TopFiles:          cfg.topFiles,
		TopGroups:         cfg.topGroups,
		TopEntries:        cfg.topEntries,
		MinCandidateSize:  cfg.minCandidateSize,
		IncludeCategories: append([]string(nil), cfg.includeCategories...),
		ExcludeCategories: append([]string(nil), cfg.excludeCategories...),
		ExcludeGroups:     append([]string(nil), cfg.excludeGroups...),
		ExcludePaths:      append([]string(nil), cfg.excludePaths...),
	}
}

func AnalyzeWithOptions(command string, opts AnalyzeOptions, logger *slog.Logger) (Report, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	rootInfo, err := os.Lstat(opts.Root)
	if err != nil {
		return Report{}, err
	}
	if !rootInfo.IsDir() {
		return Report{}, fmt.Errorf("%s is not a directory", opts.Root)
	}

	report := Report{
		Command: command,
		Root:    opts.Root,
	}
	filesystemBytes, freeBytes, availableBytes := filesystemUsage(opts.Root)
	report.FilesystemBytes = filesystemBytes
	report.FreeBytes = freeBytes
	report.AvailableBytes = availableBytes

	sc := &scanContext{
		opts:           opts,
		report:         &report,
		candidateByKey: make(map[string]int),
		groupCache:     map[string]string{},
		includeSet:     listToSet(opts.IncludeCategories),
		excludeSet:     listToSet(opts.ExcludeCategories),
		logger:         logger,
	}

	entries, err := os.ReadDir(opts.Root)
	if err != nil {
		return Report{}, err
	}

	logger.Debug("scan started", "root", opts.Root, "entries", len(entries))
	for _, entry := range entries {
		childPath := filepath.Join(opts.Root, entry.Name())
		summary, err := sc.scan(childPath, false)
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				logger.Warn("skipping unreadable entry", "path", childPath)
				continue
			}
			return Report{}, err
		}
		if summary.bytes > 0 {
			report.TopEntries = append(report.TopEntries, PathSize{Path: childPath, Bytes: summary.bytes})
		}
	}

	report.TotalBytes = sumBytes(report.TopEntries)
	sortPathSizes(report.TopEntries)
	if len(report.TopEntries) > opts.TopEntries {
		report.TopEntries = report.TopEntries[:opts.TopEntries]
	}
	sortPathSizes(report.TopFiles)
	if len(report.TopFiles) > opts.TopFiles {
		report.TopFiles = report.TopFiles[:opts.TopFiles]
	}

	sortCandidates(report.Candidates)
	report.CandidateBytes = sumCandidateBytes(report.Candidates)
	report.CategorySummaries = summarizeCategories(report.Candidates)
	report.GroupSummaries = summarizeGroups(report.Candidates, opts.TopGroups)
	applySelection(&report, opts.ExcludeGroups, opts.ExcludePaths)
	report.SelectedGroupSummaries = summarizeGroups(report.SelectedCandidates, opts.TopGroups)
	logger.Info("scan completed",
		"candidates", len(report.Candidates),
		"reclaimable_bytes", report.CandidateBytes,
		"selected", len(report.SelectedCandidates),
		"selected_bytes", report.SelectedBytes,
	)
	return report, nil
}

// scanContext carries the per-invocation state for a filesystem walk so the
// recursive scan methods keep a small, stable signature.
type scanContext struct {
	opts           AnalyzeOptions
	report         *Report
	candidateByKey map[string]int
	groupCache     map[string]string
	includeSet     map[string]struct{}
	excludeSet     map[string]struct{}
	logger         *slog.Logger
}

func (sc *scanContext) scan(path string, inCandidateDir bool) (scanSummary, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return scanSummary{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		sc.logger.Debug("skipping symlink", "path", path)
		return scanSummary{}, nil
	}
	if info.IsDir() {
		return sc.scanDir(path, info, inCandidateDir)
	}
	if !info.Mode().IsRegular() {
		return scanSummary{}, nil
	}
	return sc.scanFile(path, info, inCandidateDir), nil
}

func (sc *scanContext) scanDir(path string, info os.FileInfo, inCandidateDir bool) (scanSummary, error) {
	dirCategory, dirIsCandidate := matchDirectory(info.Name())
	nextInCandidate := inCandidateDir || dirIsCandidate
	entries, err := os.ReadDir(path)
	if err != nil {
		return scanSummary{}, err
	}
	var total int64
	latestModified := info.ModTime()
	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		summary, err := sc.scan(childPath, nextInCandidate)
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				sc.logger.Warn("skipping unreadable path", "path", childPath)
				continue
			}
			return scanSummary{}, err
		}
		total += summary.bytes
		if summary.modifiedAt.After(latestModified) {
			latestModified = summary.modifiedAt
		}
	}
	if dirIsCandidate && !inCandidateDir && includeCategory(dirCategory.Key, sc.includeSet, sc.excludeSet) && total >= sc.opts.MinCandidateSize {
		sc.addCandidate(Candidate{
			Category:    dirCategory.Display,
			CategoryKey: dirCategory.Key,
			Path:        path,
			Group:       determineGroup(path, sc.opts, sc.groupCache),
			Bytes:       total,
			Description: dirCategory.Description,
			ModifiedAt:  latestModified,
			IsDir:       true,
		})
	}
	return scanSummary{bytes: total, modifiedAt: latestModified}, nil
}

func (sc *scanContext) scanFile(path string, info os.FileInfo, inCandidateDir bool) scanSummary {
	size := info.Size()
	sc.report.TopFiles = pushTop(sc.report.TopFiles, PathSize{Path: path, Bytes: size}, sc.opts.TopFiles)

	fileCategory, fileIsCandidate := matchFile(path)
	if fileIsCandidate && !inCandidateDir && includeCategory(fileCategory.Key, sc.includeSet, sc.excludeSet) && size >= sc.opts.MinCandidateSize {
		sc.addCandidate(Candidate{
			Category:    fileCategory.Display,
			CategoryKey: fileCategory.Key,
			Path:        path,
			Group:       determineGroup(path, sc.opts, sc.groupCache),
			Bytes:       size,
			Description: fileCategory.Description,
			ModifiedAt:  info.ModTime(),
			IsDir:       false,
		})
	}
	return scanSummary{bytes: size, modifiedAt: info.ModTime()}
}

func (sc *scanContext) addCandidate(candidate Candidate) {
	key := candidate.CategoryKey + ":" + candidate.Path
	if _, exists := sc.candidateByKey[key]; exists {
		return
	}
	sc.candidateByKey[key] = len(sc.report.Candidates)
	sc.report.Candidates = append(sc.report.Candidates, candidate)
	sc.logger.Debug("candidate found",
		"category", candidate.CategoryKey,
		"path", candidate.Path,
		"bytes", candidate.Bytes,
	)
}

func determineGroup(path string, opts AnalyzeOptions, cache map[string]string) string {
	if opts.GroupMode == "repo" {
		if repoRoot := findRepoRoot(path, opts.Root, cache); repoRoot != "" {
			return repoRoot
		}
	}
	return ancestorGroup(path, opts.Root, opts.GroupDepth)
}

func findRepoRoot(path, root string, cache map[string]string) string {
	cursor := path
	info, err := os.Lstat(cursor)
	if err == nil && !info.IsDir() {
		cursor = filepath.Dir(cursor)
	}

	for {
		if value, ok := cache[cursor]; ok {
			return value
		}
		if _, err := os.Stat(filepath.Join(cursor, ".git")); err == nil {
			cache[cursor] = cursor
			return cursor
		}
		if cursor == root {
			cache[cursor] = ""
			return ""
		}
		parent := filepath.Dir(cursor)
		if parent == cursor {
			cache[cursor] = ""
			return ""
		}
		cursor = parent
	}
}

func ancestorGroup(path, root string, depth int) string {
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == "." {
		return root
	}
	parts := strings.Split(relative, string(filepath.Separator))
	if len(parts) == 0 {
		return root
	}
	if len(parts) < depth {
		depth = len(parts)
	}
	groupParts := parts[:depth]
	return filepath.Join(append([]string{root}, groupParts...)...)
}
