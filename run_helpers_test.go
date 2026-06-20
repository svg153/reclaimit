package reclaimit

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestRenderDeletionPreviewEmpty validates empty candidates message.
func TestRenderDeletionPreviewEmpty(t *testing.T) {
	out := renderDeletionPreview(nil)
	if !strings.Contains(out, "No cleanup candidates selected") {
		t.Fatalf("expected empty message, got: %s", out)
	}

	out = renderDeletionPreview([]Candidate{})
	if !strings.Contains(out, "No cleanup candidates selected") {
		t.Fatalf("expected empty message for empty slice, got: %s", out)
	}
}

// TestRenderDeletionPreviewSingle validates single candidate preview.
func TestRenderDeletionPreviewSingle(t *testing.T) {
	candidates := []Candidate{
		{Path: "/tmp/node_modules", Bytes: 1024},
	}
	out := renderDeletionPreview(candidates)
	if !strings.Contains(out, "Deleting 1 candidates") {
		t.Fatalf("expected 'Deleting 1 candidates', got: %s", out)
	}
	if !strings.Contains(out, "/tmp/node_modules") {
		t.Fatalf("expected path in preview, got: %s", out)
	}
	if !strings.Contains(out, "1.0 KiB") {
		t.Fatalf("expected '1.0 KiB', got: %s", out)
	}
}

// TestRenderDeletionPreviewMultiple validates multi-candidate preview with truncation.
func TestRenderDeletionPreviewMultiple(t *testing.T) {
	candidates := make([]Candidate, 25)
	for i := 0; i < 25; i++ {
		candidates[i] = Candidate{
			Path:  filepath.Join("/tmp", "node_modules", "pkg", string(rune('a'+i))),
			Bytes: int64(1000 + i),
		}
	}
	out := renderDeletionPreview(candidates)
	if !strings.Contains(out, "Deleting 25 candidates") {
		t.Fatalf("expected 'Deleting 25 candidates', got: %s", out)
	}
	if !strings.Contains(out, "... and 5 more") {
		t.Fatalf("expected truncation message, got: %s", out)
	}
}

// TestRenderDeletionPreviewTotal validates total bytes in preview.
func TestRenderDeletionPreviewTotal(t *testing.T) {
	candidates := []Candidate{
		{Path: "/tmp/a", Bytes: 1024},
		{Path: "/tmp/b", Bytes: 2048},
		{Path: "/tmp/c", Bytes: 3072},
	}
	out := renderDeletionPreview(candidates)
	if !strings.Contains(out, "6.0 KiB") {
		t.Fatalf("expected total '6.0 KiB', got: %s", out)
	}
}

// TestExitf validates exitf returns 1 and writes to writer.
func TestExitf(t *testing.T) {
	var buf strings.Builder
	code := exitf(&buf, "error: %s\n", "test")
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(buf.String(), "error: test") {
		t.Fatalf("expected error message, got: %s", buf.String())
	}
}

// TestWriteString validates writeString.
func TestWriteString(t *testing.T) {
	var buf strings.Builder
	err := writeString(&buf, "hello")
	if err != nil {
		t.Fatalf("writeString returned error: %v", err)
	}
	if buf.String() != "hello" {
		t.Fatalf("expected 'hello', got: %s", buf.String())
	}
}

// TestWritef validates writef with formatting.
func TestWritef(t *testing.T) {
	var buf strings.Builder
	err := writef(&buf, "count: %d\n", 42)
	if err != nil {
		t.Fatalf("writef returned error: %v", err)
	}
	if buf.String() != "count: 42\n" {
		t.Fatalf("expected 'count: 42\\n', got: %s", buf.String())
	}
}

// TestSumCandidateBytes validates sum function.
func TestSumCandidateBytes(t *testing.T) {
	candidates := []Candidate{
		{Bytes: 100},
		{Bytes: 200},
		{Bytes: 300},
	}
	if got := sumCandidateBytes(candidates); got != 600 {
		t.Fatalf("expected 600, got %d", got)
	}

	if got := sumCandidateBytes(nil); got != 0 {
		t.Fatalf("expected 0 for nil, got %d", got)
	}

	if got := sumCandidateBytes([]Candidate{}); got != 0 {
		t.Fatalf("expected 0 for empty, got %d", got)
	}
}

// TestSumBytes validates sum function for PathSize.
func TestSumBytes(t *testing.T) {
	items := []PathSize{
		{Bytes: 1000},
		{Bytes: 2000},
	}
	if got := sumBytes(items); got != 3000 {
		t.Fatalf("expected 3000, got %d", got)
	}
}

// TestListToSet validates set creation.
func TestListToSet(t *testing.T) {
	set := listToSet([]string{"a", "b", "c"})
	for _, key := range []string{"a", "b", "c"} {
		if _, ok := set[key]; !ok {
			t.Fatalf("expected set to contain %q", key)
		}
	}
	if _, ok := set["d"]; ok {
		t.Fatalf("expected set not to contain 'd'")
	}
}

// TestHasPathPrefix validates prefix matching.
func TestHasPathPrefix(t *testing.T) {
	if !hasPathPrefix("/tmp/root", "/tmp/root") {
		t.Fatalf("expected exact path to match prefix")
	}
	if !hasPathPrefix("/tmp/root/repo", "/tmp/root") {
		t.Fatalf("expected child path to match prefix")
	}
	if !hasPathPrefix("/tmp/root/repo/node_modules", "/tmp/root/repo") {
		t.Fatalf("expected grandchild to match prefix")
	}
	if hasPathPrefix("/tmp/other", "/tmp/root") {
		t.Fatalf("expected unrelated path not to match")
	}
	if hasPathPrefix("/tmp/root2", "/tmp/root") {
		t.Fatalf("expected similar prefix not to match")
	}
}

// TestIsGroupExcluded validates group exclusion logic.
func TestIsGroupExcluded(t *testing.T) {
	candidate := Candidate{Group: "/tmp/project", Path: "/tmp/project/node_modules"}

	if !isGroupExcluded(candidate, []string{"/tmp/project"}) {
		t.Fatalf("expected group to be excluded")
	}
	if isGroupExcluded(candidate, []string{"/tmp/other"}) {
		t.Fatalf("expected unrelated group not to exclude")
	}
	if isGroupExcluded(candidate, nil) {
		t.Fatalf("expected nil exclusions not to exclude")
	}
}

// TestIsGroupExcludedPathPrefix validates path prefix in group exclusion.
func TestIsGroupExcludedPathPrefix(t *testing.T) {
	candidate := Candidate{Path: "/tmp/excluded/node_modules"}

	if isGroupExcluded(candidate, []string{"/tmp/excluded"}) {
		// Should match because path starts with /tmp/excluded
	} else {
		t.Fatalf("expected path prefix to trigger exclusion")
	}
}
