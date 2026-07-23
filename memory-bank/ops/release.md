---
title: memory-bank-cli Release
doc_kind: ops
doc_function: canonical
purpose: "Canonical owner confirmed release-build facts and release unknowns."
derived_from:
  - ../engineering/architecture.md
  - ../features/FT-001-migrate-cli-source-and-rename-executable/brief.md
status: active
---

# Release

The repository has `.goreleaser.yml`; the release configuration builds only
`memory-bank-cli`. A prior public module version predates the current
executable identity and is not installation evidence for it.

[FT-003](../features/FT-003-establish-memory-bank-cli-releases/brief.md) is the
active release-delivery design for the first `memory-bank-cli` publication:
validation before a `v1.0.0` GitHub release, Go-install verification and
install/upgrade documentation. Its `design.md` owns the selected pipeline;
this document remains the owner of observed release configuration facts.

## Verification Gap

The repository workflow runs the target validation path for pull requests,
`main` and manually dispatched releases. Its publication job depends on
successful validation and the required-reviewer protection rule on the GitHub
`release` environment. No `v1.0.0` publication evidence exists yet. Public
publication also requires every credential declared by `.goreleaser.yml`;
artifact platforms and configured destinations remain unchanged.
