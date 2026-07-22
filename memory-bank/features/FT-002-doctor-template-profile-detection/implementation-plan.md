---
title: "FT-002: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Grounded execution plan for marker-based template-profile detection."
derived_from:
  - brief.md
  - design.md
  - ../../engineering/testing-policy.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_002_scope
  - ft_002_selected_solution
  - ft_002_acceptance_criteria
  - ft_002_validation_profile
---

# FT-002: Implementation Plan

## Goal

Realize `brief.md` requirements through accepted design `SOL-01`/`SOL-02`, then publish a durable issue backlink before execution evidence is claimed.

## Grounding / Support References

| Document | Role | Facts reused |
| --- | --- | --- |
| `brief.md` | canonical problem and verify owner | `REQ-*`, `SC-*`, `CHK-*`, `EVID-*`, standard profile |
| `design.md` | canonical solution owner | `SOL-*`, `CTR-01`, `INV-*`, `FM-01`, `RB-01` |
| `../../engineering/testing-policy.md` | test baseline | Go test, vet, contract regression |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `internal/doctor/doctor.go` | `detectProfile` reads lock then obsolete `tools/go.mod` | implementation target | retain lock-first precedence and local read-only behavior |
| `internal/doctor/doctor_test.go` | auto-profile and read-only regression tests | test target | extend existing fixture-driven table tests |
| `internal/doctor/testdata/template` | template fixture currently contains `tools/go.mod` | must prove module independence | replace with root marker fixture |
| `internal/doctor/testdata/downstream` | downstream-with-lock fixture | required regression case | retain unchanged behavior |
| `internal/doctor/testdata/downstream-no-lock` | absent today | required lookalike case | add only the needed fixture |

## Test Strategy

| Test surface | Canonical refs | Planned automated coverage | Required local suites | Manual-only gap |
| --- | --- | --- | --- | --- |
| profile classifier | `REQ-01`–`REQ-04`, `SOL-01`, `CTR-01`, `INV-*` | template/no module, lock downstream, no-lock lookalike, explicit profiles | `go test -count=1 ./internal/doctor`, `go test -count=1 -race ./...`, `go vet ./...` | none |
| cross-repository marker/docs | `REQ-05`, `RB-01` | documentation and issue #52 review | repository/document review | source marker addition is executed by issue #52, not this CLI checkout |

## Open Questions / Ambiguities

none; `DEC-01` through `DEC-04` in `decision-log.md` resolve the prior gates.

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Go module at repository root; fixture files are regular files | `STEP-01`–`STEP-04` | Go test cannot load package or fixture |
| test | Go 1.21-compatible environment runs the listed commands | `CHK-01`–`CHK-03` | tests/vet fail or cannot execute |
| access | GitHub issue update occurs only after a durable remote branch/commit URL exists | `STEP-05` | backlink would be broken; stop and publish first |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps |
| --- | --- | --- | --- |
| `PRE-01` | `DEC-01`, `CTR-01` | marker contract accepted | `STEP-01`–`STEP-04` |
| `PRE-02` | `DEC-04` | durable published URL available before tracker write | `STEP-05` |

## Design Realization Mapping

| Canonical solution refs | Owner | Realization target | Steps | Checks | Evidence |
| --- | --- | --- | --- | --- | --- |
| `SOL-01`, `SD-01`, `CTR-01` | `design.md` | `internal/doctor/doctor.go` and template fixture | `STEP-01`, `STEP-02` | `CHK-01` | `EVID-01` |
| `SOL-02`, `INV-01`, `INV-02`, `FM-01` | `design.md` | detector tests and downstream fixtures | `STEP-02`, `STEP-03` | `CHK-02`, `CHK-03` | `EVID-02`, `EVID-03` |
| `TRD-01`, `RB-01` | `design.md` | docs and issue #52 dependency | `STEP-04`, `STEP-05` | `CHK-04` | `EVID-04` |

## Workstreams

| Workstream | Implements | Result | Dependencies |
| --- | --- | --- | --- |
| `WS-01` | `REQ-01`, `REQ-02`, `SOL-01` | detector uses exact root marker | `PRE-01` |
| `WS-02` | `REQ-03`, `REQ-04`, `SOL-02` | complete regression fixtures/tests | `WS-01` |
| `WS-03` | `REQ-05`, `RB-01` | contract docs and durable tracker edge | `WS-01`, `WS-02`, `PRE-02` |

## Steps

| Step ID | Implements | Goal | Touchpoints | Verifies | Evidence IDs | Blocked by |
| --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | `REQ-01`, `REQ-02`, `SOL-01`, `CTR-01` | introduce private marker constants and exact regular-file detector; remove Go-module read | `internal/doctor/doctor.go` | `CHK-01` | `EVID-01` | `PRE-01` |
| `STEP-02` | `REQ-01`, `REQ-03`, `SOL-01`, `INV-02` | replace template fixture `tools/go.mod` with root marker and assert auto template result | `internal/doctor/testdata/template`, `doctor_test.go` | `CHK-01` | `EVID-01` | `STEP-01` |
| `STEP-03` | `REQ-03`, `REQ-04`, `SOL-02`, `INV-01`, `FM-01` | add no-lock lookalike fixture and explicit-profile regressions | `internal/doctor/testdata`, `doctor_test.go` | `CHK-02`, `CHK-03` | `EVID-02`, `EVID-03` | `STEP-01` |
| `STEP-04` | `REQ-05`, `TRD-01`, `RB-01` | document detector contract and issue #52 handoff in CLI-owned docs | `memory-bank/engineering/architecture.md`, feature docs | `CHK-04` | `EVID-04` | `STEP-01` |
| `STEP-05` | `REQ-05`, `RB-01` | after branch publication, add issue #2 backlink to canonical brief and cross-link issue #52 marker owner | GitHub issues | `CHK-04` | `EVID-04` | `PRE-02` |

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`–`STEP-03`, `CHK-01`–`CHK-03` | marker classifier and all fixture branches pass Go checks | `EVID-01`, `EVID-02`, `EVID-03` |
| `CP-02` | `STEP-04`, `STEP-05`, `CHK-04` | durable documentation and tracker links exist | `EVID-04` |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `CTR-01`, `FM-01` | a fixture shows lock precedence or explicit profile regression | do not publish; restore accepted behavior and reopen design | existing released detector behavior |
| `STOP-02` | `DEC-04` | no durable URL for issue link | do not post a broken backlink | local documentation branch pending publication |
