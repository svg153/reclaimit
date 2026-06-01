package reclaimit

func usageText(topic string) string {
	switch topic {
	case "analyze":
		return `reclaimit analyze

Generate a disk usage report and highlight reclaimable space.

Production usage:
  ./bin/reclaimit analyze --root "$HOME" --format markdown --out report.md

Important flags:
  --root PATH
  --format plain|markdown
  --group-mode repo|depth
  --group-depth N
  --exclude-group PATH
  --exclude-path PATH
  --out FILE
`
	case "clean":
		return `reclaimit clean

Remove the currently selected cleanup candidates.

Production usage:
  ./bin/reclaimit clean --root "$HOME" --include-category python-venv --yes

The command prints a deletion preview first and then emits a fresh post-clean report.

Important flags:
  --root PATH
  --include-category KEY
  --exclude-group PATH
  --exclude-path PATH
  --yes
`
	case "tui":
		return `reclaimit tui

Interactive terminal UI to review cleanup candidates as a path tree.
Context folders are shown separately from real deletion candidates.

Production usage:
  ./bin/reclaimit tui --root "$HOME" --format markdown --out report.md

Shortcuts:
  ↑/↓ or j/k   Move cursor
  →/l/Enter    Expand node
  ←/h          Collapse node / jump to parent
  Space        Toggle current node
  a            Toggle all
  q            Save selection and exit
  Esc          Exit without saving
`
	default:
		return `reclaimit

Analyze disk usage, detect reclaimable folders and interactively choose what to keep.

Usage:
  ./bin/reclaimit analyze [flags]
  ./bin/reclaimit tui [flags]
  ./bin/reclaimit clean [flags] --yes
  ./bin/reclaimit help [analyze|tui|clean]
  ./bin/reclaimit --version

Commands:
  analyze   Generate a plain-text or Markdown report
  tui       Open the interactive tree UI
  clean     Delete the currently selected candidates
  help      Show help for the CLI or a subcommand

Global flags:
  --root PATH
  --format plain|markdown
  --group-mode repo|depth
  --group-depth N
  --include-category KEY
  --exclude-category KEY
  --exclude-group PATH
  --exclude-path PATH
  --min-candidate-size BYTES
  --out FILE

Examples:
  ./bin/reclaimit analyze --root "$HOME" --format markdown --out report.md
  ./bin/reclaimit tui --root "$HOME"
  ./bin/reclaimit clean --root "$HOME" --include-category node-modules --yes
`
	}
}
