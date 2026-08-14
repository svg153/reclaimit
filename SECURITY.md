# Security Policy

## Reporting a Vulnerability

Please do not open a public issue for security-sensitive problems.

Instead, report vulnerabilities through a GitHub security advisory or contact the maintainer privately before disclosure.

## Release Integrity

Official releases are built by the repository's GoReleaser workflow. Each
release publishes a `reclaimit_<version>_checksums.txt` manifest containing the
SHA-256 digest of every archive.

`install.sh` requires that manifest, selects the entry whose filename exactly
matches the requested platform archive, and verifies it with `sha256sum` on
Linux or `shasum -a 256` on macOS. A missing, malformed, or mismatched checksum
stops installation before the destination directory is created. When
downloading manually, verify the same manifest entry before extracting or
running the binary.

Checksums detect corruption and disagreement between the archive and release
manifest; they are not a separate cryptographic signature. The trust boundary
still includes GitHub release delivery and the repository's release workflow.
Workflow actions are pinned to full commit SHAs, with version comments retained
for review and Dependabot updates.

## Supported Versions

| Version | Supported |
| --- | --- |
| 0.1.x | yes |
