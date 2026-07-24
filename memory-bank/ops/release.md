---
title: memory-bank-cli Release
doc_kind: ops
doc_function: canonical
purpose: "Canonical release procedure and verification requirements."
derived_from:
  - ../engineering/architecture.md
  - ../features/FT-001-migrate-cli-source-and-rename-executable/brief.md
status: active
---

# Release

The repository has `.goreleaser.yml`; the release configuration builds only
`memory-bank-cli` and publishes GitHub and Homebrew artifacts. Releases are
created by manually dispatching `.github/workflows/release.yml` from `main`
with a complete `v`-prefixed semantic version.

## Procedure

1. Select the next semantic version from user-visible changes since the latest
   tag.
2. Update `CHANGELOG.md`: move the release changes out of `Unreleased`, add a
   dated `## [X.Y.Z]` section, and update comparison links.
3. Commit and merge the changelog with the exact source commit to be released.
4. Run `go test -count=1 -race ./...`, `go vet ./...`, `goreleaser check`, a
   GoReleaser snapshot build, and the release-scoped local E2E validation.
5. Dispatch the release workflow with the selected version. The workflow
   verifies the changelog section before creating the tag and publishing.
6. Wait for successful publication, install the exact tag with `go install`,
   and verify the reported version.

`AGENTS.md` makes the changelog update an automatic part of every agent-driven
release request. `scripts/check-release-changelog.sh` prevents a manually
dispatched release from bypassing the same requirement.
