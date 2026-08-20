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
