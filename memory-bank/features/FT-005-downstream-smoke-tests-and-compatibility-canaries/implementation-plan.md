---
title: "FT-005: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for FT-005 downstream smoke and canary implementation; refines canonical facts without redefining them."
derived_from:
  - brief.md
  - design.md
  - ../../engineering/testing-policy.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_005_scope
  - ft_005_selected_design
  - ft_005_acceptance_criteria
  - ft_005_validation_profile
---

# FT-005: Implementation Plan

## Grounding / Current State

| Path | Role | Reuse |
| --- | --- | --- |
| `README.md` | documented user install path | exact `go install` command |
| `internal/cli/cli.go` | public init/update/lint/doctor contract | invoke public commands only |
| `internal/ownership/source.go` | source cleanliness/SHA validation | clone and resolve before calling CLI |
| `internal/doctor/testdata/downstream/` | existing downstream-shaped fixture | discovery only; do not mutate/reuse as CI repo |
| `.github/workflows/release.yml` | CI permission/Go setup pattern | `contents: read`, setup Go |

## Test Strategy

| Surface | Canonical refs | Automated coverage | Suites |
| --- | --- | --- | --- |
| lifecycle | `REQ-01`–`REQ-03`, `SOL-01`, `INV-02` | temp repo, init, adaptation/user-owned preservation, two updates, lint, doctor, no diff | fixture + stable CI |
| canary report | `REQ-04`–`REQ-05`, `SOL-02`, `SOL-04` | requested/resolved refs and terminal boundary | canary CI |
| integrity | `REQ-06`, `SOL-03`, `CTR-02` | conditional checksum match | fixture/CI |
| regression | `standard` profile | repository regression | `go test -count=1 -race ./...`; `go vet ./...` |

## Open Questions / Ambiguities

`none`: FPF decisions `DEC-01`–`DEC-03` are accepted in `decision-log.md`.

## Environment Contract

| Area | Contract | Failure symptom |
| --- | --- | --- |
| setup | Go, Git, POSIX shell and new temporary directories for source/downstream. | classified setup failure |
| network | release Go install/template clone; asset download only when applicable. | classified boundary failure |
| access | `contents: read`, no secrets/write token. | security review failure |

## Preconditions

| ID | Canonical ref | Required state |
| --- | --- | --- |
| `PRE-01` | `SOL-01`–`SOL-04` | design active and release available |
| `PRE-02` | `INV-03` | runner can access public Go/Git with read-only checkout |

## Design Realization Mapping

| Refs | Target | Steps | Checks | Evidence |
| --- | --- | --- | --- | --- |
| `SOL-01`, `CTR-01`, `INV-01`–`INV-02` | fixture/testdata | `STEP-01` | `CHK-01` | `EVID-01` |
| `SOL-02`, `SOL-04`, `CTR-03`, `FM-01`–`FM-02` | canary/report | `STEP-02` | `CHK-03` | `EVID-03` |
| `SOL-03`, `CTR-02`, `RB-01` | integrity phase | `STEP-02` | `CHK-04` | `EVID-04` |
| `INV-03` | stable workflow | `STEP-03` | `CHK-02` | `EVID-02` |

## Workstreams and Steps

| ID | Implements | Goal | Touchpoints | Verifies | Blocked by |
| --- | --- | --- | --- | --- | --- |
| `STEP-01` | `REQ-01`–`REQ-03`, `SOL-01` | reusable isolated lifecycle fixture | new fixture/testdata | `CHK-01` / `EVID-01` | `PRE-01` |
| `STEP-02` | `REQ-05`–`REQ-06`, `SOL-02`–`SOL-04` | ref resolution, phase report, integrity tests | fixture/tests | `CHK-03`, `CHK-04` / `EVID-03`, `EVID-04` | `STEP-01` |
| `STEP-03` | `REQ-04`, `INV-03` | blocking stable and scheduled/manual canary workflows; run standard suites | workflows/fixture | `CHK-02`–`CHK-04` / `EVID-02`–`EVID-04` | `STEP-01`, `STEP-02`, `PRE-02` |

## Checkpoints and Stop Conditions

| ID | Condition / trigger | Action |
| --- | --- | --- |
| `CP-01` | stable pair proves fixture lifecycle | retain `EVID-01` |
| `CP-02` | stable is read-only/blocking and canary reports resolved refs/boundary | retain `EVID-02`, `EVID-03` |
| `STOP-01` | ref unavailable/different or report has no deterministic boundary | stop affected run; retain evidence; do not substitute input |
| `STOP-02` | a write token, secret or persistent repo is needed | stop and reopen canonical design |
