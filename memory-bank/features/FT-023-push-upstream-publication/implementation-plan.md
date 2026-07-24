---
title: "FT-023: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution and verification record for FT-023; refines canonical brief/design facts without redefining them."
derived_from:
  - brief.md
  - design.md
status: archived
audience: humans_and_agents
must_not_define:
  - ft_023_scope
  - ft_023_selected_design
  - ft_023_acceptance_criteria
  - ft_023_validation_profile
---

# FT-023: Implementation Plan

## Grounding / Current State

| Path | Role | Reused pattern |
| --- | --- | --- |
| `internal/cli/cli.go` | command dispatch and report output | `init`, `update`, `github` flag/output convention |
| `internal/ownership/classify.go` | existing ownership boundary | exact `managed` class used by `SD-01` |
| `internal/ownership/secure_path_*.go` | path-safety reference | nested checkout safety requirement |
| `internal/cli/cli_test.go` | CLI and hermetic Git fixture pattern | local test helper convention |

## Test Strategy

| Surface | Canonical refs | Automated coverage | Command | Manual-only gap |
| --- | --- | --- | --- | --- |
| selection and dry-run | `SD-01`, `CTR-01`, `INV-01`, `INV-03` | managed inclusion, adapted exclusion and state immutability | `go test ./internal/push` | none |
| preflight | `CTR-02`, `FM-01` | dirty, unsafe, conflicted and invalid-remote rejection | `go test ./internal/push` | none |
| CLI regression | `SOL-01`, `SOL-03` | command integration and root/push help | `go test ./internal/cli` | none |
| repository regression | `SD-03` | full suite, vet, navigation and approved live PR | `go test -count=1 -race ./...`; `go vet ./...`; `go run ./cmd/memory-bank-cli lint --repo-root .` | none; live carrier is dapi/memory-bank#78 |

## Open Questions / Ambiguities

`none`: implementation must escalate any change to `SD-01`–`SD-03` instead of deciding it locally.

## Preconditions

| Precondition ID | Canonical ref | Required state |
| --- | --- | --- |
| `PRE-01` | `design.md` `SD-01`–`SD-03` | active accepted design |
| `PRE-02` | `design.md` `RB-01` | explicit owner approval and real upstream/GitHub credentials before live validation |

## Design Realization Mapping

| Canonical refs | Realization target | Steps | Checks | Evidence |
| --- | --- | --- | --- | --- |
| `SOL-01`, `SD-01`, `CTR-01`, `INV-01`, `FM-03` | `internal/push/push.go` | `STEP-01` | `CHK-03` | `EVID-03` |
| `SOL-02`, `SD-02`, `CTR-02`, `INV-02`, `FM-01`–`FM-02`, `RB-01` | `internal/push/push.go` | `STEP-02` | `CHK-02`, `CHK-04` | `EVID-02`, `EVID-04` |
| `SOL-03`, `CTR-03`, `INV-03`–`INV-04` | `internal/cli/cli.go`, `README.md` | `STEP-03` | `CHK-01`, `CHK-05` | `EVID-01`, `EVID-05` |

## Steps and Checkpoints

| Step ID | Goal | Verifies | Evidence |
| --- | --- | --- | --- |
| `STEP-01` | Add managed-only planner and dry-run report. | `CHK-03`, `CHK-05` | `EVID-03`, `EVID-05` |
| `STEP-02` | Add preflight, upstream branch transaction and failure diagnostics. | `CHK-02`, `CHK-04` | `EVID-02`, `EVID-04` |
| `STEP-03` | Add CLI/help/README and run local validation. | `CHK-01`–`CHK-05` | `EVID-01`–`EVID-05` |
| `STEP-04` | Run approved real-upstream validation. | `CHK-01`, `CHK-04` | `EVID-01`, `EVID-04` |

| Checkpoint ID | Condition |
| --- | --- |
| `CP-01` | **PASS** — managed-only, dry-run and preflight failure tests pass. |
| `CP-02` | **PASS** — full race suite, vet, navigation audit and PR CI pass. |
| `CP-03` | **PASS** — live [dapi/memory-bank#78](https://github.com/dapi/memory-bank/pull/78) confirmed the canonical `template/memory-bank/` target and default-branch preservation; details are recorded in `brief.md`. |

## Approval Gate

`AG-01`: the repository owner's completion request authorized controlled live
validation. [dapi/memory-bank#78](https://github.com/dapi/memory-bank/pull/78)
was created against `main`, verified, closed without merge and cleaned up.

## Stop Conditions / Fallback

| Stop ID | Trigger | Immediate action | Safe fallback |
| --- | --- | --- | --- |
| `STOP-01` | path, state, remote or conflict preflight fails | stop before mutation and print remediation | upstream/default branch unchanged |
| `STOP-02` | commit, push or PR failure | restore local checkout, attempt bounded remote cleanup and report residual | default branch unchanged |

## Plan-local Evidence

- `EVID-06`: local and CI implementation verification is recorded in
  [completion PR #34](https://github.com/dapi/memory-bank-cli/pull/34).

## Archive Note

All checkpoints and the manual approval gate are closed. Canonical acceptance
evidence remains in `brief.md`; this execution plan is archived with
`delivery_status: done`.
