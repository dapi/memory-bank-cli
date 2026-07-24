---
title: "FT-023: Feature Review Report"
doc_kind: feature-support
doc_function: derived
purpose: "Records bounded review–improve findings and a human gate for FT-023 without becoming a canonical owner."
derived_from:
  - brief.md
  - decision-log.md
  - ../../flows/feature.md
status: active
audience: humans_and_agents
must_not_define:
  - requirements
  - selected_solution
  - implementation_sequence
---

# FT-023: Feature Review Report

## Cycle 1

### Review summary

The bootstrap package accurately represents issue #23 as one delivery unit and keeps problem, solution and execution ownership separated. `README.md` routes only to existing documents; `brief.md` contains the required What, Verify, requirement-to-scenario/check/evidence traceability, Design Requirement Decision and Validation Profile Decision; no `design.md` or `implementation-plan.md` exists prematurely. Review found three material unresolved decisions that prevent Problem Ready and any safe selected design.

### Critical and important findings

| Severity | Finding | Conflicting / affected documents | Required action |
| --- | --- | --- | --- |
| `critical` | No evidence-backed rule identifies which downstream paths are upstream-publishable or how ambiguity is resolved. | Issue #23's required selection/exclusion outcome; `brief.md` `REQ-03`/`DEC-01`; `internal/ownership/classify.go` discovery facts | Human approves a precise selection policy and ambiguous-path behavior. |
| `critical` | No evidence-backed recovery/cleanup contract reconciles required local checkout mutation, remote branch push and GitHub PR creation with the issue's prohibition on partially applied results. | Issue #23 transaction outcome; `brief.md` `REQ-02`/`REQ-04`/`DEC-02` | Human approves failure and compensation semantics, including remote-branch and local-checkout handling. |
| `important` | No applicable canonical validation profile defines the minimum evidence/CI obligations for this Git/GitHub workflow. | `brief.md` Validation Profile Decision/`DEC-03`; `engineering/validation-profiles.md` | Human approves a profile and its minimum checks/evidence. |

No `minor` finding was changed.

### Open questions closed through FPF

None. FPF B.5 and A.10 were applied and demonstrated that the issue, current classifier and validation document do not entail any of the three material decisions. The evidence-backed non-decision is recorded in [decision-log.md](decision-log.md).

### Changes made

- Created the feature-flow bootstrap package: `README.md`, canonical `brief.md`, `decision-log.md` and this review report.
- Mapped all explicit issue outcomes to `REQ-01`–`REQ-05`, `SC-01`–`SC-05`, `CHK-01`–`CHK-05` and `EVID-01`–`EVID-05` without claiming an unapproved executable contract.
- Recorded the three blockers in the canonical `brief.md` and their FPF provenance in `decision-log.md`.

### Human gate

**Yes — stop auto-improvement here.**

| Question | Available facts | Options | Risk of a wrong choice | Needed from a human |
| --- | --- | --- | --- | --- |
| What exact policy selects publishable upstream paths? | Issue requires upstream-suitable changes only and excludes project-specific artifacts, lock/state and `.repo`; current classifier has `managed`, `adapted`, `user-owned`, `generated` and unknown→`user-owned`, but no upstream mapping. | (1) publish only a named allowlist; (2) map named ownership classes with explicit exceptions; (3) another reviewable policy that names inclusions, exclusions and ambiguity behavior. | Wrong selection can publish project data or omit canonical template changes. | Approve the policy, treatment of every ownership class/unknown path, and whether ambiguity always fails. |
| What is the failure/cleanup contract across local Git, remote Git and GitHub? | Issue requires branch→commit→push→PR and says failures must not leave partial state; no existing document defines compensation after any individual external effect. | (1) retain a successfully pushed branch and report it for manual recovery when PR creation fails; (2) attempt compensating cleanup under named preconditions; (3) another explicit, observable recovery policy. | Wrong recovery can delete useful work, leave unintended remote state, or falsely claim atomicity. | Approve allowed effects, stop points, rollback/cleanup actions and user-facing recovery diagnostics. |
| Which validation profile applies? | Issue requires success, dry-run and key failure tests; the current validation document does not define a general profile/CI matrix. | (1) approve a named existing/project profile and obligations; (2) define an FT-023 minimum (targeted tests, integration boundary, full suite, vet, lint, CI/approval); (3) provide another governed policy. | Under-validation can publish unsafe changes; an invented gate can block delivery without authority. | Name/approve the profile and exact required local, CI and external evidence. |

## Interim Gate Record

- **Status:** `stopped_by_human_gate`
- **Cycles completed:** 1
- **Critical/important findings closed:** none; FPF established that all three require a human decision rather than an invented solution.
- **Critical/important findings remaining:** `DEC-01` publish-selection policy, `DEC-02` cross-boundary failure/recovery protocol, `DEC-03` validation profile/evidence minimum.
- **Decision log:** [decision-log.md](decision-log.md)

## Cycle 2

### Review summary

After the user directed FPF-based selection, the package reaches Solution Ready: `brief.md` is active, `design.md` is active, requirements remain in the brief, and selected solution facts moved to `SD-01`–`SD-03`. The prior human gate is superseded. `implementation-plan.md` remains absent because this task authorizes documentation decisions, not execution; Feature Flow permits its creation only when implementation starts.

### Critical and important findings

| Severity | Finding | Conflicting documents | Resolution |
| --- | --- | --- | --- |
| `important` | The decision-log ownership note still described `design.md` as future, although it now owns accepted solution facts. | `decision-log.md` vs active `design.md` | Updated the decision-log dependency and ownership statement. |

No critical findings were found. Minor findings were not changed.

### Open questions closed through FPF

- `DEC-01` → `design.md` `SD-01`: only exactly-`managed` normalized paths may be published; all others are reported as exclusions, and normalization/classification failure stops.
- `DEC-02` → `design.md` `SD-02`: preflight plus fresh-branch compensating transaction, default-branch protection and residual-state diagnostics.
- `DEC-03` → `brief.md` Validation Profile / `design.md` `SD-03`: `standard` evidence contract with targeted tests, full Go suite, vet, navigation audit and approved live PR evidence.

FPF B.5 records each as an abductive choice with explicit consequences and required `CHK-*` evidence; A.7/A.10 distinguish the proposed protocol from future execution evidence.

### Changes made

- Created active canonical `design.md` with C1 context, contracts, invariants, failures, backout and risk-based verification.
- Reconciled `brief.md` to remove resolved `DEC-*`, select validation, and activate downstream routing without creating an execution plan prematurely.
- Updated `decision-log.md` to point to its active canonical solution owner.

### Human gate

No. The user explicitly delegated these choices to FPF; no further material ambiguity was found in the documentation scope.

## Cycle 3

### Review summary

Navigation validation found one important inconsistency: the active decision log linked to a deliberately absent implementation plan. The plan is deferred by the Artifact Routing Decision because implementation is outside the current task. The route was replaced with plain future-state text; no canonical contract changed.

### Critical and important findings

| Severity | Finding | Conflicting documents | Resolution |
| --- | --- | --- | --- |
| `important` | `decision-log.md` linked to absent `implementation-plan.md`. | `decision-log.md` `DEC-03` vs `brief.md` Artifact Routing Decision and filesystem | Removed the broken route and retained an explicit non-link future-plan reference. |

No critical findings were found. Minor findings were not changed.

### Open questions closed through FPF

None. This was a documentation-navigation correction, not a decision.

### Changes made

- Repaired the deferred-plan route in `decision-log.md`.

### Human gate

No.

## Final Report

- **Status:** `done`
- **Cycles completed:** 3
- **Critical findings closed:**
  - selected an auditable managed-only publish policy with fail-closed ambiguity handling;
  - selected a bounded compensating transaction and default-branch protection.
- **Important findings closed:**
  - selected the `standard` validation/evidence contract;
  - reconciled the decision-log ownership statement with the active design owner.
  - removed the broken link to the intentionally deferred execution plan.
- **Critical/important findings remaining:** none.
- **Decision log:** [decision-log.md](decision-log.md)
