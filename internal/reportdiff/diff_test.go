package reportdiff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareAddedRemovedChangedAndUnchanged(t *testing.T) {
	before := report{SchemaVersion: SchemaVersion, SelectedBytes: 30, SelectedCandidates: []candidate{
		{CategoryKey: "node-modules", Path: "/old", Bytes: 10, ModifiedAt: "2026-01-01T00:00:00Z", IsDir: true},
		{CategoryKey: "python-venv", Path: "/changed", Bytes: 10, ModifiedAt: "2026-01-01T00:00:00Z", IsDir: true},
		{CategoryKey: "rust-target", Path: "/same", Bytes: 10, ModifiedAt: "2026-01-01T00:00:00Z", IsDir: true},
	}}
	after := report{SchemaVersion: SchemaVersion, SelectedBytes: 45, SelectedCandidates: []candidate{
		{CategoryKey: "python-venv", Path: "/changed", Bytes: 20, ModifiedAt: "2026-01-02T00:00:00Z", IsDir: true},
		{CategoryKey: "rust-target", Path: "/same", Bytes: 10, ModifiedAt: "2026-01-01T00:00:00Z", IsDir: true},
		{CategoryKey: "gradle-cache", Path: "/new", Bytes: 15, ModifiedAt: "2026-01-02T00:00:00Z", IsDir: true},
	}}

	result := compare(before, after)
	if result.Added != 1 || result.Removed != 1 || result.Changed != 1 || result.Unchanged != 1 {
		t.Fatalf("unexpected candidate changes: %+v", result)
	}
	if result.AfterBytes-result.BeforeBytes != 15 || len(result.Categories) != 4 {
		t.Fatalf("unexpected totals: %+v", result)
	}
	if len(result.AddedPaths) != 1 || result.AddedPaths[0] != "/new" || len(result.RemovedPaths) != 1 || result.RemovedPaths[0] != "/old" {
		t.Fatalf("unexpected candidate paths: %+v", result)
	}
	output := Render(result)
	if !strings.Contains(output, `added    "/new"`) || !strings.Contains(output, `removed  "/old"`) || !strings.Contains(output, `changed  "/changed"`) {
		t.Fatalf("unexpected rendered candidate changes: %s", output)
	}
}

func TestCompareAnonymousReportsOmitsCandidateChanges(t *testing.T) {
	result := compare(
		report{SelectedCandidates: []candidate{{CategoryKey: "node-modules", Path: "<redacted>", Bytes: 10}}},
		report{SelectedCandidates: []candidate{{CategoryKey: "node-modules", Path: "<redacted>", Bytes: 20}}},
	)
	if !result.PathsRedacted || !strings.Contains(Render(result), "omitted") {
		t.Fatalf("expected anonymous candidate changes to be omitted: %+v", result)
	}
}

func TestCompareFilesRejectsUnsupportedOrMalformedReports(t *testing.T) {
	dir := t.TempDir()
	valid := filepath.Join(dir, "valid.json")
	unsupported := filepath.Join(dir, "unsupported.json")
	malformed := filepath.Join(dir, "malformed.json")
	if err := os.WriteFile(valid, []byte(`{"schema_version":1,"selected_bytes":0,"selected_candidates":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(unsupported, []byte(`{"schema_version":2}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(malformed, []byte(`{"schema_version":1} {}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := CompareFiles(unsupported, valid); err == nil || !strings.Contains(err.Error(), "unsupported report schema") {
		t.Fatalf("expected schema error, got %v", err)
	}
	if _, err := CompareFiles(valid, malformed); err == nil || !strings.Contains(err.Error(), "one JSON document") {
		t.Fatalf("expected trailing document error, got %v", err)
	}
}

func TestCompareFilesRejectsMissingAndOversizedReports(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.json")
	if _, err := CompareFiles(missing, missing); err == nil || !strings.Contains(err.Error(), "open report") {
		t.Fatalf("expected open error, got %v", err)
	}

	oversized := filepath.Join(dir, "oversized.json")
	file, err := os.Create(oversized)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(maxReportBytes + 1); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := CompareFiles(oversized, oversized); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected size error, got %v", err)
	}
}

func TestCompareFilesRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := CompareFiles(path, path); err == nil || !strings.Contains(err.Error(), "decode report") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestCompareFilesReadsValidReports(t *testing.T) {
	dir := t.TempDir()
	before := filepath.Join(dir, "before.json")
	after := filepath.Join(dir, "after.json")
	content := []byte(`{"schema_version":1,"selected_bytes":7,"selected_candidates":[]}`)
	if err := os.WriteFile(before, content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(after, content, 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := CompareFiles(before, after)
	if err != nil {
		t.Fatal(err)
	}
	if result.BeforeBytes != 7 || result.AfterBytes != 7 {
		t.Fatalf("unexpected result: %+v", result)
	}
}
