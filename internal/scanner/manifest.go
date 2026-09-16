package scanner

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const SelectionManifestSchemaVersion = 1
const CleanupPlanSchemaVersion = 1

var marshalSelectionManifest = json.MarshalIndent

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

// CleanupPlan is an explicit, reviewable description of candidates that a
// later clean command may remove. It is deliberately separate from a report
// so applying it cannot be implicit.
type CleanupPlan struct {
	SchemaVersion int                 `json:"schema_version"`
	Root          string              `json:"root"`
	CreatedAt     time.Time           `json:"created_at"`
	Action        string              `json:"action"`
	Candidates    []ManifestCandidate `json:"candidates"`
}

func WriteSelectionManifest(path string, root string, candidates []Candidate, exclusions SelectionExclusions) error {
	manifest, err := NewSelectionManifest(root, candidates, exclusions)
	if err != nil {
		return err
	}
	data, err := marshalSelectionManifest(manifest, "", "  ")
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

func WriteCleanupPlan(path string, root string, candidates []Candidate) error {
	plan, err := NewCleanupPlan(root, candidates)
	if err != nil {
		return err
	}
	data, err := marshalSelectionManifest(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("encode cleanup plan: %w", err)
	}
	if err := os.WriteFile(path, append(data, 10), 0o600); err != nil {
		return fmt.Errorf("write cleanup plan: %w", err)
	}
	return nil
}

func ReadCleanupPlan(path string) (CleanupPlan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return CleanupPlan{}, fmt.Errorf("read cleanup plan: %w", err)
	}
	var plan CleanupPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return CleanupPlan{}, fmt.Errorf("decode cleanup plan: %w", err)
	}
	if plan.SchemaVersion != CleanupPlanSchemaVersion {
		return CleanupPlan{}, fmt.Errorf("unsupported cleanup plan schema version %d", plan.SchemaVersion)
	}
	if plan.Action != "delete" {
		return CleanupPlan{}, fmt.Errorf("unsupported cleanup plan action %q", plan.Action)
	}
	if plan.Root == "" {
		return CleanupPlan{}, errors.New("cleanup plan root is required")
	}
	return plan, nil
}

func NewCleanupPlan(root string, candidates []Candidate) (CleanupPlan, error) {
	manifest, err := NewSelectionManifest(root, candidates, SelectionExclusions{})
	if err != nil {
		return CleanupPlan{}, err
	}
	return CleanupPlan{
		SchemaVersion: CleanupPlanSchemaVersion,
		Root:          manifest.Root,
		CreatedAt:     manifest.CreatedAt,
		Action:        "delete",
		Candidates:    manifest.Candidates,
	}, nil
}

func ValidateCleanupPlan(plan CleanupPlan, root string, current []Candidate) ([]Candidate, []SelectionMismatch, error) {
	if plan.SchemaVersion != CleanupPlanSchemaVersion {
		return nil, nil, fmt.Errorf("unsupported cleanup plan schema version %d", plan.SchemaVersion)
	}
	if plan.Action != "delete" {
		return nil, nil, fmt.Errorf("unsupported cleanup plan action %q", plan.Action)
	}
	for _, candidate := range plan.Candidates {
		if !pathWithinRoot(candidate.Path, plan.Root) {
			return nil, nil, fmt.Errorf("cleanup plan candidate %q escapes plan root", candidate.Path)
		}
	}
	return validateManifestCandidates(plan.Root, plan.Candidates, root, current)
}

func NewSelectionManifest(root string, candidates []Candidate, exclusions SelectionExclusions) (SelectionManifest, error) {
	canonicalRoot, err := canonicalPath(root)
	if err != nil {
		return SelectionManifest{}, err
	}
	entries := make([]ManifestCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		entries = append(entries, ManifestCandidate{
			Path:        candidate.Path,
			Group:       candidate.Group,
			CategoryKey: candidate.CategoryKey,
			Bytes:       candidate.Bytes,
			ModifiedAt:  candidate.ModifiedAt,
			IsDir:       candidate.IsDir,
		})
	}
	return SelectionManifest{
		SchemaVersion: SelectionManifestSchemaVersion,
		Root:          canonicalRoot,
		CreatedAt:     time.Now().UTC(),
		Exclusions:    exclusions,
		Candidates:    entries,
	}, nil
}

func ValidateSelectionManifest(manifest SelectionManifest, root string, current []Candidate) ([]Candidate, []SelectionMismatch, error) {
	if manifest.SchemaVersion != SelectionManifestSchemaVersion {
		return nil, nil, fmt.Errorf("unsupported selection manifest schema version %d", manifest.SchemaVersion)
	}
	return validateManifestCandidates(manifest.Root, manifest.Candidates, root, current)
}

func validateManifestCandidates(manifestRootValue string, expectedCandidates []ManifestCandidate, root string, current []Candidate) ([]Candidate, []SelectionMismatch, error) {
	canonicalRoot, err := canonicalPath(root)
	if err != nil {
		return nil, nil, err
	}
	manifestRoot, err := canonicalPath(manifestRootValue)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid selection manifest root: %w", err)
	}
	if canonicalRoot != manifestRoot {
		return nil, nil, fmt.Errorf("selection manifest root %q does not match scan root %q", manifestRootValue, canonicalRoot)
	}
	byPath := make(map[string]Candidate, len(current))
	for _, candidate := range current {
		byPath[filepath.Clean(candidate.Path)] = candidate
	}
	selected := make([]Candidate, 0, len(expectedCandidates))
	mismatches := make([]SelectionMismatch, 0)
	for _, expected := range expectedCandidates {
		if !pathWithinRoot(expected.Path, canonicalRoot) {
			return nil, nil, fmt.Errorf("selection manifest candidate %q escapes scan root", expected.Path)
		}
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

func pathWithinRoot(path, root string) bool {
	canonicalPathValue, err := canonicalPath(path)
	if err != nil {
		return false
	}
	canonicalRoot, err := canonicalPath(root)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(canonicalRoot, canonicalPathValue)
	return err == nil && relative != ".." && !filepath.IsAbs(relative) && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

func canonicalPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("canonicalize path %q: %w", path, err)
	}
	return filepath.Clean(absolute), nil
}
