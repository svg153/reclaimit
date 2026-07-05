package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/rivo/tview"
)

func TestBuildSelectionTreeSortsGroupsByBytes(t *testing.T) {
	report := Report{
		Root: "/root",
		Candidates: []Candidate{
			{Path: "/root/a/node_modules", Bytes: 100, IsDir: true},
			{Path: "/root/b/.venv", Bytes: 200, IsDir: true},
		},
	}

	roots := buildSelectionTree(report)
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	if roots[0].path != "/root/b" {
		t.Fatalf("expected /root/b first, got %q", roots[0].path)
	}
	if roots[0].kind != nodeFolder {
		t.Fatalf("expected root node to be a folder, got %q", roots[0].kind)
	}
	if len(roots[0].children) != 1 || roots[0].children[0].kind != nodeCandidate {
		t.Fatalf("expected /root/b to have one candidate child")
	}
	if roots[0].candidateCount != 1 {
		t.Fatalf("expected candidate count 1, got %d", roots[0].candidateCount)
	}
}

func TestDescribeNodeShowsPathAndType(t *testing.T) {
	node := &selectionNode{
		kind:       nodeCandidate,
		path:       "/root/repo/node_modules",
		bytes:      1024,
		modifiedAt: time.Date(2026, 5, 21, 1, 2, 0, 0, time.UTC),
		cand:       &Candidate{CategoryKey: "node-modules", IsDir: true, Description: "safe to reinstall"},
	}

	info := describeNode(node)
	for _, want := range []string{"Deletion candidate", "/root/repo/node_modules", "node-modules", "Last modified"} {
		if !strings.Contains(info, want) {
			t.Fatalf("expected info string to contain %q, got %q", want, info)
		}
	}
}

func TestDescribeContextNodeClarifiesNoRepoDeletion(t *testing.T) {
	node := &selectionNode{
		kind:           nodeFolder,
		label:          "repo",
		path:           "/root/repo",
		bytes:          2048,
		modifiedAt:     time.Date(2026, 5, 21, 1, 2, 0, 0, time.UTC),
		candidateCount: 3,
	}

	info := describeNode(node)
	for _, want := range []string{"Context folder", "never deletes the folder itself", "/root/repo", "Latest modification"} {
		if !strings.Contains(info, want) {
			t.Fatalf("expected info string to contain %q, got %q", want, info)
		}
	}
}

func TestRenderNodeLabelUsesIcons(t *testing.T) {
	folder := &selectionNode{
		kind:           nodeFolder,
		label:          "repo",
		bytes:          2048,
		selected:       true,
		candidateCount: 3,
		children:       []*selectionNode{{selected: true}},
	}
	candidate := &selectionNode{
		kind:     nodeCandidate,
		label:    ".venv",
		bytes:    1024,
		selected: true,
		cand:     &Candidate{CategoryKey: "python-venv", IsDir: true},
	}

	if got := renderNodeLabel(folder); !strings.Contains(got, "📁") || !strings.Contains(got, "●") {
		t.Fatalf("expected folder label to contain folder and selection icons, got %q", got)
	}
	if got := renderNodeLabel(candidate); !strings.Contains(got, "🧹") || !strings.Contains(got, "●") {
		t.Fatalf("expected candidate label to contain candidate and selection icons, got %q", got)
	}
}

func TestToggleSelectionAndCollectSelection(t *testing.T) {
	roots := []*selectionNode{
		{
			kind:     nodeFolder,
			label:    "repo",
			path:     "/root/repo",
			selected: true,
			children: []*selectionNode{
				{kind: nodeCandidate, label: ".venv", path: "/root/repo/.venv", bytes: 100, selected: true},
				{kind: nodeCandidate, label: "node_modules", path: "/root/repo/node_modules", bytes: 200, selected: true},
			},
		},
	}
	for _, child := range roots[0].children {
		child.parent = roots[0]
	}

	toggleSelection(roots[0].children[1], false)

	if allSelected(roots) {
		t.Fatalf("expected roots not all selected after toggling one child off")
	}
	if got := selectionPrefix(roots[0]); got != "◐" {
		t.Fatalf("expected partial selection prefix, got %q", got)
	}

	snapshot := collectSelection(roots)
	if snapshot.SelectedBytes != 100 {
		t.Fatalf("expected selected bytes 100, got %d", snapshot.SelectedBytes)
	}
	if len(snapshot.ExcludedGroups) != 0 || len(snapshot.ExcludedPaths) != 1 {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
}

func TestSelectionHelpers(t *testing.T) {
	root := &selectionNode{
		kind: nodeFolder,
		children: []*selectionNode{
			{kind: nodeCandidate, selected: true},
			{kind: nodeCandidate, selected: false},
		},
	}
	if !hasAnySelectedChild(root) {
		t.Fatalf("expected any selected child")
	}
	if hasAllSelectedChildren(root) {
		t.Fatalf("expected not all selected")
	}
	if !hasAnySelectedLeaf(root) {
		t.Fatalf("expected selected leaf")
	}
}

func TestPathHelpersAndTrim(t *testing.T) {
	parts := splitRelativePath("/root", "/root/a/b")
	if len(parts) != 2 || parts[0] != "a" || parts[1] != "b" {
		t.Fatalf("unexpected path parts: %#v", parts)
	}
	if got := joinPath("/root", []string{"a", "b"}); got != "/root/a/b" {
		t.Fatalf("unexpected joined path %q", got)
	}
	if got := trimForTree("0123456789", 5); got != "0123…" {
		t.Fatalf("unexpected trimmed value %q", got)
	}
}

func TestTreeNodeHelpers(t *testing.T) {
	root := tview.NewTreeNode("root")
	child := tview.NewTreeNode("child")
	grandchild := tview.NewTreeNode("grandchild")
	child.AddChild(grandchild)
	root.AddChild(child)

	if got := firstSelectableTreeNode(root); got != child {
		t.Fatalf("expected first child as first selectable node")
	}
	if got := findParentTreeNode(root, grandchild); got != child {
		t.Fatalf("expected child as parent node")
	}
}

func TestNodeColorAndStateText(t *testing.T) {
	node := &selectionNode{
		kind:     nodeCandidate,
		selected: true,
		cand:     &Candidate{IsDir: false},
	}
	if state := selectionStateText(node); state != "selected" {
		t.Fatalf("expected selected, got %q", state)
	}
	if color := nodeColor(node); color == 0 {
		t.Fatalf("expected non-zero color")
	}
}

func TestBuildTreeNodeAndRefresh(t *testing.T) {
	root := &selectionNode{
		kind:           nodeFolder,
		label:          "repo",
		path:           "/root/repo",
		bytes:          500,
		selected:       true,
		candidateCount: 2,
		children: []*selectionNode{
			{
				kind:     nodeCandidate,
				label:    ".venv",
				path:     "/root/repo/.venv",
				bytes:    300,
				selected: true,
				cand:     &Candidate{CategoryKey: "python-venv", IsDir: true},
			},
			{
				kind:     nodeCandidate,
				label:    "node_modules",
				path:     "/root/repo/node_modules",
				bytes:    200,
				selected: true,
				cand:     &Candidate{CategoryKey: "node-modules", IsDir: true},
			},
		},
	}
	root.children[0].parent = root
	root.children[1].parent = root

	treeNode := buildTreeNode(root)
	if len(treeNode.GetChildren()) != 2 {
		t.Fatalf("expected two tree children")
	}

	toggleSelection(root.children[1], false)
	refreshTreeNodes(treeNode)
	if got := treeNode.GetText(); !strings.Contains(got, "◐") {
		t.Fatalf("expected refreshed tree text to show partial selection, got %q", got)
	}
}

func TestSortTreeFoldersBeforeCandidates(t *testing.T) {
	parent := &selectionNode{
		kind: nodeFolder,
		children: []*selectionNode{
			{kind: nodeCandidate, label: "b", bytes: 10, path: "/b"},
			{kind: nodeFolder, label: "a", bytes: 5, path: "/a"},
		},
	}
	sortTree(parent)
	if parent.children[0].kind != nodeFolder {
		t.Fatalf("expected folder before candidate after sort")
	}
}
