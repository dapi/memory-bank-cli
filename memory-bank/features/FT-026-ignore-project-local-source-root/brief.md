---
title: "FT-026: Ignore Project-Local Source Root"
doc_kind: feature
doc_function: canonical
purpose: "Canonical problem, scope, validation profile and verify contract for source selection when a checkout contains a target template payload and a locked project-local copy."
derived_from:
  - ../../flows/feature.md
  - ../../engineering/validation-profiles.md
  - "https://github.com/dapi/memory-bank-cli/issues/26"
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_solution
  - implementation_sequence
---

# FT-026: Ignore Project-Local Source Root

## What

### Problem

Issue #26 describes a clean source checkout with generic `template/memory-bank/` and a project-local adapted `memory-bank/` containing `memory-bank/.lock`. Current selection treats every recognized root equally and rejects that checkout as ambiguous, preventing `init` and `update` from reading the generic template.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Dual-root source selection | Current selector rejects multiple recognized roots. | Clean target-plus-locked-local checkout completes init/update from target. | Focused Go selection and init/update regressions. |

### Scope

- `REQ-01` When `template/memory-bank/` is present, select it rather than treating it as ambiguous with legacy roots.
- `REQ-02` A project-local `memory-bank/` with `memory-bank/.lock` must not participate in target-root source selection.
- `REQ-03` Init/update regression coverage must prove downstream remains `memory-bank/` and derives from target, not the local copy.
- `REQ-04` Without target root, preserve documented bounded legacy-root compatibility.

### Non-Scope

- `NS-01` Change downstream payload name, lock format, ownership semantics, or payload content.
- `NS-02` Remove or broaden bounded legacy-root compatibility.
- `NS-03` Modify the project-local adapted copy.

### Constraints / Assumptions

- `ASM-01` Issue #26 supplies required behavior; it has no additional comments.
- `ASM-02` Current `source.go` recognizes `memory-bank`, `memory-bank-template`, and `template/memory-bank`, then rejects any multiple-root set.
- `ASM-03` Current init/update translate selected source payloads to downstream `memory-bank/`.
- `CON-01` Clean checkout and pinned Git-object validation remain required; only candidate selection changes.
- `CON-02` Git-tree and filesystem discovery must agree, because source verification and planning use both.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | This changes a source-selection contract at Git/filesystem boundaries and needs compatibility/failure reasoning. | `design.md` |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- | --- |
| `decision-log.md` | selected | Candidate-selection decision needs auditable provenance. | `decision-log.md` |
| Separate contract/diagram/use-case | omitted | One local code boundary; compact design tables are sufficient. | none |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `standard` | Existing CLI source-selection/init/update behavior changes with focused regressions; no release or security-boundary change is in scope. | none |

## Verify

### Exit Criteria

- `EC-01` A clean pinned source checkout with target plus locked local copy succeeds for init and update.
- `EC-02` Both operations install target content under downstream `memory-bank/`, never local-copy content.
- `EC-03` Target-absent checkouts retain legacy selector behavior.

### Traceability Matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01`–`REQ-02` | `ASM-01`–`ASM-02`, `CON-01`–`CON-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `ASM-03`, `CON-01` | `EC-01`–`EC-02`, `SC-01`–`SC-02` | `CHK-02` | `EVID-02` |
| `REQ-04` | `ASM-01`–`ASM-02` | `EC-03`, `SC-03` | `CHK-01`, `CHK-03` | `EVID-01`, `EVID-03` |

### Acceptance Scenarios

- `SC-01` A clean source checkout with target plus locked project-local root is accepted and uses target payload.
- `SC-02` Init/update install target content into downstream paths and never install local-copy content.
- `SC-03` Without target, existing single legacy-root acceptance and neither/multiple-root rejection remain.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-03`, `SC-01`, `SC-03` | Run focused source-selection tests. | Target wins; target-absent matrix remains compatible. | `artifacts/ft-026/verify/chk-01/` |
| `CHK-02` | `EC-01`–`EC-02`, `SC-01`–`SC-02` | Run clean pinned dual-root init/update regression. | Both operations install only target content. | `artifacts/ft-026/verify/chk-02/` |
| `CHK-03` | `EC-01`–`EC-03` | Run relevant Go suite. | Focused and full relevant tests pass. | `artifacts/ft-026/verify/chk-03/` |

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Focused source-selection test output | Go test runner | `artifacts/ft-026/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Dual-root init/update regression output | Go test runner | `artifacts/ft-026/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Relevant Go suite output | Go test runner | `artifacts/ft-026/verify/chk-03/` | `CHK-03` |
