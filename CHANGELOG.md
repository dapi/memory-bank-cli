# Changelog

All notable changes to `memory-bank-cli` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.1.0] - 2026-08-15

### Added

- Add reviewed `pull --plan FILE` and `pull --apply-plan FILE` workflows for
  resolving two-sided adapted template conflicts atomically.
- Recover and verify historical merge bases from the immutable source ref in
  the ownership lock, with deterministic non-overlapping merge candidates.

## [2.0.1] - 2026-08-14

### Fixed

- Allow `pull --ask` to restore a missing user-owned template path from the
  canonical source and return it to managed ownership.
- Preserve adapted downstream files that were removed upstream while still
  applying unrelated safe template updates.

## [2.0.0] - 2026-08-14

### Added

- Add `memory-bank-cli update` as a verified self-update command for released
  macOS and Linux binaries.
- Add `memory-bank-cli pull` for managed Memory Bank template synchronization.

### Changed

- **Breaking:** `update` no longer synchronizes a project template. Use `pull`
  with the previous template and ownership flags instead.

## [1.8.1] - 2026-07-26

### Fixed

- Publish managed changes from an isolated temporary Git worktree and remove
  the destructive rollback of the user's clean upstream checkout.
- Reject path-swap and symlink redirection while publishing managed files or
  installing the GitHub adapter, including on Windows.
- Avoid false executable-mode drift findings from `doctor` on Windows.

## [1.8.0] - 2026-07-26

### Added

- Add interactive `memory-bank-cli update --ask` resolution for user-owned
  managed-file collisions, with keep or source-overwrite choices collected
  before the update plan is applied atomically.

## [1.7.3] - 2026-07-24

### Fixed

- Reduce human-readable `init` and `update` output to planned changes and
  conflicts, with a concise final result summary; retain the complete plan in
  JSON output.

## [1.7.2] - 2026-07-24

### Fixed

- Preserve existing conflicting managed files during `init` as downstream-owned
  instead of rejecting adoption or overwriting project-specific content.

## [1.7.1] - 2026-07-24

### Fixed

- Let `memory-bank-cli doctor --fix` resolve the default upstream without
  requiring explicit source provenance, matching `init` and `update`.

## [1.7.0] - 2026-07-24

### Added

- Add `memory-bank-cli doctor --repair` support for reconstructing a missing
  ownership lock from the managed template state, with diagnostics for
  unresolved conflicts.

## [1.6.0] - 2026-07-24

### Added

- Allow `memory-bank-cli init` to resolve the configured default source without
  requiring `--source`, `--template-version`, and `--source-ref`, matching
  `update` behavior. Explicit source flags remain supported as overrides.

## [1.5.0] - 2026-07-24

### Added

- Treat every tracked regular file below an upstream `template/` directory as
  managed payload during `init`, `update`, and `push`, including dotfiles and
  executable files outside `template/memory-bank/`.
- Migrate compatible legacy payload ownership locks conservatively, preserving
  locally customized files for explicit resolution.

### Fixed

- Keep managed template-owned agent files intact while publishing downstream
  changes upstream.
- Reject absolute paths in ownership locks and strengthen preflight validation
  for managed payload locations.

## [1.4.1] - 2026-07-24

### Fixed

- Publish to the canonical upstream `template/memory-bank/` payload root,
  retaining legacy-root fallback for compatible upstream repositories.
- Reject all unresolved downstream Git conflict statuses before an upstream
  publication plan can mutate state.
- Reject symlinked `memory-bank/` and `.repo` checkout ancestry, and provide
  corrective next steps in push preflight diagnostics.

## [1.4.0] - 2026-07-24

### Added

- Add `memory-bank-cli push` to publish managed downstream Memory Bank changes
  through a dedicated upstream branch and GitHub pull request.

### Fixed

- Prefer the canonical template payload over a project-local `memory-bank`
  copy when resolving managed source files.
- Harden upstream publication with base validation, reserved branch names,
  rollback-safe failure handling, and a pinned GitHub repository target.

## [1.3.0] - 2026-07-24

### Added

- Support templates whose Memory Bank payload is nested under
  `template/memory-bank`.

## [1.2.2] - 2026-07-24

### Changed

- Republish the template source-root support later released as `v1.3.0`.

## [1.2.1] - 2026-07-24

### Added

- Add hermetic local `init` and `update` end-to-end release validation.

### Changed

- Allow manually dispatched releases to publish after automated validation
  without a separate GitHub environment approval.

## [1.2.0] - 2026-07-24

### Changed

- Derive the template profile from the selected payload root.

## [1.1.0] - 2026-07-23

### Added

- Translate template source payload paths when applying downstream updates.
- Add stable downstream smoke tests and scheduled compatibility canaries.

## [1.0.1] - 2026-07-23

### Fixed

- Preserve the project prompt catalog while updating managed blocks.

## [1.0.0] - 2026-07-23

### Added

- Publish the standalone `memory-bank-cli` executable with `init`, `update`,
  `lint`, and read-only `doctor` workflows.
- Add opt-in GitHub workflow integration and managed Memory Bank blocks.
- Add Go install, Homebrew, and platform-specific release artifacts.

[Unreleased]: https://github.com/dapi/memory-bank-cli/compare/v2.1.0...HEAD
[2.1.0]: https://github.com/dapi/memory-bank-cli/compare/v2.0.1...v2.1.0
[2.0.1]: https://github.com/dapi/memory-bank-cli/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/dapi/memory-bank-cli/compare/v1.8.1...v2.0.0
[1.8.1]: https://github.com/dapi/memory-bank-cli/compare/v1.8.0...v1.8.1
[1.4.0]: https://github.com/dapi/memory-bank-cli/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/dapi/memory-bank-cli/compare/v1.2.2...v1.3.0
[1.2.2]: https://github.com/dapi/memory-bank-cli/compare/v1.2.1...v1.2.2
[1.2.1]: https://github.com/dapi/memory-bank-cli/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/dapi/memory-bank-cli/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/dapi/memory-bank-cli/compare/v1.0.1...v1.1.0
[1.0.1]: https://github.com/dapi/memory-bank-cli/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/dapi/memory-bank-cli/releases/tag/v1.0.0
