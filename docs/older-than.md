# Age-based candidate filtering

`reclaimit analyze --older-than DURATION` keeps candidates whose latest verified modification time is strictly older than the requested duration.

Accepted duration forms are Go durations such as `24h`, `720h`, and `1h30m`, plus whole days such as `30d` (`30 × 24 hours`). The value must be positive. The comparison is made in UTC against the scan start time:

- a candidate exactly on the cutoff is not selected;
- only a candidate strictly before the cutoff is selected;
- a candidate with an unknown/zero modification time is excluded when the filter is active.

The filter applies before category and group summaries are rebuilt, so reports and cleanup operate on the same reviewed set. It does not authorize deletion: `clean` still requires `--dry-run` or `--yes`, and all normal cleanup verification remains active.

Example:

```sh
reclaimit analyze --root ~/projects --older-than 30d --format markdown
reclaimit clean --root ~/projects --older-than 30d --dry-run
```

This option is intentionally a filter over verified candidate metadata, not a claim that the data is disposable or that a typical amount of space will be recovered.
