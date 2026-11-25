package scanner

import (
	"fmt"
	"path/filepath"
	"sort"
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
	reviewCategory(newDirCategory("js-build", "dist / build", "Build artifacts that can be regenerated from source.", "dist", "build")),
	newDirCategory("rust-target", "target", "Rust build output that cargo rebuilds.", "target"),
	newDirCategory("next-cache", ".next / .nuxt", "Frontend framework build caches.", ".next", ".nuxt"),
	reviewCategory(newDirCategory("generic-cache", ".cache", "Generic caches. Review first because some tools keep useful offline assets here.", ".cache")),
	newDirPathCategory("bun-cache", "~/.bun/install/cache", "Bun download cache. Packages are downloaded again when needed.", ".bun", "install", "cache"),
	newDirPathCategory("pip-cache", "~/.cache/pip", "pip download and wheel cache. Packages are downloaded again when needed.", ".cache", "pip"),
	reviewCategory(newDirPathCategory("pipx-data", "~/.local/pipx", "pipx-managed applications and environments. Removing this requires reinstalling those applications.", ".local", "pipx")),
	newFileCategory("ds-store", ".DS_Store", "macOS desktop storage file that is regenerated automatically.", ".DS_Store"),
	newDirCategory("spotlight-index", ".Spotlight-V100", "macOS Spotlight indexing database.", ".Spotlight-V100"),
	reviewCategory(newDirCategory("macos-trash", ".Trashes", "macOS trash folder.", ".Trashes")),

	// Field evidence (2026-11-25): see docs/field-evidence-2026-11-25.md.
	newDirPathSuffixesCategory("npm-cache", "npm cache", "npm download and npx cache. Packages are downloaded again when needed.", "AppData/Local/npm-cache", ".npm/_cacache", ".npm/_npx"),
	newDirPathSuffixesCategory("go-mod-cache", "Go module cache", "Go module download cache. Modules are downloaded again by go commands and removed with `go clean -modcache`.", "go/pkg/mod"),
	newDirPathSuffixesCategory("uv-cache", "uv cache", "uv package cache. Packages are downloaded again on the next sync or install.", ".cache/uv", "AppData/Local/uv/cache"),
	newDirPathSuffixesCategory("playwright-browsers", "Playwright browsers", "Browser binaries installed by Playwright. Reinstall with `npx playwright install`.", ".cache/ms-playwright", "AppData/Local/ms-playwright"),
	newDirPathSuffixesCategory("puppeteer-cache", "Puppeteer browsers", "Browser binaries downloaded by Puppeteer. Downloaded again on demand.", ".cache/puppeteer"),
	newDirPathSuffixesCategory("pre-commit-cache", "pre-commit cache", "pre-commit hook environments. Cloned and rebuilt automatically on the next hook run.", ".cache/pre-commit"),
	reviewCategory(newDirPathSuffixesCategory("pnpm-store", "pnpm store", "pnpm content-addressable store. node_modules directories may be hard-linked into it; prefer `pnpm store prune` over deleting the store.", ".local/share/pnpm/store", "Library/pnpm/store", "AppData/Local/pnpm/store")),
	reviewCategory(newDirPathSuffixesCategory("vscode-vsix-cache", "VS Code extension VSIX cache", "Downloaded extension packages cached by VS Code. Downloaded again from the marketplace when needed.", "Code/CachedExtensionVSIXs", "Code - Insiders/CachedExtensionVSIXs", "Code - OSS/CachedExtensionVSIXs")),
	optInCategory(newDirPathSuffixesCategory("agent-worktrees", "AI agent git worktrees", "Git worktrees created by AI coding agents. They may contain uncommitted work; enable explicitly and review before deleting.", ".copilot/copilot-worktrees", ".codex/worktrees")),
	optInCategory(newDirPathSuffixesCategory("agent-session-state", "AI agent session state", "Conversation, plan, and artifact state kept by AI coding agents. It may include active sessions and history; enable explicitly and review before deleting.", ".copilot/session-state", ".codex/sessions")),
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
		SafetyNote:     description,
		RiskLevel:      RiskLow,
		DirectoryNames: dirNames,
		DirectoryPaths: map[string]struct{}{},
		FileExtensions: map[string]struct{}{},
	}
}

func newDirPathCategory(key, display, description string, pathParts ...string) Category {
	return Category{
		Key:            key,
		Display:        display,
		Description:    description,
		SafetyNote:     description,
		RiskLevel:      RiskLow,
		DirectoryNames: map[string]struct{}{},
		DirectoryPaths: map[string]struct{}{filepath.Join(pathParts...): {}},
		FileExtensions: map[string]struct{}{},
	}
}

// newDirPathSuffixesCategory builds a path-suffix category with several
// home-relative suffixes (forward slashes are normalized to the OS separator).
func newDirPathSuffixesCategory(key, display, description string, suffixes ...string) Category {
	paths := make(map[string]struct{}, len(suffixes))
	for _, suffix := range suffixes {
		paths[filepath.Join(strings.Split(filepath.ToSlash(suffix), "/")...)] = struct{}{}
	}
	return Category{
		Key:            key,
		Display:        display,
		Description:    description,
		SafetyNote:     description,
		RiskLevel:      RiskLow,
		DirectoryNames: map[string]struct{}{},
		DirectoryPaths: paths,
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
		SafetyNote:     description,
		RiskLevel:      RiskLow,
		DirectoryNames: map[string]struct{}{},
		DirectoryPaths: map[string]struct{}{},
		FileExtensions: fileExts,
	}
}

func reviewCategory(category Category) Category {
	category.RiskLevel = RiskReview
	return category
}

func optInCategory(category Category) Category {
	category.RiskLevel = RiskOptIn
	return category
}

func IncludeRiskProfile(category Category, profile string) bool {
	if category.RiskLevel == RiskOptIn {
		return false
	}
	if profile == "" || profile == "balanced" || profile == "expanded" {
		return true
	}
	return profile != "conservative" || category.RiskLevel == RiskLow
}

// IncludeCandidate combines the category include/exclude filters with the risk
// profile. Opt-in categories never pass the profile filters on their own; they
// are considered only when explicitly listed in the include set.
func IncludeCandidate(category Category, includeSet, excludeSet map[string]struct{}, profile string) bool {
	if !IncludeCategory(category.Key, includeSet, excludeSet) {
		return false
	}
	if category.RiskLevel == RiskOptIn {
		_, explicit := includeSet[category.Key]
		return explicit
	}
	return IncludeRiskProfile(category, profile)
}

func validateRiskProfile(profile string) error {
	switch profile {
	case "", "conservative", "balanced", "expanded":
		return nil
	default:
		return fmt.Errorf("unknown risk profile %q (supported: conservative, balanced, expanded)", profile)
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

// CategoryKeys returns the stable, sorted set accepted by category filters.
func CategoryKeys() []string {
	keys := make([]string, 0, len(categories))
	for _, category := range categories {
		keys = append(keys, category.Key)
	}
	sort.Strings(keys)
	return keys
}

func validateCategoryFilters(include, exclude []string) error {
	keys := CategoryKeys()
	known := ListToSet(keys)
	for _, filter := range append(append([]string{}, include...), exclude...) {
		if _, ok := known[filter]; !ok {
			return fmt.Errorf("unknown category %q (supported: %s)", filter, strings.Join(keys, ", "))
		}
	}
	return nil
}

func MatchDirectory(path string) (Category, bool) {
	name := filepath.Base(path)
	for _, category := range categories {
		if _, ok := category.DirectoryNames[name]; ok {
			return category, true
		}
		for relativePath := range category.DirectoryPaths {
			if hasPathSuffix(path, relativePath) {
				return category, true
			}
		}
	}
	return Category{}, false
}

func hasPathSuffix(path, suffix string) bool {
	path = filepath.ToSlash(filepath.Clean(path))
	suffix = strings.Trim(filepath.ToSlash(filepath.Clean(suffix)), "/")
	return path == suffix || strings.HasSuffix(path, "/"+suffix)
}

func MatchFile(path string) (Category, bool) {
	ext := filepath.Ext(path)
	name := filepath.Base(path)

	// First: try case-insensitive extension match for real extensions (.pyc, .pyo, etc.)
	for _, category := range categories {
		if len(category.DirectoryNames) == 0 && len(category.DirectoryPaths) == 0 && len(category.FileExtensions) > 0 {
			for catExt := range category.FileExtensions {
				if ext != "" && len(ext) > 1 && strings.EqualFold(ext, catExt) {
					return category, true
				}
			}
		}
	}

	// Second: exact name match for dotfiles where Ext returns the full name (e.g., .DS_Store)
	if ext == name && strings.HasPrefix(name, ".") {
		for _, category := range categories {
			if len(category.DirectoryNames) == 0 && len(category.DirectoryPaths) == 0 && len(category.FileExtensions) > 0 {
				if _, ok := category.FileExtensions[ext]; ok {
					return category, true
				}
			}
		}
	}

	return Category{}, false
}
