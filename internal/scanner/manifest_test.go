package scanner

import (
	"path/filepath"
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
