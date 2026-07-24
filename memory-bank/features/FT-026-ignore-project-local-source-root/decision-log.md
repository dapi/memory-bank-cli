---
title: "FT-026: Decision Log"
doc_kind: feature-support
doc_function: derived
purpose: "Audit ledger for FT-026 provenance and FPF reasoning; canonical facts remain in brief/design."
derived_from:
  - brief.md
  - design.md
  - ../../flows/feature.md
status: active
audience: humans_and_agents
must_not_define:
  - requirements
  - selected_solution
  - implementation_sequence
---

# FT-026: Decision Log

## Ownership

`brief.md` owns problem scope and verification. `design.md` owns selected behavior. This ledger records evidence and reasoning; update a canonical owner first if they conflict.

## Decisions and Open Questions

| ID | Status | Record | Facts considered | FPF reasoning | Canonical owner |
| --- | --- | --- | --- | --- | --- |
| `DEC-01` | accepted | Treat target `template/memory-bank` as authoritative when present; only when absent evaluate bounded legacy roots. | Issue #26 requires target selection when present, exclusion of locked local `memory-bank/`, and legacy compatibility only without target. Current `source.go` treats all three roots equally and rejects a multiple-root set; init/update already translate selected source to downstream `memory-bank/`. | FPF bounded-context and strict-distinction reasoning separates source-template identity from downstream/project-local identity. Candidate comparison rejects equal priority (fails acceptance), lock-only exception (does not implement target precedence), and downstream rename (violates scope). The selected minimal boundary change explains all facts. Evidence provenance is issue #26 and current source/tests; `CHK-*` remain future run-time evidence, not claimed results. | [design.md](design.md) `SOL-01`–`SOL-03`, `CTR-01`, `INV-01`–`INV-02` |

## Open Questions

`none`: issue #26 supplies target-present and target-absent behavior. It does not require new lock interpretation, downstream naming, or a legacy-retirement date; none is inferred.
