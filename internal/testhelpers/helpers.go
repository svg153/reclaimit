package testhelpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func BuildScanTree(tb testing.TB, repos, filesPerRepo int) string {
	tb.Helper()
	root := tb.TempDir()
	payload := strings.Repeat("a", 4096)
	for r := 0; r < repos; r++ {
		repo := filepath.Join(root, fmt.Sprintf("repo-%03d", r))
		if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
			tb.Fatalf("mkdir: %v", err)
		}
		for f := 0; f < filesPerRepo; f++ {
			file := filepath.Join(repo, fmt.Sprintf("file-%03d.go", f))
			if err := os.WriteFile(file, []byte(payload), 0o644); err != nil {
				tb.Fatalf("write: %v", err)
			}
		}
		if err := os.MkdirAll(filepath.Join(repo, "node_modules"), 0o755); err != nil {
			tb.Fatalf("mkdir node_modules: %v", err)
		}
		for m := 0; m < 3; m++ {
			lib := filepath.Join(repo, "node_modules", fmt.Sprintf("lib-%03d", m))
			if err := os.WriteFile(lib, []byte(payload), 0o644); err != nil {
				tb.Fatalf("write lib: %v", err)
			}
		}
	}
	return root
}
