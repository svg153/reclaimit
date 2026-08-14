package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanRemovesCandidates(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "repo", "node_modules")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "dep.js"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	deleted, err := Clean([]Candidate{{Path: target, Bytes: 1, IsDir: true}})
	if err != nil {
		t.Fatalf("Clean returned error: %v", err)
	}
	if deleted != 123 {
		t.Fatalf("expected deleted bytes 1, got %d", deleted)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected target to be deleted, stat err=%v", err)
	}
}


func TestCleanSkipsChangedCandidateAndContinues(t *testing.T) {
	root := t.TempDir()
	changed := filepath.Join(root, "changed.bin")
	unchanged := filepath.Join(root, "unchanged.bin")
	if err := os.WriteFile(changed, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(unchanged, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(changed, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := CleanWithOptions([]Candidate{
		{Path: changed, Bytes: 3},
		{Path: unchanged, Bytes: 1},
	}, CleanOptions{})
	if err != nil {
		t.Fatalf("CleanWithOptions returned error: %v", err)
	}
	if result.SkippedCandidates != 1 || result.DeletedCandidates != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.DeletedBytes != 1 {
		t.Fatalf("DeletedBytes = %d, want 1", result.DeletedBytes)
	}
	if _, err := os.Stat(changed); err != nil {
		t.Fatalf("changed candidate should remain: %v", err)
	}
	if _, err := os.Stat(unchanged); !os.IsNotExist(err) {
		t.Fatalf("unchanged candidate should be deleted, stat err=%v", err)
	}
}

func TestCleanCollapsesNestedCandidates(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "node_modules")
	child := filepath.Join(parent, "package")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(child, "dep.js"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := CleanWithOptions([]Candidate{
		{Path: parent, Bytes: 1, IsDir: true},
		{Path: child, Bytes: 1, IsDir: true},
	}, CleanOptions{})
	if err != nil {
		t.Fatalf("CleanWithOptions returned error: %v", err)
	}
	if result.Candidates != 1 || result.DeletedCandidates != 1 || result.DeletedBytes != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestCleanDryRunDoesNotDeleteVerifiedCandidates(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "node_modules")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "dep.js"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := CleanWithOptions([]Candidate{{Path: target, Bytes: 1, IsDir: true}}, CleanOptions{DryRun: true})
	if err != nil {
		t.Fatalf("CleanWithOptions dry-run returned error: %v", err)
	}
	if result.VerifiedBytes != 1 || result.DeletedBytes != 0 {
		t.Fatalf("unexpected dry-run result: %#v", result)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("dry-run deleted target: %v", err)
	}
}

func TestCleanSkipsMissingAndTypeChangedCandidates(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := CleanWithOptions([]Candidate{
		{Path: filepath.Join(root, "missing"), Bytes: 1},
		{Path: file, Bytes: 1, IsDir: true},
	}, CleanOptions{})
	if err != nil {
		t.Fatalf("CleanWithOptions returned error: %v", err)
	}
	if result.SkippedCandidates != 2 || result.DeletedCandidates != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatalf("type-changed candidate should remain: %v", err)
	}
}
