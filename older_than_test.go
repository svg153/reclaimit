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
	code := Run([]string{"analyze", "--root", root, "--older-than", "24h", "--min-candidate-size", "0", "--format", "json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run returned %d: %s", code, stderr.String())
	}
	selectionStart := strings.Index(stdout.String(), "\"selected_candidates\"")
	if selectionStart < 0 {
		t.Fatalf("selected_candidates missing from report: %s", stdout.String())
	}
	selection := stdout.String()[selectionStart:]
	if !strings.Contains(selection, filepath.Join(root, "stale", "node_modules")) {
		t.Fatalf("stale candidate missing from selected candidates: %s", stdout.String())
	}
	if strings.Contains(selection, filepath.Join(root, "active", "node_modules")) {
		t.Fatalf("active candidate unexpectedly selected: %s", stdout.String())
	}
}
