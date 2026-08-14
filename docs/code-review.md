# Repository Audit — 2026-08-14

This audit covers correctness, tests, coverage, static analysis, dependency
security, release automation, installation, documentation, GitHub Pages, and
project discoverability.

## Executive Summary

The product has a clear niche and a compact Go codebase, but the repository was
not release-ready at the start of this audit:

- `main` did not compile because `sumCandidateBytes` was declared twice.
- Two tests encoded behavior that no longer matched the implementation.
- Total scanned bytes were calculated from the truncated top-N list and could
  silently under-report disk usage.
- GitHub Pages returned 404 because the workflow uploaded Markdown while
  disabling Jekyll.
- The tag workflow built the root library package and injected the version into
  the wrong symbol.
- Public documentation advertised distribution channels and cleanup categories
  that were not implemented or published.
- Overall statement coverage was 68.3%, not the 90% target.

The current change fixes the correctness, lint, release, Pages, installer,
documentation, coverage, public-path deletion race, traversal concurrency, and
release supply-chain defects. Publishing the corrected release and growth
assets remain operational follow-up work.

## Validation Results

| Check | Result |
| --- | --- |
| `go test ./...` | Pass after fixes |
| Coverage gate | 94.1% overall; every production package and file is at least 90% |
| `go test -race ./...` | Pass after fixes |
| `go vet ./...` | Pass |
| `golangci-lint v2.1.6` | Pass, 0 issues |
| `govulncheck` | No known vulnerability reaches called code |
| Module verification and tidy diff | Pass |
| Scanner fuzz target | 829,709 executions in 5 seconds, pass |
| GoReleaser configuration | Valid; snapshot archives and Linux packages build |
| GitHub Actions syntax | `actionlint` pass |
| Installer | ShellCheck pass; latest Linux archive installs and runs |
| Landing page HTML | `html-validate` pass |
| Repository Markdown | `markdownlint-cli2` pass |

`govulncheck` reported three known advisories in dependencies or the standard
library that are not called by this program: one in `golang.org/x/text`, one
in `golang.org/x/sys` limited to Windows code, and one in a newer standard
library toolchain. They should continue to be handled through dependency
updates, but they are not currently reachable findings.

## Coverage

The current overall statement coverage is **94.1%**.

| Package | Coverage |
| --- | ---: |
| root command orchestration package | 91.7% |
| `internal/cli` | 92.0% |
| `internal/filesystem` | 100.0% |
| `internal/logger` | 100.0% |
| `internal/renderer` | 100.0% |
| `internal/scanner` | 94.8% |
| `internal/tui` | 92.3% |

The gate enforces 90% overall and in every package and source file with
executable production behavior. It discovers packages with `go list`, so a new
production package cannot silently avoid the check, and calculates per-file
statement totals from the complete profile so one high-risk file cannot hide
behind its package average. The process signal/bootstrap boundary in
`cmd/reclaimit` and test-support packages are explicitly not applicable. They
remain in the overall coverage profile.

The TUI tests use a simulated terminal and exercise rendering, navigation,
save, and cancel flows. CLI, scanner, renderer, cleanup orchestration, and
writer-failure tests cover the other former gaps. The reusable local and CI
gate is `scripts/check-coverage.sh`.

## Correctness and Performance Fixes

### Correctness

- Removed the duplicate cleanup helper that broke every build and test command.
- Updated stale category and repository-grouping tests.
- Counted all root scan results in `TotalBytes`, independently of the
  `TopEntries` display limit.
- Enabled the JSON report format in CLI validation and help, matching the
  renderer that already implemented it.
- Added regression tests for total-byte accounting, JSON CLI output, category
  validation, and top-N ordering.

### Measured top-N improvement

`PushTop` previously appended, sorted, and truncated on every item. In the
local 10,000-item benchmark:

| Implementation | Time/op | Bytes/op | Allocations/op |
| --- | ---: | ---: | ---: |
| Before | 5,578,260 ns | 961,568 B | 30,004 |
| Bounded ordered insertion | 45,423 ns | 1,640 B | 6 |

This replaces a measured hotspot with a small bounded insertion algorithm. The
rest of the scanner should only be optimized from representative profiles.

### Traversal-wide worker scheduler

The recursive per-directory worker pools were replaced by one coordinator and
one traversal-wide pool. A deterministic blocking-filesystem test proves that
`Workers=3` never exceeds three active inspections, and cancellation tests prove
that queued work stops and all workers exit.

On the local 16-repository, 512-file benchmark fixture (AMD EPYC 9V74, five
iterations), the scheduler measured 2,893,472 ns/op with one worker and
2,416,642 ns/op with eight workers. These numbers compare scheduler modes on
one cached local fixture; they are not a general filesystem performance claim.

## Release and Installation

The previous tag workflow attempted to build `.`, which is now a library
package, and used `-X main.version` instead of
`-X github.com/svg153/reclaimit.Version`. The latest published v0.1.6 Linux
binary consequently reports `reclaimit 0.1.0`.

Release automation is now consolidated around the checked-in GoReleaser
configuration:

- Linux, macOS, and Windows archives
- checksums
- `.deb`, `.rpm`, and `.apk` packages
- generated release notes
- a tagged non-root container image on GHCR

The installer gained a testable `RECLAIMIT_INSTALL_DIR` override, extracts
archives without attempting to preserve archive ownership, and refuses to
install unless the selected filename has a matching SHA-256 entry in the
GoReleaser manifest. Tests cover valid, missing, malformed, and mismatched
manifests before any destination write. Every workflow action is pinned to a
full commit SHA with a readable version comment; Dependabot's GitHub Actions
ecosystem remains enabled for updates.

## Documentation and Discoverability

### Fixed

- GitHub Pages now deploys `index.html`, assets, `robots.txt`,
  `sitemap.xml`, and `llms.txt` as an actual static site.
- The project-relative logo URL no longer navigates to the account Pages root.
- The landing page has a canonical URL, descriptive title, Open Graph and
  Twitter metadata, a 1200×630 social card, and SoftwareApplication JSON-LD.
- The sitemap contains only the canonical page; fragment URLs were removed.
- README and landing copy use the product's strongest defensible position:
  general analyzers show what is large, while reclaimit identifies known,
  regenerable developer artifacts and adds a guarded review workflow.
- Claims now match the 17 implemented cleanup categories and currently
  available release archives.
- Unsupported Homebrew, cache, benchmark, and competitor claims were removed.

### Current public baseline

At audit time the repository had 3 stars, 1 fork, and one download across the
assets of the latest release. Topics and repository description were already
relevant; the broken homepage and low-confidence product claims were the
larger discoverability problems.

### Growth work with the best evidence-to-effort ratio

1. Publish a corrected patch release so installation, version output, packages,
   and GHCR match the new documentation.
2. Record a short terminal demo showing analyze → TUI review → dry run. Put it
   above the README fold and reuse it in release and social posts.
3. Publish reproducible benchmarks on named hardware and a fixed fixture. Do
   not use generic “100K files in seconds” or “30–50 GB recovered” claims.
4. Add new categories from user issues and real scan evidence, each with a
   safety description and tests.
5. Track GitHub traffic, release downloads, stars per release, and Pages search
   impressions. Use these to decide which content and integrations to expand.

## Open Engineering Risks

### Remaining priorities

1. Publish a corrected patch release and verify all generated assets.
2. Add a reproducible end-to-end Pages smoke check for canonical assets.
3. Add categories such as npm, Yarn, pnpm, Go, and Docker only after defining
   exact safe paths and regeneration semantics.

## Merge Recommendation

The current change is suitable to merge: the complete local validation suite is
green and coverage exceeds 90% overall and per production package and file. GitHub-hosted
checks may remain unavailable when Actions minutes are exhausted; the same
commands have been run locally. Keep publication as a focused release operation
so generated archives, packages, checksums, and images can be verified together.
