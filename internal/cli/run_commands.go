package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/svg153/reclaimit/internal/scanner"
	"github.com/svg153/reclaimit/internal/tui"
	"github.com/svg153/reclaimit/internal/renderer"
)

func HandleTUIFlow(cfg Options, report scanner.Report, stdout, stderr io.Writer) (handled bool, updated scanner.Report, code int) {
	if cfg.Command != "tui" {
		return false, report, 0
	}

	selection, err := tui.Run(report)
	if err != nil {
		return true, report, exitf(stderr, "error: %v\n", err)
	}

	output, err := renderer.RenderReport(report, cfg.Format)
	if err != nil {
		return true, report, exitf(stderr, "error: %v\n", err)
	}
	if cfg.OutFile != "" {
		if err := os.WriteFile(cfg.OutFile, []byte(output), 0o644); err != nil {
			return true, report, exitf(stderr, "error: writing report: %v\n", err)
		}
	}
	if err := writeString(stdout, output); err != nil {
		return true, report, exitf(stderr, "error: writing report: %v\n", err)
	}
	if len(selection.ExcludedGroups) > 0 || len(selection.ExcludedPaths) > 0 {
		if err := writeString(stdout, "\n# Reproduce this selection\n"); err != nil {
			return true, report, exitf(stderr, "error: writing selection header: %v\n", err)
		}
		if err := writef(stdout, "./bin/reclaimit analyze --root %q", cfg.Root); err != nil {
			return true, report, exitf(stderr, "error: writing selection command: %v\n", err)
		}
		for _, group := range selection.ExcludedGroups {
			if err := writef(stdout, " --exclude-group %q", group); err != nil {
				return true, report, exitf(stderr, "error: writing selection command: %v\n", err)
			}
		}
		for _, path := range selection.ExcludedPaths {
			if err := writef(stdout, " --exclude-path %q", path); err != nil {
				return true, report, exitf(stderr, "error: writing selection command: %v\n", err)
			}
		}
		if err := writeString(stdout, "\n"); err != nil {
			return true, report, exitf(stderr, "error: writing selection command: %v\n", err)
		}
	}
	return true, report, 0
}

func HandleCleanFlow(cfg Options, report scanner.Report, stdout, stderr io.Writer) (scanner.Report, int) {
	if cfg.Command != "clean" {
		return report, 0
	}
	if !cfg.Yes && !cfg.DryRun {
		return report, exitf(stderr, "error: clean requires --yes or --dry-run\n")
	}
	if err := writeString(stdout, RenderDeletionPreview(report.SelectedCandidates)); err != nil {
		return report, exitf(stderr, "error: writing deletion preview: %v\n", err)
	}
	if cfg.DryRun {
		return handleDryRun(report, stdout, stderr)
	}
	return handleActualClean(cfg, report, stdout, stderr)
}

func handleDryRun(report scanner.Report, stdout, stderr io.Writer) (scanner.Report, int) {
	deleted, err := scanner.DryRun(report.SelectedCandidates)
	if err != nil {
		return report, exitf(stderr, "error: %v\n", err)
	}
	if err := writef(stdout, "\n[DRY RUN] Would delete %s across %d candidates\n", humanizeBytes(deleted), len(report.SelectedCandidates)); err != nil {
		return report, exitf(stderr, "error: writing dry-run summary: %v\n", err)
	}
	updated := report
	updated.DeletedBytes = deleted
	return updated, 0
}

func handleActualClean(cfg Options, report scanner.Report, stdout, stderr io.Writer) (scanner.Report, int) {
	deleted, err := scanner.Clean(report.SelectedCandidates)
	if err != nil {
		return report, exitf(stderr, "error: %v\n", err)
	}
	updated := report
	updated.DeletedBytes = deleted
	return updated, 0
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

func writeSelection(stdout io.Writer, cfg Options, selection tui.Selection) error {
	_, err := io.WriteString(stdout, "\n# Selection\n")
	if err != nil {
		return err
	}
	if len(selection.ExcludedGroups) > 0 || len(selection.ExcludedPaths) > 0 {
		if err := writeString(stdout, "# Excluded groups:\n"); err != nil {
			return err
		}
		for _, group := range selection.ExcludedGroups {
			if err := writeString(stdout, "  "+group+"\n"); err != nil {
				return err
			}
		}
		if err := writeString(stdout, "# Excluded paths:\n"); err != nil {
			return err
		}
		for _, path := range selection.ExcludedPaths {
			if err := writeString(stdout, "  "+path+"\n"); err != nil {
				return err
			}
		}
	}
	if selection.Saved {
		if err := writeString(stdout, "# To reproduce:\n"); err != nil {
			return err
		}
		if err := writef(stdout, "reclaimit analyze --root %q", cfg.Root); err != nil {
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
