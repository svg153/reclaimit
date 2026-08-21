package scanner

import (
	"context"
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
// Workers controls the traversal-wide worker pool.
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
	OlderThan         time.Duration
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

type scanTask struct {
	id             int
	parentID       int
	path           string
	inCandidateDir bool
	depth          int
}

type scanInspection struct {
	task            scanTask
	summary         scanSummary
	entries         []os.DirEntry
	dirCategory     Category
	dirIsCandidate  bool
	nextInCandidate bool
	isDirectory     bool
	err             error
}

type directoryState struct {
	task           scanTask
	summary        scanSummary
	remaining      int
	dirCategory    Category
	dirIsCandidate bool
}

type scanOperations struct {
	lstat   func(string) (os.FileInfo, error)
	readDir func(string) ([]os.DirEntry, error)
}

var defaultScanOperations = scanOperations{
	lstat:   os.Lstat,
	readDir: os.ReadDir,
}

func AnalyzeWithOptions(command string, opts AnalyzeOptions, logger *slog.Logger) (Report, error) {
	return AnalyzeWithContext(context.Background(), command, opts, logger)
}

// AnalyzeWithContext runs a scan with traversal-wide cancellation. At most
// opts.Workers filesystem entries are inspected concurrently.
func AnalyzeWithContext(ctx context.Context, command string, opts AnalyzeOptions, logger *slog.Logger) (Report, error) {
	if err := ctx.Err(); err != nil {
		return Report{}, err
	}
	if err := validateCategoryFilters(opts.IncludeCategories, opts.ExcludeCategories); err != nil {
		return Report{}, err
	}
	if opts.MinCandidateSize < 0 {
		return Report{}, fmt.Errorf("minimum candidate size must be >= 0")
	}
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
		operations:     defaultScanOperations,
	}

	entries, err := sc.operations.readDir(opts.Root)
	if err != nil {
		return Report{}, err
	}

	logger.Debug("scan started", "root", opts.Root, "entries", len(entries), "workers", opts.Workers, "max_depth", opts.MaxDepth)
	rootResults, err := sc.scanTree(ctx, opts.Root, entries)
	if err != nil {
		return Report{}, err
	}
	for _, result := range rootResults {
		report.TotalBytes += result.summary.bytes
		if result.summary.bytes > 0 {
			sc.mu.Lock()
			report.TopEntries = PushTop(report.TopEntries, PathSize{Path: result.path, Bytes: result.summary.bytes}, opts.TopEntries)
			sc.mu.Unlock()
		}
	}

	SortPathSizes(report.TopEntries)
	if len(report.TopEntries) > opts.TopEntries {
		report.TopEntries = report.TopEntries[:opts.TopEntries]
	}
	SortPathSizes(report.TopFiles)
	if len(report.TopFiles) > opts.TopFiles {
		report.TopFiles = report.TopFiles[:opts.TopFiles]
	}

	if opts.OlderThan > 0 {
		FilterCandidatesByAge(&report, opts.OlderThan, time.Now().UTC())
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
	operations     scanOperations

	mu                   sync.Mutex
	entriesScanned       atomic.Int64
	entriesSkipped       atomic.Int64
	truncatedDirectories atomic.Int64
	maxDepthReached      atomic.Int64
}

func (sc *scanContext) scanTree(ctx context.Context, parent string, entries []os.DirEntry) ([]scanChildResult, error) {
	if len(entries) == 0 {
		return nil, nil
	}
	workerCount := sc.opts.Workers
	if workerCount <= 0 {
		workerCount = defaultWorkers
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan scanTask)
	results := make(chan scanInspection, workerCount)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-jobs:
					if !ok {
						return
					}
					inspection := sc.inspect(task)
					select {
					case results <- inspection:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	nextID := 0
	queue := make([]scanTask, 0, len(entries))
	for _, entry := range entries {
		queue = append(queue, scanTask{
			id:       nextID,
			parentID: -1,
			path:     filepath.Join(parent, entry.Name()),
			depth:    1,
		})
		nextID++
	}

	states := make(map[int]*directoryState)
	collected := make([]scanChildResult, 0, len(entries))

	finalizeDirectory := func(state *directoryState) {
		if state.dirIsCandidate && (!state.task.inCandidateDir || len(state.dirCategory.DirectoryPaths) > 0) &&
			IncludeCategory(state.dirCategory.Key, sc.includeSet, sc.excludeSet) && state.summary.bytes >= sc.opts.MinCandidateSize {
			sc.addCandidate(Candidate{
				Category:    state.dirCategory.Display,
				CategoryKey: state.dirCategory.Key,
				Path:        state.task.path,
				Group:       sc.groupFor(state.task.path),
				Bytes:       state.summary.bytes,
				Description: state.dirCategory.Description,
				ModifiedAt:  state.summary.modifiedAt,
				IsDir:       true,
			})
		}
	}

	complete := func(parentID int, path string, summary scanSummary) {
		for parentID >= 0 {
			state := states[parentID]
			state.summary.bytes += summary.bytes
			if summary.modifiedAt.After(state.summary.modifiedAt) {
				state.summary.modifiedAt = summary.modifiedAt
			}
			state.remaining--
			if state.remaining > 0 {
				return
			}
			finalizeDirectory(state)
			delete(states, parentID)
			path = state.task.path
			summary = state.summary
			parentID = state.task.parentID
		}
		collected = append(collected, scanChildResult{path: path, summary: summary})
	}

	stopWorkers := func() {
		close(jobs)
		wg.Wait()
	}

	for len(collected) < len(entries) {
		var jobsOut chan<- scanTask
		var next scanTask
		if len(queue) > 0 {
			jobsOut = jobs
			next = queue[0]
		}
		select {
		case <-ctx.Done():
			cancel()
			stopWorkers()
			return nil, ctx.Err()
		case jobsOut <- next:
			queue = queue[1:]
		case inspection := <-results:
			if inspection.err != nil {
				sc.skip(inspection.task.path, inspection.err)
				complete(inspection.task.parentID, inspection.task.path, scanSummary{})
				continue
			}
			if !inspection.isDirectory {
				complete(inspection.task.parentID, inspection.task.path, inspection.summary)
				continue
			}

			state := &directoryState{
				task:           inspection.task,
				summary:        inspection.summary,
				remaining:      len(inspection.entries),
				dirCategory:    inspection.dirCategory,
				dirIsCandidate: inspection.dirIsCandidate,
			}
			if state.remaining == 0 {
				finalizeDirectory(state)
				complete(state.task.parentID, state.task.path, state.summary)
				continue
			}
			states[state.task.id] = state
			for _, entry := range inspection.entries {
				queue = append(queue, scanTask{
					id:             nextID,
					parentID:       state.task.id,
					path:           filepath.Join(state.task.path, entry.Name()),
					inCandidateDir: inspection.nextInCandidate,
					depth:          state.task.depth + 1,
				})
				nextID++
			}
		}
	}

	stopWorkers()
	return collected, nil
}

func (sc *scanContext) inspect(task scanTask) scanInspection {
	inspection := scanInspection{task: task}
	info, err := sc.operations.lstat(task.path)
	if err != nil {
		inspection.err = err
		return inspection
	}
	sc.entriesScanned.Add(1)
	sc.updateMaxDepth(task.depth)
	if info.Mode()&os.ModeSymlink != 0 {
		sc.logger.Debug("skipping symlink", "path", task.path)
		return inspection
	}
	if info.IsDir() && sc.opts.MaxDepth > 0 && task.depth >= sc.opts.MaxDepth {
		sc.truncatedDirectories.Add(1)
		inspection.summary.modifiedAt = info.ModTime()
		return inspection
	}
	if info.IsDir() {
		inspection.isDirectory = true
		inspection.summary.modifiedAt = info.ModTime()
		inspection.dirCategory, inspection.dirIsCandidate = MatchDirectory(task.path)
		inspection.nextInCandidate = task.inCandidateDir || inspection.dirIsCandidate
		inspection.entries, inspection.err = sc.operations.readDir(task.path)
		return inspection
	}
	if !info.Mode().IsRegular() {
		return inspection
	}
	inspection.summary = sc.scanFile(task.path, info, task.inCandidateDir)
	return inspection
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
