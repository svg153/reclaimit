package reportdiff

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const SchemaVersion = 1
const maxReportBytes = 64 << 20

type candidate struct {
	CategoryKey string `json:"category_key"`
	Path        string `json:"path"`
	Bytes       int64  `json:"bytes"`
	ModifiedAt  string `json:"modified_at"`
	IsDir       bool   `json:"is_dir"`
}

type report struct {
	SchemaVersion      int         `json:"schema_version"`
	SelectedBytes      int64       `json:"selected_bytes"`
	SelectedCandidates []candidate `json:"selected_candidates"`
}

type CategoryChange struct {
	Category    string
	Before      int64
	After       int64
	Delta       int64
	BeforeCount int
	AfterCount  int
}

type Result struct {
	BeforeBytes   int64
	AfterBytes    int64
	Categories    []CategoryChange
	Added         int
	Removed       int
	Changed       int
	Unchanged     int
	PathsRedacted bool
	AddedPaths    []string
	RemovedPaths  []string
	ChangedPaths  []string
}

func read(path string) (report, error) {
	f, err := os.Open(path)
	if err != nil {
		return report{}, fmt.Errorf("open report %q: %w", path, err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return report{}, fmt.Errorf("stat report %q: %w", path, err)
	}
	if info.Size() > maxReportBytes {
		return report{}, fmt.Errorf("report %q exceeds %d bytes", path, maxReportBytes)
	}
	dec := json.NewDecoder(io.LimitReader(f, maxReportBytes+1))
	var value report
	if err := dec.Decode(&value); err != nil {
		return report{}, fmt.Errorf("decode report %q: %w", path, err)
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return report{}, fmt.Errorf("decode report %q: expected one JSON document", path)
	}
	if value.SchemaVersion != SchemaVersion {
		return report{}, fmt.Errorf("unsupported report schema version %d in %q; expected %d", value.SchemaVersion, path, SchemaVersion)
	}
	return value, nil
}

func CompareFiles(beforePath, afterPath string) (Result, error) {
	before, err := read(beforePath)
	if err != nil {
		return Result{}, err
	}
	after, err := read(afterPath)
	if err != nil {
		return Result{}, err
	}
	return compare(before, after), nil
}

func compare(before, after report) Result {
	result := Result{BeforeBytes: before.SelectedBytes, AfterBytes: after.SelectedBytes}
	type totals struct {
		bytes int64
		count int
	}
	beforeCategories := map[string]totals{}
	afterCategories := map[string]totals{}
	for _, item := range before.SelectedCandidates {
		v := beforeCategories[item.CategoryKey]
		v.bytes += item.Bytes
		v.count++
		beforeCategories[item.CategoryKey] = v
		if item.Path == "<redacted>" {
			result.PathsRedacted = true
		}
	}
	for _, item := range after.SelectedCandidates {
		v := afterCategories[item.CategoryKey]
		v.bytes += item.Bytes
		v.count++
		afterCategories[item.CategoryKey] = v
		if item.Path == "<redacted>" {
			result.PathsRedacted = true
		}
	}
	keys := map[string]struct{}{}
	for key := range beforeCategories {
		keys[key] = struct{}{}
	}
	for key := range afterCategories {
		keys[key] = struct{}{}
	}
	for key := range keys {
		b, a := beforeCategories[key], afterCategories[key]
		result.Categories = append(result.Categories, CategoryChange{Category: key, Before: b.bytes, After: a.bytes, Delta: a.bytes - b.bytes, BeforeCount: b.count, AfterCount: a.count})
	}
	sort.Slice(result.Categories, func(i, j int) bool { return result.Categories[i].Category < result.Categories[j].Category })
	if result.PathsRedacted {
		return result
	}
	identity := func(item candidate) string { return fmt.Sprintf("%s\x00%s\x00%t", item.CategoryKey, item.Path, item.IsDir) }
	old := map[string]candidate{}
	for _, item := range before.SelectedCandidates {
		old[identity(item)] = item
	}
	for _, item := range after.SelectedCandidates {
		key := identity(item)
		previous, found := old[key]
		if !found {
			result.Added++
			result.AddedPaths = append(result.AddedPaths, item.Path)
			continue
		}
		if previous.Bytes != item.Bytes || previous.ModifiedAt != item.ModifiedAt {
			result.Changed++
			result.ChangedPaths = append(result.ChangedPaths, item.Path)
		} else {
			result.Unchanged++
		}
		delete(old, key)
	}
	result.Removed = len(old)
	for _, item := range old {
		result.RemovedPaths = append(result.RemovedPaths, item.Path)
	}
	sort.Strings(result.AddedPaths)
	sort.Strings(result.RemovedPaths)
	sort.Strings(result.ChangedPaths)
	return result
}

func Render(result Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Report diff\nSelected cleanup: %d -> %d bytes (%+d)\n\nCategories\n", result.BeforeBytes, result.AfterBytes, result.AfterBytes-result.BeforeBytes)
	for _, item := range result.Categories {
		fmt.Fprintf(&b, "  %-18s %d -> %d bytes (%+d), %d -> %d candidates\n", item.Category, item.Before, item.After, item.Delta, item.BeforeCount, item.AfterCount)
	}
	if result.PathsRedacted {
		b.WriteString("\nCandidate changes omitted because at least one report is anonymous.\n")
	} else {
		fmt.Fprintf(&b, "\nCandidates: %d added, %d removed, %d changed, %d unchanged\n", result.Added, result.Removed, result.Changed, result.Unchanged)
		for _, path := range result.AddedPaths {
			fmt.Fprintf(&b, "  added    %q\n", path)
		}
		for _, path := range result.RemovedPaths {
			fmt.Fprintf(&b, "  removed  %q\n", path)
		}
		for _, path := range result.ChangedPaths {
			fmt.Fprintf(&b, "  changed  %q\n", path)
		}
	}
	return b.String()
}
