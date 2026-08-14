# Reproducible terminal demo

The README demo uses a generated workspace rather than a real home directory.
It contains two repositories, one 12 MiB `node_modules` candidate, and one 8
MiB Python virtual environment candidate. Its numbers demonstrate output
formatting; they are not estimates of typical disk savings.

Build the current source and run all three stages:

```sh
go build -o ./bin/reclaimit ./cmd/reclaimit
./scripts/demo.sh ./bin/reclaimit
```

The script runs:

1. a read-only `analyze`;
2. an interactive TUI review when attached to a terminal;
3. a category-filtered `clean --dry-run` that removes nothing.

The temporary fixture is deleted when the script exits. It never reads a user
workspace and never sends telemetry or scan results.

To verify only the non-interactive stages, redirect standard input:

```sh
./scripts/demo.sh ./bin/reclaimit </dev/null
```
