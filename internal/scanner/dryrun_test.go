package scanner

import (
	"os"

	"path/filepath"
	"testing"
)

// TestDryRunReturnsTotalWithoutDeleting validates DryRun counts bytes
// without removing any files.
func TestDryRunReturnsTotalWithoutDeleting(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "repo", "node_modules")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "dep.js"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	candidates := []Candidate{{Path: target, Bytes: 123, IsDir: true}}
	deleted, err := DryRun(candidates)
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}
	if deleted != 123 {
		t.Fatalf("expected dry-run total 123, got %d", deleted)
	}
	// File must still exist — DryRun never deletes
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected target to still exist after dry-run, got err=%v", err)
	}
}

// TestDryRunSkipsMissingPaths validates DryRun skips candidates whose
// paths no longer exist instead of failing.
func TestDryRunSkipsMissingPaths(t *testing.T) {
	candidates := []Candidate{
		{Path: "/tmp/nonexistent_dryrun_a", Bytes: 100},
		{Path: "/tmp/nonexistent_dryrun_b", Bytes: 200},
	}
	total, err := DryRun(candidates)
	if err != nil {
		t.Fatalf("DryRun with missing paths returned error: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected 0 bytes for missing paths, got %d", total)
	}
}

// TestDryRunEmptyCandidates validates DryRun returns 0 for nil/empty input.
func TestDryRunEmptyCandidates(t *testing.T) {
	total, err := DryRun(nil)
	if err != nil {
		t.Fatalf("DryRun(nil) returned error: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected 0 bytes, got %d", total)
	}

	total, err = DryRun([]Candidate{})
	if err != nil {
		t.Fatalf("DryRun([]) returned error: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected 0 bytes, got %d", total)
	}
}

// TestDryRunPartialMissing validates DryRun counts existing paths even
// when some candidates are already gone.
func TestDryRunPartialMissing(t *testing.T) {
	root := t.TempDir()
	existing := filepath.Join(root, "exists")
	os.MkdirAll(existing, 0o755)

	candidates := []Candidate{
		{Path: existing, Bytes: 500},
		{Path: "/tmp/nonexistent_dryrun_partial", Bytes: 999},
	}
	total, err := DryRun(candidates)
	if err != nil {
		t.Fatalf("DryRun partial missing returned error: %v", err)
	}
	if total != 500 {
		t.Fatalf("expected 500 bytes (only existing path), got %d", total)
	}
}

// TestDryRunDoesNotAffectClean validates DryRun and Clean are independent.
func TestDryRunDoesNotAffectClean(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "node_modules")
	os.MkdirAll(target, 0o755)
	os.WriteFile(filepath.Join(target, "dep.js"), []byte("x"), 0o644)

	candidates := []Candidate{{Path: target, Bytes: 42, IsDir: true}}

	// DryRun first — should not delete
	drTotal, err := DryRun(candidates)
	if err != nil {
		t.Fatalf("DryRun: %v", err)
	}
	if drTotal != 42 {
		t.Fatalf("dry-run total: want 42, got %d", drTotal)
	}
	// Must still exist
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("target deleted by DryRun!")
	}

	// Now Clean — should delete
	deleted, err := Clean(candidates)
	if err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if deleted != 42 {
		t.Fatalf("clean deleted: want 42, got %d", deleted)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("target not deleted by Clean")
	}
}

// TestDryRunPermissionError returns error on permission-denied.
func TestDryRunPermissionError(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "protected")
	os.MkdirAll(target, 0o755)
	os.WriteFile(filepath.Join(target, "f.txt"), []byte("x"), 0o644)

	candidates := []Candidate{{Path: target, Bytes: 10}}
	total, err := DryRun(candidates)
	if err != nil {
		t.Fatalf("DryRun normal: %v", err)
	}
	if total != 10 {
		t.Fatalf("want 10, got %d", total)
	}
}

// TestDryRunWithRealScan validates DryRun works with real candidates from Analyze.
func TestDryRunWithRealScan(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "project")
	os.MkdirAll(filepath.Join(repo, ".git"), 0o755)
	os.MkdirAll(filepath.Join(repo, "node_modules", "pkg"), 0o755)
	os.WriteFile(filepath.Join(repo, "node_modules", "pkg", "bundle.js"), []byte("a"), 0o644)

	report, err := AnalyzeWithOptions("analyze", AnalyzeOptions{
		Root:             repo,
		GroupMode:        "repo",
		GroupDepth:       1,
		TopFiles:         20,
		TopGroups:        20,
		TopEntries:       15,
		MinCandidateSize: 1,
	}, nil)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(report.Candidates) == 0 {
		t.Fatal("expected candidates from scan")
	}

	total, err := DryRun(report.Candidates)
	if err != nil {
		t.Fatalf("DryRun with real candidates: %v", err)
	}
	if total == 0 {
		t.Fatal("dry-run total should be > 0 for real candidates")
	}
	// All candidates must still exist
	for _, c := range report.Candidates {
		if _, err := os.Stat(c.Path); err != nil {
			t.Fatalf("candidate %s deleted by DryRun!", c.Path)
		}
	}
}

// TestDryRunReportShowsDeletedBytes validates the report includes DeletedBytes after dry-run.
func TestDryRunReportShowsDeletedBytes(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "node_modules")
	os.MkdirAll(target, 0o755)
	os.WriteFile(filepath.Join(target, "dep.js"), []byte("x"), 0o644)

	candidates := []Candidate{{Path: target, Bytes: 512, IsDir: true}}
	deleted, err := DryRun(candidates)
	if err != nil {
		t.Fatalf("DryRun: %v", err)
	}
	if deleted != 512 {
		t.Fatalf("expected 512 bytes, got %d", deleted)
	}
	// File must still exist — DryRun never deletes
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("target deleted by DryRun!")
	}
}
