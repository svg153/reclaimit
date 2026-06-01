package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

func DetermineGroup(path string, opts AnalyzeOptions, cache map[string]string) string {
	if opts.GroupMode == "repo" {
		if repoRoot := findRepoRoot(path, opts.Root, cache); repoRoot != "" {
			return repoRoot
		}
	}
	return ancestorGroup(path, opts.Root, opts.GroupDepth)
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
