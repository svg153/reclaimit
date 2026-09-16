# Selection manifests

A selection manifest records a reviewed candidate set for a later dry run. It is versioned JSON and includes the scan root, creation time, exclusions, and each candidate's path, category, type, size, and modification time.

Export a reviewed set:

```sh
reclaimit analyze --root ~/projects --older-than 30d --export-selection selection.json --format markdown
```

Import validates the manifest before it can affect a report or cleanup:

```sh
reclaimit clean --root ~/projects --import-selection selection.json --dry-run
```

Safety rules:

- unsupported schema versions and root mismatches fail closed before filesystem mutation;
- missing or changed candidates are excluded and reported as `selection_mismatches` in JSON;
- import is never an implicit `--yes` and never bypasses the normal cleanup verification;
- the manifest is written with owner-only permissions (`0600`);
- paths are compared using the platform's canonical cleaned path semantics.

The manifest is a review aid, not proof that a candidate is disposable. Re-run a dry run and inspect the report before confirming deletion.

## Explicit cleanup plans

For a named artifact that can be reviewed and applied later, export a cleanup
plan instead of relying on a selection manifest:

```sh
reclaimit analyze --root ~/projects --export-plan cleanup-plan.json
reclaimit clean --root ~/projects --plan cleanup-plan.json --dry-run
reclaimit clean --root ~/projects --plan cleanup-plan.json --yes
```

A cleanup plan is versioned JSON with `action: "delete"`, the original scan
root, and the candidate identity snapshot. `--plan` is accepted only by
`clean`. It rejects unsupported schemas, root mismatches, paths outside the
original root, missing or changed candidates, and unsupported category changes.
Any mismatch aborts the complete plan before deletion begins. `--dry-run` is
still non-destructive, while `--yes` is the explicit apply confirmation.

Plans contain real filesystem paths and are written with owner-only permissions
(`0600`); do not publish or attach them to public reports.
