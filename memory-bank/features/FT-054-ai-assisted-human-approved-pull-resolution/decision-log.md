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
| `DEC-01` | accepted | Use a versioned JSON resolution plan whose canonical whole-plan digest is covered by independently verifiable, tamper-resistant human authorization; the CLI regenerates and compares every non-decision review-context field before it accepts the selected-action overlay and applies once atomically. | Issue #54 requires machine-readable planning, reviewable versioned plans, stale/tamper rejection and no partial mutation. Current `pull --dry-run --json` is read-only; `.lock` stores ownership/digests/modes; current ownership transaction is atomic. | FPF B.5: propose a durable plan instead of terminal-only answers. Deduction: a digest stored in an agent-writable plan does not establish approval, and matching content identities alone do not establish truthful review context. Independently verified human authorization must cover the canonical plan digest while trusted CLI regeneration binds ownership/state labels, proposed and allowed actions, reasons and other context to current inputs; the transaction protects the payload-plus-lock-and-sidecar consequence. `DEC-04` owns selection of the concrete mechanism. Induction is delegated to `CHK-01`–`CHK-04`; no run-time success is claimed yet. | [design.md](design.md) `SOL-01`–`SOL-02`, `CTR-01`, `INV-01`–`INV-02` |
| `DEC-02` | accepted | Treat every two-sided `adapted` path in the authorized affected-path set as unresolved until a human records one of the issue-required actions; a mechanical merge records verified base bytes from lock-bound historical-base sidecar state, local/upstream inputs, the `mbc-diff3-lines-v1` algorithm version, result bytes and mode, and is recomputed at apply. The sidecar alone defines absent or unavailable base state; `.lock` retains its required last-present digest/mode, and the sidecar path is reserved from payload planning. | Issue #54 names the required actions, prohibits silent adapted decisions and permits a mechanical merge only when its exact result is encoded in the reviewed plan. Current `--ask` already separates answer collection from mutation. Existing locks retain identities/digests and modes only, so a legacy or missing sidecar entry without retained base state cannot mechanically merge. | FPF strict boundary and evidence-chain reasoning separates proposal (agent/AI) from authorization (human) and deterministic execution (CLI). A result digest alone can describe arbitrary bytes; retaining base content atomically in a lock-bound sidecar, while leaving the existing strict `.lock` format readable by an older binary, and recomputing the normative algorithm proves the limited mechanical-merge claim. Reserving the sidecar path prevents internal state from overwriting or being planned as payload. | [design.md](design.md) `SOL-03`, `CTR-02`, `INV-03`, `INV-07` |
| `DEC-03` | pending | Refresh the required Feature Flow routing links to issue #54 after publishing this revised package snapshot. | Package Rule 12 requires links; commit `54cb644` predates later material changes to the canonical brief, design and plan and therefore is a superseded immutable carrier. | FPF evidence graph requires a resolvable carrier for the current canonical package, not a guessed or superseded URL. After this revision is published, update the issue routing record to the new immutable carrier and then mark this decision accepted. | [issue #54 routing record](https://github.com/dapi/memory-bank-cli/issues/54#issuecomment-5288246481) |
| `DEC-04` | pending_human_approval | Select the concrete human-authorization workflow: carrier and verification format, canonicalization procedure, authorized-reviewer trust policy/configuration, credential provisioning/revocation and reviewer authorization operation. | `REQ-02` requires an independently verifiable, tamper-resistant human authorization boundary. The required outcome alone does not establish interoperable verification, trusted-authority administration or revocation behavior. | FPF assurance reasoning: the authorization claim is only as strong as its defined trust root, message canonicalization and operational credential controls. These choices affect the public contract and security boundary, so they cannot be inferred during implementation. | `brief.md` `BD-01`; `design.md` `SOL-02`, `SD-03` |

## Open Questions

`DEC-04` is open and blocks Plan Ready. Exact flag spelling and ordinary schema field names remain implementation choices, but the authorization workflow must be selected and human-approved before implementation.
