---
title: "FT-014: Decision Log"
doc_kind: feature-support
doc_function: derived
purpose: "Audit ledger for FT-014 provenance and FPF reasoning; canonical facts remain in brief/design."
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

# FT-014: Decision Log

## Ownership

`brief.md` owns problem-space scope, validation and verify. `design.md` owns the selected source/downstream bridge and release rollout. This ledger records why decisions were accepted; if it conflicts with a canonical owner, update that owner first and then this log.

## Decisions and Open Questions

| ID | Status | Record | Facts considered | FPF reasoning | Canonical owner |
| --- | --- | --- | --- | --- | --- |
| `DEC-01` | accepted | Treat source payload and downstream payload as separate semantic contexts joined by one explicit path bridge. | Issues #14/#63 assign `memory-bank-template/` to the source repository and preserve `memory-bank/` for installed payload, lock, downstream routing and defaults. Current planner/classifier/lock and generated managed agent block operate on `memory-bank/`; current template-profile doctor independently checks a repository-local route. | FPF bounded-context discipline says locally valid meanings must not leak across the boundary and translation must be explicit. Candidate A, translate once before planning and derive template-local diagnostics from source scope, preserves both contexts. Candidate B, rename globally, violates downstream invariants. Scope fit and consistency select A. | [design.md](design.md) Grounding, `SOL-01`–`SOL-03`, `CTR-01`–`CTR-03` |
| `DEC-02` | accepted | Do not add old-source fallback; accept only `memory-bank-template/` after FT-014. | Issue #63 requires the template to have the new root and no duplicate old payload. Issue #14 names support for both roots indefinitely as a non-goal. Existing released CLI remains available during the ordered handoff. | FPF abductive loop: candidates were strict new root and dual-path discovery. Parsimony and falsifiability favour the strict root: one exact Git path yields deterministic negative tests; fallback adds an unrequested branch and can mask wrong-repository state. | [brief.md](brief.md) `NS-02`; [design.md](design.md) `ALT-02`, `TRD-02`, `SD-02` |
| `DEC-03` | accepted | Use profile-derived effective scope and matching repository-local agent route for doctor, but the existing explicit `--scope-root memory-bank-template` for template lint. | Doctor already has auto/template/downstream profiles and marker detection, but currently normalizes scope before profile and hard-codes the downstream agent link. Lint has no profile but already exposes normalized `--scope-root`. The generated downstream agent block is a separate contract and remains unchanged. | FPF candidate comparison considered adding lint profile detection versus reusing its explicit scope connector. Parsimony and public-contract scope fit favour reuse. Strict distinction keeps template-local diagnostic routing separate from init/update-generated downstream guidance. | [design.md](design.md) Grounding, `SOL-03`, `SOL-04`, `SD-03`, `CTR-03` |
| `DEC-04` | accepted | Select `release-deployment` validation and defer the exact release tag to the protected workflow input. | Issue #14 requires a released CLI before issue #63 merges. The repository has an existing validation/release workflow and protected `release` environment. Neither issue names a version. | FPF evidence discipline places tests and exact-commit validation before the external effect. Boundary discipline forbids inventing an absent version requirement; the human-approved immutable tag becomes recorded evidence rather than a feature-authored fact. | [brief.md](brief.md) Validation Profile, `CON-03`, `CON-04`; [design.md](design.md) `SD-04`, `INV-05`, `RB-02` |

## Open Questions

`none`: the issue pair, current CLI contracts and release workflow provide enough facts for Plan Ready. The release version is an approval-time workflow input, not an unresolved product or design decision.
