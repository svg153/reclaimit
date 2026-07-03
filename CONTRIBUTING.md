# Contributing to reclaimit

Thank you for your interest in contributing to `reclaimit`! This guide outlines our development workflow, coding standards, and repository conventions.

---

## Getting Started

### Prerequisites
- **Go**: Version 1.24 or 1.25+ is required (see `go.mod`).
- **Task**: We use [Taskfile](https://taskfile.dev) to manage build, lint, and test scripts. Install it via `brew install go-task` or follow their installation guide.

### Setup
1. Fork and clone the repository:
   ```bash
   git clone https://github.com/your-username/reclaimit.git
   cd reclaimit
   ```
2. Build the binary to verify your setup:
   ```bash
   task build
   ```
   The compiled binary will be placed at `./bin/reclaimit`.

---

## Development Workflow

We use standard `task` targets to validate changes locally before committing:

- `task build` — Compiles the CLI binary to `./bin/reclaimit`
- `task fmt` — Formats Go source files using `gofmt`
- `task vet` — Runs `go vet` static analysis
- `task lint` — Runs `golangci-lint` (pinned version run via `go run`)
- `task test` — Runs the test suite and generates a coverage profile (`coverage.out`)
- `task check` — Runs full quality gates (`fmt`, `vet`, `lint`, `vulncheck`, `coverage-check`, `build`)
- `task ci` — Runs local CI validations (`fmt`, `vet`, `test`, `build`)

---

## Testing

Quality and testing are central to the repository:
- **Coverage**: We enforce a minimum coverage threshold of **70%** (validated in CI and locally via `task coverage-check`).
- **Table-Driven Tests**: We prefer table-driven test structures with explicit sub-tests using `t.Run(tt.name, ...)`.
- **Test Names**: Use the format `TestFunctionName/TestCaseDescription`.
- **Isolation**: Always use `t.TempDir()` for file-system tests instead of operating on real user home directories. Mock external dependencies like filesystems or CLI executions.
- **Co-location**: Put test files (`*_test.go`) in the same package and directory as their corresponding source files.

---

## Code Style

Our Go code style conforms to the following guidelines (configured in `AGENTS.md`):

- **Variable Declarations**: Use `var` for package-level variables and short declarations (`:=`) for local variables.
- **Error Handling**: 
  - Prefer `errors.Is` and `errors.As` over raw string comparison for errors.
  - Wrap errors using `fmt.Errorf("context: %w", err)`.
- **Generics**: Do not use the `any` type — use specific types or constraints.
- **Complexity**: Functions longer than **50 lines** should be refactored and split into smaller helpers.
- **Exported Names**: Keep exported names short, clear, and descriptive. Avoid abbreviations except standard abbreviations like `ID`, `URL`, or `API`.
- **Explicit Setup**: Do not use Go `init()` functions. Use explicit setup/initialization functions instead.

---

## Pull Request Process

1. Create a descriptive branch for your changes (e.g. `feat/bun-cache` or `fix/traversal-symlink`).
2. Run validation tools locally before pushing:
   ```bash
   task ci
   ```
3. If your changes affect the CLI flags, commands, or behavior, ensure you update:
   - The `--help` messages and topic usage text.
   - The `README.md` file.
4. Keep pull requests focused on a single responsibility. Avoid combining unrelated fixes or formatting changes.
5. Fill out the Pull Request template completely, ensuring all checklist tasks pass.

---

## Adding a New Category

To add a new category (e.g., matching a new build tool cache), follow these steps:

1. **Add Category in `scanner_categories.go`**:
   Define the new category inside the `categories` slice in `scanner_categories.go`. Use `newDirCategory` for folder matches or `newFileCategory` for file matches:
   ```go
   // Example for a directory category
   newDirCategory("my-cache", ".mycache", "My tool package cache that can be rebuilt.", ".mycache"),
   ```
2. **Add Tests**:
   Create a test case in `scanner_extra_test.go` or `scanner_test.go` verifying that paths matching this category are correctly detected and sized. Use `filepath.Base` or `filepath.Clean` to keep tests platform-independent:
   ```go
   func TestAnalyzeFindsMyCache(t *testing.T) {
       root := t.TempDir()
       mustMkdir(t, filepath.Join(root, ".mycache"))
       mustWriteFile(t, filepath.Join(root, ".mycache", "file.bin"), "data")
       // ... run Analyze and assert category key and size ...
   }
   ```
3. **Update README.md**:
   Add the new category/folder name to the list of common development leftovers in `README.md`.
4. **Validate**:
   Run `task check` to ensure your tests pass and coverages are met.

---

## Releasing

- **CI Releases**: Official releases are built and published by GitHub Actions using GoReleaser when a semantic version tag (e.g., `v0.1.7`) is pushed to the repository.
- **Snapshot Releases**: To build and test a release snapshot locally, run:
  ```bash
  task dist
  ```
  *(Note: This requires `goreleaser` to be installed locally).*
