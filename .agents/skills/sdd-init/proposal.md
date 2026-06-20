---
title: "Add --dry-run to clean command + expand test coverage"
status: proposal
created_at: 2026-06-17
---

# Add --dry-run to clean command + expand test coverage

## Problem

`reclaimit clean` has two critical gaps:

1. **No dry-run mode**: `Clean()` calls `os.RemoveAll()` directly. There's no way to preview what would be deleted without actually deleting. The README promises "delete safely" but there's no simulation mode.

2. **Low test coverage**: Only 5 of ~27 Go files have tests. The `Clean()` function, `renderDeletionPreview()`, `options.go` parsing, `selection.go`, `format.go`, `scanner_categories.go`, `scanner_grouping.go`, `scanner_summaries.go`, `run_commands.go`, `run_helpers.go`, `logging.go`, `usage.go`, `filesystem_*.go` are all untested.

## Solution

### 1. Add `--dry-run` flag to clean command

- Add `--dry-run` / `-n` boolean flag to config
- When `--dry-run` is set, `Clean()` becomes `DryRun()` — it validates all paths exist and returns what would be deleted, but never calls `os.RemoveAll`
- The deletion preview (`renderDeletionPreview`) is shown regardless of dry-run
- Dry-run returns exit code 0 even if there are candidates (it's informational)
- When `--dry-run` is NOT set, existing behavior is unchanged (requires `--yes`)

### 2. Expand test coverage to ~80%

Add tests for all untested production files:
- `options.go` — flag parsing, validation, defaults
- `selection.go` — filterCandidates, isGroupExcluded, isPathExcluded
- `format.go` — renderPlain, renderMarkdown, renderHelpers
- `scanner_categories.go` — matchDirectory, matchFile, includeCategory
- `scanner_grouping.go` — determineGroup, findRepoRoot, ancestorGroup
- `scanner_summaries.go` — summarizeCategories, summarizeGroups, helpers
- `run_commands.go` — handleTUIFlow, handleCleanFlow
- `run_helpers.go` — renderDeletionPreview, exitf, writeString, writef
- `logging.go` — newLogger, validLogLevel
- `usage.go` — usageText
- `filesystem_*.go` — filesystemUsage per-OS
- `DryRun` — dry-run mode with and without candidates

## Scope

- **Files changed**: ~4 production files (options.go, run_commands.go, scanner_clean.go, logging.go), ~6 test files
- **Lines added**: ~200 production + ~400 tests
- **Breaking changes**: None. `--dry-run` is opt-in.

## Non-goals

- JSON output format (separate feature)
- Ignore rules file (.reclaimitignore) (separate feature)
- Runtime version injection (separate feature)
- Additional category definitions (separate feature)
