package reclaimit

import (
	"io"
	"os"
)

func handleTUIFlow(cfg config, report Report, stdout, stderr io.Writer) (handled bool, updated Report, code int) {
	if cfg.command != "tui" {
		return false, report, 0
	}

	updated, selection, err := RunTUI(report)
	if err != nil {
		return true, report, exitf(stderr, "error: %v\n", err)
	}
	output, err := RenderReport(updated, cfg.format)
	if err != nil {
		return true, report, exitf(stderr, "error: %v\n", err)
	}
	if cfg.outFile != "" {
		if err := os.WriteFile(cfg.outFile, []byte(output), 0o644); err != nil {
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
		if err := writef(stdout, "./bin/reclaimit analyze --root %q", cfg.root); err != nil {
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
	return true, updated, 0
}

func handleCleanFlow(cfg config, report Report, stdout, stderr io.Writer) (Report, int) {
	if cfg.command != "clean" {
		return report, 0
	}
	if !cfg.yes {
		return report, exitf(stderr, "error: clean requires --yes\n")
	}
	if err := writeString(stdout, renderDeletionPreview(report.SelectedCandidates)); err != nil {
		return report, exitf(stderr, "error: writing deletion preview: %v\n", err)
	}
	deleted, err := Clean(report.SelectedCandidates)
	if err != nil {
		return report, exitf(stderr, "error: %v\n", err)
	}
	updated, err := Analyze(cfg)
	if err != nil {
		return report, exitf(stderr, "error: %v\n", err)
	}
	updated.DeletedBytes = deleted
	return updated, 0
}
