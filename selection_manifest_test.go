package reclaimit

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/svg153/reclaimit/internal/scanner"
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

func TestCleanupPlanCLIRequiresReviewAndFailsClosed(t *testing.T) {
	root := t.TempDir()
	candidate := filepath.Join(root, "node_modules")
	if err := os.MkdirAll(candidate, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(candidate, "package.json"), []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(root, "cleanup-plan.json")
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"analyze", "--root", root, "--min-candidate-size", "0", "--export-plan", planPath}, &stdout, &stderr); code != 0 {
		t.Fatalf("plan export returned %d: %s", code, stderr.String())
	}
	var plan scanner.CleanupPlan
	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &plan); err != nil || plan.Action != "delete" || len(plan.Candidates) != 1 {
		t.Fatalf("unexpected cleanup plan: err=%v plan=%+v", err, plan)
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"clean", "--root", root, "--min-candidate-size", "0", "--plan", planPath, "--dry-run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("plan dry-run returned %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Cleanup plan validated") || strings.Contains(stdout.String(), "[CLEAN]") {
		t.Fatalf("dry-run did not describe a non-destructive plan: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"clean", "--root", root, "--min-candidate-size", "0", "--plan", planPath, "--yes"}, &stdout, &stderr); code != 0 {
		t.Fatalf("plan apply returned %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "[CLEAN]") {
		t.Fatalf("plan apply did not report cleanup: %s", stdout.String())
	}
	if _, err := os.Stat(candidate); !os.IsNotExist(err) {
		t.Fatalf("plan apply should remove candidate, stat error=%v", err)
	}

	if err := os.MkdirAll(candidate, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(candidate, "package.json"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"clean", "--root", root, "--min-candidate-size", "0", "--plan", planPath, "--yes"}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "cleanup plan validation failed") {
		t.Fatalf("changed plan should fail closed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}
