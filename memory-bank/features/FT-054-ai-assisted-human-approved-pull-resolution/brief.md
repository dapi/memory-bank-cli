---
title: "FT-054: AI-Assisted, Human-Approved Pull Resolution"
doc_kind: feature
doc_function: canonical
purpose: "Canonical problem, scope, validation profile and verify contract for reviewable pull resolution plans."
derived_from:
  - ../../flows/feature.md
  - ../../engineering/validation-profiles.md
  - ../../use-cases/UC-002-update-template.md
  - "https://github.com/dapi/memory-bank-cli/issues/54"
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - selected_solution
  - implementation_sequence
---

# FT-054: AI-Assisted, Human-Approved Pull Resolution

## What

### Problem

`pull` stops conservatively on ownership conflicts and `pull --ask` collects one-off terminal answers, but neither leaves an agent a durable, reviewable resolution plan. Projects with adapted Memory Bank documents therefore cannot prepare a proposed update without either an interactive terminal or an unsafe transfer of ownership choices to automation.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Pull conflict resolution | Conflicts stop `pull`; `--ask` resolves only interactively. | A non-mutating plan exposes every decision, and a reviewed complete plan applies atomically only when its recorded inputs still match. | Hermetic CLI/ownership E2E and plan-tamper/staleness regressions. |

### Scope

- `REQ-01` Provide a machine-readable, non-mutating `pull` plan containing path, ownership, base identity, local/upstream state, safe proposed action, human-decision requirement, and concise conflict context.
- `REQ-02` Define a versioned, reviewable resolution-plan file that records accepted actions and any reviewed non-overlapping merge result.
- `REQ-03` Apply a resolution plan only after validating its lock, source identity, local payload and template payload; apply all accepted actions and `.lock` atomically.
- `REQ-04` Require an explicit human-selected `keep-local`, `take-upstream`, or `apply-reviewed-merge` action for every two-sided `adapted` change; an unresolved path blocks apply.
- `REQ-05` Preserve deterministic safe actions, including managed no-drift update, user-owned missing-template `keep-and-detach`, one-sided adapted behavior, and a plan-encoded non-overlapping mechanical merge.
- `REQ-06` Keep default `pull` conservative and backward-compatible; document the agent/human boundary and `keep-and-detach` policy.
- `REQ-07` Add hermetic E2E coverage for the issue acceptance matrix, including executable modes and all-or-nothing failure.

### Non-Scope

- `NS-01` Automatically infer a semantic merge or resolve a two-sided adapted conflict.
- `NS-02` Delete a user-owned file merely because it disappeared upstream.
- `NS-03` Send document contents to an external AI provider, create a GitHub PR, or replace Git review.

### Constraints / Assumptions

- `ASM-01` Issue #54 is the sole product request and has no comments or attached implementation decision.
- `ASM-02` Current `pull --dry-run --json` already derives a non-mutating ownership report; `pull --ask` collects all terminal answers before mutation.
- `ASM-03` Current ownership planning validates an existing `.lock` and writes destination payload plus lock through an atomic transaction; lock entries record ownership, digests and modes.
- `CON-01` A plan must be bound to the reviewed source, lock and payload identities so stale or tampered input fails before mutation.
- `CON-02` The CLI must remain deterministic and must never call an LLM or treat an AI proposal as approval.
- `CON-03` The public `pull` contract and the stable update use case must remain conservative when no plan is applied.

## Blocking Decisions

`none`: the issue fixes the mandatory human boundary and current ownership/transaction facts support the feature-local plan contract. Exact flag spelling and JSON field names are intentionally implementation details, provided the selected contract and `REQ-*` remain satisfied.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature changes a public CLI/file-format contract and adds a trust boundary, state validation and atomic rollback semantics. | `design.md` |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- | --- |
| `decision-log.md` | selected | Plan representation and approval boundary require fact-bounded FPF provenance. | `decision-log.md` |
| `design.md` | selected | CLI, plan-file, ownership and atomicity contracts require explicit solution reasoning. | `design.md` |
| `implementation-plan.md` | selected | Execution crosses CLI, ownership, lock, transaction, documentation and E2E surfaces. | `implementation-plan.md` |
| Separate contract / sequence diagram | omitted | The one CLI-to-local-file contract is reviewable in compact design tables. | none |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `standard` | Current project policy calls for Go test/vet and contract-focused regressions; the change is local, has no release/publication action, and its higher integrity risk is covered by targeted hermetic E2E. | none |

## Verify

### Exit Criteria

- `EC-01` Planning is read-only and reports enough context for an agent or reviewer without deciding ambiguous adapted paths.
- `EC-02` A complete plan containing only deterministic actions applies repeatably and atomically.
- `EC-03` Every two-sided adapted conflict remains unapplied until an explicit allowed human action is recorded; each selected action gives its specified content and mode result.
- `EC-04` Stale, malformed or tampered plans fail before mutating any downstream file or `.lock`.
- `EC-05` Default `pull` behavior and documentation preserve the conservative agent/human boundary.

### Traceability Matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01`–`REQ-02` | `ASM-01`–`ASM-02`, `CON-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `ASM-03`, `CON-01` | `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |
| `REQ-04`–`REQ-05` | `ASM-02`–`ASM-03`, `CON-01`–`CON-02` | `EC-03`, `SC-03`–`SC-04` | `CHK-03` | `EVID-03` |
| `REQ-03`, `REQ-07` | `CON-01` | `EC-04`, `SC-05` | `CHK-04` | `EVID-04` |
| `REQ-06` | `CON-02`–`CON-03` | `EC-05`, `SC-06` | `CHK-05` | `EVID-05` |

### Acceptance Scenarios

- `SC-01` An agent reads a JSON pull plan for a changed template without filesystem or lock mutation and can identify each required human decision.
- `SC-02` A reviewer applies a complete deterministic plan against unchanged inputs and receives its specified file and lock updates as one committed result.
- `SC-03` A plan with an unresolved two-sided adapted path is rejected without mutation.
- `SC-04` After a human records each allowed action, apply produces the recorded local, upstream, or reviewed-merge content and executable mode.
- `SC-05` A changed lock, source ref, local payload, template payload, malformed plan, or injected write failure leaves all downstream files and `.lock` unchanged.
- `SC-06` A user following help and README understands that an AI may prepare a proposal but cannot approve it; ordinary `pull` stays conservative.

### Negative Coverage

- `NEG-01` Reject an unknown schema/action, a plan with unrecognized fields, or a path outside the payload boundary.
- `NEG-02` Reject a purported reviewed merge whose encoded result or digest does not match the applied bytes.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Run CLI and ownership plan serialization/read-only tests. | Planning has no mutation and exposes required identities/context. | `artifacts/ft-054/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-02` | Run hermetic deterministic plan/apply E2E. | Complete safe plan applies atomically and repeatably. | `artifacts/ft-054/verify/chk-02/` |
| `CHK-03` | `EC-03`, `SC-03`–`SC-04`, `NEG-02` | Run adapted conflict and mode-preservation E2E. | Unresolved conflicts fail; only explicit allowed actions yield recorded results. | `artifacts/ft-054/verify/chk-03/` |
| `CHK-04` | `EC-04`, `SC-05`, `NEG-01` | Run stale/tamper/path and injected-failure transaction regressions. | Validation fails before mutation; interrupted apply leaves payload and lock unchanged. | `artifacts/ft-054/verify/chk-04/` |
| `CHK-05` | `EC-05`, `SC-06` | Review help/README and run relevant Go test and vet suites. | Boundary and `keep-and-detach` policy agree; required suites pass. | `artifacts/ft-054/verify/chk-05/` |

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Read-only plan test output and sample JSON | Go test runner | `artifacts/ft-054/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Deterministic apply E2E output | Go test runner | `artifacts/ft-054/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Adapted-action and mode E2E output | Go test runner | `artifacts/ft-054/verify/chk-03/` | `CHK-03` |
| `EVID-04` | Staleness/tamper and transaction-failure output | Go test runner | `artifacts/ft-054/verify/chk-04/` | `CHK-04` |
| `EVID-05` | Documentation audit plus applicable Go test/vet output | Implementer / Go toolchain | `artifacts/ft-054/verify/chk-05/` | `CHK-05` |

