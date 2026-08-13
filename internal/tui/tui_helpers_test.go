package tui

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

// ── pure helpers ──────────────────────────────────────────────────

func TestSplitRelativePath(t *testing.T) {
	tests := []struct {
		root, target string
		want         []string
	}{
		{"/a", "/a/b", []string{"b"}},
		{"/a", "/a/b/c", []string{"b", "c"}},
		{"/a", "/a", nil},
		{"/a", "/a/x/y/z", []string{"x", "y", "z"}},
	}
	for _, tc := range tests {
		got := splitRelativePath(tc.root, tc.target)
		if fmt.Sprint(got) != fmt.Sprint(tc.want) {
			t.Errorf("splitRelativePath(%q, %q) = %v, want %v",
				tc.root, tc.target, got, tc.want)
		}
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		root     string
		segments []string
		want     string
	}{
		{"/a", []string{"b"}, "/a/b"},
		{"/a", []string{"b", "c"}, "/a/b/c"},
		{"/a", []string{}, "/a"},
	}
	for _, tc := range tests {
		got := joinPath(tc.root, tc.segments)
		if got != tc.want {
			t.Errorf("joinPath(%q, %v) = %q, want %q",
				tc.root, tc.segments, got, tc.want)
		}
	}
}

func TestTrimForTree(t *testing.T) {
	tests := []struct {
		val  string
		max  int
		want string
	}{
		{"hello", 10, "hello"},         // shorter than max
		{"hello world", 5, "hell…"},    // 5-1=4 chars + …
		{"hi", 1, "h"},                  // max=1 → first char only
		{"a", 0, ""},                    // max=0 → empty
	}
	for _, tc := range tests {
		got := trimForTree(tc.val, tc.max)
		if got != tc.want {
			t.Errorf("trimForTree(%q, %d) = %q, want %q",
				tc.val, tc.max, got, tc.want)
		}
	}
}

func TestHumanizeBytes(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
		{1099511627776, "1.0 TiB"},
	}
	for _, tc := range tests {
		got := humanizeBytes(tc.size)
		if got != tc.want {
			t.Errorf("humanizeBytes(%d) = %q, want %q", tc.size, got, tc.want)
		}
	}
}

func TestHumanizeTimestamp(t *testing.T) {
	mar1 := time.Date(2026, 3, 1, 14, 30, 0, 0, time.UTC)
	if got := humanizeTimestamp(mar1); got != "2026-03-01 14:30" {
		t.Errorf("got %q", got)
	}
	zero := time.Time{}
	if got := humanizeTimestamp(zero); got != "-" {
		t.Errorf("got %q", got)
	}
}

// ── selectionNode helpers ─────────────────────────────────────────

func makeLeaf(path string, bytes int64, selected bool) *selectionNode {
	return &selectionNode{
		kind:     nodeCandidate,
		label:    filepath.Base(path),
		path:     path,
		bytes:    bytes,
		selected: selected,
	}
}

func makeFolder(path string, children ...*selectionNode) *selectionNode {
	f := &selectionNode{
		kind:     nodeFolder,
		label:    filepath.Base(path),
		path:     path,
		children: children,
	}
	for _, c := range children {
		c.parent = f
	}
	return f
}

func TestAllSelected(t *testing.T) {
	allOn := []*selectionNode{makeLeaf("/a/x", 1, true), makeLeaf("/a/y", 2, true)}
	allOff := []*selectionNode{makeLeaf("/a/x", 1, false)}
	mixed := []*selectionNode{makeLeaf("/a/x", 1, true), makeLeaf("/a/y", 2, false)}

	if !allSelected(allOn) {
		t.Error("allSelected(all on) should be true")
	}
	if allSelected(allOff) {
		t.Error("allSelected(all off) should be false")
	}
	if allSelected(mixed) {
		t.Error("allSelected(mixed) should be false")
	}
}

func TestHasAnySelectedLeaf(t *testing.T) {
	leafOn := makeLeaf("/a/x", 1, true)
	leafOff := makeLeaf("/a/x", 1, false)

	if !hasAnySelectedLeaf(leafOn) {
		t.Error("leaf on should have selected leaf")
	}
	if hasAnySelectedLeaf(leafOff) {
		t.Error("leaf off should not have selected leaf")
	}

	// folder with mixed children
	folder := makeFolder("/a",
		makeLeaf("/a/x", 1, true),
		makeLeaf("/a/y", 2, false),
	)
	if !hasAnySelectedLeaf(folder) {
		t.Error("folder with one selected child should return true")
	}

	// folder with all children selected
	folder2 := makeFolder("/a",
		makeLeaf("/a/x", 1, true),
		makeLeaf("/a/y", 2, true),
	)
	if !hasAnySelectedLeaf(folder2) {
		t.Error("folder with all selected children should return true")
	}

	// folder with no children selected
	folder3 := makeFolder("/a",
		makeLeaf("/a/x", 1, false),
		makeLeaf("/a/y", 2, false),
	)
	if hasAnySelectedLeaf(folder3) {
		t.Error("folder with no selected children should return false")
	}
}

func TestHasAnySelectedChild(t *testing.T) {
	// Leaf has no children → hasAnySelectedChild = false (loop doesn't run)
	leafOn := makeLeaf("/a/x", 1, true)
	if hasAnySelectedChild(leafOn) {
		t.Error("leaf should not have any selected child (no children)")
	}

	// Folder with selected leaf child
	folder := makeFolder("/a", makeLeaf("/a/x", 1, true))
	if !hasAnySelectedChild(folder) {
		t.Error("folder with selected child should return true")
	}

	// Folder with unselected leaf child
	folder2 := makeFolder("/a", makeLeaf("/a/x", 1, false))
	if hasAnySelectedChild(folder2) {
		t.Error("folder with unselected child should return false")
	}
}

func TestHasAllSelectedChildren(t *testing.T) {
	// Leaf has no children → vacuous truth → true
	leafSel := makeLeaf("/a/x", 1, true)
	if !hasAllSelectedChildren(leafSel) {
		t.Error("leaf (no children) should be vacuously all selected")
	}

	leafUnsel := makeLeaf("/a/x", 1, false)
	if !hasAllSelectedChildren(leafUnsel) {
		t.Error("leaf (no children) is vacuously all selected regardless of selected state")
	}

	// Folder: all children selected
	folderAll := makeFolder("/a",
		makeLeaf("/a/x", 1, true),
		makeLeaf("/a/y", 2, true),
	)
	if !hasAllSelectedChildren(folderAll) {
		t.Error("folder with all selected children should return true")
	}

	// Folder: one not selected
	folderMix := makeFolder("/a",
		makeLeaf("/a/x", 1, true),
		makeLeaf("/a/y", 2, false),
	)
	if hasAllSelectedChildren(folderMix) {
		t.Error("folder with one unselected should return false")
	}
}

func TestSelectionPrefix(t *testing.T) {
	// leaf selected
	leafSel := makeLeaf("/a/x", 1, true)
	if selectionPrefix(leafSel) != "●" {
		t.Errorf("selected leaf prefix = %q", selectionPrefix(leafSel))
	}

	// leaf unselected
	leafOff := makeLeaf("/a/x", 1, false)
	if selectionPrefix(leafOff) != "○" {
		t.Errorf("unselected leaf prefix = %q", selectionPrefix(leafOff))
	}

	// folder all selected
	folderAll := makeFolder("/a",
		makeLeaf("/a/x", 1, true),
		makeLeaf("/a/y", 2, true),
	)
	if selectionPrefix(folderAll) != "●" {
		t.Errorf("folder all selected prefix = %q", selectionPrefix(folderAll))
	}

	// folder mixed
	folderMix := makeFolder("/a",
		makeLeaf("/a/x", 1, true),
		makeLeaf("/a/y", 2, false),
	)
	if selectionPrefix(folderMix) != "◐" {
		t.Errorf("folder mixed prefix = %q", selectionPrefix(folderMix))
	}

	// folder all unselected
	folderOff := makeFolder("/a",
		makeLeaf("/a/x", 1, false),
		makeLeaf("/a/y", 2, false),
	)
	if selectionPrefix(folderOff) != "○" {
		t.Errorf("folder all unselected prefix = %q", selectionPrefix(folderOff))
	}
}

func TestToggleSelection(t *testing.T) {
	n := makeLeaf("/a/x", 100, true)
	toggleSelection(n, false)
	if n.selected {
		t.Error("leaf should be unselected after toggleSelection(false)")
	}

	n2 := makeLeaf("/a/y", 50, false)
	toggleSelection(n2, true)
	if !n2.selected {
		t.Error("leaf should be selected after toggleSelection(true)")
	}
}

func TestCollectSelection(t *testing.T) {
	// Build: /a (folder) ── /a/x (leaf, selected)
	//                  ── /a/y (leaf, unselected)
	root := makeFolder("/a",
		makeLeaf("/a/x", 100, true),
		makeLeaf("/a/y", 50, false),
	)
	sel := collectSelection([]*selectionNode{root})
	if sel.SelectedBytes != 100 {
		t.Errorf("SelectedBytes = %d, want 100", sel.SelectedBytes)
	}
	if len(sel.ExcludedPaths) != 1 || sel.ExcludedPaths[0] != "/a/y" {
		t.Errorf("ExcludedPaths = %v, want [/a/y]", sel.ExcludedPaths)
	}
}

func TestCollectSelectionAllSelected(t *testing.T) {
	root := makeFolder("/a",
		makeLeaf("/a/x", 100, true),
		makeLeaf("/a/y", 200, true),
	)
	sel := collectSelection([]*selectionNode{root})
	if sel.SelectedBytes != 300 {
		t.Errorf("SelectedBytes = %d, want 300", sel.SelectedBytes)
	}
	if len(sel.ExcludedPaths) != 0 || len(sel.ExcludedGroups) != 0 {
		t.Errorf("no exclusions expected, got paths=%v groups=%v",
			sel.ExcludedPaths, sel.ExcludedGroups)
	}
}

func TestRefreshAncestors(t *testing.T) {
	// /a ── /a/x (folder) ── /a/x/y (leaf, on)
	leaf := makeLeaf("/a/x/y", 10, true)
	mid := makeFolder("/a/x", leaf)
	root := makeFolder("/a", mid)

	// leaf off → parents should be unselected
	leaf.selected = false
	toggleSelection(leaf, false)
	if root.selected || mid.selected {
		t.Error("unselected leaf should propagate false to ancestors")
	}

	// leaf on → ancestors should be selected
	toggleSelection(leaf, true)
	if !root.selected || !mid.selected {
		t.Error("selected leaf should propagate true to ancestors")
	}
}

func TestCollectSelectionExcludedGroup(t *testing.T) {
	// All children unselected → parent becomes excluded group
	root := makeFolder("/a",
		makeLeaf("/a/x", 100, false),
		makeLeaf("/a/y", 50, false),
	)
	sel := collectSelection([]*selectionNode{root})
	if sel.SelectedBytes != 0 {
		t.Errorf("SelectedBytes = %d, want 0", sel.SelectedBytes)
	}
	if len(sel.ExcludedGroups) != 1 || sel.ExcludedGroups[0] != "/a" {
		t.Errorf("ExcludedGroups = %v, want [/a]", sel.ExcludedGroups)
	}
}

func TestSelectionStateText(t *testing.T) {
	tests := []struct {
		node *selectionNode
		want string
	}{
		{makeLeaf("/x", 1, true), "selected"},
		{makeLeaf("/x", 1, false), "not selected"},
	}
	for _, tc := range tests {
		got := selectionStateText(tc.node)
		if got != tc.want {
			t.Errorf("selectionStateText(%v) = %q, want %q", tc.node.path, got, tc.want)
		}
	}
}
