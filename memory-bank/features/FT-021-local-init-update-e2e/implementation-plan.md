---
title: "FT-021: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for FT-021 without redefining its canonical problem or selected design."
derived_from:
  - brief.md
  - design.md
  - use-cases/README.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_021_scope
  - ft_021_selected_design
  - ft_021_acceptance_criteria
  - ft_021_validation_profile
---

# FT-021: Implementation Plan

## Grounding / Current State

| Path | Current role | Reuse / implication |
| --- | --- | --- |
| `internal/cli/cli.go` | Public command dispatch. | Run only the produced executable, not internal Go functions. |
| `internal/ownership/source.go` | Clean source/ref validation. | Local fixture must supply clean checkout, version and full SHA. |
| `internal/ownership/transaction_test.go` | Existing atomicity/path-type regression patterns. | Mirror safety assertions at the CLI boundary. |
| `scripts/downstream-smoke.sh` | Networked release/canary fixture. | Do not reuse for local lane; preserve its external purpose. |
| `.github/workflows/release.yml` | `validate` builds a snapshot; separate `release` needs `validate`. | Insert candidate E2E in `validate` after snapshot, so release cannot start first. |
| `.github/workflows/downstream-canary.yml` | Schedule/manual external lane. | Keep it non-PR and retain its evidence behavior. |

## Test Strategy

| Surface | Canonical refs | Automated coverage | Required local / CI suite | Manual-only gap |
| --- | --- | --- | --- | --- |
| local CLI/Git lifecycle | `REQ-01`–`REQ-03`, `SOL-01`–`SOL-02`, `CTR-01`–`CTR-02`, `INV-01`–`INV-02` | named shell scenarios and snapshots | local runner; dedicated CI job; `go test -count=1 -race ./...`; `go vet ./...` | none |
| merge policy | `REQ-04`, `SOL-03`, `CTR-03`, `FM-03` | workflow/job inspection | `CHK-02` | admin binding, `AG-01` |
| release/canary | `REQ-05`–`REQ-06`, `SOL-04`, `INV-03` | candidate binary run; trigger/dependency inspection | `CHK-03`–`CHK-04` | none |

## Open Questions / Ambiguities

`none`: FPF decisions `DEC-01`–`DEC-05` resolve the feature-local choices. `AG-01` is an external execution approval, not an unresolved semantics question.

## Environment Contract

| Area | Contract | Failure symptom |
| --- | --- | --- |
| setup | POSIX shell, Go, Git and temporary writable filesystem. | Fixture cannot create/tag/clone local Git repositories. |
| local E2E network | A separate setup step builds the binary; the runner receives it through `E2E_BINARY` and uses no network service. | Missing/non-executable `E2E_BINARY` or any non-local URL/tool dependency invalidates `CHK-01`. |
| CI governance | Repository administrator can configure a required status check on `main`. | `EC-02` remains incomplete. |

## Preconditions

| ID | Canonical ref | Required state | Used by |
| --- | --- | --- | --- |
| `PRE-01` | `SOL-01`–`SOL-02` | Current checkout builds and has Git/Go test environment. | `STEP-01`–`STEP-02` |
| `PRE-02` | `SOL-03`, `CTR-03` | Stable job name selected in workflow. | `STEP-03` |
| `PRE-03` | `SOL-04` | Snapshot artifact path is available in the `validate` job immediately after its build step. | `STEP-04` |

## Design Realization Mapping

| Refs | Realization target | Steps | Checks | Evidence |
| --- | --- | --- | --- | --- |
| `SOL-01`, `CTR-01`, `INV-01`, `FM-01` | local shell runner/fixtures | `STEP-01` | `CHK-01` | `EVID-01` |
| `SOL-02`, `CTR-02`, `INV-02`, `FM-02` | scenario assertions | `STEP-02` | `CHK-01` | `EVID-01` |
| `SOL-03`, `CTR-03`, `FM-03`, `RB-01` | dedicated workflow and merge policy | `STEP-03` | `CHK-02` | `EVID-02` |
| `SOL-04`, `INV-03`, `RB-01` | release/canary workflows | `STEP-04` | `CHK-03`, `CHK-04` | `EVID-03`, `EVID-04` |

## Workstreams and Steps

| ID | Implements | Goal | Touchpoints | Verifies | Blocked by |
| --- | --- | --- | --- | --- | --- |
| `STEP-01` | `REQ-01`–`REQ-02`, `SOL-01` | Build binary and construct independent local Git lifecycle/profile scenarios. | new E2E runner/testdata | `CHK-01` | `PRE-01` |
| `STEP-02` | `REQ-03`, `SOL-02` | Add grouped local-edit/deletion/collision/atomicity assertions. | same runner/testdata | `CHK-01` | `STEP-01` |
| `STEP-03` | `REQ-04`, `SOL-03` | Add PR/main job, retain evidence and bind stable check as required. | workflow + repository policy | `CHK-02` | `STEP-02`, `PRE-02`, `AG-01` |
| `STEP-04` | `REQ-05`–`REQ-06`, `SOL-04` | Run candidate E2E in `validate` after snapshot (therefore before dependent release); preserve non-blocking external canary. | release/canary workflows | `CHK-03`, `CHK-04` | `STEP-01`, `PRE-03` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | Stable local-E2E job is green and its exact check name is known. | `STEP-03` | Changing `main` merge policy is an external repository-governance action. | Repository administrator; protection/ruleset URL or export in `EVID-02`. |

## Parallelizable Work

- `PAR-01` After `STEP-01`, `STEP-02` scenario expansion and `STEP-04` release/canary wiring can proceed in parallel.
- `PAR-02` `STEP-03` waits for a stable job name from the completed local suite.

## Checkpoints and Stop Conditions

| ID | Condition / trigger | Action |
| --- | --- | --- |
| `CP-01` | Every local scenario passes with only local Git inputs. | Retain `EVID-01`; proceed to CI/release wiring. |
| `CP-02` | Candidate E2E runs before any publish action and canary remains non-blocking. | Retain `EVID-03`/`EVID-04`. |
| `STOP-01` | A scenario requires network or cannot determine a full local SHA. | Stop; update canonical design before adding a workaround. |
| `STOP-02` | Required-check binding needs a different policy or cannot be configured. | Stop `STEP-03`, retain evidence and seek a new `AG-01` decision. |
