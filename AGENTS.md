# Code Review Rules — reclaimit (Go CLI)

## Go

- Use `var` for package-level vars, `:=` for local
- Prefer `errors.Is`/`errors.As` over raw comparison
- Prefer specific types; use `any` when it is the appropriate generic constraint
- Keep functions cohesive and split them when it improves the API or testability
- Exported names: short, clear, no abbreviations (except standard: ID, URL, API)
- Errors should be wrapped with `fmt.Errorf("...: %w", err)`
- `init()` functions discouraged — use explicit setup

## Testing

- Table-driven tests preferred over single-test functions
- Test names: `TestFunctionName/TestCaseDescription`
- Use `t.TempDir()` and simulation adapters for filesystem and terminal tests
- Coverage ratchet: 67.5% minimum; target: 90% with no untested critical paths

## CLI

- Use `flag` or `cobra` consistently
- `--help` should be clear and actionable
- Exit codes: 0 = success, 1 = error
- Never panic in production code — return errors

## File Structure

- `cmd/reclaimit/main.go` — entrypoint only
- `internal/` — all business logic
- `*_test.go` — co-located with source files
- Keep packages focused: one responsibility per package
