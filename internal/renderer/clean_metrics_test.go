package renderer

import (
	"strings"
	"testing"

	"github.com/svg153/reclaimit/internal/scanner"
)

func TestRenderJSONIncludesCleanupVerification(t *testing.T) {
	output, err := RenderReport(scanner.Report{
		Command:                "clean",
		ExpectedDeletedBytes:   12,
		VerifiedDeletedBytes:   10,
		DeletedBytes:           8,
		SkippedCleanCandidates: 1,
		FailedCleanCandidates: 1,
	}, "json")
	if err != nil {
		t.Fatalf("RenderReport: %v", err)
	}
	for _, field := range []string{
		"expected_deleted_bytes",
		"verified_deleted_bytes",
		"skipped_clean_candidates",
		"failed_clean_candidates",
	} {
		if !strings.Contains(output, field) {
			t.Fatalf("JSON output missing %q: %s", field, output)
		}
	}
}
