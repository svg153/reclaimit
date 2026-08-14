package scanner

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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
	if deleted != 1 {
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

func TestCleanQuarantinesReplacementInsteadOfDeletingIt(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "candidate.bin")
	stashed := filepath.Join(root, "candidate-before-swap.bin")
	if err := os.WriteFile(target, []byte("safe"), 0o644); err != nil {
		t.Fatal(err)
	}

	operations := defaultCleanOperations
	operations.rename = func(oldPath, quarantinePath string) error {
		if err := os.Rename(oldPath, stashed); err != nil {
			return err
		}
		if err := os.WriteFile(oldPath, []byte("replacement"), 0o644); err != nil {
			return err
		}
		return os.Rename(oldPath, quarantinePath)
	}

	result, err := cleanWithOperations([]Candidate{{Path: target, Bytes: 4}}, CleanOptions{}, operations)
	if err != nil {
		t.Fatal(err)
	}
	if result.DeletedCandidates != 0 || result.FailedCandidates != 1 || len(result.Issues) != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	issue := result.Issues[0]
	if issue.QuarantinePath == "" || !strings.Contains(issue.Reason, "identity changed") {
		t.Fatalf("replacement quarantine was not reported: %#v", issue)
	}
	data, err := os.ReadFile(issue.QuarantinePath)
	if err != nil {
		t.Fatalf("replacement was not preserved: %v", err)
	}
	if string(data) != "replacement" {
		t.Fatalf("quarantined content = %q", data)
	}
	if data, err := os.ReadFile(stashed); err != nil || string(data) != "safe" {
		t.Fatalf("original object was lost: data=%q err=%v", data, err)
	}
}

func TestCleanPreservesCandidateChangedInsideQuarantine(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "node_modules")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "dep.js"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	operations := defaultCleanOperations
	operations.rename = func(oldPath, quarantinePath string) error {
		if err := os.Rename(oldPath, quarantinePath); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(quarantinePath, "new.js"), []byte("changed"), 0o644)
	}
	result, err := cleanWithOperations([]Candidate{{Path: target, Bytes: 1, IsDir: true}}, CleanOptions{}, operations)
	if err != nil {
		t.Fatal(err)
	}
	if result.FailedCandidates != 1 || result.DeletedCandidates != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
	issue := result.Issues[0]
	if issue.QuarantinePath == "" || !strings.Contains(issue.Reason, "contents changed") {
		t.Fatalf("changed quarantine was not reported: %#v", issue)
	}
	if _, err := os.Stat(filepath.Join(issue.QuarantinePath, "new.js")); err != nil {
		t.Fatalf("changed data was deleted: %v", err)
	}
}

func TestCleanReportsQuarantineFailureAndContinues(t *testing.T) {
	root := t.TempDir()
	blocked := filepath.Join(root, "blocked.bin")
	deletable := filepath.Join(root, "deletable.bin")
	for _, path := range []string{blocked, deletable} {
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	operations := defaultCleanOperations
	operations.rename = func(oldPath, quarantinePath string) error {
		if oldPath == blocked {
			return errors.New("rename denied")
		}
		return os.Rename(oldPath, quarantinePath)
	}
	result, err := cleanWithOperations([]Candidate{{Path: blocked, Bytes: 1}, {Path: deletable, Bytes: 1}}, CleanOptions{}, operations)
	if err != nil {
		t.Fatal(err)
	}
	if result.FailedCandidates != 1 || result.DeletedCandidates != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if _, err := os.Stat(blocked); err != nil {
		t.Fatalf("failed candidate should remain: %v", err)
	}
	if _, err := os.Stat(deletable); !os.IsNotExist(err) {
		t.Fatalf("remaining candidate was not deleted: %v", err)
	}
}

func TestCleanReportsRecoverableDeleteFailure(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "candidate.bin")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	operations := defaultCleanOperations
	operations.removeAll = func(string) error { return errors.New("delete denied") }
	result, err := cleanWithOperations([]Candidate{{Path: target, Bytes: 1}}, CleanOptions{}, operations)
	if err != nil {
		t.Fatal(err)
	}
	issue := result.Issues[0]
	if result.FailedCandidates != 1 || issue.QuarantinePath == "" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if _, err := os.Stat(issue.QuarantinePath); err != nil {
		t.Fatalf("failed deletion should remain recoverable: %v", err)
	}
}

func TestCleanSkipsSymbolicLinks(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.bin")
	link := filepath.Join(root, "candidate-link")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symbolic links are unavailable: %v", err)
	}
	result, err := CleanWithOptions([]Candidate{{Path: link}}, CleanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.SkippedCandidates != 1 || !strings.Contains(result.Issues[0].Reason, "symbolic links") {
		t.Fatalf("unexpected result: %#v", result)
	}
	if _, err := os.Lstat(link); err != nil {
		t.Fatalf("symbolic link was removed: %v", err)
	}
}

func TestCleanProtectsFilesystemRootsAndEmptyPaths(t *testing.T) {
	for _, path := range []string{"", ".", string(filepath.Separator)} {
		_, _, issue := verifyCandidate(Candidate{Path: path})
		if issue == nil || issue.Status != "skipped" || !strings.Contains(issue.Reason, "protected") {
			t.Fatalf("path %q was not protected: %#v", path, issue)
		}
	}
}

func TestCleanReportsEmptyQuarantineCleanupWarning(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "candidate.bin")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	operations := defaultCleanOperations
	operations.remove = func(string) error { return errors.New("directory busy") }
	result, err := cleanWithOperations([]Candidate{{Path: target, Bytes: 1}}, CleanOptions{}, operations)
	if err != nil {
		t.Fatal(err)
	}
	if result.DeletedCandidates != 1 || result.FailedCandidates != 0 || len(result.Issues) != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.Issues[0].Status != "warning" || result.Issues[0].QuarantinePath == "" {
		t.Fatalf("cleanup warning was not reported: %#v", result.Issues[0])
	}
}
