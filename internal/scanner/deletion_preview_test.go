package scanner

import (
	"fmt"
	"strings"
	"testing"
)

func TestRenderDeletionPreviewEmpty(t *testing.T) {
	if got := RenderDeletionPreview(nil); got != "No cleanup candidates selected.\n" {
		t.Fatalf("unexpected empty preview: %q", got)
	}
}

func TestRenderDeletionPreviewListsAndLimitsCandidates(t *testing.T) {
	candidates := make([]Candidate, 22)
	for i := range candidates {
		candidates[i] = Candidate{Path: fmt.Sprintf("/tmp/candidate-%02d", i), Bytes: 1024}
	}
	preview := RenderDeletionPreview(candidates)
	for _, want := range []string{
		"Deleting 22 candidates totaling 22.0 KiB",
		"/tmp/candidate-00",
		"/tmp/candidate-19",
		"... and 2 more",
	} {
		if !strings.Contains(preview, want) {
			t.Fatalf("preview missing %q: %s", want, preview)
		}
	}
	if strings.Contains(preview, "/tmp/candidate-20") {
		t.Fatalf("preview exceeded display limit: %s", preview)
	}
}

func TestDeletionPreviewHumanizeBytes(t *testing.T) {
	for size, want := range map[int64]string{
		0:       "0 B",
		1:       "1 B",
		1 << 20: "1.0 MiB",
		1 << 40: "1.0 TiB",
	} {
		if got := humanizeBytes(size); got != want {
			t.Fatalf("humanizeBytes(%d) = %q, want %q", size, got, want)
		}
	}
}
