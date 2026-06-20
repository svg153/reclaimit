# Code Review Rules — reclaimit (Go CLI)

## Go

- Use `var` for package-level vars, `:=` for local
- Prefer `errors.Is`/`errors.As` over raw comparison
- No `any` — use specific types or generics
- Functions > 50 lines should be split
- Exported names: short, clear, no abbreviations (except standard: ID, URL, API)
- Errors should be wrapped with `fmt.Errorf("...: %w", err)`
- `init()` functions discouraged — use explicit setup

## Testing

- Table-driven tests preferred over single-test functions
- Test names: `TestFunctionName/TestCaseDescription`
- Mock external deps (filesystem, CLI) — test logic, not I/O
- Coverage threshold: 70% minimum

## CLI

- Use `flag` or `cobra` consistently
- `--help` should be clear and actionable
- Exit codes: 0 = success, 1 = error, 2 = usage error
- Never panic in production code — return errors

## File Structure

- `cmd/reclaimit/main.go` — entrypoint only
- `internal/` — all business logic
- `*_test.go` — co-located with source files
- Keep packages focused: one responsibility per package
