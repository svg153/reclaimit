package main

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

	deleted, err := Clean([]Candidate{{Path: target, Bytes: 123, IsDir: true}})
	if err != nil {
		t.Fatalf("Clean returned error: %v", err)
	}
	if deleted != 123 {
		t.Fatalf("expected deleted bytes 123, got %d", deleted)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected target to be deleted, stat err=%v", err)
	}
}
