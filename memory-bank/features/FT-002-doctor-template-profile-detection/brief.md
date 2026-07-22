---
title: "FT-002: Doctor Template-Profile Detection"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief for replacing `doctor --profile auto` template-source detection based on `tools/go.mod` with an explicit source-repository marker."
derived_from:
  - ../../flows/feature.md
  - ../../engineering/validation-profiles.md
  - ../../product/context.md
  - ../../domain/rules.md
source_refs:
  - "https://github.com/dapi/memory-bank-cli/issues/2"
  - "https://github.com/dapi/memory-bank/issues/51"
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_solution
  - implementation_sequence
---

# FT-002: Doctor Template-Profile Detection

## What

### Problem

`doctor --profile auto` currently classifies a repository as the template source only when its root contains `tools/go.mod` with module `github.com/dapi/memory-bank/tools`. Issue #2 states that this implementation detail disappears after CLI extraction, so the template source needs an explicit, stable identity that is outside the copied `memory-bank/` payload.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Auto-profile template detection | reads and matches `tools/go.mod` | reads an explicit marker in the template source repository | Go tests using a template fixture without `tools/go.mod` |
| `MET-02` | Downstream classification safety | current heuristic is unrelated to a marker namespace | a downstream repository is not classified as template merely because it contains a similarly named file | Go tests using downstream-with-lock and downstream-without-lock fixtures |

### Scope

- `REQ-01` Define an explicit, stable template-source marker, its owner, and its location in `dapi/memory-bank`; the marker must be outside the copied `memory-bank/` payload.
- `REQ-02` Make `mb-cli doctor --profile auto` classify the template source from that marker rather than from `tools/go.mod`.
- `REQ-03` Add and maintain fixtures covering: template source without `tools/go.mod`, downstream repository with `memory-bank/.lock`, and downstream repository without that lock but with a similarly named non-marker file.
- `REQ-04` Preserve explicit `--profile template` and `--profile downstream` behavior.
- `REQ-05` Document the detection rules and marker contract in the repositories that own them.

### Non-Scope

- `NS-01` Restore, retain, or otherwise depend on `tools/go.mod` in the template source repository.
- `NS-02` Change the semantics of explicit `--profile template` or `--profile downstream`.
- `NS-03` Change the copied `memory-bank/` payload or use its contents as the source-repository marker.
- `NS-04` Redesign downstream ownership-lock format or `doctor` findings unrelated to profile classification.

### Constraints / Assumptions

- `ASM-01` The current detector first classifies any repository with `memory-bank/.lock` as downstream, then classifies as template only when `tools/go.mod` contains `module github.com/dapi/memory-bank/tools`; all other repositories are downstream.
- `ASM-02` Existing fixtures are `internal/doctor/testdata/template` (with `tools/go.mod`) and `internal/doctor/testdata/downstream` (with `memory-bank/.lock`).
- `CON-01` Issue #2 requires the marker to live in `dapi/memory-bank`, outside the copied `memory-bank/` payload; this CLI checkout cannot establish the marker in that separate repository.
- `CON-02` The issue requires a downstream repository not be misclassified merely for a similarly named file, so the marker namespace and validation rule are part of the external contract.
- `DEC-01` Accepted marker contract: `dapi/memory-bank/.memory-bank-template` is a UTF-8 root file, outside `memory-bank/`, containing exactly one line `memory-bank-template-v1`, terminated by LF or CRLF. Auto detection recognizes the template only when this exact path and logical line are present; issue `dapi/memory-bank#52` owns adding and documenting the source marker.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature changes a CLI profile-selection contract and introduces a cross-repository file/configuration contract that needs an explicit validation rule and false-positive boundary. | [design.md](design.md) |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- | --- |
| Feature-local decision log | selected | Records FPF resolution or human gate for the cross-repository marker contract. | [decision-log.md](decision-log.md); canonical facts remain in this brief and later `design.md` |
| `design.md` | selected | Required by the CLI/configuration contract and now enabled by `DEC-01`. | [design.md](design.md) |
| `implementation-plan.md` | selected | Execution is in scope after problem and solution readiness. | [implementation-plan.md](implementation-plan.md) |
| Separate interaction contract / diagram / ADR | omitted for now | No existing evidence establishes a need beyond the future feature-local design; revisit after marker contract selection. | none |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `standard` | The feature is an ordinary CLI code, fixture, and documentation change: testing policy requires Go test/vet and contract-focused regression; there is no release, security-boundary, migration, or capacity trigger. | none |

## Verify

### Exit Criteria

- `EC-01` `doctor --profile auto` reports template profile for a template-source fixture that has the accepted marker and no `tools/go.mod`.
- `EC-02` Auto detection reports downstream profile for fixtures both with and without `memory-bank/.lock`; a similarly named non-marker file does not change that result.
- `EC-03` Explicit `--profile template` and `--profile downstream` retain their current classifications independently of marker presence.
- `EC-04` The marker contract and detection rule are documented by their respective owners, including the marker's location outside `memory-bank/`.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `CON-01`, `CON-02`, `DEC-01` | `EC-01`, `EC-04` | `CHK-01`, `CHK-04` | `EVID-01`, `EVID-04` |
| `REQ-02` | `ASM-01`, `DEC-01` | `EC-01` | `CHK-01`, `CHK-03` | `EVID-01`, `EVID-03` |
| `REQ-03` | `ASM-02`, `CON-02` | `EC-01`, `EC-02` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-04` | issue #2 | `EC-03` | `CHK-03` | `EVID-03` |
| `REQ-05` | `CON-01`, `DEC-01` | `EC-04` | `CHK-04` | `EVID-04` |

### Acceptance Scenarios

- `SC-01` A contributor runs `mb-cli doctor --profile auto` in the template-source fixture after the CLI source has been extracted; the fixture lacks `tools/go.mod`, contains the accepted marker, and is classified as template.
- `SC-02` A contributor runs auto profile in a downstream fixture with `memory-bank/.lock`; it is classified as downstream.
- `SC-03` A contributor runs auto profile in a downstream fixture without `memory-bank/.lock` but with a similarly named non-marker file; it remains downstream.
- `SC-04` A user supplies each explicit profile; the selected explicit profile is honored independently of auto detection.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Run the focused Go profile tests against the accepted template fixture. | template profile; no `tools/go.mod` dependency | `artifacts/ft-002/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-02`, `SC-03` | Run focused Go profile tests against both downstream fixtures. | both report downstream; lookalike is ignored | `artifacts/ft-002/verify/chk-02/` |
| `CHK-03` | `EC-01`, `EC-03`, `SC-04` | Run focused Go profile tests for auto and both explicit values. | auto follows marker; explicit values unchanged | `artifacts/ft-002/verify/chk-03/` |
| `CHK-04` | `EC-04` | Review owning documentation in CLI and template repositories. | exact marker contract, owner, and outside-payload location are documented | `artifacts/ft-002/verify/chk-04/` |

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | template-fixture focused-test output | implementer/CI | `artifacts/ft-002/verify/chk-01/` | `CHK-01` |
| `EVID-02` | downstream-fixture focused-test output | implementer/CI | `artifacts/ft-002/verify/chk-02/` | `CHK-02` |
| `EVID-03` | explicit-profile regression-test output | implementer/CI | `artifacts/ft-002/verify/chk-03/` | `CHK-03` |
| `EVID-04` | marker-contract documentation review | implementer/reviewer | `artifacts/ft-002/verify/chk-04/` | `CHK-04` |
