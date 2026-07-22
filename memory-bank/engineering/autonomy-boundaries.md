---
title: memory-bank-cli Autonomy Boundaries
doc_kind: engineering
doc_function: canonical
purpose: "Canonical owner границ безопасной automation и mutation behaviour."
derived_from:
  - architecture.md
  - ../domain/rules.md
status: active
---

# Autonomy Boundaries

`init` and `update` may mutate only the resolved downstream repository after source validation and full planning. `--dry-run` is the non-mutating preview. `doctor` is explicitly read-only; lint only audits.

The implementation rejects source/repository overlap, repository-relative path escape, forbidden lock payload, unsafe symlink ancestry and unsafe topology changes. It manages one specified agent instruction file rather than arbitrarily rewriting project instructions.

Release publication, tagging, credentials, CI configuration and installation documentation are not implemented or authorised by the FT-001 scope recorded in this repository.
