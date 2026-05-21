package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Category struct {
	Key            string
	Display        string
	Description    string
	DirectoryNames map[string]struct{}
	FileExtensions map[string]struct{}
}

type PathSize struct {
	Path  string
	Bytes int64
}

type Candidate struct {
	Category    string
	CategoryKey string
	Path        string
	Group       string
	Bytes       int64
	Description string
	ModifiedAt  time.Time
	IsDir       bool
}

type CategorySummary struct {
	Category    string
	CategoryKey string
	Bytes       int64
	Count       int
	Description string
}

type GroupSummary struct {
	Group      string
	Bytes      int64
	Count      int
	ModifiedAt time.Time
}

type scanSummary struct {
	bytes      int64
	modifiedAt time.Time
}

type Report struct {
	Command                   string
	Root                      string
	TotalBytes                int64
	FreeBytes                 int64
	AvailableBytes            int64
	FilesystemBytes           int64
	TopEntries                []PathSize
	TopFiles                  []PathSize
	Candidates                []Candidate
	SelectedCandidates        []Candidate
	CandidateBytes            int64
	SelectedBytes             int64
	CategorySummaries         []CategorySummary
	GroupSummaries            []GroupSummary
	SelectedCategorySummaries []CategorySummary
	SelectedGroupSummaries    []GroupSummary
	DeletedBytes              int64
}

var categories = []Category{
	newDirCategory("node-modules", "node_modules", "JavaScript dependencies that are safe to reinstall from lockfiles.", "node_modules"),
	newDirCategory("python-venv", ".venv / venv", "Python virtual environments that can usually be recreated from requirements or pyproject files.", ".venv", "venv"),
	newDirCategory("python-cache", "__pycache__", "Compiled Python caches that are regenerated automatically.", "__pycache__"),
	newFileCategory("python-bytecode", "*.pyc", "Compiled Python bytecode files regenerated on demand.", ".pyc", ".pyo"),
	newDirCategory("pytest-cache", ".pytest_cache", "Pytest execution cache that is safe to remove.", ".pytest_cache"),
	newDirCategory("mypy-cache", ".mypy_cache", "Mypy cache that is safe to remove.", ".mypy_cache"),
	newDirCategory("tox", ".tox", "Tox virtualenvs that can be recreated.", ".tox"),
	newDirCategory("js-build", "dist / build", "Build artifacts that can be regenerated from source.", "dist", "build"),
	newDirCategory("rust-target", "target", "Rust build output that cargo rebuilds.", "target"),
	newDirCategory("next-cache", ".next / .nuxt", "Frontend framework build caches.", ".next", ".nuxt"),
	newDirCategory("generic-cache", ".cache", "Generic caches. Review first because some tools keep useful offline assets here.", ".cache"),
}

func newDirCategory(key, display, description string, names ...string) Category {
	dirNames := make(map[string]struct{}, len(names))
	for _, name := range names {
		dirNames[name] = struct{}{}
	}
	return Category{
		Key:            key,
		Display:        display,
		Description:    description,
		DirectoryNames: dirNames,
		FileExtensions: map[string]struct{}{},
	}
}

func newFileCategory(key, display, description string, exts ...string) Category {
	fileExts := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		fileExts[ext] = struct{}{}
	}
	return Category{
		Key:            key,
		Display:        display,
		Description:    description,
		DirectoryNames: map[string]struct{}{},
		FileExtensions: fileExts,
	}
}

func Analyze(cfg config) (Report, error) {
	rootInfo, err := os.Lstat(cfg.root)
	if err != nil {
		return Report{}, err
	}
	if !rootInfo.IsDir() {
		return Report{}, fmt.Errorf("%s is not a directory", cfg.root)
	}

	report := Report{
		Command: cfg.command,
		Root:    cfg.root,
	}
	filesystemBytes, freeBytes, availableBytes := filesystemUsage(cfg.root)
	report.FilesystemBytes = filesystemBytes
	report.FreeBytes = freeBytes
	report.AvailableBytes = availableBytes

	candidateByKey := make(map[string]int)
	groupCache := map[string]string{}
	includeSet := listToSet(cfg.includeCategories)
	excludeSet := listToSet(cfg.excludeCategories)

	entries, err := os.ReadDir(cfg.root)
	if err != nil {
		return Report{}, err
	}

	for _, entry := range entries {
		childPath := filepath.Join(cfg.root, entry.Name())
		summary, err := scanNode(childPath, cfg, false, &report, candidateByKey, groupCache, includeSet, excludeSet)
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				continue
			}
			return Report{}, err
		}
		if summary.bytes > 0 {
			report.TopEntries = append(report.TopEntries, PathSize{Path: childPath, Bytes: summary.bytes})
		}
	}

	report.TotalBytes = sumBytes(report.TopEntries)
	sortPathSizes(report.TopEntries)
	if len(report.TopEntries) > cfg.topEntries {
		report.TopEntries = report.TopEntries[:cfg.topEntries]
	}
	sortPathSizes(report.TopFiles)
	if len(report.TopFiles) > cfg.topFiles {
		report.TopFiles = report.TopFiles[:cfg.topFiles]
	}

	sortCandidates(report.Candidates)
	report.CandidateBytes = sumCandidateBytes(report.Candidates)
	report.CategorySummaries = summarizeCategories(report.Candidates)
	report.GroupSummaries = summarizeGroups(report.Candidates, cfg.topGroups)
	applySelection(&report, cfg.excludeGroups, cfg.excludePaths)
	report.SelectedGroupSummaries = summarizeGroups(report.SelectedCandidates, cfg.topGroups)
	return report, nil
}

func Clean(candidates []Candidate) (int64, error) {
	var deleted int64
	for _, candidate := range candidates {
		if err := os.RemoveAll(candidate.Path); err != nil {
			return deleted, fmt.Errorf("deleting %s: %w", candidate.Path, err)
		}
		deleted += candidate.Bytes
	}
	return deleted, nil
}

func scanNode(path string, cfg config, inCandidateDir bool, report *Report, candidateByKey map[string]int, groupCache map[string]string, includeSet, excludeSet map[string]struct{}) (scanSummary, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return scanSummary{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return scanSummary{}, nil
	}

	if info.IsDir() {
		dirCategory, dirIsCandidate := matchDirectory(info.Name())
		nextInCandidate := inCandidateDir || dirIsCandidate
		entries, err := os.ReadDir(path)
		if err != nil {
			return scanSummary{}, err
		}
		var total int64
		latestModified := info.ModTime()
		for _, entry := range entries {
			childPath := filepath.Join(path, entry.Name())
			summary, err := scanNode(childPath, cfg, nextInCandidate, report, candidateByKey, groupCache, includeSet, excludeSet)
			if err != nil {
				if errors.Is(err, os.ErrPermission) {
					continue
				}
				return scanSummary{}, err
			}
			total += summary.bytes
			if summary.modifiedAt.After(latestModified) {
				latestModified = summary.modifiedAt
			}
		}
		if dirIsCandidate && !inCandidateDir && includeCategory(dirCategory.Key, includeSet, excludeSet) && total >= cfg.minCandidateSize {
			addCandidate(report, candidateByKey, Candidate{
				Category:    dirCategory.Display,
				CategoryKey: dirCategory.Key,
				Path:        path,
				Group:       determineGroup(path, cfg, groupCache),
				Bytes:       total,
				Description: dirCategory.Description,
				ModifiedAt:  latestModified,
				IsDir:       true,
			})
		}
		return scanSummary{bytes: total, modifiedAt: latestModified}, nil
	}

	if !info.Mode().IsRegular() {
		return scanSummary{}, nil
	}

	size := info.Size()
	report.TopFiles = pushTop(report.TopFiles, PathSize{Path: path, Bytes: size}, cfg.topFiles)

	fileCategory, fileIsCandidate := matchFile(path)
	if fileIsCandidate && !inCandidateDir && includeCategory(fileCategory.Key, includeSet, excludeSet) && size >= cfg.minCandidateSize {
		addCandidate(report, candidateByKey, Candidate{
			Category:    fileCategory.Display,
			CategoryKey: fileCategory.Key,
			Path:        path,
			Group:       determineGroup(path, cfg, groupCache),
			Bytes:       size,
			Description: fileCategory.Description,
			ModifiedAt:  info.ModTime(),
			IsDir:       false,
		})
	}

	return scanSummary{bytes: size, modifiedAt: info.ModTime()}, nil
}

func addCandidate(report *Report, candidateByKey map[string]int, candidate Candidate) {
	key := candidate.CategoryKey + ":" + candidate.Path
	if _, exists := candidateByKey[key]; exists {
		return
	}
	candidateByKey[key] = len(report.Candidates)
	report.Candidates = append(report.Candidates, candidate)
}

func includeCategory(category string, includeSet, excludeSet map[string]struct{}) bool {
	if _, blocked := excludeSet[category]; blocked {
		return false
	}
	if len(includeSet) == 0 {
		return true
	}
	_, allowed := includeSet[category]
	return allowed
}

func matchDirectory(name string) (Category, bool) {
	for _, category := range categories {
		if _, ok := category.DirectoryNames[name]; ok {
			return category, true
		}
	}
	return Category{}, false
}

func matchFile(path string) (Category, bool) {
	ext := strings.ToLower(filepath.Ext(path))
	for _, category := range categories {
		if _, ok := category.FileExtensions[ext]; ok {
			return category, true
		}
	}
	return Category{}, false
}

func determineGroup(path string, cfg config, cache map[string]string) string {
	if cfg.groupMode == "repo" {
		if repoRoot := findRepoRoot(path, cfg.root, cache); repoRoot != "" {
			return repoRoot
		}
	}
	return ancestorGroup(path, cfg.root, cfg.groupDepth)
}

func findRepoRoot(path, root string, cache map[string]string) string {
	cursor := path
	info, err := os.Lstat(cursor)
	if err == nil && !info.IsDir() {
		cursor = filepath.Dir(cursor)
	}

	for {
		if value, ok := cache[cursor]; ok {
			return value
		}
		if _, err := os.Stat(filepath.Join(cursor, ".git")); err == nil {
			cache[cursor] = cursor
			return cursor
		}
		if cursor == root {
			cache[cursor] = ""
			return ""
		}
		parent := filepath.Dir(cursor)
		if parent == cursor {
			cache[cursor] = ""
			return ""
		}
		cursor = parent
	}
}

func ancestorGroup(path, root string, depth int) string {
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == "." {
		return root
	}
	parts := strings.Split(relative, string(filepath.Separator))
	if len(parts) == 0 {
		return root
	}
	if len(parts) < depth {
		depth = len(parts)
	}
	groupParts := parts[:depth]
	return filepath.Join(append([]string{root}, groupParts...)...)
}

func pushTop(list []PathSize, item PathSize, limit int) []PathSize {
	list = append(list, item)
	sortPathSizes(list)
	if len(list) > limit {
		list = list[:limit]
	}
	return list
}

func sortPathSizes(items []PathSize) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Bytes == items[j].Bytes {
			return items[i].Path < items[j].Path
		}
		return items[i].Bytes > items[j].Bytes
	})
}

func sortCandidates(items []Candidate) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Bytes == items[j].Bytes {
			return items[i].Path < items[j].Path
		}
		return items[i].Bytes > items[j].Bytes
	})
}

func summarizeCategories(candidates []Candidate) []CategorySummary {
	index := map[string]*CategorySummary{}
	for _, candidate := range candidates {
		entry, ok := index[candidate.CategoryKey]
		if !ok {
			entry = &CategorySummary{
				Category:    candidate.Category,
				CategoryKey: candidate.CategoryKey,
				Description: candidate.Description,
			}
			index[candidate.CategoryKey] = entry
		}
		entry.Bytes += candidate.Bytes
		entry.Count++
	}

	summaries := make([]CategorySummary, 0, len(index))
	for _, summary := range index {
		summaries = append(summaries, *summary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Bytes == summaries[j].Bytes {
			return summaries[i].Category < summaries[j].Category
		}
		return summaries[i].Bytes > summaries[j].Bytes
	})
	return summaries
}

func summarizeGroups(candidates []Candidate, limit int) []GroupSummary {
	index := map[string]*GroupSummary{}
	for _, candidate := range candidates {
		entry, ok := index[candidate.Group]
		if !ok {
			entry = &GroupSummary{Group: candidate.Group}
			index[candidate.Group] = entry
		}
		entry.Bytes += candidate.Bytes
		entry.Count++
		if candidate.ModifiedAt.After(entry.ModifiedAt) {
			entry.ModifiedAt = candidate.ModifiedAt
		}
	}

	summaries := make([]GroupSummary, 0, len(index))
	for _, summary := range index {
		summaries = append(summaries, *summary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Bytes == summaries[j].Bytes {
			return summaries[i].Group < summaries[j].Group
		}
		return summaries[i].Bytes > summaries[j].Bytes
	})
	if len(summaries) > limit {
		summaries = summaries[:limit]
	}
	return summaries
}

func isGroupExcluded(candidate Candidate, excludedGroups []string) bool {
	for _, group := range excludedGroups {
		if hasPathPrefix(candidate.Group, group) || hasPathPrefix(candidate.Path, group) {
			return true
		}
	}
	return false
}

func hasPathPrefix(path, prefix string) bool {
	if path == prefix {
		return true
	}
	rel, err := filepath.Rel(prefix, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func sumBytes(items []PathSize) int64 {
	var total int64
	for _, item := range items {
		total += item.Bytes
	}
	return total
}

func sumCandidateBytes(items []Candidate) int64 {
	var total int64
	for _, item := range items {
		total += item.Bytes
	}
	return total
}

func listToSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

// filesystemUsage is implemented per-OS in filesystem_*.go
