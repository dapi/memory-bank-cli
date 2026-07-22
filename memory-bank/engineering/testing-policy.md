---
title: memory-bank-cli Testing Policy
doc_kind: engineering
doc_function: canonical
purpose: "Canonical owner observed test coverage and verification expectations."
derived_from:
  - architecture.md
  - ../features/FT-001-migrate-cli-source-and-rename-executable/brief.md
status: active
---

# Testing Policy

Tests are Go package tests under `internal/` and cover CLI contracts, ownership classification/source validation/transactions/topology/symlink handling, doctor governance/report behavior, lint reports and repository root discovery. Fixtures live beside relevant tests under `testdata/`.

For FT-001, the established verification contract is:

```sh
go test -count=1 -race ./...
go vet ./...
```

Tests should preserve command semantics, exit codes and versioned JSON fields. Security-sensitive filesystem behavior needs platform-aware tests where applicable. The repository has no documented CI workflow, coverage threshold, benchmark policy or release gate beyond the FT-001 evidence.
