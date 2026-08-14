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
| `DEC-01` | accepted | Use a versioned JSON resolution plan whose canonical whole-plan digest is covered by a detached attestation from an authorized human-controlled signing key; per-path inputs bind lock, source, local and upstream identities, and accepted actions/results are revalidated before one atomic apply. | Issue #54 requires machine-readable planning, reviewable versioned plans, stale/tamper rejection and no partial mutation. Current `pull --dry-run --json` is read-only; `.lock` stores ownership/digests/modes; current ownership transaction is atomic. | FPF B.5: propose a durable plan instead of terminal-only answers. Deduction: a digest stored in an agent-writable plan does not establish approval, so an independently verified human attestation must cover the canonical plan digest; content digests and source/lock identities remain validation inputs, and a transaction protects the payload-plus-lock consequence. Induction is delegated to `CHK-01`–`CHK-04`; no run-time success is claimed yet. | [design.md](design.md) `SOL-01`–`SOL-02`, `CTR-01`, `INV-01`–`INV-02` |
| `DEC-02` | accepted | Treat every two-sided `adapted` path as unresolved until a human records one of the issue-required actions; a mechanical merge records verified base bytes, local/upstream inputs, algorithm version, result bytes and mode, and is recomputed at apply. | Issue #54 names the required actions, prohibits silent adapted decisions and permits a mechanical merge only when its exact result is encoded in the reviewed plan. Current `--ask` already separates answer collection from mutation. | FPF strict boundary and evidence-chain reasoning separates proposal (agent/AI) from authorization (human) and deterministic execution (CLI). A result digest alone can describe arbitrary bytes; binding base content to its lock identity and recomputing the specified deterministic three-way merge proves the limited mechanical-merge claim. | [design.md](design.md) `SOL-03`, `CTR-02`, `INV-03` |
| `DEC-03` | accepted | Added the required Feature Flow routing links to issue #54 after publishing commit `54cb644`. | Package Rule 12 requires links; the published commit supplies immutable artifact URLs. | FPF evidence graph requires a resolvable carrier, not a guessed URL. The issue comment now links the brief, design and plan at the published carrier. | [issue #54 routing record](https://github.com/dapi/memory-bank-cli/issues/54#issuecomment-5288246481) |

## Open Questions

`none`: issue #54 deliberately leaves exact flag spelling and schema field names as implementation choices. They do not alter the accepted human boundary, identity validation, action set or atomicity contract.
