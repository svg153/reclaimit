package reclaimit

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

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
	ignoreFile        string
	includeCategories stringList
	excludeCategories stringList
	excludeGroups     stringList
	excludePaths      stringList
	yes               bool
	dryRun            bool
	quiet             bool
	logLevel          string
	logger            *slog.Logger
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
		logLevel:         "warn",
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
	fs.StringVar(&cfg.ignoreFile, "ignore-file", "", "path to a .reclaimitignore file with exclusion rules")
	fs.BoolVar(&cfg.yes, "yes", false, "confirm destructive cleanup when using clean")
	fs.BoolVar(&cfg.dryRun, "dry-run", false, "preview cleanup without deleting files")
	fs.BoolVar(&cfg.quiet, "quiet", false, "suppress non-essential output (sets log level to error)")
	fs.StringVar(&cfg.logLevel, "log-level", cfg.logLevel, "log verbosity sent to stderr: debug, info, warn or error")
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

	switch cfg.format {
	case "plain", "markdown", "json":
	default:
		return cfg, fmt.Errorf("unsupported format %q", cfg.format)
	}

	// Load ignore file if provided
	if cfg.ignoreFile != "" {
		patterns, err := loadIgnoreFile(cfg.ignoreFile)
		if err != nil {
			return cfg, fmt.Errorf("loading ignore file: %w", err)
		}
		for _, p := range patterns {
			absPath, err := filepath.Abs(p)
			if err != nil {
				return cfg, err
			}
			cfg.excludePaths = append(cfg.excludePaths, filepath.Clean(absPath))
		}
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
	if !validLogLevel(cfg.logLevel) {
		return cfg, fmt.Errorf("unsupported log level %q", cfg.logLevel)
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

// loadIgnoreFile reads a .reclaimitignore file and returns a list of absolute paths.
// Each non-empty, non-comment line is treated as a path to exclude.
func loadIgnoreFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, nil
}
