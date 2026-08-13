package scanner

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/svg153/reclaimit/internal/filesystem"
)

const defaultWorkers = 8

// AnalyzeOptions are scanner-specific options independent from CLI routing.
// MaxDepth is the maximum traversal depth below Root; zero means unlimited.
// Workers controls the bounded worker pool used for directory entries.
type AnalyzeOptions struct {
	Root              string
	GroupMode         string
	GroupDepth        int
	TopFiles          int
	TopGroups         int
	TopEntries        int
	MinCandidateSize  int64
	MaxDepth          int
	Workers           int
	IncludeCategories []string
	ExcludeCategories []string
	ExcludeGroups     []string
	ExcludePaths      []string
}

type scanSummary struct {
	bytes      int64
	modifiedAt time.Time
}

type scanChildResult struct {
	path    string
	summary scanSummary
}

func AnalyzeWithOptions(command string, opts AnalyzeOptions, logger *slog.Logger) (Report, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if opts.MaxDepth < 0 {
		return Report{}, fmt.Errorf("max depth must be >= 0")
	}
	if opts.Workers <= 0 {
		opts.Workers = defaultWorkers
	}

	rootInfo, err := os.Lstat(opts.Root)
	if err != nil {
		return Report{}, err
	}
	if !rootInfo.IsDir() {
		return Report{}, fmt.Errorf("%s is not a directory", opts.Root)
	}

	started := time.Now()
	report := Report{
		Command: command,
		Root:    opts.Root,
	}
	filesystemBytes, freeBytes, availableBytes := filesystem.FilesystemUsage(opts.Root)
	report.FilesystemBytes = filesystemBytes
	report.FreeBytes = freeBytes
	report.AvailableBytes = availableBytes

	sc := &scanContext{
		opts:           opts,
		report:         &report,
		candidateByKey: make(map[string]int),
		groupCache:     map[string]string{},
		includeSet:     ListToSet(opts.IncludeCategories),
		excludeSet:     ListToSet(opts.ExcludeCategories),
		logger:         logger,
	}

	entries, err := os.ReadDir(opts.Root)
	if err != nil {
		return Report{}, err
	}

	logger.Debug("scan started", "root", opts.Root, "entries", len(entries), "workers", opts.Workers, "max_depth", opts.MaxDepth)
	rootResults := sc.scanEntries(opts.Root, entries, false, 1)
	for _, result := range rootResults {
		if result.summary.bytes > 0 {
			sc.mu.Lock()
			report.TopEntries = PushTop(report.TopEntries, PathSize{Path: result.path, Bytes: result.summary.bytes}, opts.TopEntries)
			sc.mu.Unlock()
		}
	}

	report.TotalBytes = SumBytes(report.TopEntries)
	SortPathSizes(report.TopEntries)
	if len(report.TopEntries) > opts.TopEntries {
		report.TopEntries = report.TopEntries[:opts.TopEntries]
	}
	SortPathSizes(report.TopFiles)
	if len(report.TopFiles) > opts.TopFiles {
		report.TopFiles = report.TopFiles[:opts.TopFiles]
	}

	SortCandidates(report.Candidates)
	report.CandidateBytes = SumCandidateBytes(report.Candidates)
	report.CategorySummaries = SummarizeCategories(report.Candidates)
	report.GroupSummaries = SummarizeGroups(report.Candidates, opts.TopGroups)
	ApplySelection(&report, opts.ExcludeGroups, opts.ExcludePaths)
	report.SelectedGroupSummaries = SummarizeGroups(report.SelectedCandidates, opts.TopGroups)
	report.EntriesScanned = sc.entriesScanned.Load()
	report.EntriesSkipped = sc.entriesSkipped.Load()
	report.TruncatedDirectories = sc.truncatedDirectories.Load()
	report.MaxDepthReached = int(sc.maxDepthReached.Load())
	logger.Info("scan completed",
		"candidates", len(report.Candidates),
		"reclaimable_bytes", report.CandidateBytes,
		"selected", len(report.SelectedCandidates),
		"selected_bytes", report.SelectedBytes,
		"entries_scanned", report.EntriesScanned,
		"entries_skipped", report.EntriesSkipped,
		"truncated_directories", report.TruncatedDirectories,
		"max_depth_reached", report.MaxDepthReached,
		"duration_ms", time.Since(started).Milliseconds(),
	)
	return report, nil
}

// scanContext carries the per-invocation state for a filesystem walk.
type scanContext struct {
	opts           AnalyzeOptions
	report         *Report
	candidateByKey map[string]int
	groupCache     map[string]string
	includeSet     map[string]struct{}
	excludeSet     map[string]struct{}
	logger         *slog.Logger

	mu                   sync.Mutex
	entriesScanned       atomic.Int64
	entriesSkipped       atomic.Int64
	truncatedDirectories atomic.Int64
	maxDepthReached      atomic.Int64
}

func (sc *scanContext) scanEntries(parent string, entries []os.DirEntry, inCandidateDir bool, depth int) []scanChildResult {
	if len(entries) == 0 {
		return nil
	}
	workers := sc.opts.Workers
	if workers > len(entries) {
		workers = len(entries)
	}
	jobs := make(chan os.DirEntry)
	results := make(chan scanChildResult, len(entries))
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range jobs {
				childPath := filepath.Join(parent, entry.Name())
				summary, err := sc.scan(childPath, inCandidateDir, depth)
				if err != nil {
					sc.skip(childPath, err)
					continue
				}
				results <- scanChildResult{path: childPath, summary: summary}
			}
		}()
	}
	for _, entry := range entries {
		jobs <- entry
	}
	close(jobs)
	wg.Wait()
	close(results)

	collected := make([]scanChildResult, 0, len(results))
	for result := range results {
		collected = append(collected, result)
	}
	return collected
}

func (sc *scanContext) scan(path string, inCandidateDir bool, depth int) (scanSummary, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return scanSummary{}, err
	}
	sc.entriesScanned.Add(1)
	sc.updateMaxDepth(depth)
	if info.Mode()&os.ModeSymlink != 0 {
		sc.logger.Debug("skipping symlink", "path", path)
		return scanSummary{}, nil
	}
	if info.IsDir() && sc.opts.MaxDepth > 0 && depth >= sc.opts.MaxDepth {
		sc.truncatedDirectories.Add(1)
		return scanSummary{modifiedAt: info.ModTime()}, nil
	}
	if info.IsDir() {
		return sc.scanDir(path, info, inCandidateDir, depth)
	}
	if !info.Mode().IsRegular() {
		return scanSummary{}, nil
	}
	return sc.scanFile(path, info, inCandidateDir), nil
}

func (sc *scanContext) scanDir(path string, info os.FileInfo, inCandidateDir bool, depth int) (scanSummary, error) {
	dirCategory, dirIsCandidate := MatchDirectory(path)
	nextInCandidate := inCandidateDir || dirIsCandidate
	entries, err := os.ReadDir(path)
	if err != nil {
		return scanSummary{}, err
	}
	children := sc.scanEntries(path, entries, nextInCandidate, depth+1)
	var total int64
	latestModified := info.ModTime()
	for _, child := range children {
		total += child.summary.bytes
		if child.summary.modifiedAt.After(latestModified) {
			latestModified = child.summary.modifiedAt
		}
	}
	if dirIsCandidate && (!inCandidateDir || len(dirCategory.DirectoryPaths) > 0) && IncludeCategory(dirCategory.Key, sc.includeSet, sc.excludeSet) && total >= sc.opts.MinCandidateSize {
		sc.addCandidate(Candidate{
			Category:    dirCategory.Display,
			CategoryKey: dirCategory.Key,
			Path:        path,
			Group:       sc.groupFor(path),
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
	sc.mu.Lock()
	sc.report.TopFiles = PushTop(sc.report.TopFiles, PathSize{Path: path, Bytes: size}, sc.opts.TopFiles)
	sc.mu.Unlock()

	fileCategory, fileIsCandidate := MatchFile(path)
	if fileIsCandidate && !inCandidateDir && IncludeCategory(fileCategory.Key, sc.includeSet, sc.excludeSet) && size >= sc.opts.MinCandidateSize {
		sc.addCandidate(Candidate{
			Category:    fileCategory.Display,
			CategoryKey: fileCategory.Key,
			Path:        path,
			Group:       sc.groupFor(path),
			Bytes:       size,
			Description: fileCategory.Description,
			ModifiedAt:  info.ModTime(),
			IsDir:       false,
		})
	}
	return scanSummary{bytes: size, modifiedAt: info.ModTime()}
}

func (sc *scanContext) groupFor(path string) string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return DetermineGroup(path, sc.opts, sc.groupCache)
}

func (sc *scanContext) addCandidate(candidate Candidate) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
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

func (sc *scanContext) skip(path string, err error) {
	sc.entriesSkipped.Add(1)
	sc.logger.Warn("skipping unreadable entry", "path", path, "error", err)
}

func (sc *scanContext) updateMaxDepth(depth int) {
	for {
		current := sc.maxDepthReached.Load()
		if int64(depth) <= current || sc.maxDepthReached.CompareAndSwap(current, int64(depth)) {
			return
		}
	}
}
