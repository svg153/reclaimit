package renderer

import (
	"testing"

	"github.com/svg153/reclaimit/internal/scanner"
)

func TestRenderReport_Plain(t *testing.T) {
	report := scanner.Report{
		Root:             "/tmp",
		TotalBytes:       1000,
		FilesystemBytes:  10000,
		AvailableBytes:   5000,
		Candidates:       []scanner.Candidate{},
		CandidateBytes:   0,
		SelectedBytes:    0,
	}

	out, err := RenderReport(report, "plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"Disk usage report",
		"Filesystem:",
		"Cleanup by category",
		"Cleanup by group",
		"Top cleanup candidates",
	}

	for _, exp := range expected {
		if !contains(out, exp) {
			t.Errorf("output missing %q\nOutput:\n%s", exp, out)
		}
	}
}

func TestRenderReport_Markdown(t *testing.T) {
	report := scanner.Report{
		Root:             "/tmp",
		TotalBytes:       1000,
		FilesystemBytes:  10000,
		AvailableBytes:   5000,
		CandidateBytes:   0,
		SelectedBytes:    0,
	}

	out, err := RenderReport(report, "markdown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !contains(out, "## Executive summary") {
		t.Errorf("markdown output missing header\nOutput:\n%s", out)
	}
}

func TestRenderReport_JSON(t *testing.T) {
	report := scanner.Report{
		Root:             "/tmp",
		TotalBytes:       1000,
		FilesystemBytes:  10000,
		AvailableBytes:   5000,
		CandidateBytes:   0,
		SelectedBytes:    0,
	}

	out, err := RenderReport(report, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !contains(out, "\"root\":") {
		t.Errorf("json output missing root field\nOutput:\n%s", out)
	}
}

func TestRenderReport_UnknownFormat(t *testing.T) {
	report := scanner.Report{Root: "/tmp"}
	_, err := RenderReport(report, "xml")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
