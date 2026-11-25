package scanner

import (
	"slices"
	"strings"
	"testing"
)

// TestMatchDirectory validates directory matching for all categories.
func TestMatchDirectory(t *testing.T) {
	for _, name := range []string{"node_modules", ".venv", "venv", "__pycache__",
		".pytest_cache", ".mypy_cache", ".tox", "dist", "build", "target",
		".next", ".nuxt", ".cache", ".cache/pip", ".local/pipx", ".bun/install/cache",
		".npm/_cacache", ".npm/_npx", "go/pkg/mod", "AppData/Local/npm-cache",
		".cache/uv", "AppData/Local/uv/cache", ".cache/ms-playwright",
		"AppData/Local/ms-playwright", ".cache/puppeteer", ".cache/pre-commit",
		".local/share/pnpm/store", "Library/pnpm/store", "AppData/Local/pnpm/store",
		"Code/CachedExtensionVSIXs", "Code - Insiders/CachedExtensionVSIXs",
		"Code - OSS/CachedExtensionVSIXs",
		".copilot/copilot-worktrees", ".copilot/session-state"} {
		cat, ok := MatchDirectory(name)
		if !ok {
			t.Fatalf("expected %q to match a category", name)
		}
		if cat.Key == "" {
			t.Fatalf("matched category has empty key for %q", name)
		}
	}

	for name, want := range map[string]string{
		".cache/pip":                           "pip-cache",
		".local/pipx":                          "pipx-data",
		".bun/install/cache":                   "bun-cache",
		".npm/_cacache":                        "npm-cache",
		"AppData/Local/npm-cache":              "npm-cache",
		"go/pkg/mod":                           "go-mod-cache",
		".cache/uv":                            "uv-cache",
		"AppData/Local/uv/cache":               "uv-cache",
		".cache/ms-playwright":                 "playwright-browsers",
		"AppData/Local/ms-playwright":          "playwright-browsers",
		".cache/puppeteer":                     "puppeteer-cache",
		".cache/pre-commit":                    "pre-commit-cache",
		".local/share/pnpm/store":              "pnpm-store",
		"Library/pnpm/store":                   "pnpm-store",
		"Code/CachedExtensionVSIXs":            "vscode-vsix-cache",
		"Code - Insiders/CachedExtensionVSIXs": "vscode-vsix-cache",
		".copilot/copilot-worktrees":           "agent-worktrees",
		".copilot/session-state":               "agent-session-state",
	} {
		cat, ok := MatchDirectory(name)
		if !ok || cat.Key != want {
			t.Fatalf("MatchDirectory(%q) = (%q, %v), want category %q", name, cat.Key, ok, want)
		}
	}

	if _, ok := MatchDirectory("unknown_dir"); ok {
		t.Fatalf("expected unknown_dir not to match")
	}
	if _, ok := MatchDirectory(".bun"); ok {
		t.Fatal("the Bun home must not be treated as a cache")
	}
}

// TestMatchDirectoryHomePrefixesNotCandidates guards against flagging the
// tool homes themselves instead of their cache subdirectories.
func TestMatchDirectoryHomePrefixesNotCandidates(t *testing.T) {
	for _, name := range []string{
		".npm", ".npm/_logs", "AppData/Local/npm-cache-mcp", ".cache/puppeteer-mcp",
		"ms-playwright-mcp", ".cache/ms-playwright-mcp",
		"go", "go/pkg", ".cache/uvx", "Library/pnpm", ".local/share/pnpm",
		".copilot", ".codex",
	} {
		if _, ok := MatchDirectory(name); ok {
			t.Fatalf("expected %q not to match any category", name)
		}
	}
}

// TestMatchFile validates file extension matching.
func TestMatchFile(t *testing.T) {
	cat, ok := MatchFile("/tmp/main.pyc")
	if !ok {
		t.Fatalf("expected .pyc to match")
	}
	if cat.Key != "python-bytecode" {
		t.Fatalf("expected 'python-bytecode', got %q", cat.Key)
	}

	cat, ok = MatchFile("/tmp/main.pyo")
	if !ok {
		t.Fatalf("expected .pyo to match")
	}
	if cat.Key != "python-bytecode" {
		t.Fatalf("expected 'python-bytecode', got %q", cat.Key)
	}

	if _, ok := MatchFile("/tmp/main.go"); ok {
		t.Fatalf("expected .go not to match")
	}
}

// TestMatchFileCaseInsensitive validates case-insensitive matching.
func TestMatchFileCaseInsensitive(t *testing.T) {
	cat, ok := MatchFile("/tmp/main.PYC")
	if !ok {
		t.Fatalf("expected .PYC to match (case insensitive)")
	}
	if cat.Key != "python-bytecode" {
		t.Fatalf("expected 'python-bytecode', got %q", cat.Key)
	}
}

// TestCategoriesList validates all categories have required fields.
func TestCategoriesList(t *testing.T) {
	for _, cat := range categories {
		if cat.Key == "" {
			t.Fatalf("category has empty key")
		}
		if cat.Display == "" {
			t.Fatalf("category %q has empty display", cat.Key)
		}
		if cat.Description == "" {
			t.Fatalf("category %q has empty description", cat.Key)
		}
		if cat.SafetyNote == "" {
			t.Fatalf("category %q has empty safety note", cat.Key)
		}
		if cat.RiskLevel != RiskLow && cat.RiskLevel != RiskReview && cat.RiskLevel != RiskOptIn {
			t.Fatalf("category %q has unsupported risk level %q", cat.Key, cat.RiskLevel)
		}
		if len(cat.DirectoryNames) == 0 && len(cat.DirectoryPaths) == 0 && len(cat.FileExtensions) == 0 {
			t.Fatalf("category %q has no directory names, directory paths, or file extensions", cat.Key)
		}
	}
}

func TestIncludeRiskProfile(t *testing.T) {
	low := Category{RiskLevel: RiskLow}
	review := Category{RiskLevel: RiskReview}
	optIn := Category{RiskLevel: RiskOptIn}

	if !IncludeRiskProfile(low, "conservative") {
		t.Fatal("conservative profile should include low-risk categories")
	}
	if IncludeRiskProfile(review, "conservative") {
		t.Fatal("conservative profile should exclude review categories")
	}
	for _, profile := range []string{"", "balanced", "expanded"} {
		if !IncludeRiskProfile(review, profile) {
			t.Fatalf("profile %q should include review categories", profile)
		}
	}
	for _, profile := range []string{"", "conservative", "balanced", "expanded"} {
		if IncludeRiskProfile(optIn, profile) {
			t.Fatalf("profile %q must exclude opt-in categories", profile)
		}
	}
	if err := validateRiskProfile("reckless"); err == nil {
		t.Fatal("expected invalid risk profile error")
	}
}

func TestIncludeCandidate(t *testing.T) {
	low := Category{Key: "node-modules", RiskLevel: RiskLow}
	review := Category{Key: "pipx-data", RiskLevel: RiskReview}
	optIn := Category{Key: "agent-worktrees", RiskLevel: RiskOptIn}

	// Defaults: no include, no exclude, balanced.
	for _, cat := range []Category{low, review} {
		if !IncludeCandidate(cat, nil, nil, "") {
			t.Fatalf("default scan must include %q category", cat.Key)
		}
	}
	if IncludeCandidate(optIn, nil, nil, "") {
		t.Fatal("opt-in category must be excluded by default")
	}

	// Risk profiles still apply to non opt-in categories.
	if IncludeCandidate(review, nil, nil, "conservative") {
		t.Fatal("conservative profile must exclude review categories")
	}

	// Explicit include enables opt-in categories at every profile.
	for _, profile := range []string{"", "conservative", "balanced", "expanded"} {
		if !IncludeCandidate(optIn, ListToSet([]string{"agent-worktrees"}), nil, profile) {
			t.Fatalf("explicit include must enable opt-in category at profile %q", profile)
		}
	}

	// The explicit include acts through the include filter, so a matching
	// exclusion always wins and a non-opt-in include does not unlock it.
	if IncludeCandidate(optIn, ListToSet([]string{"agent-worktrees"}), ListToSet([]string{"agent-worktrees"}), "") {
		t.Fatal("exclude filter must win over include for opt-in categories")
	}
	if IncludeCandidate(optIn, ListToSet([]string{"node-modules"}), nil, "") {
		t.Fatal("including a different category must not enable opt-in categories")
	}

	// Explicit include cannot bypass a risk profile for non opt-in categories.
	if !IncludeCandidate(low, ListToSet([]string{"node-modules"}), nil, "conservative") {
		t.Fatal("explicit include should keep low-risk categories enabled under conservative")
	}
	if IncludeCandidate(review, ListToSet([]string{"pipx-data"}), nil, "conservative") {
		t.Fatal("explicit include must not bypass the conservative profile for review categories")
	}
}

// TestCategoryDescriptions validates all descriptions are non-empty and meaningful.
func TestCategoryDescriptions(t *testing.T) {
	for _, cat := range categories {
		if len(cat.Description) < 5 {
			t.Fatalf("category %q description too short: %q", cat.Key, cat.Description)
		}
	}
}

func TestCategoryKeysAreSortedAndFiltersRejectTypos(t *testing.T) {
	keys := CategoryKeys()
	if len(keys) != len(categories) || !slices.IsSorted(keys) {
		t.Fatalf("CategoryKeys() = %#v", keys)
	}
	if err := validateCategoryFilters([]string{"node-modules"}, []string{"python-cache"}); err != nil {
		t.Fatalf("known category filters failed: %v", err)
	}
	err := validateCategoryFilters([]string{"node-module"}, nil)
	if err == nil || !strings.Contains(err.Error(), "node-modules") {
		t.Fatalf("unknown category error = %v", err)
	}
}

// TestNewDirCategory validates directory category creation.
func TestNewDirCategory(t *testing.T) {
	cat := newDirCategory("test", "Test Dir", "A test category", "test_dir", "test-dir")
	if cat.Key != "test" {
		t.Fatalf("expected key 'test', got %q", cat.Key)
	}
	if cat.Display != "Test Dir" {
		t.Fatalf("expected display 'Test Dir', got %q", cat.Key)
	}
	if _, ok := cat.DirectoryNames["test_dir"]; !ok {
		t.Fatalf("expected DirectoryNames to contain 'test_dir'")
	}
	if _, ok := cat.DirectoryNames["test-dir"]; !ok {
		t.Fatalf("expected DirectoryNames to contain 'test-dir'")
	}
	if len(cat.FileExtensions) != 0 {
		t.Fatalf("expected empty FileExtensions")
	}
}

// TestNewFileCategory validates file category creation.
func TestNewFileCategory(t *testing.T) {
	cat := newFileCategory("test", "Test File", "A test file category", ".test", ".txt")
	if cat.Key != "test" {
		t.Fatalf("expected key 'test', got %q", cat.Key)
	}
	if _, ok := cat.FileExtensions[".test"]; !ok {
		t.Fatalf("expected FileExtensions to contain '.test'")
	}
	if _, ok := cat.FileExtensions[".txt"]; !ok {
		t.Fatalf("expected FileExtensions to contain '.txt'")
	}
	if len(cat.DirectoryNames) != 0 {
		t.Fatalf("expected empty DirectoryNames")
	}
}
