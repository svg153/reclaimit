# Code Review and Improvement Opportunities

This review uses the architecture in `docs/architecture.md` to assess dead code, shallow modules, duplication, maintainability, and localized performance opportunities.

## Executive Summary

The codebase is small, readable, and organized around a useful core seam: `Analyze` builds a report that other command paths reuse. Scanner responsibilities are now split across `scanner_types.go`, `scanner_categories.go`, `scanner_analyze.go`, `scanner_summaries.go`, and `scanner_clean.go`, which improves locality.

## What Looks Good

- `main.go` keeps command routing simple and reuses the same analysis path for `analyze`, `tui`, and `clean`.
- `selection.go` is a clean seam for post-scan filtering.
- `tui.go` is isolated from the filesystem traversal logic.
- Platform-specific filesystem usage is already behind a small adapter.

## Dead Code and Redundant Surfaces

### No confirmed dead feature branch

There is no obvious unreachable command path in the current binary. The main commands are exercised through parsing and report generation, and the repository already contains tests for scanner, cleanup, help, rendering helpers, and TUI behavior.

### Effectively redundant helper: `pushTop`

`pushTop` in `scanner.go` is not dead, but it is shallow. It only appends an item, sorts the entire slice, and truncates to a limit. Because it is used in only one place, its interface currently adds indirection without much leverage.

Recommendation:

- Either inline it inside the `scanNode` file path where it is used.
- Or deepen it by making it the single generic top-N helper reused by both `TopFiles` and `TopEntries` logic.

## Duplication That Should Be Consolidated First

### Sorting helpers

`sortPathSizes` and `sortCandidates` encode the same ordering policy: descending bytes, then ascending path. The behavior is duplicated across different element types.

Impact:

- Changing ordering semantics means touching multiple helpers.
- The common sort policy is conceptually one rule but implemented in two places.

### Summation helpers

`sumBytes` and `sumCandidateBytes` are identical except for the element type.

Impact:

- Low runtime cost, but avoidable cognitive duplication.

### Aggregation patterns

`summarizeCategories` and `summarizeGroups` repeat the same aggregation structure:

- create an index map
- update aggregate entries while iterating
- convert map to slice
- sort the slice
- optionally truncate

Impact:

- Today this is still manageable.
- If summary behavior grows, the duplication will raise the maintenance cost quickly.

## Maintainability Friction

### Traversal module remains the main deepening seam

The recursive traversal now lives behind `scanContext.scan` and related helpers in `scanner_analyze.go`.

- `cfg`
- `inCandidateDir`
- `report`
- `candidateByKey`
- `groupCache`
- `includeSet`
- `excludeSet`

This is still workable, but it indicates the traversal has become the real center of behavior while its interface keeps widening.

Recommendation:

- Introduce a dedicated scan context struct when new traversal behavior is added.
- Do not rush this into a large refactor in the current pass unless the signature changes again.

### Grouping logic is spread across helper functions

`determineGroup`, `findRepoRoot`, and `ancestorGroup` are still local and understandable, but they already form a single grouping policy module in practice.

Recommendation:

- Keep them together conceptually in docs and in future refactors.
- If grouping rules expand, turn them into a deeper module rather than adding more conditionals to `scanContext.scan`.

### Windows adapter now returns real filesystem metrics

`filesystem_windows.go` now uses `GetDiskFreeSpaceExW` and performs overflow-safe clamping before converting to `int64`.

Impact:

- Report fidelity is consistent across Unix and Windows.
- Cross-compile behavior is preserved because implementation is build-tagged.

## Performance Review

### Repeated sort in `pushTop`

Every regular file pushes into `TopFiles` by appending, sorting, and truncating. For the current repository size and expected usage this is acceptable, but it is the clearest micro-optimization target if scans get large.

Recommendation:

- Keep as-is for now if simplicity is the priority.
- Consider a bounded min-heap only if profiling shows this path matters on very large trees.

### Repository root detection cache is local but narrow

`findRepoRoot` caches only exact cursor lookups. It avoids some repeated work, but there is still room for a more aggressive ancestor cache if scans spend time in large monorepos.

Recommendation:

- Do not optimize preemptively.
- If real scans on large trees are slow, measure this path before changing it.

## Test Gaps Worth Tracking

The existing tests already cover most of the core workflow. The most useful additions would be:

- permission-denied traversal inside nested directories
- symlink-heavy directory trees
- selection behavior combining `--exclude-group` and `--exclude-path`
- Windows filesystem usage expectations for the Win32 adapter

## Prioritized Improvement Backlog

### 1. Strong recommendation

Consolidate the duplicated helper logic in `scanner.go`.

Why first:

- Small blast radius
- Easy to verify with the current tests
- Improves locality without changing product behavior

### 2. Worth exploring

Replace the repeated top-N full sort with a bounded structure only if profiling on real data shows it matters.

Why second:

- It is the main visible micro-performance hotspot.
- It is not yet a proven bottleneck.

### 3. Worth exploring

Introduce a scan context struct if traversal behavior grows further.

Why third:

- It would deepen the traversal module.
- It adds abstraction, so it should be justified by actual growth.

### 4. Completed

Real Windows filesystem usage is implemented via Win32 API.

## Implemented in This Pass

This pass applies two low-risk improvements:

- consolidated duplicated sorting and summation helper logic in `scanner.go`
- made Markdown report timestamps deterministic across environments by formatting them in UTC in `render_helpers.go`

These changes stay local, preserve the existing control flow, and are covered by the current test suite.
