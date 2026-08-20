package reclaimit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunAnalyzeOlderThanFiltersCandidates(t *testing.T) {
	root := t.TempDir()
	stale := filepath.Join(root, "stale", "node_modules")
	active := filepath.Join(root, "active", "node_modules")
	for _, path := range []string{stale, active} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "package.json"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now()
	for _, path := range []string{stale, filepath.Join(stale, "package.json")} {
		if err := os.Chtimes(path, now.Add(-48*time.Hour), now.Add(-48*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"analyze", "--root", root, "--older-than", "24h", "--min-candidate-size", "0"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run returned %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), filepath.Join(root, "stale", "node_modules")) {
		t.Fatalf("stale candidate missing from report: %s", stdout.String())
	}
	if strings.Contains(stdout.String(), filepath.Join(root, "active", "node_modules")) {
		t.Fatalf("active candidate unexpectedly selected: %s", stdout.String())
	}
}
