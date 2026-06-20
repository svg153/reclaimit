package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetermineGroup_RepoMode(t *testing.T) {
	opts := AnalyzeOptions{GroupMode: "repo", GroupDepth: 1}
	cache := make(map[string]string)
	// Create a temp dir with a .git to simulate a repo root
	tmp, err := os.MkdirTemp("", "group-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	gitPath := filepath.Join(tmp, ".git")
	if err := os.MkdirAll(gitPath, 0o755); err != nil {
		t.Fatal(err)
	}

	result := DetermineGroup(filepath.Join(tmp, "src", "main.go"), opts, cache)
	if result != tmp {
		t.Errorf("expected repo root %s, got %s", tmp, result)
	}
}

func TestDetermineGroup_RepoMode_NoGit(t *testing.T) {
	opts := AnalyzeOptions{GroupMode: "repo", GroupDepth: 1}
	cache := make(map[string]string)

	tmp, err := os.MkdirTemp("", "group-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	result := DetermineGroup(filepath.Join(tmp, "src", "main.go"), opts, cache)
	if result != "" {
		t.Errorf("expected empty group for non-git path, got %s", result)
	}
}

func TestDetermineGroup_DepthMode(t *testing.T) {
	opts := AnalyzeOptions{GroupMode: "depth", GroupDepth: 2, Root: "/root"}
	cache := make(map[string]string)

	result := DetermineGroup("/root/a/b/c/file.txt", opts, cache)
	expected := "/root/a/b"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestDetermineGroup_DepthMode_ShortPath(t *testing.T) {
	opts := AnalyzeOptions{GroupMode: "depth", GroupDepth: 5, Root: "/root"}
	cache := make(map[string]string)

	// Path shorter than depth - the code uses min(len(parts), depth)
	// relative = "a/file.txt" -> parts = [a, file.txt] (2 parts)
	// depth = min(2, 5) = 2, groupParts = [a, file.txt]
	// result = /root/a/file.txt
	result := DetermineGroup("/root/a/file.txt", opts, cache)
	if result != "/root/a/file.txt" {
		t.Errorf("expected /root/a/file.txt, got %s", result)
	}
}

func TestFindRepoRoot_CacheHit(t *testing.T) {
	cache := map[string]string{"/repo": "/repo"}
	result := findRepoRoot("/repo/src/main.go", "/repo", cache)
	if result != "/repo" {
		t.Errorf("expected /repo from cache, got %s", result)
	}
}

func TestFindRepoRoot_CacheMiss_GitFound(t *testing.T) {
	tmp, err := os.MkdirTemp("", "repo-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	gitPath := filepath.Join(tmp, ".git")
	if err := os.MkdirAll(gitPath, 0o755); err != nil {
		t.Fatal(err)
	}

	cache := make(map[string]string)
	result := findRepoRoot(filepath.Join(tmp, "src", "main.go"), tmp, cache)
	if result != tmp {
		t.Errorf("expected %s, got %s", tmp, result)
	}
	// Should be cached
	if _, ok := cache[tmp]; !ok {
		t.Error("expected result to be cached")
	}
}

func TestFindRepoRoot_NoGit_ReturnsEmpty(t *testing.T) {
	tmp, err := os.MkdirTemp("", "no-git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	cache := make(map[string]string)
	result := findRepoRoot(filepath.Join(tmp, "file.txt"), tmp, cache)
	if result != "" {
		t.Errorf("expected empty, got %s", result)
	}
}

func TestFindRepoRoot_AtRoot_ReturnsEmpty(t *testing.T) {
	tmp, err := os.MkdirTemp("", "root-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	cache := make(map[string]string)
	result := findRepoRoot(tmp, tmp, cache)
	if result != "" {
		t.Errorf("expected empty at root, got %s", result)
	}
}

func TestAncestorGroup_DefaultDepth(t *testing.T) {
	result := ancestorGroup("/root/a/file.txt", "/root", 1)
	if result != "/root/a" {
		t.Errorf("expected /root/a, got %s", result)
	}
}

func TestAncestorGroup_CustomDepth(t *testing.T) {
	result := ancestorGroup("/root/a/b/c/file.txt", "/root", 2)
	if result != "/root/a/b" {
		t.Errorf("expected /root/a/b, got %s", result)
	}
}

func TestAncestorGroup_SameAsRoot(t *testing.T) {
	// relative = "file.txt", parts = [file.txt], depth = 1
	// groupParts = [file.txt], result = /root/file.txt
	result := ancestorGroup("/root/file.txt", "/root", 1)
	if result != "/root/file.txt" {
		t.Errorf("expected /root/file.txt, got %s", result)
	}
}

func TestAncestorGroup_RelativeError(t *testing.T) {
	// Path not under root -> Rel returns error -> returns root
	result := ancestorGroup("/other/file.txt", "/root", 1)
	if result != "/" {
		t.Errorf("expected / for out-of-tree path, got %s", result)
	}
}
