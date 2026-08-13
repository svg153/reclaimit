package tui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/svg153/reclaimit/internal/scanner"
)

type selectionNodeKind string

const (
	nodeFolder    selectionNodeKind = "folder"
	nodeCandidate selectionNodeKind = "candidate"
)


// Aliases to keep the TUI interface clean.
type Candidate = scanner.Candidate
type Report = scanner.Report

type selectionNode struct {
	kind           selectionNodeKind
	label          string
	path           string
	bytes          int64
	selected       bool
	parent         *selectionNode
	children       []*selectionNode
	cand           *Candidate
	modifiedAt     time.Time
	candidateCount int
}

type Selection struct {
	SelectedBytes int64
	ExcludedGroups []string
	ExcludedPaths  []string
	Saved          bool
}

func Run(report Report) (Selection, error) {
	roots := buildSelectionTree(report)

	app := tview.NewApplication()
	tree := tview.NewTreeView().SetGraphics(true)
	tree.SetGraphicsColor(tcell.ColorCadetBlue)
	tree.SetBorder(true)
	tree.SetBorderPadding(0, 0, 1, 1)
	tree.SetTitle(" Cleanup tree ")
	details := tview.NewTextView()
	details.SetDynamicColors(true)
	details.SetWrap(true)
	details.SetBorder(true)
	details.SetBorderPadding(0, 0, 1, 1)
	details.SetTitle(" Details ")
	footer := tview.NewTextView()
	footer.SetDynamicColors(true)
	footer.SetWrap(true)
	footer.SetBorder(true)
	footer.SetBorderPadding(0, 0, 1, 1)
	footer.SetTitle(" Legend / keys ")
	footer.SetText("States:  ● selected   ◐ partial   ○ off\nNodes:   📁 context folder   🧹 directory target   📄 file target\nKeys:    j/k or arrows move   space toggle   enter/right expand   left collapse   a all   q save   esc cancel")

	rootNode := tview.NewTreeNode("🏠 " + report.Root).SetColor(tcell.ColorLightSlateGray).SetExpanded(true)
	for _, root := range roots {
		rootNode.AddChild(buildTreeNode(root))
	}
	tree.SetRoot(rootNode).SetCurrentNode(firstSelectableTreeNode(rootNode))

	updateDetails := func(treeNode *tview.TreeNode) {
		if treeNode == nil {
			return
		}
		ref, ok := treeNode.GetReference().(*selectionNode)
		if !ok || ref == nil {
			details.SetText("Select a node to inspect it.")
			return
		}
		details.SetText(describeNode(ref))
	}
	updateDetails(tree.GetCurrentNode())

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		ref, ok := node.GetReference().(*selectionNode)
		if ok && ref != nil && len(ref.children) > 0 {
			node.SetExpanded(!node.IsExpanded())
		}
		updateDetails(node)
	})

	tree.SetChangedFunc(func(node *tview.TreeNode) {
		updateDetails(node)
	})

	mainFlex := tview.NewFlex().
		AddItem(tree, 0, 2, true).
		AddItem(details, 0, 1, false)
	rootFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainFlex, 0, 1, true).
		AddItem(footer, 5, 0, false)

	var snapshot Selection

	refreshUI := func() {
		refreshTreeNodes(rootNode)
		updateDetails(tree.GetCurrentNode())
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		current := tree.GetCurrentNode()
		ref, _ := current.GetReference().(*selectionNode)

		switch event.Key() {
		case tcell.KeyEscape:
			app.Stop()
			return nil
		case tcell.KeyRight:
			if current != nil {
				current.SetExpanded(true)
			}
			return nil
		case tcell.KeyLeft:
			if current != nil && current.IsExpanded() {
				current.SetExpanded(false)
				return nil
			}
			parent := findParentTreeNode(rootNode, current)
			if parent != nil {
				tree.SetCurrentNode(parent)
			}
			return nil
		case tcell.KeyEnter:
			if current != nil {
				current.SetExpanded(!current.IsExpanded())
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				snapshot = collectSelection(roots)
				snapshot.Saved = true
				app.Stop()
				return nil
			case ' ':
				if ref != nil {
					toggleSelection(ref, !ref.selected)
					refreshUI()
				}
				return nil
			case 'a':
				next := !allSelected(roots)
				for _, root := range roots {
					toggleSelection(root, next)
				}
				refreshUI()
				return nil
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			case 'l':
				if current != nil {
					current.SetExpanded(true)
				}
				return nil
			case 'h':
				if current != nil && current.IsExpanded() {
					current.SetExpanded(false)
					return nil
				}
				parent := findParentTreeNode(rootNode, current)
				if parent != nil {
					tree.SetCurrentNode(parent)
				}
				return nil
			}
		}
		return event
	})

	if err := app.SetRoot(rootFlex, true).Run(); err != nil {
		return Selection{}, err
	}
	return snapshot, nil
}

func buildSelectionTree(report Report) []*selectionNode {
	roots := map[string]*selectionNode{}
	for _, candidate := range report.Candidates {
		candidateCopy := candidate
		segments := splitRelativePath(report.Root, candidate.Path)
		if len(segments) == 0 {
			segments = []string{filepath.Base(candidate.Path)}
		}

		rootLabel := segments[0]
		rootNode := roots[rootLabel]
		if rootNode == nil {
			rootNode = &selectionNode{
				kind:       nodeFolder,
				label:      rootLabel,
				path:       filepath.Join(report.Root, rootLabel),
				selected:   true,
				modifiedAt: candidate.ModifiedAt,
			}
			roots[rootLabel] = rootNode
		}

		current := rootNode
		current.bytes += candidate.Bytes
		current.candidateCount++
		if candidate.ModifiedAt.After(current.modifiedAt) {
			current.modifiedAt = candidate.ModifiedAt
		}
		current.path = joinPath(report.Root, segments[:1])

		for depth := 1; depth < len(segments); depth++ {
			segment := segments[depth]
			child := findOrCreateChild(current, segment, joinPath(report.Root, segments[:depth+1]), depth == len(segments)-1)
			child.bytes += candidate.Bytes
			child.candidateCount++
			if candidate.ModifiedAt.After(child.modifiedAt) {
				child.modifiedAt = candidate.ModifiedAt
			}
			current = child
		}
		current.kind = nodeCandidate
		current.selected = true
		current.cand = &candidateCopy
		current.modifiedAt = candidate.ModifiedAt
	}

	rootList := make([]*selectionNode, 0, len(roots))
	for _, root := range roots {
		sortTree(root)
		rootList = append(rootList, root)
	}
	sort.Slice(rootList, func(i, j int) bool {
		if rootList[i].bytes == rootList[j].bytes {
			return rootList[i].path < rootList[j].path
		}
		return rootList[i].bytes > rootList[j].bytes
	})
	return rootList
}

func buildTreeNode(node *selectionNode) *tview.TreeNode {
	treeNode := tview.NewTreeNode(renderNodeLabel(node)).
		SetReference(node).
		SetSelectable(true).
		SetExpanded(node.kind == nodeFolder && node.parent == nil)
	treeNode.SetColor(nodeColor(node))
	nodeTreeRef(node, treeNode)

	for _, child := range node.children {
		treeNode.AddChild(buildTreeNode(child))
	}
	return treeNode
}

func nodeTreeRef(node *selectionNode, treeNode *tview.TreeNode) {
	node.label = strings.TrimSpace(node.label)
	treeNode.SetReference(node)
}

func refreshTreeNodes(treeNode *tview.TreeNode) {
	ref, _ := treeNode.GetReference().(*selectionNode)
	if ref != nil {
		treeNode.SetText(renderNodeLabel(ref))
		treeNode.SetColor(nodeColor(ref))
	}
	for _, child := range treeNode.GetChildren() {
		refreshTreeNodes(child)
	}
}

func renderNodeLabel(node *selectionNode) string {
	state := selectionPrefix(node)
	if node.kind == nodeFolder {
		name := trimForTree(node.label+"/", 28)
		return fmt.Sprintf("%s 📁 %-28s  %10s  %2d tgt", state, name, humanizeBytes(node.bytes), node.candidateCount)
	}

	icon := "🧹"
	if node.cand != nil && !node.cand.IsDir {
		icon = "📄"
	}
	category := ""
	if node.cand != nil {
		category = node.cand.CategoryKey
	}
	name := trimForTree(node.label, 28)
	return fmt.Sprintf("%s %s %-28s  %10s  %s", state, icon, name, humanizeBytes(node.bytes), trimForTree(category, 14))
}

func findOrCreateChild(parent *selectionNode, label, path string, isLeaf bool) *selectionNode {
	for _, child := range parent.children {
		if child.label == label && child.path == path {
			return child
		}
	}

	kind := nodeFolder
	if isLeaf {
		kind = nodeCandidate
	}
	child := &selectionNode{
		kind:     kind,
		label:    label,
		path:     path,
		selected: true,
		parent:   parent,
	}
	parent.children = append(parent.children, child)
	return child
}

func sortTree(node *selectionNode) {
	for _, child := range node.children {
		sortTree(child)
	}
	sort.Slice(node.children, func(i, j int) bool {
		if node.children[i].kind != node.children[j].kind {
			return node.children[i].kind == nodeFolder
		}
		if node.children[i].bytes == node.children[j].bytes {
			return node.children[i].path < node.children[j].path
		}
		return node.children[i].bytes > node.children[j].bytes
	})
}

func splitRelativePath(root, target string) []string {
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == "." {
		return nil
	}
	return strings.Split(relative, string(filepath.Separator))
}

func joinPath(root string, segments []string) string {
	all := append([]string{root}, segments...)
	return filepath.Join(all...)
}

func toggleSelection(node *selectionNode, selected bool) {
	node.selected = selected
	for _, child := range node.children {
		toggleSelection(child, selected)
	}
	refreshAncestors(node.parent)
}

func refreshAncestors(node *selectionNode) {
	if node == nil {
		return
	}
	node.selected = true
	for _, child := range node.children {
		if !child.selected {
			node.selected = false
			break
		}
	}
	refreshAncestors(node.parent)
}

func allSelected(roots []*selectionNode) bool {
	for _, root := range roots {
		if !root.selected {
			return false
		}
	}
	return true
}

func collectSelection(roots []*selectionNode) Selection {
	var snapshot Selection
	for _, root := range roots {
		collectNodeSelection(root, &snapshot)
	}
	sort.Strings(snapshot.ExcludedGroups)
	sort.Strings(snapshot.ExcludedPaths)
	return snapshot
}

func collectNodeSelection(node *selectionNode, snapshot *Selection) {
	if len(node.children) == 0 {
		if node.selected {
			snapshot.SelectedBytes += node.bytes
			return
		}
		snapshot.ExcludedPaths = append(snapshot.ExcludedPaths, node.path)
		return
	}

	if !hasAnySelectedLeaf(node) {
		snapshot.ExcludedGroups = append(snapshot.ExcludedGroups, node.path)
		return
	}

	for _, child := range node.children {
		collectNodeSelection(child, snapshot)
	}
}

func hasAnySelectedLeaf(node *selectionNode) bool {
	if len(node.children) == 0 {
		return node.selected
	}
	for _, child := range node.children {
		if hasAnySelectedLeaf(child) {
			return true
		}
	}
	return false
}

func hasAnySelectedChild(node *selectionNode) bool {
	for _, child := range node.children {
		if child.selected || hasAnySelectedLeaf(child) {
			return true
		}
	}
	return false
}

func hasAllSelectedChildren(node *selectionNode) bool {
	for _, child := range node.children {
		if !child.selected {
			return false
		}
		if len(child.children) > 0 && !hasAllSelectedChildren(child) {
			return false
		}
	}
	return true
}

func selectionPrefix(node *selectionNode) string {
	if len(node.children) == 0 {
		if node.selected {
			return "●"
		}
		return "○"
	}
	if hasAllSelectedChildren(node) {
		return "●"
	}
	if hasAnySelectedChild(node) {
		return "◐"
	}
	return "○"
}

func describeNode(node *selectionNode) string {
	var b strings.Builder
	if node.kind == nodeFolder {
		fmt.Fprintf(&b, "[deepskyblue::b]📁 Context folder[-:-:-]\n\n")
		fmt.Fprintf(&b, "[aqua]Selection[-]: [white]%s[-]\n", selectionStateText(node))
		fmt.Fprintf(&b, "[aqua]Path[-]: [white]%s[-]\n", node.path)
		fmt.Fprintf(&b, "[aqua]Candidates below[-]: [white]%d[-]\n", node.candidateCount)
		fmt.Fprintf(&b, "[aqua]Reclaimable below[-]: [white]%s[-]\n", humanizeBytes(node.bytes))
		fmt.Fprintf(&b, "[aqua]Latest modification[-]: [white]%s[-]\n\n", humanizeTimestamp(node.modifiedAt))
		b.WriteString("[white]Important:[-] toggling this node [white]never deletes the folder itself[-]. It only toggles the descendant cleanup candidates under it.\n")
		return b.String()
	}

	kind := "directory"
	if node.cand != nil && !node.cand.IsDir {
		kind = "file"
	}
	icon := "🧹"
	if kind == "file" {
		icon = "📄"
	}
	fmt.Fprintf(&b, "[mediumpurple::b]%s Deletion candidate[-:-:-]\n\n", icon)
	fmt.Fprintf(&b, "[aqua]Selection[-]: [white]%s[-]\n", selectionStateText(node))
	fmt.Fprintf(&b, "[aqua]Exact target path[-]: [white]%s[-]\n", node.path)
	if node.cand != nil {
		fmt.Fprintf(&b, "[aqua]Category[-]: [white]%s[-]\n", node.cand.CategoryKey)
		fmt.Fprintf(&b, "[aqua]Candidate type[-]: [white]%s[-]\n", kind)
		fmt.Fprintf(&b, "[aqua]Reason[-]: [white]%s[-]\n", node.cand.Description)
	}
	fmt.Fprintf(&b, "[aqua]Candidate size[-]: [white]%s[-]\n", humanizeBytes(node.bytes))
	fmt.Fprintf(&b, "[aqua]Last modified[-]: [white]%s[-]\n\n", humanizeTimestamp(node.modifiedAt))
	b.WriteString("[white]If selected for cleanup:[-] this [white]exact candidate path[-] is eligible for deletion, not the whole repository root.\n")
	return b.String()
}

func selectionStateText(node *selectionNode) string {
	switch selectionPrefix(node) {
	case "●":
		return "selected"
	case "◐":
		return "partially selected"
	default:
		return "not selected"
	}
}

func nodeColor(node *selectionNode) tcell.Color {
	state := selectionPrefix(node)
	if node.kind == nodeFolder {
		switch state {
		case "●":
			return tcell.ColorDeepSkyBlue
		case "◐":
			return tcell.ColorLightSkyBlue
		default:
			return tcell.ColorSlateGray
		}
	}
	if node.cand != nil && !node.cand.IsDir {
		switch state {
		case "●":
			return tcell.ColorMediumPurple
		default:
			return tcell.ColorSlateGray
		}
	}
	switch state {
	case "●":
		return tcell.ColorTurquoise
	default:
		return tcell.ColorSlateGray
	}
}

func trimForTree(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}

func firstSelectableTreeNode(root *tview.TreeNode) *tview.TreeNode {
	for _, child := range root.GetChildren() {
		return child
	}
	return root
}

func findParentTreeNode(root, target *tview.TreeNode) *tview.TreeNode {
	if root == nil || target == nil {
		return nil
	}
	for _, child := range root.GetChildren() {
		if child == target {
			return root
		}
		if parent := findParentTreeNode(child, target); parent != nil {
			return parent
		}
	}
	return nil
}

func humanizeBytes(size int64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	value := float64(size)
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%d %s", size, units[unit])
	}
	return fmt.Sprintf("%.1f %s", value, units[unit])
}

func humanizeTimestamp(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.UTC().Format("2006-01-02 15:04")
}
