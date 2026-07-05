package reclaimit

import (
	"fmt"
	"io"
	"os"

	"github.com/svg153/reclaimit/internal/cli"
	"github.com/svg153/reclaimit/internal/logger"
	"github.com/svg153/reclaimit/internal/renderer"
	"github.com/svg153/reclaimit/internal/scanner"
	"github.com/svg153/reclaimit/internal/tui"
)

var Version = "dev"


func Run(args []string, stdout, stderr io.Writer) int {
	cfg, err := cli.ParseConfig(args)
	if err != nil {
		return exitf(stderr, "error: %v\n", err)
	}

	if cfg.Command == "help" {
		if err := writeString(stdout, cli.UsageText(cfg.HelpTopic)); err != nil {
			return exitf(stderr, "error: writing help: %v\n", err)
		}
		return 0
	}
	if cfg.Command == "version" {
		if err := writef(stdout, "reclaimit %s\n", Version); err != nil {
			return exitf(stderr, "error: writing version: %v\n", err)
		}
		return 0
	}

	if cfg.Quiet {
		cfg.LogLevel = "error"
	}
	cfg.Logger = logger.NewLogger(cfg.LogLevel, stderr)

	report, err := scanner.AnalyzeWithOptions(
		cfg.Command,
		toScannerOpts(cfg),
		cfg.Logger,
	)
	if err != nil {
		return exitf(stderr, "error: %v\n", err)
	}

	if cfg.Command == "tui" {
		selection, err := tui.Run(report)
		if err != nil {
			return exitf(stderr, "error: %v\n", err)
		}
		output, err := renderer.RenderReport(report, cfg.Format)
		if err != nil {
			return exitf(stderr, "error: %v\n", err)
		}
		if err := writeSelection(stdout, cfg, selection); err != nil {
			return exitf(stderr, "error: %v\n", err)
		}
		return writeOutput(stdout, cfg.OutFile, output)
	}

	if cfg.Command == "clean" {
		if !cfg.Yes && !cfg.DryRun {
			return exitf(stderr, "error: clean requires --yes or --dry-run\n")
		}

		if preview := cli.RenderDeletionPreview(report.SelectedCandidates); preview != "" {
			if err := writeString(stdout, preview); err != nil {
				return exitf(stderr, "error: %v\n", err)
			}
		}

		if cfg.DryRun {
			deleted, err := scanner.DryRun(report.SelectedCandidates)
			if err != nil {
				return exitf(stderr, "error: %v\n", err)
			}
			report.DeletedBytes = deleted
			if err := writef(stdout, "\n[DRY RUN] Would delete %s across %d candidates\n", humanizeBytes(deleted), len(report.SelectedCandidates)); err != nil {
				return exitf(stderr, "error: %v\n", err)
			}
		} else {
			deleted, err := scanner.Clean(report.SelectedCandidates)
			if err != nil {
				return exitf(stderr, "error: %v\n", err)
			}
			report.DeletedBytes = deleted
		}

		return writeOutput(stdout, cfg.OutFile, "")
	}

	output, err := renderer.RenderReport(report, cfg.Format)
	if err != nil {
		return exitf(stderr, "error: %v\n", err)
	}

	return writeOutput(stdout, cfg.OutFile, output)
}

func toScannerOpts(cfg cli.Options) scanner.AnalyzeOptions {
	return scanner.AnalyzeOptions{
		Root:              cfg.Root,
		GroupMode:         cfg.GroupMode,
		GroupDepth:        cfg.GroupDepth,
		TopFiles:          cfg.TopFiles,
		TopGroups:         cfg.TopGroups,
		TopEntries:        cfg.TopEntries,
		MinCandidateSize:  cfg.MinCandidateSize,
		IncludeCategories: cfg.IncludeCategories,
		ExcludeCategories: cfg.ExcludeCategories,
		ExcludeGroups:     cfg.ExcludeGroups,
		ExcludePaths:      cfg.ExcludePaths,
	}
}

func writeSelection(stdout io.Writer, cfg cli.Options, selection tui.Selection) error {
	if err := writeString(stdout, "\n# Selection\n"); err != nil {
		return err
	}
	if len(selection.ExcludedGroups) > 0 || len(selection.ExcludedPaths) > 0 {
		if err := writeString(stdout, "# Reproduce this selection:\n"); err != nil {
			return err
		}
		if err := writef(stdout, "./bin/reclaimit analyze --root %q", cfg.Root); err != nil {
			return err
		}
		for _, group := range selection.ExcludedGroups {
			if err := writef(stdout, " --exclude-group %q", group); err != nil {
				return err
			}
		}
		for _, path := range selection.ExcludedPaths {
			if err := writef(stdout, " --exclude-path %q", path); err != nil {
				return err
			}
		}
		if err := writeString(stdout, "\n"); err != nil {
			return err
		}
	}
	return nil
}

func writeOutput(stdout io.Writer, outFile string, output string) int {
	if outFile != "" {
		if err := os.WriteFile(outFile, []byte(output), 0o644); err != nil {
			exitf(stdout, "error: writing report: %v\n", err)
		}
		return 0
	}
	if err := writeString(stdout, output); err != nil {
		return 1
	}
	return 0
}

func exitf(w io.Writer, format string, args ...any) int {
	_, _ = fmt.Fprintf(w, format, args...)
	return 1
}

func writeString(w io.Writer, value string) error {
	_, err := io.WriteString(w, value)
	return err
}

func writef(w io.Writer, format string, args ...any) error {
	_, err := fmt.Fprintf(w, format, args...)
	return err
}

func humanizeBytes(bytes int64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	if bytes == 0 {
		return "0 B"
	}
	idx := 0
	size := float64(bytes)
	for size >= 1024 && idx < len(units)-1 {
		size /= 1024
		idx++
	}
	if idx == 0 {
		return fmt.Sprintf("%d %s", int(size), units[idx])
	}
	return fmt.Sprintf("%.1f %s", size, units[idx])
}
