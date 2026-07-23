---
title: "FT-005: Decision Log"
doc_kind: feature-support
doc_function: derived
purpose: "Audit ledger for FT-005 provenance and FPF reasoning. It links canonical owners and does not define requirements, selected solution or implementation sequence."
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

# FT-005: Decision Log

## Ownership

`brief.md` owns problem-space facts and verify; `design.md` owns selected feature-local solution facts. If this ledger conflicts with a canonical owner, update that owner first and then this log.

## Decisions and Open Questions

| ID | Status | Record | Facts considered | FPF reasoning | Canonical owner |
| --- | --- | --- | --- | --- | --- |
| `DEC-01` | accepted | Stable pins CLI and template to `v1.0.0` / `06f88e4638c6fd5f07cd05caf93e0601927faea3`. The scheduled/manual canary accepts separate refs, resolves them to immutable commits, installs the CLI by its resolved commit and clones the template at its resolved commit. Each run reports requested/resolved inputs and creates fresh source/downstream repos. | Issue #5 requires pinned stable and scheduled/manual canary; the CLI requires a clean local source plus version/full SHA; `v1.0.0` is published. | FPF B.5.2 prompt: reproducible stable evidence plus detectable compatibility drift. Candidates: moving `latest`; explicit refs resolved to SHAs; a new manifest. Explicit resolved refs win by parsimony, consistency with CLI provenance, and probeability: each result is replayable without adding a policy artifact. | [design.md](design.md) `SOL-01`, `SOL-02`, `CTR-01` |
| `DEC-02` | accepted | The fixture uses documented Go installation and separately checks every downloaded release binary against `checksums.txt` when both exist. The binary never replaces the Go-installed fixture CLI. | README documents Go installation; issue requests integrity when assets are introduced; `v1.0.0` has binaries and `checksums.txt`. | FPF candidates: Go only, binaries only, both checks. Both has the highest explanatory reach while separating user-path behavior from packaging evidence; its two predictions are directly testable. | [design.md](design.md) `SOL-03`, `CTR-02` |
| `DEC-03` | accepted | Select `standard`: targeted fixture/workflow coverage plus established full Go test/vet regression and CI evidence. | The feature changes CI and consumes a release, but FT-003 alone publishes/releases. | FPF compares existing `standard`, `release-deployment`, and a new profile. `standard` is the smallest existing vocabulary that covers fixture/workflow risk; the other options either exceed scope or invent new policy. | [brief.md](brief.md) Validation Profile; [implementation-plan.md](implementation-plan.md) |
