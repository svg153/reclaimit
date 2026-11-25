# Field evidence: cache, store, and agent-artifact categories (2026-11-25)

This document records the measurements that motivated the categories added by
`feat(scanner): field-evidence categories`. It follows the
[category safety contract](category-safety-contract.md): every category below
has an exact, tool-owned path suffix, an observed size scale, and a risk level
justified by what deletion actually costs.

## Privacy by design

This document is intentionally coarse. It contains **no** exact sizes, no
usernames, hostnames, or project names, and no timeline of any particular
machine. Sizes are recorded as buckets (`<1 GB`, `1-5 GB`, `5-10 GB`,
`>10 GB`) because the evidential point is the order of magnitude and the
regrowth behavior, never the identity of the system that was measured.

All paths are anonymized: `~` is the user home and `%LocalAppData%` is the
Windows per-user local application data root. The suffixes themselves are
documented tool defaults, not machine-specific state.

## Method

Three read-only passes spread over several weeks on a Windows host with a WSL
distribution used for day-to-day development:

- WSL directories: `du -sb` on individual tool directories.
- Windows directories: `robocopy <dir> NULL /L /E /BYTES /NJH /NJS /NFL /NDL`
  to sum sizes without touching data.

No directory was deleted during the study. The repeated passes exist to answer
one question: does the directory grow back, and how fast?

## Observed evidence

| Category key | Path suffix(s) | Where | Size bucket | Regrowth behavior |
| --- | --- | --- | --- | --- |
| `npm-cache` | `.npm/_cacache`, `.npm/_npx` | WSL | >10 GB | Stable cache size; refilled quickly after manual pruning |
| `npm-cache` | `%LocalAppData%/npm-cache` | Windows | 5-10 GB | Stable across passes |
| `go-mod-cache` | `go/pkg/mod` | WSL | 5-10 GB | Monotonic growth (new module versions append) |
| `uv-cache` | `.cache/uv` | WSL | 5-10 GB | Steady growth between passes |
| `uv-cache` | `%LocalAppData%/uv/cache` | Windows | 1-5 GB | Stable across passes |
| `playwright-browsers` | `.cache/ms-playwright` | WSL | 5-10 GB | Growth tied to browser version bumps |
| `playwright-browsers` | `%LocalAppData%/ms-playwright` | Windows | 1-5 GB | Stable; two browser major versions coexisting is normal |
| `puppeteer-cache` | `.cache/puppeteer` | WSL | 1-5 GB | Stable; bounded by downloaded Chromium builds |
| `pre-commit-cache` | `.cache/pre-commit` | WSL | <1 GB | Stable across passes |
| `pnpm-store` | `.local/share/pnpm/store` | WSL | 1-5 GB | Grows with project churn |
| `pnpm-store` | `%LocalAppData%/pnpm/store` | Windows | <1 GB | Stable across passes |
| `vscode-vsix-cache` | `%LocalAppData%/Code/CachedExtensionVSIXs` (+ Insiders/OSS variants) | Windows | 1-5 GB | Stale VSIXs accumulate on every extension update |
| `agent-worktrees` | `~/.copilot/copilot-worktrees`, `~/.codex/worktrees` | Windows | >10 GB | Grows on every agent coding session; worktrees outlive their sessions |
| `agent-session-state` | `~/.copilot/session-state`, `~/.codex/sessions` | Windows | >10 GB | Largest observed agent-artifact growth source; conversation records and checkpoints accumulate per session |

## Risk-level decisions

- **low** — `npm-cache`, `go-mod-cache`, `uv-cache`, `playwright-browsers`,
  `puppeteer-cache`, `pre-commit-cache`: exact tool-owned caches under a
  tool-specific root. Deletion costs only re-downloading, and every tool
  documents how to repopulate (`npm install`, `go clean -modcache`,
  `uv sync`, `npx playwright install`, and so on). The suffixes never match
  the tool home itself: `.npm` alone is not a candidate because it also holds
  `_logs`, and scope is limited to the documented cache subtrees.
- **review** — `pnpm-store`: the store is content-addressable and project
  `node_modules` directories may be hardlinked into it; deleting the store can
  break working checkouts. The recommendable operation is `pnpm store prune`,
  not directory deletion, so the category surfaces for review only.
  **review** — `vscode-vsix-cache`: re-downloadable from the marketplace, but
  extension installers double as offline install sources; surfacing it requires
  explicit review.
- **opt-in** — `agent-worktrees`, `agent-session-state`: these are not caches.
  `copilot-worktrees` is a git worktree root: a worktree may contain
  uncommitted work that exists nowhere else. `session-state` holds
  conversation records, checkpoints, and session artifacts that may include an
  active session. Both are excluded from every risk profile and appear only
  when explicitly enabled with `--include-category`, per the contract's
  "excluded from normal cleanup unless the user explicitly enables the
  category" rule.

## Negative evidence encoded in tests

- `ms-playwright-mcp` must not match `playwright-browsers`; matching compares
  complete path segments, not substrings.
- `.npm`, `go`, `go/pkg`, `Library/pnpm`, `.local/share/pnpm`, `.copilot`, and
  `.codex` must not match by themselves: tool homes contain state beyond the
  documented caches.
- `.cache/uvx`, `AppData/Local/npm-cache-mcp`, `.cache/puppeteer-mcp` and
  similar near-misses must not match a suffix at a segment boundary.

## What this changes for users

Aggregated across the measured passes, these categories reach double-digit-GB
totals on agent-heavy development machines, dominated by agent worktrees,
agent session state, and package-manager caches. `analyze` with the default
`balanced` profile now surfaces the cache categories automatically; the two
agent categories stay invisible unless explicitly requested, which keeps
normal reports focused on true regenerable artifacts.
