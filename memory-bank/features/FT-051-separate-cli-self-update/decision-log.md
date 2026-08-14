---
title: "FT-051: Decision Log"
doc_kind: feature-support
doc_function: derived
purpose: "Audit ledger for FT-051 FPF reasoning and accepted self-update delivery decision; canonical facts remain in brief.md/design.md."
derived_from:
  - brief.md
  - ../../flows/feature.md
status: active
audience: humans_and_agents
must_not_define:
  - requirements
  - selected_solution
  - implementation_sequence
---

# FT-051: Decision Log

## Ownership

`brief.md` owns problem scope and verification. `design.md` owns the selected self-update solution. This ledger records FPF reasoning and external-project precedent; it does not duplicate canonical solution facts.

## Decisions and Open Questions

| ID | Status | Record | Facts considered | FPF reasoning | Canonical owner |
| --- | --- | --- | --- | --- | --- |
| `DEC-01` | accepted | Use release-backed replacement of the invoked executable on macOS/Linux; Windows reports manual replacement. | This repository already publishes matching raw platform binaries and `checksums.txt` through GoReleaser. `code-converge` uses the GitHub Release API, compatible assets, SHA-256 manifest and atomic replacement. `start-issue` adds staged-version verification, upgrade-only comparison and a Windows manual-update boundary. | FPF bounded-context reasoning keeps Memory Bank synchronization (`pull`) separate from distribution (`update`). Candidate comparison rejects Go-toolchain reinstallation (not all installed CLIs have Go), unverified replacement (violates `CON-01`), and multi-channel detection (unnecessary scope). The selected contract reuses this repository's release artifacts and both validated project patterns while preserving a safe no-op/failure boundary. | [design.md](design.md) `SOL-01`–`SOL-03`, `CTR-01`–`CTR-02`, `INV-01`–`INV-03` |
