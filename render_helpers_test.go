package main

import (
	"strings"
	"testing"
	"time"
)

func TestCandidateRowsIncludeMetadata(t *testing.T) {
	rows := candidateRows([]Candidate{{
		CategoryKey: "python-venv",
		Group:       "/root/repo",
		Path:        "/root/repo/.venv",
		Bytes:       1024,
		ModifiedAt:  time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
		IsDir:       true,
	}}, 10)

	row := strings.Join(rows, "")
	for _, want := range []string{"python-venv", "dir", "2026-05-20", "/root/repo/.venv"} {
		if !strings.Contains(row, want) {
			t.Fatalf("expected candidate row to contain %q, got %q", want, row)
		}
	}
}

func TestGroupRowsIncludeModifiedAt(t *testing.T) {
	rows := groupRows([]GroupSummary{{
		Group:      "/root/repo",
		Bytes:      2048,
		Count:      3,
		ModifiedAt: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
	}})

	row := strings.Join(rows, "")
	if !strings.Contains(row, "2026-05-20 12:00") {
		t.Fatalf("expected group row to include timestamp, got %q", row)
	}
}

func TestHumanizeTimestampZero(t *testing.T) {
	if got := humanizeTimestamp(time.Time{}); got != "-" {
		t.Fatalf("expected -, got %q", got)
	}
}
