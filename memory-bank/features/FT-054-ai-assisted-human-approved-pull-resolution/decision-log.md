---
title: "FT-054: Decision Log"
doc_kind: feature-support
doc_function: derived
purpose: "Audit ledger for FT-054 FPF reasoning; canonical facts remain in brief.md and design.md."
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

# FT-054: Decision Log

## Ownership

`brief.md` owns scope and verification; `design.md` owns the selected plan/apply solution. This ledger records facts, FPF reasoning and consequences only.

## Decisions and Open Questions

| ID | Status | Record | Facts considered | FPF reasoning | Canonical owner |
| --- | --- | --- | --- | --- | --- |
| `DEC-01` | accepted | Use a versioned JSON resolution plan whose per-path review inputs bind lock, source, local and upstream identities, and whose accepted action/result is revalidated before one atomic apply. | Issue #54 requires machine-readable planning, reviewable versioned plans, stale/tamper rejection and no partial mutation. Current `pull --dry-run --json` is read-only; `.lock` stores ownership/digests/modes; current ownership transaction is atomic. | FPF B.5: propose a durable plan instead of terminal-only answers. Deduction: an approval is meaningful only for the exact reviewed bytes/identities, so content digests and source/lock identities must be validation inputs; a transaction then protects the payload-plus-lock consequence. Induction is delegated to `CHK-01`–`CHK-04`; no run-time success is claimed yet. This is the smallest representation that gives review provenance without an LLM or external service. | [design.md](design.md) `SOL-01`, `CTR-01`, `INV-01`–`INV-02` |
| `DEC-02` | accepted | Treat every two-sided `adapted` path as unresolved until a human records one of the issue-required actions; a reviewed merge is bytes, mode and digest in the plan, never an AI authorization. | Issue #54 names the required actions, prohibits silent adapted decisions and permits a mechanical merge only when its exact result is encoded in the reviewed plan. Current `--ask` already separates answer collection from mutation. | FPF strict boundary and evidence-chain reasoning separates proposal (agent/AI) from authorization (human) and from deterministic execution (CLI). Candidates that auto-select an action or accept an unbound merge violate the stated boundary; a plan-encoded result preserves a reviewable carrier for the later validation. | [design.md](design.md) `SOL-02`, `CTR-02`, `INV-03` |
| `DEC-03` | deferred | Add the required Feature Flow routing links to issue #54 after the feature branch containing these documents is published. | Package Rule 12 requires links. `git ls-remote` shows no remote branch named `feature/issue-54-feat-ai-assisted-human-approved-pull-res`, so no durable URL exists yet. | FPF evidence graph requires a resolvable carrier, not a guessed URL. Publishing is an external GitHub action outside this documentation change; the link is a traceability follow-up, not an unresolved solution fact. | [brief.md](brief.md) source issue; future issue #54 comment |

## Open Questions

`none`: issue #54 deliberately leaves exact flag spelling and schema field names as implementation choices. They do not alter the accepted human boundary, identity validation, action set or atomicity contract.

