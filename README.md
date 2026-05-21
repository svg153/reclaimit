[![CI](https://github.com/svg153/reclaimit/actions/workflows/ci.yml/badge.svg)](https://github.com/svg153/reclaimit/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/svg153/reclaimit)](https://github.com/svg153/reclaimit/releases/latest) [![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://golang.org) [![Coverage](https://codecov.io/gh/svg153/reclaimit/branch/main/graph/badge.svg?token=)](https://codecov.io/gh/svg153/reclaimit)

# reclaimit

Install via Homebrew once a release is published:

```bash
brew install svg153/reclaimit/reclaimit
```


![Go 1.21+](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)
![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)
![TUI](https://img.shields.io/badge/UI-Terminal-6f42c1)

`reclaimit` is a Go CLI for finding reclaimable disk space with a bias toward **developer workstations**.

It scans a root directory, identifies well-known cleanup targets such as `node_modules`, virtual environments, caches and build artifacts, and lets you review them either as:

- a **plain-text report**
- a **Markdown report** with tables, Mermaid and PlantUML blocks
- an **interactive TUI** with a path tree, context folders, and exact delete targets

## Why this tool exists

Classic disk analyzers are good at answering **“what is big?”**.

This tool is optimized for **“what is safe enough to review for deletion?”**.

That means it knows about common development leftovers like:

- `node_modules`
- `.venv`, `venv`, `.tox`
- `__pycache__`, `.pyc`, `.pyo`
- `.pytest_cache`, `.mypy_cache`
- `dist`, `build`, `target`
- `.next`, `.nuxt`
- `.cache`

## Key capabilities

- **Fast local analysis** using a single Go binary
- **Candidate-aware cleanup** instead of raw size reporting only
- **Repository-aware grouping** with `--group-mode repo`
- **Path-tree TUI** built with `tview` / `tcell`
- **Context nodes** that never imply deleting the repository root itself
- **Exact-path exclusions** via `--exclude-path`
- **Prefix-based group exclusions** via `--exclude-group`
- **Last-modified timestamps** in Markdown output and TUI details
- **Safe clean flow** with deletion preview before destructive actions

## Installation

### Build locally

```bash
git clone <repo-url>
cd reclaimit
task build
./bin/reclaimit --version
```

### Install into your user PATH

```bash
task install
reclaimit --version
```

By default `task install` builds the binary into:

```bash
$HOME/.local/bin/reclaimit
```

## Quick start

### 1. Generate a Markdown report

```bash
./bin/reclaimit analyze --root "$HOME" --format markdown --out report.md
```

### 2. Explore cleanup candidates interactively

```bash
./bin/reclaimit tui --root "$HOME" --format markdown
```

### 3. Delete a reviewed subset

```bash
./bin/reclaimit clean --root "$HOME" --include-category python-venv --yes
```

## Commands

### `analyze`

Generate a plain-text or Markdown report.

```bash
./bin/reclaimit analyze --root "$HOME" --format markdown --out report.md
```

Useful flags:

- `--root PATH`
- `--format plain|markdown`
- `--group-mode repo|depth`
- `--group-depth N`
- `--exclude-group PATH`
- `--exclude-path PATH`
- `--out FILE`

### `tui`

Open the interactive tree UI.

```bash
./bin/reclaimit tui --root "$HOME"
```

TUI semantics:

- **Context folders** (`📁`) are grouping nodes only.
- Toggling a context folder **does not mean deleting that folder itself**.
- **Deletion candidates** are explicit target nodes:
  - `🧹` directory target
  - `📄` file target

Default shortcuts:

- `j/k` or `↑/↓` — move
- `Enter` or `→` — expand
- `←` — collapse / move to parent
- `Space` — toggle current node
- `a` — toggle all
- `q` — save selection and exit
- `Esc` — discard changes and exit

### `clean`

Delete the currently selected candidate set.

```bash
./bin/reclaimit clean --root "$HOME" --include-category node-modules --yes
```

Behavior:

- prints a **preview** of what will be deleted
- deletes the selected candidates
- prints a **fresh post-clean report**

## Exclusions and selection rules

### `--exclude-group`

Excludes a whole subtree by prefix match.

Example:

```bash
./bin/reclaimit analyze --root "$HOME" --exclude-group "$HOME/REPOS/project-a"
```

### `--exclude-path`

Excludes **one exact candidate path**.

It is **not** a prefix match.

Example:

```bash
./bin/reclaimit analyze --root "$HOME" \
  --exclude-path "$HOME/REPOS/project-a/.venv"
```

## Markdown report output

The Markdown report includes:

- executive summary table
- Mermaid charts
- PlantUML mindmap block
- collapsible sections for large tables
- candidate and group **last-modified timestamps**

This makes the output useful for:

- sharing in issues or pull requests
- archiving cleanup snapshots
- post-processing in other tools

## Safety model

This tool is opinionated, but intentionally conservative:

- it only flags known cleanup categories
- it separates **context** from **actual delete targets**
- it supports exclusions before cleaning
- it requires `--yes` for destructive cleanup

Still, this is a cleanup tool: review output before deletion, especially for generic cache directories.

## Development

Taskfile targets:

```bash
task fmt
task vet
task test
task coverage-html
task build
task install
task report
task tui
task ci
```

## Testing and coverage

The repository includes:

- scanner tests
- cleanup tests
- render helper tests
- TUI tree/helper tests
- help/usage tests

Generate coverage locally:

```bash
task test
task coverage-html
```

## CI

GitHub Actions workflow is included for:

- Go `1.21`
- Go `1.22`
- Go `1.23`

## Roadmap ideas

- richer filtering by age / last-modified thresholds
- export/import of reviewed selections
- release pipeline for prebuilt binaries
- ignore rules file
- optional JSON output for automation

## Contributing

Contributions are welcome.

If you open a change, prefer:

- focused PRs
- tests for new behavior
- updated help/README when CLI behavior changes

## License

MIT — see [`LICENSE`](./LICENSE).
