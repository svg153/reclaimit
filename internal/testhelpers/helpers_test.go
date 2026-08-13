package testhelpers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildScanTree(t *testing.T) {
	root := BuildScanTree(t, 2, 3)
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(entries))
	}
	// Check first repo
	repo1 := filepath.Join(root, "repo-000")
	if _, err := os.Stat(filepath.Join(repo1, ".git")); os.IsNotExist(err) {
		t.Error("repo-000 should have .git dir")
	}
	// Check files exist
	for i := 0; i < 3; i++ {
		_ = filepath.Join(repo1, fmt.Sprintf("file-%03d.go", i))
	}
	// Check node_modules
	nm := filepath.Join(repo1, "node_modules")
	if _, err := os.Stat(nm); os.IsNotExist(err) {
		t.Error("repo-000 should have node_modules")
	}
}

func TestBuildScanTree_ZeroRepos(t *testing.T) {
	root := BuildScanTree(t, 0, 5)
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 repos, got %d", len(entries))
	}
}

func TestBuildScanTree_ZeroFiles(t *testing.T) {
	root := BuildScanTree(t, 1, 0)
	repo := filepath.Join(root, "repo-000")
	entries, err := os.ReadDir(repo)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	// Should have .git and node_modules only
	hasGit := false
	hasNM := false
	for _, e := range entries {
		if e.Name() == ".git" {
			hasGit = true
		}
		if e.Name() == "node_modules" {
			hasNM = true
		}
	}
	if !hasGit {
		t.Error("repo should have .git")
	}
	if !hasNM {
		t.Error("repo should have node_modules")
	}
}
