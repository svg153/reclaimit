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
	newDirCategory("gradle-cache", ".gradle", "Gradle build cache and dependency downloads that can be rebuilt.", ".gradle"),
	newDirCategory("ide-config", ".idea / .vscode", "IDE configuration and index caches that can be regenerated.", ".idea", ".vscode"),
	newDirCategory("vendor", "vendor", "Go vendor directory that can be rebuilt with go mod vendor.", "vendor"),
	newDirCategory("go-mod-cache", "pkg / mod", "Go module cache directories.", "pkg", "mod"),
	newDirCategory("npm-cache", ".npm", "npm package cache that can be rebuilt.", ".npm"),
	newDirCategory("yarn-cache", ".yarn", "Yarn package cache that can be rebuilt.", ".yarn"),
	newDirCategory("pnpm-store", ".pnpm-store", "pnpm global store that can be rebuilt.", ".pnpm-store"),
	newDirCategory("bun-cache", ".bun", "Bun package cache that can be rebuilt.", ".bun"),
	newDirCategory("cargo-cache", ".cargo", "Cargo registry cache that can be rebuilt.", ".cargo"),
	newDirCategory("go-sum-cache", "go.sum", "Go sum cache file.", "go.sum"),
	newDirCategory("node-modules-cache", "node_modules/.cache", "npm package build cache inside node_modules.", "node_modules/.cache"),
	newDirCategory("terraform-state", ".terraform", "Terraform provider cache and state that can be rebuilt.", ".terraform"),
	newDirCategory("docker-buildx", ".buildx", "Docker Buildx cache.", ".buildx"),
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
