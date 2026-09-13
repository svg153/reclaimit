package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSelectionManifestRoundTripAndValidation(t *testing.T) {
	root := t.TempDir()
	modified := time.Date(2026, 8, 20, 12, 0, 0, 123, time.UTC)
	candidates := []Candidate{{Path: filepath.Join(root, "node_modules"), Group: root, CategoryKey: "node-modules", Bytes: 42, ModifiedAt: modified, IsDir: true}}
	path := filepath.Join(root, "selection.json")
	if err := WriteSelectionManifest(path, root, candidates, SelectionExclusions{Groups: []string{filepath.Join(root, "vendor")}}); err != nil {
		t.Fatal(err)
	}
	manifest, err := ReadSelectionManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	selected, mismatches, err := ValidateSelectionManifest(manifest, root, candidates)
	if err != nil || len(mismatches) != 0 || len(selected) != 1 {
		t.Fatalf("validation failed: selected=%+v mismatches=%+v err=%v", selected, mismatches, err)
	}
}

func TestSelectionManifestFailsClosed(t *testing.T) {
	root := t.TempDir()
	modified := time.Now().UTC()
	candidate := Candidate{Path: filepath.Join(root, "candidate"), Bytes: 1, ModifiedAt: modified}
	manifest, err := NewSelectionManifest(root, []Candidate{candidate}, SelectionExclusions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := ValidateSelectionManifest(manifest, filepath.Join(root, "other"), []Candidate{candidate}); err == nil {
		t.Fatal("expected root mismatch")
	}
	candidate.Bytes = 2
	selected, mismatches, err := ValidateSelectionManifest(manifest, root, []Candidate{candidate})
	if err != nil || len(selected) != 0 || len(mismatches) != 1 || mismatches[0].Status != "changed" {
		t.Fatalf("expected changed audit entry: selected=%+v mismatches=%+v err=%v", selected, mismatches, err)
	}
}

func TestSelectionManifestRejectsUnsupportedSchema(t *testing.T) {
	root := t.TempDir()
	manifest, err := NewSelectionManifest(root, nil, SelectionExclusions{})
	if err != nil {
		t.Fatal(err)
	}
	manifest.SchemaVersion++
	if _, _, err := ValidateSelectionManifest(manifest, root, nil); err == nil {
		t.Fatal("expected unsupported schema error")
	}
}

func TestReadSelectionManifestRejectsInvalidFiles(t *testing.T) {
	root := t.TempDir()
	for name, tt := range map[string]struct {
		content string
		want    string
	}{
		"invalid-json": {content: "{", want: "decode selection manifest"},
		"missing-root": {content: `{"schema_version":1}`, want: "selection manifest root is required"},
		"bad-schema":   {content: `{"schema_version":2,"root":"/tmp"}`, want: "unsupported selection manifest schema version 2"},
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(root, name+".json")
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := ReadSelectionManifest(path); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("ReadSelectionManifest error = %v, want %q", err, tt.want)
			}
		})
	}

	if _, err := ReadSelectionManifest(filepath.Join(root, "missing.json")); err == nil || !strings.Contains(err.Error(), "read selection manifest") {
		t.Fatalf("expected read selection manifest error, got %v", err)
	}
}

func TestWriteSelectionManifestReportsRootAndWriteErrors(t *testing.T) {
	root := t.TempDir()
	if err := WriteSelectionManifest(filepath.Join(root, "missing", "selection.json"), root, nil, SelectionExclusions{}); err == nil ||
		!strings.Contains(err.Error(), "write selection manifest") {
		t.Fatalf("expected write selection manifest error, got %v", err)
	}
}

func TestSelectionManifestReportsMissingCandidate(t *testing.T) {
	root := t.TempDir()
	candidate := Candidate{Path: filepath.Join(root, "node_modules"), Bytes: 1, ModifiedAt: time.Now().UTC(), IsDir: true}
	manifest, err := NewSelectionManifest(root, []Candidate{candidate}, SelectionExclusions{})
	if err != nil {
		t.Fatal(err)
	}

	selected, mismatches, err := ValidateSelectionManifest(manifest, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 0 || len(mismatches) != 1 || mismatches[0].Status != "missing" {
		t.Fatalf("expected missing mismatch: selected=%+v mismatches=%+v", selected, mismatches)
	}
}
