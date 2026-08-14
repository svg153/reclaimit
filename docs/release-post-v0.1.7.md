# Draft release post: reclaimit v0.1.7

> Publication guard: this is a draft. Publish it only after the `v0.1.7` tag,
> archives, packages, checksum manifest, and container image are public and
> independently verified. Remove any channel that was not actually published.

reclaimit helps developers find known, regenerable artifacts and review them by
project before cleanup. Unlike a general disk analyzer, it focuses on the
question behind the size: what created this directory, and how would I recreate
it?

This patch release repairs the distribution path and strengthens the cleanup
contract:

- archives for Linux, macOS, and Windows on amd64 and arm64;
- Linux DEB, RPM, and APK packages;
- a checksum-verifying installer;
- a non-root GHCR image for amd64 and arm64;
- same-filesystem quarantine and pre-delete revalidation;
- traversal-wide worker limits and cancellation;
- plain-text, Markdown, and JSON cleanup results.

See the synthetic
[analyze → TUI → dry-run demo](https://github.com/svg153/reclaimit#reclaimit--find-developer-files-you-can-safely-clean-up)
and its [reproduction steps](https://github.com/svg153/reclaimit/blob/main/docs/demo.md).
The 20 MiB shown there belongs to generated test files; it is not a typical
savings claim.

Install after publication:

```sh
curl -fsSL https://raw.githubusercontent.com/svg153/reclaimit/main/install.sh | bash
```

Then start read-only:

```sh
reclaimit analyze --root "$HOME/code"
```

Category ideas and anonymized aggregate observations are welcome in
[issue #54](https://github.com/svg153/reclaimit/issues/54). Participation is
optional, and reclaimit does not transmit scan data.
