---
title: "FT-005: Feature Review Report"
doc_kind: feature-support
doc_function: derived
purpose: "Records the bounded review-improve result and human gate for FT-005 without becoming a canonical owner."
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

# FT-005: Feature Review Report

## Cycle 1

### Review summary

Issue #5 had no feature package. This bootstrap package maps each explicit issue acceptance area to `REQ-01`–`REQ-06`, keeps selected CI topology and execution sequencing out of the brief, and does not create downstream artifacts before the Problem Ready gate. `README.md`, `brief.md`, the decision log and this report consistently identify the package as draft/planned.

### Findings

| Severity | Finding | Evidence | Required action |
| --- | --- | --- | --- |
| `critical` | The issue does not define how the stable pinned reference and the scheduled/manual canary candidate are selected or how the local clean template checkout required by the current CLI is acquired. A solution would otherwise invent a compatibility policy and make failure attribution ambiguous. | Issue #5 scope; [brief.md](brief.md) `ASM-02`, `DEC-01`; [decision-log.md](decision-log.md) `DEC-01`. | Human selects the lane-input policy. |
| `important` | The conditional phrase “including artifact integrity when assets are introduced” has no defined asset set, trigger, or integrity procedure, while the documented user path is Go installation and the current GitHub release also exposes binaries/checksums. | Issue #5; [brief.md](brief.md) `ASM-01`, `DEC-02`; [decision-log.md](decision-log.md) `DEC-02`. | Human defines the integrity boundary. |
| `important` | No documented validation profile/minimum evidence contract covers this CI-plus-release-install integration feature. | [validation profiles](../../engineering/validation-profiles.md); [brief.md](brief.md) `DEC-03`; [decision-log.md](decision-log.md) `DEC-03`. | Human approves the profile/evidence minimum. |

No `minor` findings were acted on.

### Open questions closed through FPF

None. FPF B.5/B.5.2 was used to state the evidence-backed prompt, retain materially different candidate policies and apply scope fit, consistency and probeability filters. The current documents do not privilege a candidate without inventing policy; the resulting provenance is recorded in [decision-log.md](decision-log.md).

### Changes made

- Created the FT-005 bootstrap package: `README.md`, canonical `brief.md`, `decision-log.md` and this report.
- Recorded the issue-derived requirements, non-scope, verification/evidence contract and blocking decisions.
- Did not create `design.md` or `implementation-plan.md`, because that would violate Feature Flow before Problem Ready.

### Human gate

**Yes — stop auto-improvement here.**

| Question | Available facts | Options | Risk of a wrong choice | Needed from a human |
| --- | --- | --- | --- | --- |
| What selects stable and canary CLI/template references, including the clean source checkout? | Stable must be pinned and blocking; canary must be scheduled/manual; CLI requires an explicit clean local template checkout/version/full SHA; the forward-only `v1.0.1` replacement for the poisoned `v1.0.0` module is published. | Pin stable to `v1.0.1` and vary canary via moving latest; use explicit workflow-dispatch inputs for both; use a version manifest/lock as the source of both. | False compatibility claims, non-hermetic stable CI, or unclassifiable failures. | Select the policy and the allowed source acquisition/credentials. |
| What does conditional release-asset integrity cover? | Users install through `go install`; `v1.0.1` additionally has binary release assets and `checksums.txt`; issue says integrity when assets are introduced. | Go-install only; download/check release binaries and checksums now; add a named future packaging trigger. | Missing required supply-chain coverage or adding a test surface outside issue intent. | Name assets, trigger, verifier and expected evidence. |
| Which validation profile/evidence minimum applies? | Current policy covers Go test/vet and FT-003 release checks; profiles document no general catalogue for this feature. | Approve `standard`; approve `release-deployment`; define a feature-specific integration minimum. | Under-testing a blocking CI/release-path change or imposing unapproved delivery obligations. | Select and state the required suites/CI evidence. |

## Final status

`stopped_by_human_gate` after 1 cycle. No automatic follow-up cycle was run, as required by the gate.

## Cycle 2

### Review summary

The FPF decisions were promoted from the former blockers to their canonical owners: the active brief owns the `standard` profile and canonical verify contract, the active design owns stable/canary, integrity and attribution semantics, and the active plan maps every applicable design reference to a step/check/evidence chain. `README.md` routes only to existing downstream documents.

### Findings

| Severity | Finding | Resolution |
| --- | --- | --- |
| `important` | The initial canary wording resolved refs but did not state that the installed CLI and cloned template use those resolved commits; integrity wording did not state whether all binary assets were covered. | `SOL-02` now installs/clones at resolved commits; `SOL-03` validates every binary asset. |

No `critical` or remaining `important` findings. No `minor` findings were acted on.

### Open questions closed through FPF

- `DEC-01`: selected explicit refs resolved to immutable commits, with fixed forward-only `v1.0.1` stable pair replacing the poisoned `v1.0.0` module.
- `DEC-02`: selected separate Go-install and conditional all-binary checksum evidence.
- `DEC-03`: selected `standard` rather than release-deployment or an invented profile.

The candidate sets, filters (parsimony, constraint consistency, explanatory reach and probeability) and evidence-backed facts are in [decision-log.md](decision-log.md).

### Changes made

- Activated `brief.md`; added canonical validation decision.
- Created active `design.md` and `implementation-plan.md`; updated package routes.
- Repaired the canary input-to-install binding and all-binary integrity coverage.

### Human gate

No. The next lawful evidence step is implementation, not further document repair.

## Final status (supersedes Cycle 1 status)

`done` after 2 cycles. The document package is Problem Ready, Solution Ready and Plan Ready; feature execution/evidence remains intentionally pending.
