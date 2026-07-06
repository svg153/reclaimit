package scanner

import (
	"fmt"

	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildScanTree creates a deterministic synthetic workspace with repos, nested
// dependency directories, and large files so benchmarks exercise a realistic walk.
func buildScanTree(tb testing.TB, repos, filesPerRepo int) string {
	tb.Helper()
	root := tb.TempDir()
	payload := strings.Repeat("a", 4096)
	for r := 0; r < repos; r++ {
		repo := filepath.Join(root, fmt.Sprintf("repo-%03d", r))
		if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
			tb.Fatalf("mkdir: %v", err)
		}
		nm := filepath.Join(repo, "node_modules", "pkg")
		if err := os.MkdirAll(nm, 0o755); err != nil {
			tb.Fatalf("mkdir: %v", err)
		}
		for f := 0; f < filesPerRepo; f++ {
			file := filepath.Join(nm, fmt.Sprintf("chunk-%04d.js", f))
			if err := os.WriteFile(file, []byte(payload), 0o644); err != nil {
				tb.Fatalf("write: %v", err)
			}
		}
	}
	return root
}

func BenchmarkAnalyzeWithOptions(b *testing.B) {
	root := buildScanTree(b, 40, 25)
	cfg := analyzeConfig(root)
	cfg.MinCandidateSize = 1
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := AnalyzeWithOptions("analyze", cfg, nil); err != nil {
			b.Fatalf("AnalyzeWithOptions: %v", err)
		}
	}
}

func BenchmarkPushTop(b *testing.B) {
	const limit = 20
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var top []PathSize
		for j := 0; j < 10000; j++ {
			top = PushTop(top, PathSize{Path: "p", Bytes: int64(j % 997)}, limit)
		}
		if len(top) > limit {
			b.Fatalf("PushTop exceeded limit: %d", len(top))
		}
	}
}

func BenchmarkSummarizeGroups(b *testing.B) {
	candidates := make([]Candidate, 0, 5000)
	for i := 0; i < 5000; i++ {
		candidates = append(candidates, Candidate{
			Group:       fmt.Sprintf("/tmp/group-%d", i%250),
			CategoryKey: "node-modules",
			Bytes:       int64(i),
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := SummarizeGroups(candidates, 20); len(got) == 0 {
			b.Fatal("expected summaries")
		}
	}
}

// FuzzMatchFile ensures the file matcher never panics on arbitrary paths.
func FuzzMatchFile(f *testing.F) {
	for _, seed := range []string{"", "main.pyc", "/tmp/a.PYO", "weird..name.", "nested/dir/file.js"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, path string) {
		_, _ = MatchFile(path)
	})
}

// FuzzMatchDirectory ensures the directory matcher never panics on arbitrary names.
func FuzzMatchDirectory(f *testing.F) {
	for _, seed := range []string{"", "node_modules", ".venv", "__pycache__", "random-name"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, name string) {
		_, _ = MatchDirectory(name)
	})
}

// FuzzHumanizeBytes ensures byte formatting is panic-free across the int64 range.
func FuzzHumanizeBytes(f *testing.F) {
	for _, seed := range []int64{0, 1, 1023, 1024, 1 << 30, -1} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, size int64) {
		if got := humanizeBytes(size); got == "" {
			t.Fatalf("humanizeBytes(%d) returned empty string", size)
		}
	})
}


