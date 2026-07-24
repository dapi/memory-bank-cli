---
title: "FT-026: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Grounded execution plan for FT-026 without redefining its canonical problem or selected design."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_026_scope
  - ft_026_selected_design
  - ft_026_acceptance_criteria
  - ft_026_validation_profile
---

# FT-026: Implementation Plan

## Grounding / Current State

| Path | Current role | Reuse / implication |
| --- | --- | --- |
| `internal/ownership/source.go` | Defines roots; Git/filesystem discovery call cardinality selector. | Add target precedence once; retain target-absent legacy matrix. |
| `internal/ownership/source_test.go` | Pinned-root translation and clean-Git fixture coverage. | Extend with distinguishable target/local fixture and init/update assertions. |
| `internal/ownership/update_test.go` | Update behavior at ownership API boundary. | Add/update regression there if fixture organization requires it. |
| `internal/ownership/types.go` | Defines downstream payload and lock contract. | Assert downstream `memory-bank/` remains unchanged. |

## Test Strategy

| Surface | Canonical refs | Automated coverage | Required local / CI suite | Manual-only gap |
| --- | --- | --- | --- | --- |
| source selection | `REQ-01`–`REQ-02`, `SOL-01`–`SOL-02`, `CTR-01`, `INV-01` | Git/filesystem selector matrix | focused ownership tests | none |
| init/update translation | `REQ-03`, `SOL-03`, `INV-02` | clean pinned dual-root init/update regression | focused ownership tests | none |
| compatibility/safety | `REQ-04`, `SOL-02`, `FM-*` | existing legacy/missing/multiple tests + relevant suite | `go test ./...` | none |

## Open Questions / Ambiguities

`none`: `DEC-01` records the target-present versus target-absent rule from issue #26. The package does not infer new lock semantics or a legacy-retirement date.

## Environment Contract

| Area | Contract | Failure symptom |
| --- | --- | --- |
| tests | Go/Git available; tests create temporary local Git repositories. | Fixture cannot commit or inspect source tree. |
| source fixture | Target/local files have distinguishable content; local copy contains `.lock`. | Regression cannot prove provenance. |

## Preconditions

| ID | Canonical ref | Required state | Used by |
| --- | --- | --- | --- |
| `PRE-01` | `SOL-01`–`SOL-02` | Existing selection/legacy tests pass before change. | `STEP-01`–`STEP-03` |
| `PRE-02` | `SOL-03`, `CTR-01` | Fixture can create a clean committed source checkout. | `STEP-02` |

## Workstreams

| Step ID | Owner | Realization target | Work | Paths | Check / evidence | Depends on |
| --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `SOL-01`, `SOL-02`, `CTR-01`, `INV-01` | Refactor candidate selection: target presence returns target; target absence delegates to bounded legacy matrix for both discovery callers. | `internal/ownership/source.go` | `CHK-01`, `EVID-01` | `PRE-01` |
| `STEP-02` | agent | `SOL-03`, `INV-01`–`INV-02`, `FM-01` | Add selector and clean pinned init/update regressions for target plus locked local roots; assert downstream paths. | `internal/ownership/source_test.go`, `internal/ownership/update_test.go` if needed | `CHK-01`, `CHK-02`; `EVID-01`, `EVID-02` | `PRE-02`, `STEP-01` |
| `STEP-03` | agent | `SOL-02`, `FM-01`–`FM-02`, `RB-01` | Run focused and relevant full Go tests; reconcile evidence to brief Verify. | affected tests, `artifacts/ft-026/verify/` | `CHK-03`, `EVID-03` | `STEP-02` |

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `SOL-01`–`SOL-02` | Both discovery paths share target precedence and target-absent compatibility. | `EVID-01` |
| `CP-02` | `STEP-02`, `SOL-03`, `INV-02` | Init/update install only target bytes into downstream `memory-bank/`. | `EVID-02` |
| `CP-03` | `STEP-03`, `FM-*` | Focused and relevant full suites are green. | `EVID-03` |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `FM-01` | Git/filesystem discovery cannot share one documented matrix. | Stop and update canonical design. | Existing equal-priority selector. |
| `STOP-02` | `FM-02`, `INV-02` | Regression shows changed pinning/downstream namespace. | Stop; do not merge selector change. | Existing pinned validation/layout. |

## Done Conditions

- `CHK-01`–`CHK-03` have concrete `EVID-01`–`EVID-03` carriers.
- Applicable `SOL-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` map to steps/checks/evidence.
- Final acceptance follows `brief.md#verify`.
