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

## Verification Gap

Current sources do not document the release trigger, artifact platforms, signing, publication destination, rollback or installation procedure. These must remain open until an authoritative release source is added.
