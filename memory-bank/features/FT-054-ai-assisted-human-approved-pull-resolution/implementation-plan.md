---
title: "FT-054: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Grounded execution plan for FT-054 without redefining canonical problem or solution facts."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_054_scope
  - ft_054_selected_design
  - ft_054_acceptance_criteria
  - ft_054_blocker_state
  - ft_054_validation_profile
---

# FT-054: Implementation Plan

## Goal

Deliver the opt-in, human-approved resolution-plan workflow specified by `brief.md` and `design.md`, with a plan/apply validation boundary before the existing atomic ownership transaction.

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
| `internal/ownership/lock.go`, `types.go` | Lock decoding is strict and lock entries hold ownership, digests and modes. | Bind plan identity checks and mechanical-merge base verification to the same validated state. |
| `internal/cli/cli_test.go`, `internal/ownership/*_test.go` | Existing tests cover JSON dry run, ask behavior, planning and transactions. | Add codec, stale/tamper, adapted-action/mode and all-or-nothing cases beside local patterns. |
| `README.md` | Documents conservative `pull --ask`. | Explain plan/apply, AI proposal limit and `keep-and-detach`. |

## Open Questions

`none`: `DEC-01` and `DEC-02` resolve the feature-local representation and authorization boundary. Exact flag spelling/field names are constrained by `SOL-01`/`CTR-01` and may be selected during implementation without changing the contract.

## Preconditions

| Precondition ID | Requirement | Evidence / action |
| --- | --- | --- |
| `PRE-01` | A clean working tree and hermetic fixture source/downstream repos are available. | Follow existing CLI ownership test setup. |
| `PRE-02` | No apply is attempted until strict plan validation, a valid authorized detached human approval attestation and all required human decisions succeed. | `INV-02`, `INV-03`; `CHK-03`–`CHK-04`. |

## Workstreams

| Workstream | Covers | Owner | Depends on |
| --- | --- | --- | --- |
| `WS-01` | `SOL-01`–`SOL-02`, `CTR-01`, `INV-01` | either | `PRE-01` |
| `WS-02` | `SOL-02`–`SOL-04`, `CTR-02`, `INV-02`–`INV-05`, `FM-*` | either | `WS-01` |
| `WS-03` | `REQ-06`, use-case/docs and complete evidence | either | `WS-02` |

## Design Realization Mapping

| Design refs | Realization target | Steps | Checks / evidence |
| --- | --- | --- | --- |
| `SOL-01`–`SOL-02`, `SD-01`, `SD-03`, `CTR-01`, `INV-01` | ownership plan model/canonical digest, strict JSON codec and detached-attestation verification; CLI planning flag/output | `STEP-01` | `CHK-01`, `CHK-04` / `EVID-01`, `EVID-04` |
| `SOL-03`–`SOL-04`, `SD-02`, `CTR-02`, `INV-02`–`INV-05`, `FM-01`–`FM-02` | apply validation, mechanical-merge derivation/recomputation and adapted/deterministic action resolution | `STEP-02` | `CHK-02`–`CHK-04` / `EVID-02`–`EVID-04` |
| `SOL-04`, `FM-03`, `RB-01` | reuse atomic transaction; help/README/use case; regression suites | `STEP-03`–`STEP-04` | `CHK-04`–`CHK-05` / `EVID-04`–`EVID-05` |

## Steps

| Step ID | Actor | Implements | Goal | Touchpoints | Verifies | Evidence IDs | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | either | `REQ-01`–`REQ-02`, `SOL-01`–`SOL-02`, `CTR-01` | Add versioned strict plan model, canonical digest and detached authorized-human-attestation verification with read-only planning output. | `internal/ownership`, `internal/cli`, focused tests | `CHK-01`, `CHK-04` | `EVID-01`, `EVID-04` | `PRE-01` | trusted reviewer-key configuration and signing operation | representation cannot bind all required review identities or distinguish a human attestation from an agent-written plan |
| `STEP-02` | either | `REQ-03`–`REQ-05`, `SOL-03`–`SOL-04`, `CTR-02`, `INV-*` | Add validated apply and action resolution, including verified base bytes, deterministic three-way merge recomputation, reviewed result/mode and deterministic keep-and-detach. | `internal/ownership`, CLI tests/E2E | `CHK-02`, `CHK-03` | `EVID-02`, `EVID-03` | `STEP-01`, `PRE-02` | human decision and detached attestation exist before apply; no runtime prompt required | an action or mechanical merge cannot be validated before staging |
| `STEP-03` | either | `REQ-03`, `REQ-07`, `SOL-04`, `FM-*` | Route valid results through existing atomic transaction and add stale/tamper/injected-failure E2E. | ownership transaction and regression tests | `CHK-04` | `EVID-04` | `STEP-02` | none | payload and lock cannot be proved unchanged after failure |
| `STEP-04` | either | `REQ-06`, `RB-01` | Update help, README and UC-002; run standard profile suites and retain evidence. | `README.md`, `memory-bank/use-cases/UC-002-update-template.md`, Go tests/vet | `CHK-05` | `EVID-05` | `STEP-03` | none | documentation implies AI approval or default pull changes |

## Checkpoints and Stop Conditions

| Checkpoint ID | After | Pass criterion | Evidence |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01` | Plan is read-only, strict, has a canonical digest and accepts only a detached attestation from an authorized human reviewer. | `EVID-01`, `EVID-04` |
| `CP-02` | `STEP-02` | Every adapted two-sided path requires an explicit valid action; every mechanical merge is reproducible from validated inputs and has the exact recorded output. | `EVID-02`, `EVID-03` |
| `CP-03` | `STEP-03` | Stale/tampered/failed apply preserves all payload and lock state. | `EVID-04` |
| `CP-04` | `STEP-04` | Documentation and standard suite evidence agree with canonical contract. | `EVID-05` |

| Stop ID | Trigger | Action |
| --- | --- | --- |
| `STOP-01` | A necessary change would let AI inference authorize a path or weaken identity validation. | Stop and return to `design.md`; require a human decision for any changed safety boundary. |
| `STOP-02` | A failure can leave only payload or only `.lock` mutated. | Do not expose apply; restore/use the existing transaction boundary and re-run `CHK-04`. |

## Test Strategy

Automated coverage: plan codec/canonical-digest/read-only CLI output (`CHK-01`); deterministic safe apply on identical pre-state fixtures plus stale replay rejection (`CHK-02`); all adapted actions, mechanical-merge recomputation, unresolved rejection and modes (`CHK-03`); stale/tamper/path/attestation and transaction interruption (`CHK-04`); help/docs plus applicable Go test/vet (`CHK-05`). These are hermetic local fixtures and are also the required CI evidence under the selected standard profile. Tests use an authorized fixture signing key, while production approval requires the corresponding human-controlled key; real review is a human operational act, not plan data an agent can author.
