---
title: memory-bank-cli Validation Profiles
doc_kind: engineering
doc_function: canonical
purpose: "Canonical owner validation profile facts recorded by project delivery documentation."
derived_from:
  - testing-policy.md
  - ../features/FT-001-migrate-cli-source-and-rename-executable/brief.md
status: active
---

# Validation Profiles

FT-001 selected `release-deployment` because it changes release-build artifact configuration. The feature explicitly does not publish a release; issue #3 owns publication/tag/install documentation evidence.

For ordinary code changes, this repository only confirms Go test/vet checks and contract-focused regression tests. No general profile catalogue or CI matrix is documented.
