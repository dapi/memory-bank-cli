---
title: "FT-005: Downstream Smoke Tests and Compatibility Canaries"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для downstream smoke tests и compatibility canaries. Фиксирует problem space, scope, selected validation profile и verify contract без выбора CI topology или execution sequence."
derived_from:
  - ../../flows/feature.md
  - ../../engineering/testing-policy.md
  - ../../engineering/validation-profiles.md
  - ../FT-003-establish-memory-bank-cli-releases/brief.md
  - "https://github.com/dapi/memory-bank-cli/issues/5"
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_solution
  - implementation_sequence
---

# FT-005: Downstream Smoke Tests and Compatibility Canaries

## What

### Problem

Issue #5 requires delivery-path evidence that the standalone CLI and Memory Bank template work together in an isolated downstream repository. Existing repository tests cover CLI, ownership, doctor and lint behavior, but they do not constitute the requested downstream fixture, blocking stable smoke gate or scheduled/manual compatibility canary.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Downstream acceptance coverage | Repository-local Go tests and doctor fixtures exist; issue-required release-path downstream smoke is absent. | The issue's stable smoke and canary outcomes are demonstrated. | Canonical scenarios, CI evidence and failure attribution evidence. |

### Scope

- `REQ-01` Create a hermetic downstream fixture that uses the supported released CLI installation path and installs the Memory Bank template through the CLI.
- `REQ-02` In that fixture, verify `memory-bank-cli init`, a minimal adaptation, repeatable `memory-bank-cli update`, `memory-bank-cli lint`, and `memory-bank-cli doctor`.
- `REQ-03` Verify that adapted and user-owned files survive update and that a repeat run yields no diff.
- `REQ-04` Add a pinned stable smoke lane as a blocking CI gate, using isolated temporary repositories and least-privilege permissions.
- `REQ-05` Add a scheduled/manual canary lane that reports whether a failure belongs to the CLI, template, packaging, or external tooling.
- `REQ-06` Exercise the same release installation path users receive, including integrity checking when release assets are introduced.

### Non-Scope

- `NS-01` Change CLI ownership semantics, template content, release publishing, or the external template-repository CI switch owned by [template issue #52](https://github.com/dapi/memory-bank/issues/52).
- `NS-02` Claim a compatibility guarantee, select a canary version/source policy, or define an artifact-integrity mechanism before it is explicitly decided.
- `NS-03` Add non-hermetic network or credential dependencies to the stable blocking lane.

### Constraints / Assumptions

- `ASM-01` The supported user installation path is `go install github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@vX.Y.Z`; the forward-only `v1.0.1` replacement for the poisoned `v1.0.0` module is published at commit `b0d8ca47cdb3315df4b755a704d01ffb139754e7` and contains binaries plus `checksums.txt`.
- `ASM-02` The current CLI `init` and `update` require a clean local template Git checkout, `--source`, `--template-version`, and a full-commit `--source-ref`; it does not fetch a template itself.
- `ASM-03` Current tests and fixtures cover repository-local ownership/doctor/lint behavior; no issue #5 downstream CI workflow or fixture is present.
- `CON-01` Stable smoke must be hermetic, isolated in temporary repositories, least-privilege, and blocking.
- `CON-02` The issue requires the canary to distinguish CLI, template, packaging and external-tooling failures.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature adds CI topology, release-install and template-source boundaries, a compatibility policy, failure attribution and conditional integrity semantics. | `design.md` |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- | --- |
| `decision-log.md` | selected | Blocking decisions require auditable FPF provenance. | `decision-log.md`; canonical facts remain in this brief. |
| `design.md` | selected | Required CI/fixture topology, integrity and attribution semantics. | `design.md` |
| `implementation-plan.md` | selected | Execution and test sequencing after selected design. | `implementation-plan.md` |
| Separate fixture/contract support docs | omitted | No evidence yet shows a separate review boundary beyond future design. | none |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `standard` | The feature changes CI and consumes, but does not publish, a released installation path. It needs targeted fixture/workflow coverage plus the established Go test/vet regression contract; release-deployment is owned by FT-003. | none |

## Verify

### Exit Criteria

- `EC-01` A blocking stable CI lane demonstrates the issue-required downstream fixture lifecycle in an isolated temporary repository with least-privilege permissions.
- `EC-02` A scheduled/manual canary produces evidence that classifies a failure as CLI, template, packaging, or external tooling.
- `EC-03` The selected user release-install path and any in-scope release-asset integrity behavior are exercised with traceable evidence.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01`–`REQ-03` | `ASM-02`, `ASM-03`, `CON-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `CON-01` | `EC-01`, `SC-02` | `CHK-02` | `EVID-02` |
| `REQ-05` | `CON-02` | `EC-02`, `SC-03` | `CHK-03` | `EVID-03` |
| `REQ-06` | `ASM-01` | `EC-03`, `SC-04` | `CHK-04` | `EVID-04` |

### Acceptance Scenarios

- `SC-01` An isolated downstream repository installs the released CLI, supplies a clean local template checkout to `init`, makes a minimal adaptation and user-owned addition, then runs `update` twice; the adaptation and user-owned file survive and the second run has no diff.
- `SC-02` The pinned stable lane executes `SC-01` plus `lint` and `doctor` as a blocking CI gate with temporary repository isolation and least-privilege permissions.
- `SC-03` A scheduled or manually started canary executes the selected compatibility candidate and emits evidence sufficient to classify any failure as CLI, template, packaging, or external tooling.
- `SC-04` The fixture installs through the supported user release path and, when an in-scope release asset is used, records the agreed integrity result.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Run the implemented hermetic fixture lifecycle. | init/adaptation/update/lint/doctor conditions pass; repeated update leaves no diff. | `artifacts/ft-005/verify/chk-01/` |
| `CHK-02` | `EC-01`, `SC-02` | Inspect and run the stable CI lane. | It is blocking, pinned, temporary-repository isolated and least-privilege. | `artifacts/ft-005/verify/chk-02/` |
| `CHK-03` | `EC-02`, `SC-03` | Dispatch or schedule the canary and inspect its report. | Its selected inputs and failure attribution are unambiguous. | `artifacts/ft-005/verify/chk-03/` |
| `CHK-04` | `EC-03`, `SC-04` | Execute the selected release-install path and any agreed integrity procedure. | The user path works and integrity result is recorded when applicable. | `artifacts/ft-005/verify/chk-04/` |

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Fixture lifecycle log and diff result | stable fixture | `artifacts/ft-005/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Stable CI definition and run URL/log | CI | `artifacts/ft-005/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Canary inputs and classification report | CI | `artifacts/ft-005/verify/chk-03/` | `CHK-03` |
| `EVID-04` | Release-install and conditional integrity log | CI | `artifacts/ft-005/verify/chk-04/` | `CHK-04` |
