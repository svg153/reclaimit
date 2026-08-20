package reclaimit

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelectionManifestImportReportsChangedCandidate(t *testing.T) {
	root := t.TempDir()
	candidate := filepath.Join(root, "node_modules")
	if err := os.MkdirAll(candidate, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(candidate, "package.json")
	if err := os.WriteFile(file, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, "selection.json")
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"analyze", "--root", root, "--min-candidate-size", "0", "--export-selection", manifestPath}, &stdout, &stderr); code != 0 {
		t.Fatalf("export returned %d: %s", code, stderr.String())
	}
	if err := os.WriteFile(file, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"analyze", "--root", root, "--min-candidate-size", "0", "--import-selection", manifestPath, "--format", "json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("import returned %d: %s", code, stderr.String())
	}
	var report struct {
		SelectionMismatches []struct {
			Status string `json:"status"`
		} `json:"selection_mismatches"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.SelectionMismatches) != 1 || report.SelectionMismatches[0].Status != "changed" {
		t.Fatalf("expected changed audit entry, got %s", stdout.String())
	}
	if strings.Contains(stdout.String(), "cleaned") {
		t.Fatal("import must remain read-only")
	}
}
