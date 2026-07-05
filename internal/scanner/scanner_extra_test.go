package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func analyzeConfig(root string) AnalyzeOptions {
	return AnalyzeOptions{
		Root:             root,
		GroupMode:        "repo",
		GroupDepth:       1,
		TopFiles:         20,
		TopGroups:        20,
		TopEntries:       15,
		MinCandidateSize: 1,
	}
}

func TestAnalyzeSkipsPermissionDeniedDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permission semantics")
	}
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}
	root := t.TempDir()
	repo := filepath.Join(root, "project")
	mustMkdir(t, filepath.Join(repo, ".git"))
	mustMkdir(t, filepath.Join(repo, "node_modules"))
	mustWriteFile(t, filepath.Join(repo, "node_modules", "dep.js"), strings.Repeat("a", 2<<20))

	locked := filepath.Join(root, "locked")
	mustMkdir(t, locked)
	mustWriteFile(t, filepath.Join(locked, "secret.bin"), strings.Repeat("z", 1<<20))
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

	report, err := AnalyzeWithOptions("analyze", analyzeConfig(root), nil)
	if err != nil {
		t.Fatalf("Analyze should skip unreadable dirs, got error: %v", err)
	}
	if len(report.Candidates) != 1 {
		t.Fatalf("expected 1 candidate from the readable tree, got %d", len(report.Candidates))
	}
}

func TestAnalyzeIgnoresSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is unreliable on Windows CI")
	}
	root := t.TempDir()
	repo := filepath.Join(root, "project")
	mustMkdir(t, filepath.Join(repo, ".git"))
	mustMkdir(t, filepath.Join(repo, "node_modules", "pkg"))
	mustWriteFile(t, filepath.Join(repo, "node_modules", "pkg", "bundle.js"), strings.Repeat("a", 2<<20))

	// A symlink that both mirrors a candidate and forms a cycle back to the repo.
	if err := os.Symlink(filepath.Join(repo, "node_modules"), filepath.Join(repo, "mirror")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	if err := os.Symlink(repo, filepath.Join(repo, "self")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	report, err := AnalyzeWithOptions("analyze", analyzeConfig(root), nil)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if len(report.Candidates) != 1 {
		t.Fatalf("expected symlinks to be skipped (1 candidate), got %d", len(report.Candidates))
	}
}

func TestExcludeGroupAndPathCombined(t *testing.T) {
	root := t.TempDir()
	repoA := filepath.Join(root, "repo-a")
	repoB := filepath.Join(root, "repo-b")
	repoC := filepath.Join(root, "repo-c")
	for _, repo := range []string{repoA, repoB, repoC} {
		mustMkdir(t, filepath.Join(repo, ".git"))
		mustMkdir(t, filepath.Join(repo, "node_modules"))
		mustWriteFile(t, filepath.Join(repo, "node_modules", "dep.js"), strings.Repeat("x", 2<<20))
	}

	cfg := analyzeConfig(root)
	cfg.ExcludeGroups = []string{repoA}
	cfg.ExcludePaths = []string{filepath.Join(repoB, "node_modules")}

	report, err := AnalyzeWithOptions("analyze", cfg, nil)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if len(report.Candidates) != 3 {
		t.Fatalf("expected 3 detected candidates, got %d", len(report.Candidates))
	}
	if len(report.SelectedCandidates) != 1 {
		t.Fatalf("expected exclusions to leave 1 selected candidate, got %d", len(report.SelectedCandidates))
	}
	if report.SelectedCandidates[0].Group != repoC {
		t.Fatalf("expected repo-c to survive exclusions, got %s", report.SelectedCandidates[0].Group)
	}
}

func TestAnalyzeFindsBunCache(t *testing.T) {
	root := t.TempDir()

	// Bun global cache: ~/.bun/install/cache (matching .bun directory name)
	bunGlobal := filepath.Join(root, ".bun", "install", "cache")
	mustMkdir(t, bunGlobal)
	mustWriteFile(t, filepath.Join(bunGlobal, "cache.json"), strings.Repeat("x", 1024))

	// Scan Global Cache (root is scan target)
	report, err := AnalyzeWithOptions("analyze", analyzeConfig(root), nil)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	foundGlobal := false
	for _, c := range report.Candidates {
		if c.CategoryKey == "bun-cache" && filepath.Base(c.Path) == ".bun" {
			foundGlobal = true
			if c.Bytes != 1024 {
				t.Errorf("expected global cache candidate size 1024, got %d", c.Bytes)
			}
		}
	}
	if !foundGlobal {
		t.Error("expected to find Bun global cache (.bun) as candidate")
	}
}

func TestAnalyzeFindsMacOSCandidates(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".git"), 0o755)

	// 1. .DS_Store file candidate
	mustWriteFile(t, filepath.Join(root, ".DS_Store"), strings.Repeat("d", 1024))

	// 2. .Spotlight-V100 directory candidate
	spotlightDir := filepath.Join(root, ".Spotlight-V100")
	mustMkdir(t, spotlightDir)
	mustWriteFile(t, filepath.Join(spotlightDir, "store.db"), strings.Repeat("s", 2048))

	// 3. .Trashes directory candidate
	trashesDir := filepath.Join(root, ".Trashes")
	mustMkdir(t, trashesDir)
	mustWriteFile(t, filepath.Join(trashesDir, "trash_file"), strings.Repeat("t", 4096))

	report, err := AnalyzeWithOptions("analyze", analyzeConfig(root), nil)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	var foundDSStore, foundSpotlight, foundTrashes bool
	for _, c := range report.Candidates {
		switch c.CategoryKey {
		case "ds-store":
			if filepath.Base(c.Path) == ".DS_Store" {
				foundDSStore = true
				if c.Bytes != 1024 {
					t.Errorf("expected .DS_Store size 1024, got %d", c.Bytes)
				}
			}
		case "spotlight-index":
			if filepath.Base(c.Path) == ".Spotlight-V100" {
				foundSpotlight = true
				if c.Bytes != 2048 {
					t.Errorf("expected .Spotlight-V100 size 2048, got %d", c.Bytes)
				}
			}
		case "macos-trash":
			if filepath.Base(c.Path) == ".Trashes" {
				foundTrashes = true
				if c.Bytes != 4096 {
					t.Errorf("expected .Trashes size 4096, got %d", c.Bytes)
				}
			}
		}
	}

	if !foundDSStore {
		t.Error("expected to find .DS_Store file as candidate")
	}
	if !foundSpotlight {
		t.Error("expected to find .Spotlight-V100 directory as candidate")
	}
	if !foundTrashes {
		t.Error("expected to find .Trashes directory as candidate")
	}
}
