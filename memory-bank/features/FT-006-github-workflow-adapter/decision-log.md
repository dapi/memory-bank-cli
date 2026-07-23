---
title: "FT-006: Decision Log"
doc_kind: feature-support
doc_function: derived
purpose: "Audit ledger for FT-006 provenance and FPF reasoning. It links to canonical owners and does not define requirements or a selected solution."
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

# FT-006: Decision Log

## Ownership

`brief.md` owns problem-space facts and blockers. A future `design.md` will own selected feature-local solution facts. If this ledger conflicts with a canonical owner, update that owner first and then this log.

## Decisions and Open Questions

| ID | Status | Record | Facts considered | FPF reasoning | Canonical owner |
| --- | --- | --- | --- | --- | --- |
| `DEC-01` | accepted | Use the explicit `memory-bank-cli github init|update` boundary with embedded assets and marker-owned blocks. | Issue #40 permits an opt-in command/config, requires no core dependency, specifies managed-marker/manifest ownership and asks that existing templates remain user-owned. Existing generic `init`/`update` already have a separate Memory Bank lock contract. | FPF B.5: a separate subcommand is the smallest hypothesis preserving the generic contract. Tests deductively verify opt-in invocation, preservation, conflict and dry-run predictions. | [design.md](design.md) `SOL-01`–`SOL-03` |
| `DEC-02` | accepted | Use `standard` validation: targeted adapter/CLI tests, full Go suite, vet, navigation audit and PR CI. | The project documents ordinary Go test/vet checks; the feature changes a CLI/filesystem contract but has no remote API or deployment. Issue #6 explicitly requires fixtures. | FPF B.5: this is the minimum evidence set that tests the selected hypothesis' observable safety properties without claiming release/deployment obligations. | [brief.md](brief.md) Validation Profile; [implementation-plan.md](implementation-plan.md) |
