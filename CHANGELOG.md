# Changelog

All notable changes to `memory-bank-cli` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

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

[Unreleased]: https://github.com/dapi/memory-bank-cli/compare/v1.4.0...HEAD
[1.4.0]: https://github.com/dapi/memory-bank-cli/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/dapi/memory-bank-cli/compare/v1.2.2...v1.3.0
[1.2.2]: https://github.com/dapi/memory-bank-cli/compare/v1.2.1...v1.2.2
[1.2.1]: https://github.com/dapi/memory-bank-cli/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/dapi/memory-bank-cli/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/dapi/memory-bank-cli/compare/v1.0.1...v1.1.0
[1.0.1]: https://github.com/dapi/memory-bank-cli/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/dapi/memory-bank-cli/releases/tag/v1.0.0
