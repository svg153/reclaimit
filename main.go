package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const version = "0.1.0"

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	if value == "" {
		return errors.New("empty value is not allowed")
	}
	*s = append(*s, value)
	return nil
}

type config struct {
	command           string
	helpTopic         string
	root              string
	format            string
	groupMode         string
	groupDepth        int
	topFiles          int
	topGroups         int
	topEntries        int
	minCandidateSize  int64
	outFile           string
	includeCategories stringList
	excludeCategories stringList
	excludeGroups     stringList
	excludePaths      stringList
	yes               bool
}

func main() {
	cfg, err := parseConfig(os.Args[1:])
	if err != nil {
		exitf("error: %v\n", err)
	}

	if cfg.command == "help" {
		fmt.Print(usageText(cfg.helpTopic))
		return
	}
	if cfg.command == "version" {
		fmt.Printf("reclaimit %s\n", version)
		return
	}

	report, err := Analyze(cfg)
	if err != nil {
		exitf("error: %v\n", err)
	}

	if cfg.command == "tui" {
		report, selection, err := RunTUI(report)
		if err != nil {
			exitf("error: %v\n", err)
		}
		output, err := RenderReport(report, cfg.format)
		if err != nil {
			exitf("error: %v\n", err)
		}
		if cfg.outFile != "" {
			if err := os.WriteFile(cfg.outFile, []byte(output), 0o644); err != nil {
				exitf("error: writing report: %v\n", err)
			}
		}
		fmt.Print(output)
		if len(selection.ExcludedGroups) > 0 || len(selection.ExcludedPaths) > 0 {
			fmt.Println("\n# Reproduce this selection")
			fmt.Printf("./bin/reclaimit analyze --root %q", cfg.root)
			for _, group := range selection.ExcludedGroups {
				fmt.Printf(" --exclude-group %q", group)
			}
			for _, path := range selection.ExcludedPaths {
				fmt.Printf(" --exclude-path %q", path)
			}
			fmt.Println()
		}
		return
	}

	if cfg.command == "clean" {
		if !cfg.yes {
			exitf("error: clean requires --yes\n")
		}
		fmt.Print(renderDeletionPreview(report.SelectedCandidates))
		deleted, err := Clean(report.SelectedCandidates)
		if err != nil {
			exitf("error: %v\n", err)
		}
		report, err = Analyze(cfg)
		if err != nil {
			exitf("error: %v\n", err)
		}
		report.DeletedBytes = deleted
	}

	output, err := RenderReport(report, cfg.format)
	if err != nil {
		exitf("error: %v\n", err)
	}

	if cfg.outFile != "" {
		if err := os.WriteFile(cfg.outFile, []byte(output), 0o644); err != nil {
			exitf("error: writing report: %v\n", err)
		}
	}

	fmt.Print(output)
}

func parseConfig(args []string) (config, error) {
	cfg := config{
		command:          "analyze",
		root:             ".",
		format:           "plain",
		groupMode:        "repo",
		groupDepth:       1,
		topFiles:         20,
		topGroups:        20,
		topEntries:       15,
		minCandidateSize: 1 << 20,
	}

	if len(args) > 0 {
		switch args[0] {
		case "help":
			cfg.command = "help"
			if len(args) > 1 {
				cfg.helpTopic = args[1]
			}
			return cfg, nil
		case "-h", "--help":
			cfg.command = "help"
			return cfg, nil
		case "--version", "version":
			cfg.command = "version"
			return cfg, nil
		case "analyze", "clean", "tui":
			cfg.command = args[0]
			args = args[1:]
		}
	}
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			cfg.helpTopic = cfg.command
			cfg.command = "help"
			return cfg, nil
		}
	}

	fs := flag.NewFlagSet("reclaimit", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, usageText(cfg.command))
	}

	fs.StringVar(&cfg.root, "root", cfg.root, "path to scan")
	fs.StringVar(&cfg.format, "format", cfg.format, "output format: plain or markdown")
	fs.StringVar(&cfg.groupMode, "group-mode", cfg.groupMode, "group candidates by repo or depth")
	fs.IntVar(&cfg.groupDepth, "group-depth", cfg.groupDepth, "depth to use when group-mode=depth")
	fs.IntVar(&cfg.topFiles, "top-files", cfg.topFiles, "number of largest files to show")
	fs.IntVar(&cfg.topGroups, "top-groups", cfg.topGroups, "number of candidate groups to show")
	fs.IntVar(&cfg.topEntries, "top-entries", cfg.topEntries, "number of largest direct children under root to show")
	fs.Int64Var(&cfg.minCandidateSize, "min-candidate-size", cfg.minCandidateSize, "minimum candidate size in bytes")
	fs.StringVar(&cfg.outFile, "out", "", "write the report to a file")
	fs.BoolVar(&cfg.yes, "yes", false, "confirm destructive cleanup when using clean")
	fs.Var(&cfg.includeCategories, "include-category", "limit to a category (repeatable)")
	fs.Var(&cfg.excludeCategories, "exclude-category", "exclude a category (repeatable)")
	fs.Var(&cfg.excludeGroups, "exclude-group", "exclude a group path prefix (repeatable)")
	fs.Var(&cfg.excludePaths, "exclude-path", "exclude a specific candidate path (repeatable)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			cfg.helpTopic = cfg.command
			cfg.command = "help"
			return cfg, nil
		}
		return cfg, err
	}

	if cfg.format != "plain" && cfg.format != "markdown" {
		return cfg, fmt.Errorf("unsupported format %q", cfg.format)
	}
	if cfg.groupMode != "repo" && cfg.groupMode != "depth" {
		return cfg, fmt.Errorf("unsupported group mode %q", cfg.groupMode)
	}
	if cfg.groupDepth < 1 {
		return cfg, errors.New("group-depth must be >= 1")
	}
	if cfg.topFiles < 1 || cfg.topGroups < 1 || cfg.topEntries < 1 {
		return cfg, errors.New("top limits must be >= 1")
	}

	absRoot, err := filepath.Abs(cfg.root)
	if err != nil {
		return cfg, err
	}
	cfg.root = filepath.Clean(absRoot)

	for i, group := range cfg.excludeGroups {
		absGroup, err := filepath.Abs(group)
		if err != nil {
			return cfg, err
		}
		cfg.excludeGroups[i] = filepath.Clean(absGroup)
	}
	for i, path := range cfg.excludePaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return cfg, err
		}
		cfg.excludePaths[i] = filepath.Clean(absPath)
	}

	return cfg, nil
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

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

func renderDeletionPreview(candidates []Candidate) string {
	var b strings.Builder
	if len(candidates) == 0 {
		return "No cleanup candidates selected.\n"
	}

	total := sumCandidateBytes(candidates)
	fmt.Fprintf(&b, "Deleting %d candidates totaling %s:\n", len(candidates), humanizeBytes(total))
	limit := len(candidates)
	if limit > 20 {
		limit = 20
	}
	for _, candidate := range candidates[:limit] {
		fmt.Fprintf(&b, "  - %s  %s\n", humanizeBytes(candidate.Bytes), candidate.Path)
	}
	if len(candidates) > limit {
		fmt.Fprintf(&b, "  ... and %d more\n", len(candidates)-limit)
	}
	b.WriteString("\n")
	return b.String()
}
