package scanner

import (
	"testing"
)

func TestMatchDirectory_NodeModules(t *testing.T) {
	cat, ok := MatchDirectory("node_modules")
	if !ok {
		t.Fatal("expected node_modules to match")
	}
	if cat.Key != "node-modules" {
		t.Errorf("expected key node-modules, got %s", cat.Key)
	}
}

func TestMatchDirectory_PythonVenv(t *testing.T) {
	cat, ok := MatchDirectory(".venv")
	if !ok {
		t.Fatal("expected .venv to match")
	}
	if cat.Key != "python-venv" {
		t.Errorf("expected key python-venv, got %s", cat.Key)
	}
}

func TestMatchDirectory_Venv(t *testing.T) {
	_, ok := MatchDirectory("venv")
	if !ok {
		t.Fatal("expected venv to match")
	}
}

func TestMatchDirectory_PythonCache(t *testing.T) {
	cat, ok := MatchDirectory("__pycache__")
	if !ok {
		t.Fatal("expected __pycache__ to match")
	}
	if cat.Key != "python-cache" {
		t.Errorf("expected key python-cache, got %s", cat.Key)
	}
}

func TestMatchDirectory_PytestCache(t *testing.T) {
	_, ok := MatchDirectory(".pytest_cache")
	if !ok {
		t.Fatal("expected .pytest_cache to match")
	}
}

func TestMatchDirectory_MypyCache(t *testing.T) {
	_, ok := MatchDirectory(".mypy_cache")
	if !ok {
		t.Fatal("expected .mypy_cache to match")
	}
}

func TestMatchDirectory_Tox(t *testing.T) {
	_, ok := MatchDirectory(".tox")
	if !ok {
		t.Fatal("expected .tox to match")
	}
}

func TestMatchDirectory_JsBuild(t *testing.T) {
	cat, ok := MatchDirectory("dist")
	if !ok {
		t.Fatal("expected dist to match")
	}
	if cat.Key != "js-build" {
		t.Errorf("expected key js-build, got %s", cat.Key)
	}
}

func TestMatchDirectory_Build(t *testing.T) {
	_, ok := MatchDirectory("build")
	if !ok {
		t.Fatal("expected build to match")
	}
}

func TestMatchDirectory_RustTarget(t *testing.T) {
	_, ok := MatchDirectory("target")
	if !ok {
		t.Fatal("expected target to match")
	}
}

func TestMatchDirectory_NextCache(t *testing.T) {
	_, ok := MatchDirectory(".next")
	if !ok {
		t.Fatal("expected .next to match")
	}
}

func TestMatchDirectory_NuxtCache(t *testing.T) {
	_, ok := MatchDirectory(".nuxt")
	if !ok {
		t.Fatal("expected .nuxt to match")
	}
}

func TestMatchDirectory_GenericCache(t *testing.T) {
	_, ok := MatchDirectory(".cache")
	if !ok {
		t.Fatal("expected .cache to match")
	}
}

func TestMatchDirectory_Unknown(t *testing.T) {
	_, ok := MatchDirectory("unknown_folder")
	if ok {
		t.Fatal("expected unknown_folder not to match")
	}
}

func TestMatchDirectory_Empty(t *testing.T) {
	_, ok := MatchDirectory("")
	if ok {
		t.Fatal("expected empty string not to match")
	}
}

func TestMatchFile_Pyc(t *testing.T) {
	cat, ok := MatchFile("/path/to/file.pyc")
	if !ok {
		t.Fatal("expected .pyc to match")
	}
	if cat.Key != "python-bytecode" {
		t.Errorf("expected key python-bytecode, got %s", cat.Key)
	}
}

func TestMatchFile_Pyo(t *testing.T) {
	_, ok := MatchFile("/path/to/file.pyo")
	if !ok {
		t.Fatal("expected .pyo to match")
	}
}

func TestMatchFile_Py(t *testing.T) {
	_, ok := MatchFile("/path/to/file.py")
	if ok {
		t.Fatal("expected .py not to match (only .pyc/.pyo)")
	}
}

func TestMatchFile_UnknownExt(t *testing.T) {
	_, ok := MatchFile("/path/to/file.go")
	if ok {
		t.Fatal("expected .go not to match")
	}
}

func TestIncludeCategory_AllowAll(t *testing.T) {
	// Empty include set means allow all
	result := IncludeCategory("node-modules", nil, nil)
	if !result {
		t.Error("expected allow all with empty include set")
	}
}

func TestIncludeCategory_IncludeSet(t *testing.T) {
	inc := map[string]struct{}{"node-modules": {}}
	result := IncludeCategory("node-modules", inc, nil)
	if !result {
		t.Error("expected node-modules to be allowed")
	}
	result = IncludeCategory("python-venv", inc, nil)
	if result {
		t.Error("expected python-venv to be blocked")
	}
}

func TestIncludeCategory_ExcludeSet(t *testing.T) {
	exc := map[string]struct{}{"node-modules": {}}
	result := IncludeCategory("node-modules", nil, exc)
	if result {
		t.Error("expected node-modules to be blocked")
	}
	result = IncludeCategory("python-venv", nil, exc)
	if !result {
		t.Error("expected python-venv to be allowed")
	}
}

func TestIncludeCategory_ExcludeWins(t *testing.T) {
	inc := map[string]struct{}{"node-modules": {}}
	exc := map[string]struct{}{"node-modules": {}}
	result := IncludeCategory("node-modules", inc, exc)
	if result {
		t.Error("expected exclude to win over include")
	}
}

func TestNewDirCategory(t *testing.T) {
	cat := newDirCategory("test-cat", "Test", "A test category", "testdir1", "testdir2")
	if cat.Key != "test-cat" {
		t.Errorf("expected key test-cat, got %s", cat.Key)
	}
	if cat.Display != "Test" {
		t.Errorf("expected display Test, got %s", cat.Display)
	}
	if _, ok := cat.DirectoryNames["testdir1"]; !ok {
		t.Error("expected testdir1 in DirectoryNames")
	}
	if _, ok := cat.DirectoryNames["testdir2"]; !ok {
		t.Error("expected testdir2 in DirectoryNames")
	}
	if len(cat.FileExtensions) != 0 {
		t.Error("expected empty FileExtensions")
	}
}

func TestNewFileCategory(t *testing.T) {
	cat := newFileCategory("test-cat", "Test", "A test category", ".ext1", ".ext2")
	if cat.Key != "test-cat" {
		t.Errorf("expected key test-cat, got %s", cat.Key)
	}
	if _, ok := cat.FileExtensions[".ext1"]; !ok {
		t.Error("expected .ext1 in FileExtensions")
	}
	if _, ok := cat.FileExtensions[".ext2"]; !ok {
		t.Error("expected .ext2 in FileExtensions")
	}
	if len(cat.DirectoryNames) != 0 {
		t.Error("expected empty DirectoryNames")
	}
}
