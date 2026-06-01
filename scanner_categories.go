package reclaimit

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
