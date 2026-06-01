package reclaimit

import (
	"io"
	"os"
)

var Version = "dev"

func Run(args []string, stdout, stderr io.Writer) int {
	cfg, err := parseConfig(args)
	if err != nil {
		return exitf(stderr, "error: %v\n", err)
	}

	if cfg.command == "help" {
		if err := writeString(stdout, usageText(cfg.helpTopic)); err != nil {
			return exitf(stderr, "error: writing help: %v\n", err)
		}
		return 0
	}
	if cfg.command == "version" {
		if err := writef(stdout, "reclaimit %s\n", Version); err != nil {
			return exitf(stderr, "error: writing version: %v\n", err)
		}
		return 0
	}

	cfg.logger = newLogger(cfg.logLevel, stderr)

	report, err := Analyze(cfg)
	if err != nil {
		return exitf(stderr, "error: %v\n", err)
	}

	handled, updatedReport, code := handleTUIFlow(cfg, report, stdout, stderr)
	if handled {
		return code
	}
	report = updatedReport

	updatedReport, code = handleCleanFlow(cfg, report, stdout, stderr)
	if code != 0 {
		return code
	}
	report = updatedReport

	output, err := RenderReport(report, cfg.format)
	if err != nil {
		return exitf(stderr, "error: %v\n", err)
	}

	if cfg.outFile != "" {
		if err := os.WriteFile(cfg.outFile, []byte(output), 0o644); err != nil {
			return exitf(stderr, "error: writing report: %v\n", err)
		}
	}

	if err := writeString(stdout, output); err != nil {
		return exitf(stderr, "error: writing report: %v\n", err)
	}
	return 0
}
