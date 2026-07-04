# reclaimit — Reclaimable Disk Space Analyzer for Developer Workstations

**reclaimit** is a fast, safe Go CLI that finds reclaimable disk space on developer machines — with a bias toward **what is safe enough to delete**.

Unlike classic disk analyzers that just answer "what is big?", reclaimit answers **"what can I safely clean up?"** — by understanding developer-specific leftovers like `node_modules`, `.venv`, build caches, Docker layers, and more.

## Quick Install

```bash
# macOS / Linux
brew install svg153/reclaimit/reclaimit

# Or the universal install script
curl -fsSL https://raw.githubusercontent.com/svg153/reclaimit/main/install.sh | bash

# Or go install
go install github.com/svg153/reclaimit@latest
```

## Why reclaimit?

Most disk space tools (`du`, `ncdu`, `gdu`) show you what's taking space. But deleting blindly is risky — you might nuke a project's `node_modules` or a Python virtual environment you still need.

reclaimit is different. It **categorizes** cleanup targets by type and lets you review before deleting:

- **Scan** any directory → get a structured report
- **Review** in a terminal UI (TUI) with a path tree
- **Delete** only what you've verified

## Key Features

- **Developer-aware**: Knows about `node_modules`, `.venv`, `__pycache__`, `.next`, `dist/`, Docker layers, and 20+ more patterns
- **Three output modes**: plain text, Markdown (with tables, Mermaid diagrams), and interactive TUI
- **Repository-aware grouping**: Groups candidates by project, not just by size
- **Safe by design**: Preview before delete, `--yes` flag required, context nodes that never imply deleting the repo root
- **Single binary**: No dependencies, no Python, no Node — just a ~10MB Go binary
- **Docker-ready**: Non-root distroless image on GHCR
- **Cross-platform**: Linux, macOS, Windows

- `node_modules`
- `.venv`, `venv`, `.tox`
- `__pycache__`, `.pyc`, `.pyo`
- `.pytest_cache`, `.mypy_cache`
- `dist`, `build`, `target`
- `.next`, `.nuxt`
- `.cache` (generic caches), `.npm`, `.yarn`, `.pnpm-store`, `.bun` (package manager caches)
- `.DS_Store`, `.Spotlight-V100`, `.Trashes` (macOS Finder metadata, index caches, and trash folders)

## Usage

### Generate a report

```bash
reclaimit analyze --root "$HOME" --format markdown --out cleanup-report.md
```

### Interactive cleanup

```bash
reclaimit tui --root "$HOME"
```

### Clean reviewed targets

```bash
reclaimit clean --root "$HOME" --include-category python-venv --yes
```

## Comparison

| Feature | reclaimit | `du` / `ncdu` | `gdu` | `du-dust` |
|---------|-----------|---------------|-------|-----------|
| Developer-aware patterns | ✅ 20+ patterns | ❌ Raw sizes | ❌ Raw sizes | ❌ Raw sizes |
| Safe delete preview | ✅ Review before delete | ❌ Delete immediately | ❌ Delete immediately | ❌ Delete immediately |
| Markdown reports | ✅ Tables + Mermaid | ❌ | ❌ | ❌ |
| Repository grouping | ✅ Group by project | ❌ Flat tree | ❌ Flat tree | ❌ Flat tree |
| Docker image | ✅ GHCR distroless | ❌ | ❌ | ❌ |
| Homebrew tap | ✅ | ❌ | ❌ | ❌ |
| Cross-platform | ✅ Go + Windows | ✅ | ✅ | ✅ |

## Installation

### Homebrew (macOS / Linux)

```bash
brew install svg153/reclaimit/reclaimit
```

### Universal install script

```bash
curl -fsSL https://raw.githubusercontent.com/svg153/reclaimit/main/install.sh | bash
```

Installs to `$HOME/.local/bin/reclaimit`.

### Go install

```bash
go install github.com/svg153/reclaimit@latest
```

### Linux packages

Release builds publish `.deb`, `.rpm`, and `.apk` packages. Download from the [latest release](https://github.com/svg153/reclaimit/releases/latest).

### Docker

```bash
docker run --rm ghcr.io/svg153/reclaimit:latest analyze --root /scan --format markdown
```

### Build from source

```bash
git clone https://github.com/svg153/reclaimit.git
cd reclaimit
task build
./bin/reclaimit --version
```

## Commands

| Command | Description |
|---------|-------------|
| `analyze` | Generate plain-text or Markdown report |
| `tui` | Interactive terminal UI with path tree |
| `clean` | Delete reviewed cleanup targets |

### `analyze` flags

`--root PATH` — Directory to scan (default: current directory)
`--format plain\|markdown` — Output format
`--group-mode repo\|depth` — Grouping strategy
`--group-depth N` — Depth for grouping (when using `--group-mode depth`)
`--exclude-group PATH` — Exclude by path prefix
`--exclude-path PATH` — Exact path exclusion
`--out FILE` — Write report to file
`--log-level debug\|info\|warn\|error` — Log verbosity

### `tui` flags

`--root PATH` — Directory to scan
`--format markdown` — Include Mermaid diagrams in TUI

### `clean` flags

`--root PATH` — Directory to clean
`--include-category CATEGORY` — Only clean this category (e.g., `python-venv`)
`--exclude-category CATEGORY` — Skip this category
`--yes` — Confirm deletion (required for safety)

## Architecture

```
reclaimit
├── analyze     → Scan + categorize → report (text/markdown)
├── tui         → Scan + categorize → interactive tree UI
└── clean       → Review list → delete with safety checks
```

See [docs/architecture.md](docs/architecture.md) for C4 diagrams and detailed design.

## Development

```bash
task fmt        # gofmt
task vet        # go vet
task lint       # golangci-lint
task test       # tests + coverage
task bench      # benchmarks
task build      # build binary
task docker-build  # build distroless image
task check      # full quality gate
```

## FAQ

**Is reclaimit safe to use?**
Yes. The `clean` command requires `--yes` to delete anything. The `analyze` and `tui` commands only read — they never modify files.

**What directories does reclaimit recognize?**
Over 20 developer-specific patterns including `node_modules`, `.venv`, `__pycache__`, `.pytest_cache`, `dist/`, `build/`, `.next`, `.nuxt`, Docker layers, Go build caches, npm/yarn/pnpm caches, and more.

**Can I exclude specific paths?**
Yes. Use `--exclude-path` for exact paths or `--exclude-group` for prefix-based exclusions. You can also create a config file for persistent exclusions.

**Does it work on Windows?**
Yes. reclaimit is written in Go and supports Linux, macOS, and Windows.

**How fast is it?**
A single binary with no external dependencies. Scans 100K+ files in seconds. Benchmarks are in [docs/architecture.md](docs/architecture.md).

**Is there a GUI?**
No — reclaimit is terminal-first. The TUI uses `tview` for an interactive tree interface. A web UI is on the roadmap.

**Can I export the report?**
Yes. `analyze --format markdown --out report.md` produces a Markdown report with tables, Mermaid diagrams, and PlantUML blocks.

## License

MIT — see [LICENSE](./LICENSE).

## Contributing

Contributions welcome! See [CONTRIBUTING](CONTRIBUTING.md) for guidelines.
