package scanner

import (
	"testing"
)

// TestMatchDirectory validates directory matching for all categories.
func TestMatchDirectory(t *testing.T) {
	for _, name := range []string{"node_modules", ".venv", "venv", "__pycache__",
		".pytest_cache", ".mypy_cache", ".tox", "dist", "build", "target",
		".next", ".nuxt", ".cache"} {
		cat, ok := MatchDirectory(name)
		if !ok {
			t.Fatalf("expected %q to match a category", name)
		}
		if cat.Key == "" {
			t.Fatalf("matched category has empty key for %q", name)
		}
	}

	if _, ok := MatchDirectory("unknown_dir"); ok {
		t.Fatalf("expected unknown_dir not to match")
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
		if len(cat.DirectoryNames) == 0 && len(cat.FileExtensions) == 0 {
			t.Fatalf("category %q has no directory names or file extensions", cat.Key)
		}
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
