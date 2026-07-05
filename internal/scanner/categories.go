package scanner

import (
	"path/filepath"
	"strings"
)

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
	newDirCategory("bun-cache", ".bun", "Bun global cache that is safe to remove.", ".bun"),
	newFileCategory("ds-store", ".DS_Store", "macOS desktop storage file that is regenerated automatically.", ".DS_Store"),
	newDirCategory("spotlight-index", ".Spotlight-V100", "macOS Spotlight indexing database.", ".Spotlight-V100"),
	newDirCategory("macos-trash", ".Trashes", "macOS trash folder.", ".Trashes"),
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

func IncludeCategory(category string, includeSet, excludeSet map[string]struct{}) bool {
	if _, blocked := excludeSet[category]; blocked {
		return false
	}
	if len(includeSet) == 0 {
		return true
	}
	_, allowed := includeSet[category]
	return allowed
}

func MatchDirectory(name string) (Category, bool) {
	for _, category := range categories {
		if _, ok := category.DirectoryNames[name]; ok {
			return category, true
		}
	}
	return Category{}, false
}

func MatchFile(path string) (Category, bool) {
	ext := filepath.Ext(path)
	name := filepath.Base(path)

	// First: try case-insensitive extension match for real extensions (.pyc, .pyo, etc.)
	for _, category := range categories {
		if len(category.DirectoryNames) == 0 && len(category.FileExtensions) > 0 {
			for catExt := range category.FileExtensions {
				if ext != "" && len(ext) > 1 && strings.EqualFold(ext, catExt) {
					return category, true
				}
			}
		}
	}

	// Second: exact name match for dotfiles where Ext returns the full name (e.g., .DS_Store)
	// filepath.Ext(".DS_Store") == ".DS_Store" — ext equals name for these
	if ext == name && strings.HasPrefix(name, ".") {
		for _, category := range categories {
			if len(category.DirectoryNames) == 0 && len(category.FileExtensions) > 0 {
				if _, ok := category.FileExtensions[ext]; ok {
					return category, true
				}
			}
		}
	}

	return Category{}, false
}
