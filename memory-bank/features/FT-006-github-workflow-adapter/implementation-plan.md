---
title: "FT-006: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution record for FT-006 implementation and verification; it refines canonical brief/design facts without redefining them."
derived_from:
  - brief.md
  - design.md
status: archived
audience: humans_and_agents
must_not_define:
  - ft_006_scope
  - ft_006_selected_design
  - ft_006_acceptance_criteria
  - ft_006_validation_profile
---

# FT-006: Implementation Plan

## Grounding / Current State

| Path | Role | Reused pattern |
| --- | --- | --- |
| `internal/cli/cli.go` | public command dispatch and JSON/text output | existing `init`, `update`, `doctor` command style |
| `internal/ownership/` | existing safety vocabulary and dry-run report behavior | ownership-aware decisions, without changing generic lock contract |
| `internal/doctor/doctor.go` | downstream CI guidance | pinned CLI recommendation |

## Test Strategy

| Surface | Canonical refs | Automated coverage | Command |
| --- | --- | --- | --- |
| adapter planner | `REQ-01`–`REQ-05`, `SOL-02`–`SOL-03`, `CTR-01`–`CTR-04`, `FM-01`–`FM-02` | create/idempotence, dry-run, custom template, drift conflict, symlink | `go test ./internal/githubadapter` |
| CLI boundary | `REQ-01`, `REQ-03`, `SOL-01` | JSON dry-run output and no mutation | `go test ./internal/cli` |
| repository regression | all applicable refs | full Go suite, vet, documentation navigation | `go test ./...`; `go vet ./...`; `go run ./cmd/memory-bank-cli lint --repo-root .` |

## Open Questions / Ambiguities

`none`: issue #40 supplies the required forms/PR/closeout/ownership semantics; `design.md` records the compatible feature-local interface choice.

## Preconditions

| ID | Canonical ref | Required state |
| --- | --- | --- |
| `PRE-01` | `SOL-01`–`SOL-03` | design active |

## Design Realization Mapping

| Canonical refs | Realization target | Steps | Checks | Evidence |
| --- | --- | --- | --- | --- |
| `SOL-01`, `INV-01` | `internal/cli/cli.go` | `STEP-01` | `CHK-02` | `EVID-02` |
| `SOL-02`, `CTR-01`–`CTR-03`, `INV-02` | `internal/githubadapter/adapter.go` | `STEP-02` | `CHK-01` | `EVID-01` |
| `SOL-03`, `CTR-04`, `FM-01`–`FM-02`, `RB-01` | `internal/githubadapter/adapter.go` and tests | `STEP-02` | `CHK-01` | `EVID-01` |

## Steps and Checkpoints

| Step | Goal | Verifies | Evidence |
| --- | --- | --- | --- |
| `STEP-01` | Add opt-in CLI command and report integration. | `CHK-02` | `EVID-02` |
| `STEP-02` | Add marker-owned asset planner and fixture-like unit coverage. | `CHK-01` | `EVID-01` |
| `STEP-03` | Run complete local validation and publish PR. | `CHK-03` | `EVID-03` |

| Checkpoint | Condition |
| --- | --- |
| `CP-01` | Targeted adapter/CLI tests pass. |
| `CP-02` | Full local suites and navigation audit pass; PR CI is green. |

## Evidence

- `EVID-01`: `go test ./internal/githubadapter` result.
- `EVID-02`: `go test ./internal/cli` result.
- `EVID-03`: full-suite, vet, lint and PR CI results.
