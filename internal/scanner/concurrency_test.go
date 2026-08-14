package scanner

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func scannerTestOptions(root string, workers int) AnalyzeOptions {
	return AnalyzeOptions{
		Root: root, GroupMode: "repo", GroupDepth: 1,
		TopFiles: 100, TopGroups: 100, TopEntries: 100,
		MinCandidateSize: 1, Workers: workers,
	}
}

func TestAnalyzeConcurrentTraversalIsDeterministic(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 12; i++ {
		dir := filepath.Join(root, "repo-"+strconv.Itoa(i), "node_modules")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for j := 0; j < 5; j++ {
			path := filepath.Join(dir, "dependency-"+strconv.Itoa(j)+".js")
			if err := os.WriteFile(path, []byte("dependency"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	serial, err := AnalyzeWithOptions("analyze", scannerTestOptions(root, 1), nil)
	if err != nil {
		t.Fatal(err)
	}
	parallel, err := AnalyzeWithOptions("analyze", scannerTestOptions(root, 8), nil)
	if err != nil {
		t.Fatal(err)
	}

	serialCandidates := candidateKeys(serial)
	parallelCandidates := candidateKeys(parallel)
	if !reflect.DeepEqual(serialCandidates, parallelCandidates) {
		t.Fatalf("serial and parallel candidates differ:\nserial=%v\nparallel=%v", serialCandidates, parallelCandidates)
	}
	if parallel.EntriesScanned == 0 || parallel.MaxDepthReached == 0 {
		t.Fatalf("expected traversal metrics, got %#v", parallel)
	}
}

func TestAnalyzeMaxDepthStopsTraversal(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "one", "two", "three")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deep, "payload"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := AnalyzeWithOptions("analyze", AnalyzeOptions{
		Root: root, GroupMode: "repo", GroupDepth: 1,
		TopFiles: 10, TopGroups: 10, TopEntries: 10,
		MinCandidateSize: 1, MaxDepth: 2, Workers: 2,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report.TruncatedDirectories == 0 {
		t.Fatalf("expected at least one truncated directory, got %#v", report)
	}
	if report.MaxDepthReached != 2 {
		t.Fatalf("MaxDepthReached = %d, want 2", report.MaxDepthReached)
	}
	for _, candidate := range report.Candidates {
		if candidate.Path == deep {
			t.Fatalf("did not expect candidate below max depth: %#v", candidate)
		}
	}
}

func candidateKeys(report Report) []string {
	keys := make([]string, 0, len(report.Candidates))
	for _, candidate := range report.Candidates {
		keys = append(keys, candidate.CategoryKey+":"+candidate.Path)
	}
	sort.Strings(keys)
	return keys
}

func TestTraversalWideWorkerLimit(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 12; i++ {
		if err := os.WriteFile(filepath.Join(root, "file-"+strconv.Itoa(i)), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}

	var active atomic.Int64
	var maximum atomic.Int64
	started := make(chan struct{}, len(entries))
	release := make(chan struct{})
	operations := defaultScanOperations
	operations.lstat = func(path string) (os.FileInfo, error) {
		current := active.Add(1)
		for {
			observed := maximum.Load()
			if current <= observed || maximum.CompareAndSwap(observed, current) {
				break
			}
		}
		started <- struct{}{}
		<-release
		active.Add(-1)
		return os.Lstat(path)
	}

	sc := schedulerTestContext(root, 3, operations)
	done := make(chan error, 1)
	go func() {
		_, scanErr := sc.scanTree(context.Background(), root, entries)
		done <- scanErr
	}()
	for i := 0; i < 3; i++ {
		<-started
	}
	if got := active.Load(); got != 3 {
		t.Fatalf("active workers = %d, want 3", got)
	}
	close(release)
	select {
	case scanErr := <-done:
		if scanErr != nil {
			t.Fatal(scanErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("global scheduler did not finish")
	}
	if got := maximum.Load(); got != 3 {
		t.Fatalf("maximum active workers = %d, want 3", got)
	}
	if got := active.Load(); got != 0 {
		t.Fatalf("workers still active after completion: %d", got)
	}
}

func TestTraversalCancellationStopsQueuedWork(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 32; i++ {
		if err := os.WriteFile(filepath.Join(root, "file-"+strconv.Itoa(i)), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var active atomic.Int64
	var inspected atomic.Int64
	started := make(chan struct{}, 4)
	operations := defaultScanOperations
	operations.lstat = func(string) (os.FileInfo, error) {
		inspected.Add(1)
		active.Add(1)
		defer active.Add(-1)
		started <- struct{}{}
		<-ctx.Done()
		return nil, ctx.Err()
	}
	sc := schedulerTestContext(root, 4, operations)
	done := make(chan error, 1)
	go func() {
		_, scanErr := sc.scanTree(ctx, root, entries)
		done <- scanErr
	}()
	<-started
	cancel()
	select {
	case scanErr := <-done:
		if !errors.Is(scanErr, context.Canceled) {
			t.Fatalf("scan error = %v, want context.Canceled", scanErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled scheduler leaked or blocked workers")
	}
	if got := active.Load(); got != 0 {
		t.Fatalf("workers still active after cancellation: %d", got)
	}
	if got := inspected.Load(); got >= int64(len(entries)) {
		t.Fatalf("cancellation did not stop queued work: inspected %d of %d", got, len(entries))
	}
}

func TestGlobalSchedulerPreservesUnreadableEntryMetrics(t *testing.T) {
	root := t.TempDir()
	good := filepath.Join(root, "good.bin")
	denied := filepath.Join(root, "denied.bin")
	for _, path := range []string{good, denied} {
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	operations := defaultScanOperations
	operations.lstat = func(path string) (os.FileInfo, error) {
		if path == denied {
			return nil, os.ErrPermission
		}
		return os.Lstat(path)
	}
	sc := schedulerTestContext(root, 2, operations)
	results, err := sc.scanTree(context.Background(), root, entries)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || sc.entriesSkipped.Load() != 1 || sc.entriesScanned.Load() != 1 {
		t.Fatalf("metrics changed: results=%#v scanned=%d skipped=%d", results, sc.entriesScanned.Load(), sc.entriesSkipped.Load())
	}
}

func TestAnalyzeWithContextHonorsPreCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := AnalyzeWithContext(ctx, "analyze", scannerTestOptions(t.TempDir(), 2), nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func schedulerTestContext(root string, workers int, operations scanOperations) *scanContext {
	return &scanContext{
		opts:           scannerTestOptions(root, workers),
		report:         &Report{},
		candidateByKey: make(map[string]int),
		groupCache:     make(map[string]string),
		includeSet:     make(map[string]struct{}),
		excludeSet:     make(map[string]struct{}),
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		operations:     operations,
	}
}

func BenchmarkAnalyzeManyFiles(b *testing.B) {
	root := b.TempDir()
	for i := 0; i < 16; i++ {
		dir := filepath.Join(root, "repo-"+strconv.Itoa(i), "node_modules")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatal(err)
		}
		for j := 0; j < 32; j++ {
			path := filepath.Join(dir, "dependency-"+strconv.Itoa(j)+".js")
			if err := os.WriteFile(path, []byte("dependency"), 0o644); err != nil {
				b.Fatal(err)
			}
		}
	}
	for _, workers := range []int{1, 8} {
		b.Run("workers="+strconv.Itoa(workers), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := AnalyzeWithOptions("analyze", scannerTestOptions(root, workers), nil); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
