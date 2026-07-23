---
title: memory-bank-cli Coding Style
doc_kind: engineering
doc_function: convention
purpose: "Canonical owner observed code conventions; code remains owner of implementation details."
derived_from:
  - architecture.md
  - testing-policy.md
status: active
---

# Coding Style

- Write Go packages under `internal/` for non-public implementation; keep `cmd/memory-bank-cli` thin.
- Expose command behavior through `cli.Run(arguments, version, stdout, stderr)` so tests can assert output and exit codes without spawning the binary.
- Use explicit typed report/option structures for machine-readable contracts.
- Keep filesystem and Git operations validated before mutation; preserve rollback/error paths with regression tests.
- Use table-driven tests where multiple input variants define one contract.

No repository-specific formatter, linter configuration or commit-message convention is present in sources. Standard Go formatting is expected but not a separately documented project rule.
