# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- guarded cleanup with pre-delete identity checks, same-filesystem quarantine,
  and recovery paths for changed data
- traversal-wide worker limit, cancellation, depth limits, JSON output, and
  reusable coverage enforcement
- pip, pipx, Bun download-cache, and macOS cleanup categories
- reproducible terminal demo, benchmark methodology, growth baseline, and
  repository security and community health files
- verified release archives, Linux packages, checksums, and multi-platform
  container configuration

### Changed

- release workflows inject version metadata, pin actions by commit, and publish
  true amd64/arm64 container images
- cleanup categories and CLI filters now fail closed on unsafe or unknown input
- README and landing-page copy now use verified developer-cleanup positioning
- repository no longer tracks generated binaries, editor databases, or dead
  placeholder packages

### Fixed

- total scanned bytes no longer depend on the top-N display limit
- unknown commands and positional arguments no longer trigger a default scan
- generated TUI reproduction commands are safe to paste into a POSIX shell
- Bun cleanup preserves the runtime and global tools outside
  `.bun/install/cache`
- GitHub Pages, installer verification, release targets, Windows metrics, race
  handling, and package-level coverage regressions

## [0.1.6] - 2026-05-21

### Added

- multi-platform release artifacts
- improved TUI tree and markdown reporting
- Dependabot, CodeQL and GitHub Pages automation

### Fixed

- cross-platform filesystem usage build issues
- release artifact upload flow
