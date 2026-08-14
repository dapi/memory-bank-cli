---
title: "FT-054: Blocked Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Blocked, grounded execution plan for FT-054; it is not executable until the canonical Plan Ready blocker is resolved."
derived_from:
  - brief.md
  - design.md
status: blocked
audience: humans_and_agents
must_not_define:
  - ft_054_scope
  - ft_054_selected_design
  - ft_054_acceptance_criteria
  - ft_054_blocker_state
  - ft_054_validation_profile
---

# FT-054: Blocked Implementation Plan

## Goal

After `brief.md` `BD-01` is resolved, deliver the opt-in, human-approved resolution-plan workflow specified by `brief.md` and `design.md`, with a plan/apply validation boundary before the existing atomic ownership transaction. This is a blocked planning artifact, not authorization to begin implementation or to call FT-054 Plan Ready.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `brief.md` | canonical problem and verify owner | `REQ-*`, `SC-*`, `CHK-*`, `EVID-*` | Update brief first |
| `design.md` | canonical solution owner | `SOL-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-01` | Update design first |
| `../../use-cases/UC-002-update-template.md` | stable use-case owner | conservative pull behavior and atomic outcome | Update use case if scenario changes |

## Current State / Reference Points

| Area | Current fact | Execution implication |
| --- | --- | --- |
| `internal/cli/cli.go` | `pull` dispatches to `runOwnership`; `--ask`, `--dry-run`, JSON output and help exist. | Extend parser/reporting without changing default command semantics. |
| `internal/ownership/update.go` | Planner produces `Report`/`Decision`; staged writes and lock update use the transaction boundary. | Build plan/apply around this model and reuse transaction rather than a second writer. |
| `internal/ownership/lock.go`, `types.go` | Lock decoding is strict and lock entries hold ownership, digests and modes. | Preserve the `.lock` wire format and its required last-present digest/mode when sidecar state is absent or unavailable; add an atomically written, lock-bound historical-base sidecar as the sole authority for those states and reject reserved-sidecar-path payload collisions. Legacy or missing sidecar state cannot use mechanical merge. |
| `internal/cli/cli_test.go`, `internal/ownership/*_test.go` | Existing tests cover JSON dry run, ask behavior, planning and transactions. | Add codec, stale/tamper, adapted-action/mode and all-or-nothing cases beside local patterns. |
| `README.md` | Documents conservative `pull --ask`. | Explain plan/apply, AI proposal limit and `keep-and-detach`. |

## Open Questions

`BD-01` blocks execution: a human must approve the authorization carrier and verification format, canonicalization procedure, trust policy/configuration, credential lifecycle and reviewer authorization workflow. Exact flag spelling/field names may be selected during implementation only after that authorization contract is approved.

## Preconditions

| Precondition ID | Requirement | Evidence / action |
| --- | --- | --- |
| `PRE-01` | A clean working tree and hermetic fixture source/downstream repos are available. | Follow existing CLI ownership test setup. |
| `PRE-02` | Implementation does not begin until `BD-01` is resolved. | `BD-01`. |
| `PRE-03` | Runtime apply proceeds only after strict plan validation, valid human authorization and all required human decisions succeed. | `INV-02`, `INV-03`; `CHK-03`–`CHK-04`. |

## Workstreams

| Workstream | Covers | Owner | Depends on |
| --- | --- | --- | --- |
| `WS-01` | `SOL-01`–`SOL-02`, `CTR-01`, `INV-01` | either | `PRE-01`, `BD-01` |
| `WS-02` | `SOL-02`–`SOL-05`, `CTR-02`, `INV-02`–`INV-08`, `FM-*` | either | `WS-01` |
| `WS-03` | `REQ-06`, use-case/docs and complete evidence | either | `WS-02` |

## Design Realization Mapping

| Design refs | Realization target | Steps | Checks / evidence |
| --- | --- | --- | --- |
| `SOL-01`–`SOL-02`, `SD-01`, `SD-03`, `CTR-01`, `INV-01` | ownership plan model/canonical digest, strict JSON codec and `BD-01`-selected human-authorization verification; CLI planning flag/output | `STEP-01` | `CHK-01`, `CHK-04` / `EVID-01`, `EVID-04` |
| `SOL-03`–`SOL-05`, `SD-02`, `CTR-02`, `INV-02`–`INV-08`, `FM-01`–`FM-02`, `FM-04` | bounded lock-bound historical-base sidecar and plan storage, regenerated review-context validation, mechanical-merge derivation/recomputation, exact affected-path coverage and adapted/deterministic action resolution | `STEP-02` | `CHK-02`–`CHK-04` / `EVID-02`–`EVID-04` |
| `SOL-04`, `FM-03`, `RB-01` | reuse atomic transaction; help/README/use case; regression suites | `STEP-03`–`STEP-04` | `CHK-04`–`CHK-05` / `EVID-04`–`EVID-05` |

## Steps

| Step ID | Actor | Implements | Goal | Touchpoints | Verifies | Evidence IDs | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | either | `REQ-01`–`REQ-02`, `SOL-01`–`SOL-02`, `CTR-01` | After `BD-01` approval, add a versioned strict plan model, bounded pre-decode carrier reader and strict streaming structural/metadata validation, canonical digest and selected human-authorization verification with read-only planning output. | `internal/ownership`, `internal/cli`, focused tests | `CHK-01`, `CHK-04` | `EVID-01`, `EVID-04` | `PRE-01`, `BD-01` | approved authorization workflow | representation cannot bind all required review identities or distinguish a human authorization from an agent-written plan |
| `STEP-02` | either | `REQ-03`–`REQ-05`, `SOL-03`–`SOL-05`, `CTR-02`, `INV-02`–`INV-08`, `FM-04` | Add validated apply and action resolution, including an explicit scoped affected-path selector; present/absent tombstone transitions; bounded atomic retention and digest verification in a lock-bound historical-base sidecar while preserving the strict `.lock` wire format and its required last-present digest/mode; reserve the sidecar path from payload planning and reject existing downstream/template collisions without implicit migration; schema-versioned sidecar, plan-carrier, structural/metadata, plan-field, aggregate-plan, merge-input and aggregate LCS-work quota enforcement; deterministic lexicographic aggregate snapshot allocation; an exclusive update guard; transaction-scoped recomputation of the full and scoped affected-path sets, the complete deterministic non-decision review-context projection, and final identity/action-coverage validation; `mbc-diff3-lines-v1` recomputation; reviewed result/mode; managed no-drift update; and deterministic keep-and-detach that preserves bytes/mode while removing its lock entry. Test downgrade to an older binary after sidecar creation. Mark mechanical merge unavailable for legacy, tombstone-side, missing/stale-sidecar or quota-unavailable entries without a verified present snapshot; never synthesize one from a digest. | `internal/ownership`, CLI tests/E2E | `CHK-02`, `CHK-03`, `CHK-04` | `EVID-02`, `EVID-03`, `EVID-04` | `STEP-01`, `PRE-02` | human decision and valid authorization exist before runtime apply; no runtime prompt required | an action, scope, review-context projection, affected-path set, resource bound, mechanical merge or final pre-commit state cannot be validated before staging |
| `STEP-03` | either | `REQ-03`, `REQ-07`, `SOL-04`, `FM-*` | Route valid results through the guarded atomic transaction and add stale/tamper/cooperating-writer/injected-failure E2E; record the advisory guard's non-cooperating-writer boundary. | ownership transaction and regression tests | `CHK-04` | `EVID-04` | `STEP-02` | none | payload, lock and sidecar cannot be proved unchanged after failure or a cooperating writer can interleave after final validation |
| `STEP-04` | either | `REQ-06`, `RB-01` | Update help, README and UC-002; run standard profile suites and retain evidence. | `README.md`, `memory-bank/use-cases/UC-002-update-template.md`, Go tests/vet | `CHK-05` | `EVID-05` | `STEP-03` | none | documentation implies AI approval or default pull changes |

## Checkpoints and Stop Conditions

| Checkpoint ID | After | Pass criterion | Evidence |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01` | Plan is read-only, strict, has a canonical digest and accepts only the `BD-01`-approved human authorization. | `EVID-01`, `EVID-04` |
| `CP-02` | `STEP-02` | The plan exactly matches the regenerated deterministic non-decision review context and covers the recomputed full or explicitly scoped affected-path set; every adapted two-sided path in scope requires an explicit valid action; present/absent transitions put authoritative tombstones in the sidecar while retaining strict legacy `.lock` fields; the reserved sidecar path cannot enter payload planning; every mechanical merge is reproducible from validated inputs and has the exact recorded output; strict old-lock decoding survives new sidecar creation; and every `SOL-05` carrier, structural, retention and merge resource bound has its specified safe outcome. | `EVID-02`, `EVID-03`, `EVID-04` |
| `CP-03` | `STEP-03` | Stale/tampered/failed or cooperating-writer apply preserves all payload, lock and sidecar state, with no validation-to-commit interleaving; the non-cooperating-writer boundary is documented. | `EVID-04` |
| `CP-04` | `STEP-04` | Documentation and standard suite evidence agree with canonical contract. | `EVID-05` |

| Stop ID | Trigger | Action |
| --- | --- | --- |
| `STOP-01` | A necessary change would let AI inference authorize a path or weaken identity validation. | Stop and return to `design.md`; require a human decision for any changed safety boundary. |
| `STOP-02` | A failure can leave only payload or only `.lock` mutated. | Do not expose apply; restore/use the existing transaction boundary and re-run `CHK-04`. |

## Test Strategy

Automated coverage: plan codec/canonical-digest/read-only CLI output (`CHK-01`); deterministic safe apply on identical pre-state fixtures, stale replay rejection after a state-changing apply, and idempotent no-op replay (`CHK-02`); recomputation of the entire deterministic non-decision review-context projection and full/explicit scoped affected-path sets, with altered ownership/state labels, proposed actions, allowed actions, reasons/context, omitted, duplicate and extra-action rejection; all present- and tombstone-based one-sided and two-sided adapted actions; managed no-drift update including upstream deletion; user-owned missing-template detach with content/mode/ownership/lock assertions; lock-bound historical-base sidecar and tombstone retention/verification with unchanged strict legacy lock fields; existing downstream/template reserved-sidecar-path collision rejection; legacy/missing/stale-sidecar rejection; downgrade to an older binary after sidecar creation; every `SOL-05` boundary (pre-decode carrier, structural/metadata, per-snapshot, deterministic aggregate snapshot allocation, per-field, aggregate decoded plan, merge-input lines and aggregate LCS work), including unavailable-base and no-mutation outcomes; and `mbc-diff3-lines-v1` recomputation with LCS tie-break, collision, line-ending, binary and mode-conflict cases (`CHK-03`); stale/tamper/path/misleading-context/human-authorization, cooperating-writer and transaction-interruption coverage (`CHK-04`); help/docs plus applicable Go test/vet (`CHK-05`). These are hermetic local fixtures and are also the required CI evidence under the selected standard profile. Tests use the `BD-01`-approved fixture authorization workflow, while production approval uses the approved human-controlled workflow; real review is a human operational act, not plan data an agent can author.
