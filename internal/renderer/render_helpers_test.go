package renderer

import (
	"strings"
	"testing"
	"time"

	"github.com/svg153/reclaimit/internal/scanner"
)

func TestCandidateRowsIncludeMetadata(t *testing.T) {
	items := []scanner.Candidate{
		{
			Path:        "/tmp/node_modules",
			Bytes:       1024,
			CategoryKey: "node-modules",
			ModifiedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			IsDir:       true,
		},
	}
	rows := candidateRows(items, 2)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !strings.Contains(rows[0], "node-modules") {
		t.Errorf("row missing path: %s", rows[0])
	}
	if !strings.Contains(rows[0], "node-modules") {
		t.Errorf("row missing node-modules: %s", rows[0])
	}
	if !strings.Contains(rows[0], "1.0 KiB") {
		t.Errorf("row missing bytes: %s", rows[0])
	}
}

func TestCandidateRowsEmpty(t *testing.T) {
	rows := candidateRows(nil, 10)
	if len(rows) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(rows))
	}
}

func TestCandidateRowsLimit(t *testing.T) {
	items := make([]scanner.Candidate, 5)
	for i := 0; i < 5; i++ {
		items[i] = scanner.Candidate{
			Path:  "/tmp/file-" + string(rune('0'+i)),
			Bytes: int64((i + 1) * 100),
		}
	}
	limited := candidateRows(items, 3)
	if len(limited) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(limited))
	}
}

func TestCandidateRowsSorting(t *testing.T) {
	items := []scanner.Candidate{
		{Path: "/tmp/a", Bytes: 100},
		{Path: "/tmp/b", Bytes: 500},
		{Path: "/tmp/c", Bytes: 300},
	}
	limited := candidateRows(items, 10)
	if len(limited) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(limited))
	}
	// Check sorting: largest first
	if !strings.Contains(limited[0], "/tmp/a") {
		t.Errorf("first row should be /tmp/a (sorted by path), got: %s", limited[0])
	}
	if !strings.Contains(limited[2], "/tmp/c") {
		t.Errorf("last row should be /tmp/c, got: %s", limited[2])
	}
}

func TestPathSizeRows(t *testing.T) {
	items := []scanner.PathSize{
		{Path: "/tmp/large", Bytes: 1048576},
		{Path: "/tmp/small", Bytes: 1024},
	}
	rows := pathSizeRows(items)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if !strings.Contains(rows[0], "large") {
		t.Errorf("first row missing large: %s", rows[0])
	}
}

func TestCategoryRows(t *testing.T) {
	items := []scanner.CategorySummary{
		{CategoryKey: "node-modules", Category: "node_modules", Bytes: 1000, Count: 5, Description: "JS dependencies"},
	}
	rows := categoryRows(items)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !strings.Contains(rows[0], "node-modules") {
		t.Errorf("row missing node-modules: %s", rows[0])
	}
}

func TestGroupRows(t *testing.T) {
	items := []scanner.GroupSummary{
		{Group: "/root/repo-a", Bytes: 2000, Count: 10, ModifiedAt: time.Now()},
	}
	rows := groupRows(items)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !strings.Contains(rows[0], "repo-a") {
		t.Errorf("row missing group: %s", rows[0])
	}
}

func TestCandidateKind(t *testing.T) {
	dirItem := scanner.Candidate{IsDir: true}
	if candidateKind(dirItem) != "dir" {
		t.Errorf("expected dir, got: %s", candidateKind(dirItem))
	}
	fileItem := scanner.Candidate{IsDir: false}
	if candidateKind(fileItem) != "file" {
		t.Errorf("expected file, got: %s", candidateKind(fileItem))
	}
}

func TestHumanizeTimestamp(t *testing.T) {
	zero := time.Time{}
	if humanizeTimestamp(zero) != "-" {
		t.Errorf("expected -, got: %s", humanizeTimestamp(zero))
	}
	ts := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	result := humanizeTimestamp(ts)
	if result != "2024-06-15 10:30" {
		t.Errorf("expected 2024-06-15 10:30, got: %s", result)
	}
}
