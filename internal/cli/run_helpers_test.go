package cli

import (
	"strings"
	"testing"

	"github.com/svg153/reclaimit/internal/scanner"
)

// TestRenderDeletionPreviewEmpty validates empty candidates message.
func TestRenderDeletionPreviewEmpty(t *testing.T) {
	out := RenderDeletionPreview(nil)
	if !strings.Contains(out, "No cleanup candidates selected") {
		t.Fatalf("expected empty message, got: %s", out)
	}
}

// TestRenderDeletionPreviewSingle validates single candidate preview.
func TestRenderDeletionPreviewSingle(t *testing.T) {
	out := RenderDeletionPreview([]scanner.Candidate{
		{Path: "/tmp/test", Bytes: 1024},
	})
	if !strings.Contains(out, "Deleting 1 candidate") {
		t.Fatalf("expected 'Deleting 1 candidates', got: %s", out)
	}
	if !strings.Contains(out, "/tmp/test") {
		t.Fatalf("expected path, got: %s", out)
	}
	if !strings.Contains(out, "1.0 KiB") {
		t.Fatalf("expected '1.0 KiB', got: %s", out)
	}
}

// TestRenderDeletionPreviewMultiple validates multiple candidates preview.
func TestRenderDeletionPreviewMultiple(t *testing.T) {
	out := RenderDeletionPreview([]scanner.Candidate{
		{Path: "/tmp/a", Bytes: 100},
		{Path: "/tmp/b", Bytes: 200},
		{Path: "/tmp/c", Bytes: 300},
	})
	if !strings.Contains(out, "Deleting 3 candidates") {
		t.Fatalf("expected 'Deleting 3 candidates', got: %s", out)
	}
	if !strings.Contains(out, "/tmp/a") || !strings.Contains(out, "/tmp/b") {
		t.Fatalf("expected all paths, got: %s", out)
	}
	if !strings.Contains(out, "600 B") {
		t.Fatalf("expected total bytes, got: %s", out)
	}
}
