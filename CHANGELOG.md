# Changelog

All notable changes to this project will be documented in this file.

## [0.4.2](https://github.com/svg153/reclaimit/compare/v0.4.1...v0.4.2) (2026-08-21)


### Bug Fixes

* **cli:** register selection manifest flags ([9817515](https://github.com/svg153/reclaimit/commit/98175154ce9f2f964c5c96a413279a3c291bef5c))

## [0.4.1](https://github.com/svg153/reclaimit/compare/v0.4.0...v0.4.1) (2026-08-21)


### Bug Fixes

* **scanner:** keep older-than in AnalyzeOptions ([98fec13](https://github.com/svg153/reclaimit/commit/98fec137f912eb67a0994ed81b99b5415751077a))

## [0.4.0](https://github.com/svg153/reclaimit/compare/v0.3.0...v0.4.0) (2026-08-20)


### Features

* **cli:** add reviewed selection manifests ([a087cfc](https://github.com/svg153/reclaimit/commit/a087cfc825bd3169643c137b06110320cb474109))


### Bug Fixes

* **cli:** import duration parsing dependencies ([0767d4a](https://github.com/svg153/reclaimit/commit/0767d4aa5433fcc762e2db5b4bfa15569789ab05))
* repair selection manifest CI compilation ([5589ef0](https://github.com/svg153/reclaimit/commit/5589ef02a591ed701837cf55cb41cab80a9fa9af))

## [0.3.0](https://github.com/svg153/reclaimit/compare/v0.2.2...v0.3.0) (2026-08-20)


### Features

* **scanner:** filter candidates by age ([1f45282](https://github.com/svg153/reclaimit/commit/1f4528235260ba093695a538e66e67eea2451a8e))

## [0.2.2](https://github.com/svg153/reclaimit/compare/v0.2.1...v0.2.2) (2026-08-19)


### Bug Fixes

* **pages:** remove root symlink from static site source ([4399165](https://github.com/svg153/reclaimit/commit/439916519d3e9cff9436dada9a133da8ef5376f7))

## [0.2.1](https://github.com/svg153/reclaimit/compare/v0.2.0...v0.2.1) (2026-08-19)


### Bug Fixes

* **release:** fetch tags created during release workflow ([7b9683f](https://github.com/svg153/reclaimit/commit/7b9683f52a4804cf2c585fbaf9fb4b1793347f82))
* **release:** publish artifacts only from version tags ([7dd56e7](https://github.com/svg153/reclaimit/commit/7dd56e7e5c2273c93d5a50d43a661d92853775f3))

## [0.2.0](https://github.com/svg153/reclaimit/compare/v0.1.6...v0.2.0) (2026-08-14)


### Features

* add --dry-run to clean + 70+ tests + Gentle-AI setup ([#16](https://github.com/svg153/reclaimit/issues/16)) ([cbbbf21](https://github.com/svg153/reclaimit/commit/cbbbf21a209df9d045ee2dc6f62931944d67ee70))
* add --quiet flag for non-interactive / scripted output ([15212f3](https://github.com/svg153/reclaimit/commit/15212f3771d3595a7a0fdd2e78764762588fef1b)), closes [#25](https://github.com/svg153/reclaimit/issues/25)
* add automated multi-distro release notes generation ([6a47145](https://github.com/svg153/reclaimit/commit/6a471456adaa1e56abcd66c9c79f0c333319fb5a)), closes [#26](https://github.com/svg153/reclaimit/issues/26)
* add Bun package manager cache support ([e954df7](https://github.com/svg153/reclaimit/commit/e954df77dc9a63a3658dd34a38daeb53e1f77022))
* add grouping and summarization features for disk usage analysis ([9e1e15e](https://github.com/svg153/reclaimit/commit/9e1e15e3f93a8f27b26bcfa6aac651719c106b1f))
* add JSON output, .reclaimitignore support, and 12 new categories ([7e1cb86](https://github.com/svg153/reclaimit/commit/7e1cb86ba912a846792f6a7e910b278f7fc4f24b))
* add landing page and comprehensive SEO optimization ([f2074b6](https://github.com/svg153/reclaimit/commit/f2074b602ade80ed15e751cc9d28b5e2946710e7))
* add macOS cleanup candidates support ([d9199bc](https://github.com/svg153/reclaimit/commit/d9199bcb4faed12046537844c3257adbb3c2448c))
* **scanner:** add concurrent traversal and scan limits ([be3f9ca](https://github.com/svg153/reclaimit/commit/be3f9ca1b71a9ee75fe6252679b356b3065ecfbc))
* **scanner:** detect pip and pipx data directories ([55c544a](https://github.com/svg153/reclaimit/commit/55c544a39cda8895b837df9f3829d944e695889b))
* unify gentle-ai Hermes bridge + ecosystem setup (APM) ([#20](https://github.com/svg153/reclaimit/issues/20)) ([8753cab](https://github.com/svg153/reclaimit/commit/8753cab301ae231dbd557275b4dec30a2ba715ba))


### Bug Fixes

* auto-enable GitHub Pages in docs workflow ([a2bcb35](https://github.com/svg153/reclaimit/commit/a2bcb35ae747b524122665a6f571eae36dbff250))
* **clean:** quarantine verified candidates before deletion ([#50](https://github.com/svg153/reclaimit/issues/50)) ([c22ae65](https://github.com/svg153/reclaimit/commit/c22ae654bb6dba6f8a129b9014d0afa244b38fd8))
* **clean:** verify candidates before deletion ([b9cb9ca](https://github.com/svg153/reclaimit/commit/b9cb9ca4f60effbb238aec9226b33eb81e4f1f2f))
* **cli:** reject ambiguous command input ([#57](https://github.com/svg153/reclaimit/issues/57)) ([e0e85ab](https://github.com/svg153/reclaimit/commit/e0e85abb60f1508b929629aed083667caf22a6cd))
* grant CodeQL workflow required permissions ([420dd07](https://github.com/svg153/reclaimit/commit/420dd079b4862a9655b5b1184f5bc5595e51c1a9))
* **release:** publish a true multi-platform container image ([#53](https://github.com/svg153/reclaimit/issues/53)) ([dfe6b65](https://github.com/svg153/reclaimit/commit/dfe6b654ed6c35adb973d22a6715a3b7290bbb73))
* remove conflicting deploy-pages workflow (legacy Pages mode active) ([2d5b599](https://github.com/svg153/reclaimit/commit/2d5b599a38e25141fa2b784fbbe562c0e3d4b4a0))
* **renderer:** neutralize untrusted display values ([481e4b3](https://github.com/svg153/reclaimit/commit/481e4b36a90c9662858f84977eb1e66bd9d32df3)), closes [#64](https://github.com/svg153/reclaimit/issues/64)
* restore repository quality and publish SEO-ready site ([#43](https://github.com/svg153/reclaimit/issues/43)) ([7604a29](https://github.com/svg153/reclaimit/commit/7604a29b78fafc253fd367363a8721801ba40014))
* **scanner:** preserve Bun runtime and global tools ([#61](https://github.com/svg153/reclaimit/issues/61)) ([547833f](https://github.com/svg153/reclaimit/commit/547833f372945c4a447c87238f9dc221b2968210))
* **scanner:** reject unknown category filters ([#60](https://github.com/svg153/reclaimit/issues/60)) ([4f0a736](https://github.com/svg153/reclaimit/commit/4f0a7361f445fa6b9d2c5de202579ae4862d1736))
* **security:** require patched Go toolchains ([#55](https://github.com/svg153/reclaimit/issues/55)) ([08d3fb6](https://github.com/svg153/reclaimit/commit/08d3fb687a83ab02a9e0fab55777bcea3e7c2e22))
* switch docs deployment to official Pages actions ([53f5cc7](https://github.com/svg153/reclaimit/commit/53f5cc7a15780abafb69a098974b2ac58048cc63))
* **tui:** emit a paste-safe selection command ([#59](https://github.com/svg153/reclaimit/issues/59)) ([a3ad1ca](https://github.com/svg153/reclaimit/commit/a3ad1cae63b5befbd19a586d8399ee18ba884371))
* update site URLs to GitHub Pages (svg153.github.io/reclaimit) ([dd68eb6](https://github.com/svg153/reclaimit/commit/dd68eb64045fb214ad0eff4bd77d5d11a3f769d8))
* validate Taskfile and stabilize docs publishing ([e5aac99](https://github.com/svg153/reclaimit/commit/e5aac99e9d089e071a3076e8bc88034a3843cb91))


### Performance Improvements

* **scanner:** enforce a traversal-wide worker ceiling ([#51](https://github.com/svg153/reclaimit/issues/51)) ([eb3d6f1](https://github.com/svg153/reclaimit/commit/eb3d6f19310a3a655419ac1bf6094c72ee1682db))

## [Unreleased]

### Added

- guarded cleanup with pre-delete identity checks, same-filesystem quarantine,
  and recovery paths for changed data
- traversal-wide worker limit, cancellation, depth limits, JSON output, and
  reusable coverage enforcement
- pip, pipx, Bun download-cache, and macOS cleanup categories
- reproducible terminal demo, benchmark methodology, growth baseline, and
  repository security and community health files
- verified release archives, Linux packages, checksums, and multi-platform
  container configuration

### Changed

- release workflows inject version metadata, pin actions by commit, and publish
  true amd64/arm64 container images
- cleanup categories and CLI filters now fail closed on unsafe or unknown input
- README and landing-page copy now use verified developer-cleanup positioning
- repository no longer tracks generated binaries, editor databases, or dead
  placeholder packages

### Fixed

- total scanned bytes no longer depend on the top-N display limit
- unknown commands and positional arguments no longer trigger a default scan
- generated TUI reproduction commands are safe to paste into a POSIX shell
- Bun cleanup preserves the runtime and global tools outside
  `.bun/install/cache`
- GitHub Pages, installer verification, release targets, Windows metrics, race
  handling, and package-level coverage regressions

## [0.1.6] - 2026-05-21

### Added

- multi-platform release artifacts
- improved TUI tree and markdown reporting
- Dependabot, CodeQL and GitHub Pages automation

### Fixed

- cross-platform filesystem usage build issues
- release artifact upload flow
