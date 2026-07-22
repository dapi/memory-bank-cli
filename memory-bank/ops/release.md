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

The repository has `.goreleaser.yml`; FT-001 states the release configuration must build only `mb-cli`. FT-001 also records that the target had no tags/releases at its observed baseline and explicitly hands off tag publication and installation documentation to issue #3.

[FT-003](../features/FT-003-establish-mb-cli-releases/brief.md) is the active release-delivery design: it introduces validation-before-publication, a tag-driven `v1.0.0` GitHub release, Go-install verification and install/upgrade documentation. Its `design.md` owns the selected pipeline; this document remains the owner of observed release configuration facts.

## Verification Gap

The repository workflow runs the target validation path for pull requests, `main` and semantic-version tags; its publication job is tag-triggered and depends on successful validation. No tag or release evidence exists yet. Public publication remains blocked until a release maintainer approves the external effect and confirms any credentials required by the existing configuration. Artifact platforms and configured destinations remain those in `.goreleaser.yml`; no new signing/distribution policy is inferred.
