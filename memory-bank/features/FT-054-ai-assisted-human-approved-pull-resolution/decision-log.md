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
| `DEC-01` | accepted | Use a versioned JSON resolution plan whose canonical whole-plan digest is covered by independently verifiable, tamper-resistant human authorization; per-path inputs bind lock, source, local and upstream identities, and accepted actions/results are revalidated before one atomic apply. | Issue #54 requires machine-readable planning, reviewable versioned plans, stale/tamper rejection and no partial mutation. Current `pull --dry-run --json` is read-only; `.lock` stores ownership/digests/modes; current ownership transaction is atomic. | FPF B.5: propose a durable plan instead of terminal-only answers. Deduction: a digest stored in an agent-writable plan does not establish approval, so independently verified human authorization must cover the canonical plan digest; content digests and source/lock identities remain validation inputs, and a transaction protects the payload-plus-lock consequence. `DEC-04` owns selection of the concrete mechanism. Induction is delegated to `CHK-01`–`CHK-04`; no run-time success is claimed yet. | [design.md](design.md) `SOL-01`–`SOL-02`, `CTR-01`, `INV-01`–`INV-02` |
| `DEC-02` | accepted | Treat every two-sided `adapted` path as unresolved until a human records one of the issue-required actions; a mechanical merge records verified base bytes from the lock-retained historical-base snapshot, local/upstream inputs, algorithm version, result bytes and mode, and is recomputed at apply. | Issue #54 names the required actions, prohibits silent adapted decisions and permits a mechanical merge only when its exact result is encoded in the reviewed plan. Current `--ask` already separates answer collection from mutation. Existing locks retain identities/digests and modes only, so a legacy entry without a retained snapshot cannot mechanically merge. | FPF strict boundary and evidence-chain reasoning separates proposal (agent/AI) from authorization (human) and deterministic execution (CLI). A result digest alone can describe arbitrary bytes; retaining base content atomically with the lock, binding it to its lock identity, and recomputing the specified deterministic three-way merge proves the limited mechanical-merge claim. | [design.md](design.md) `SOL-03`, `CTR-02`, `INV-03`, `INV-07` |
| `DEC-03` | pending | Refresh the required Feature Flow routing links to issue #54 after publishing this revised package snapshot. | Package Rule 12 requires links; commit `54cb644` predates later material changes to the canonical brief, design and plan and therefore is a superseded immutable carrier. | FPF evidence graph requires a resolvable carrier for the current canonical package, not a guessed or superseded URL. After this revision is published, update the issue routing record to the new immutable carrier and then mark this decision accepted. | [issue #54 routing record](https://github.com/dapi/memory-bank-cli/issues/54#issuecomment-5288246481) |
| `DEC-04` | pending_human_approval | Select the concrete human-authorization workflow: attestation/signature format, canonicalization procedure, authorized-reviewer trust-store configuration, key provisioning/revocation and reviewer signing operation. | `REQ-02` requires an independently verifiable, tamper-resistant human authorization boundary. A detached-attestation direction alone does not establish interoperable verification, trusted-key administration or revocation behavior. | FPF assurance reasoning: the authorization claim is only as strong as its defined trust root, message canonicalization and operational key controls. These choices affect the public contract and security boundary, so they cannot be inferred during implementation. | `brief.md` `BD-01`; `design.md` `SOL-02`, `SD-03` |

## Open Questions

`DEC-04` is open and blocks Plan Ready. Exact flag spelling and ordinary schema field names remain implementation choices, but the authorization workflow must be selected and human-approved before implementation.
