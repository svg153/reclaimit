# Code Review Rules for reclaimit (Go)

## Language & Style
- Go idioms: `gofmt`, `go vet`, `staticcheck`
- Conventional commits only (no AI attribution)
- Error handling: wrap with `fmt.Errorf("...", err)`, not `%w` for user-facing errors
- Interfaces: small, focused, define at call site

## Review Checklist
1. Does the code follow Go conventions?
2. Are errors handled properly?
3. Is there unnecessary complexity?
4. Are tests meaningful (not just coverage)?
5. Does it break existing behavior?

## SDD Integration
- Run `golangci-lint` before committing
- All changes must pass `go test ./...`
- PR description must reference the change spec
