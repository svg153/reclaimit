package scanner

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const SelectionManifestSchemaVersion = 1

type SelectionManifest struct {
	SchemaVersion int                 `json:"schema_version"`
	Root          string              `json:"root"`
	CreatedAt     time.Time           `json:"created_at"`
	Exclusions    SelectionExclusions `json:"exclusions"`
	Candidates    []ManifestCandidate `json:"candidates"`
}

type SelectionExclusions struct {
	Categories []string `json:"categories,omitempty"`
	Groups     []string `json:"groups,omitempty"`
	Paths      []string `json:"paths,omitempty"`
}

type ManifestCandidate struct {
	Path        string    `json:"path"`
	Group       string    `json:"group"`
	CategoryKey string    `json:"category_key"`
	Bytes       int64     `json:"bytes"`
	ModifiedAt  time.Time `json:"modified_at"`
	IsDir       bool      `json:"is_dir"`
}

type SelectionMismatch struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func WriteSelectionManifest(path string, root string, candidates []Candidate, exclusions SelectionExclusions) error {
	manifest, err := NewSelectionManifest(root, candidates, exclusions)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode selection manifest: %w", err)
	}
	if err := os.WriteFile(path, append(data, 10), 0o600); err != nil {
		return fmt.Errorf("write selection manifest: %w", err)
	}
	return nil
}

func ReadSelectionManifest(path string) (SelectionManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SelectionManifest{}, fmt.Errorf("read selection manifest: %w", err)
	}
	var manifest SelectionManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return SelectionManifest{}, fmt.Errorf("decode selection manifest: %w", err)
	}
	if manifest.SchemaVersion != SelectionManifestSchemaVersion {
		return SelectionManifest{}, fmt.Errorf("unsupported selection manifest schema version %d", manifest.SchemaVersion)
	}
	if manifest.Root == "" {
		return SelectionManifest{}, errors.New("selection manifest root is required")
	}
	return manifest, nil
}

func NewSelectionManifest(root string, candidates []Candidate, exclusions SelectionExclusions) (SelectionManifest, error) {
	canonicalRoot, err := canonicalPath(root)
	if err != nil {
		return SelectionManifest{}, err
	}
	entries := make([]ManifestCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		entries = append(entries, ManifestCandidate{
			Path: candidate.Path, Group: candidate.Group, CategoryKey: candidate.CategoryKey,
			Bytes: candidate.Bytes, ModifiedAt: candidate.ModifiedAt, IsDir: candidate.IsDir,
		})
	}
	return SelectionManifest{
		SchemaVersion: SelectionManifestSchemaVersion,
		Root: canonicalRoot,
		CreatedAt: time.Now().UTC(),
		Exclusions: exclusions,
		Candidates: entries,
	}, nil
}

func ValidateSelectionManifest(manifest SelectionManifest, root string, current []Candidate) ([]Candidate, []SelectionMismatch, error) {
	canonicalRoot, err := canonicalPath(root)
	if err != nil {
		return nil, nil, err
	}
	manifestRoot, err := canonicalPath(manifest.Root)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid selection manifest root: %w", err)
	}
	if manifest.SchemaVersion != SelectionManifestSchemaVersion {
		return nil, nil, fmt.Errorf("unsupported selection manifest schema version %d", manifest.SchemaVersion)
	}
	if canonicalRoot != manifestRoot {
		return nil, nil, fmt.Errorf("selection manifest root %q does not match scan root %q", manifest.Root, canonicalRoot)
	}
	byPath := make(map[string]Candidate, len(current))
	for _, candidate := range current {
		byPath[filepath.Clean(candidate.Path)] = candidate
	}
	selected := make([]Candidate, 0, len(manifest.Candidates))
	mismatches := make([]SelectionMismatch, 0)
	for _, expected := range manifest.Candidates {
		currentCandidate, ok := byPath[filepath.Clean(expected.Path)]
		if !ok {
			mismatches = append(mismatches, SelectionMismatch{Path: expected.Path, Status: "missing", Reason: "candidate is no longer present"})
			continue
		}
		if currentCandidate.IsDir != expected.IsDir || currentCandidate.Bytes != expected.Bytes ||
			currentCandidate.CategoryKey != expected.CategoryKey || !currentCandidate.ModifiedAt.Equal(expected.ModifiedAt) {
			mismatches = append(mismatches, SelectionMismatch{Path: expected.Path, Status: "changed", Reason: "candidate identity, type, size, category, or modification time changed"})
			continue
		}
		selected = append(selected, currentCandidate)
	}
	return selected, mismatches, nil
}

func canonicalPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("canonicalize path %q: %w", path, err)
	}
	return filepath.Clean(absolute), nil
}
