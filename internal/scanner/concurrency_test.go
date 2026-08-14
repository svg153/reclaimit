package scanner

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := AnalyzeWithOptions("analyze", scannerTestOptions(root, 8), nil); err != nil {
			b.Fatal(err)
		}
	}
}
