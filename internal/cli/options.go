package cli

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

type Options struct {
	Command           string
	HelpTopic         string
	Root              string
	Format            string
	GroupMode         string
	GroupDepth        int
	TopFiles          int
	TopGroups         int
	TopEntries        int
	MinCandidateSize  int64
	OutFile           string
	IgnoreFile        string
	IncludeCategories []string
	ExcludeCategories []string
	ExcludeGroups     []string
	ExcludePaths      []string
	Yes               bool
	DryRun            bool
	Quiet             bool
	LogLevel          string
	Logger            *slog.Logger
}

func ParseConfig(args []string) (Options, error) {
	cfg := Options{
		Command:          "analyze",
		Root:             ".",
		Format:           "plain",
		GroupMode:        "repo",
		GroupDepth:       1,
		TopFiles:         20,
		TopGroups:        20,
		TopEntries:       15,
		MinCandidateSize: 1 << 20,
		LogLevel:         "warn",
	}

	if len(args) > 0 {
		switch args[0] {
		case "help":
			cfg.Command = "help"
			if len(args) > 1 {
				cfg.HelpTopic = args[1]
			}
			return cfg, nil
		case "-h", "--help":
			cfg.Command = "help"
			return cfg, nil
		case "--version", "version":
			cfg.Command = "version"
			return cfg, nil
		case "analyze", "clean", "tui":
			cfg.Command = args[0]
			args = args[1:]
		}
	}
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			cfg.HelpTopic = cfg.Command
			cfg.Command = "help"
			return cfg, nil
		}
	}

	fs := flag.NewFlagSet("reclaimit", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, UsageText(cfg.Command))
	}

	fs.StringVar(&cfg.Root, "root", cfg.Root, "path to scan")
	fs.StringVar(&cfg.Format, "format", cfg.Format, "output format: plain or markdown")
	fs.StringVar(&cfg.GroupMode, "group-mode", cfg.GroupMode, "group candidates by repo or depth")
	fs.IntVar(&cfg.GroupDepth, "group-depth", cfg.GroupDepth, "depth to use when group-mode=depth")
	fs.IntVar(&cfg.TopFiles, "top-files", cfg.TopFiles, "number of largest files to show")
	fs.IntVar(&cfg.TopGroups, "top-groups", cfg.TopGroups, "number of candidate groups to show")
	fs.IntVar(&cfg.TopEntries, "top-entries", cfg.TopEntries, "number of largest direct children under root to show")
	fs.Int64Var(&cfg.MinCandidateSize, "min-candidate-size", cfg.MinCandidateSize, "minimum candidate size in bytes")
	fs.StringVar(&cfg.OutFile, "out", "", "write the report to a file")
	fs.StringVar(&cfg.IgnoreFile, "ignore-file", "", "path to a .reclaimitignore file with exclusion rules")
	fs.BoolVar(&cfg.Yes, "yes", false, "confirm destructive cleanup when using clean")
	fs.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "log verbosity sent to stderr: debug, info, warn or error")
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "preview cleanup without deleting files")
	fs.BoolVar(&cfg.Quiet, "quiet", false, "suppress non-essential output (sets log level to error)")

	includeCategories := stringList{}
	excludeCategories := stringList{}
	excludeGroups := stringList{}
	excludePaths := stringList{}
	fs.Var(&includeCategories, "include-category", "limit to a category (repeatable)")
	fs.Var(&excludeCategories, "exclude-category", "exclude a category (repeatable)")
	fs.Var(&excludeGroups, "exclude-group", "exclude a group path prefix (repeatable)")
	fs.Var(&excludePaths, "exclude-path", "exclude a specific candidate path (repeatable)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			cfg.HelpTopic = cfg.Command
			cfg.Command = "help"
			return cfg, nil
		}
		return cfg, err
	}

	cfg.IncludeCategories = append(cfg.IncludeCategories, includeCategories...)
	cfg.ExcludeCategories = append(cfg.ExcludeCategories, excludeCategories...)
	cfg.ExcludeGroups = append(cfg.ExcludeGroups, excludeGroups...)
	cfg.ExcludePaths = append(cfg.ExcludePaths, excludePaths...)

	if cfg.Format != "plain" && cfg.Format != "markdown" {
		return cfg, fmt.Errorf("unsupported format %q", cfg.Format)
	}
	if cfg.GroupMode != "repo" && cfg.GroupMode != "depth" {
		return cfg, fmt.Errorf("unsupported group mode %q", cfg.GroupMode)
	}
	if cfg.GroupDepth < 1 {
		return cfg, errors.New("group-depth must be >= 1")
	}
	if cfg.TopFiles < 1 || cfg.TopGroups < 1 || cfg.TopEntries < 1 {
		return cfg, errors.New("top limits must be >= 1")
	}
	// Load ignore file if provided
	if cfg.IgnoreFile != "" {
		patterns, err := loadIgnoreFile(cfg.IgnoreFile)
		if err != nil {
			return cfg, fmt.Errorf("loading ignore file: %w", err)
		}
		for _, p := range patterns {
			absPath, err := filepath.Abs(p)
			if err != nil {
				return cfg, err
			}
			cfg.ExcludePaths = append(cfg.ExcludePaths, filepath.Clean(absPath))
		}
	}
	if !ValidLogLevel(cfg.LogLevel) {
		return cfg, fmt.Errorf("unsupported log level %q", cfg.LogLevel)
	}

	absRoot, err := filepath.Abs(cfg.Root)
	if err != nil {
		return cfg, err
	}
	cfg.Root = filepath.Clean(absRoot)

	for i, group := range cfg.ExcludeGroups {
		absGroup, err := filepath.Abs(group)
		if err != nil {
			return cfg, err
		}
		cfg.ExcludeGroups[i] = filepath.Clean(absGroup)
	}
	for i, path := range cfg.ExcludePaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return cfg, err
		}
		cfg.ExcludePaths[i] = filepath.Clean(absPath)
	}

	return cfg, nil
}


func loadIgnoreFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ignore file: %w", err)
	}
	var patterns []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, nil
}
