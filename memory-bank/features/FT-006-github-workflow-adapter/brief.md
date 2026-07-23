---
title: "FT-006: Opt-in GitHub Workflow Adapter"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для opt-in GitHub adapter. Фиксирует problem space, исходный scope и blocking decisions без принятия solution или execution decisions."
derived_from:
  - ../../flows/feature.md
  - ../../engineering/architecture.md
  - "https://github.com/dapi/memory-bank-cli/issues/6"
status: active
delivery_status: done
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
  - adapter_public_contract
---

# FT-006: Opt-in GitHub Workflow Adapter

## What

### Problem

Issue #6 requests an opt-in GitHub adapter that `memory-bank-cli` installs and updates, while keeping GitHub outside the Memory Bank core dependency. The requested adapter must manage optional issue forms, a PR template, validation workflow/configuration and agent guidance; it must protect user-owned GitHub templates and report create, update, conflict and dry-run decisions.

### Outcome

The issue defines qualitative acceptance only. No baseline, numeric target or source-template version is stated; those must not be invented in this brief.

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Required adapter outcomes | not stated | All issue acceptance criteria demonstrated | Canonical scenarios and evidence after adapter contract is approved |

### Scope

- `REQ-01` Provide an explicit, opt-in adapter boundary for GitHub-specific issue forms, PR template, validation workflow/configuration and agent guidance; non-GitHub adoption remains unaffected.
- `REQ-02` Install and update the adapter through `memory-bank-cli` with ownership-aware, dry-run-capable and non-destructive handling of user-owned `.github/` templates.
- `REQ-03` Report create, update, conflict and dry-run decisions for adapter-managed paths.
- `REQ-04` Define adapter guidance that distinguishes Small Change from Feature documentation/evidence requirements and addresses delivery evidence, full closure and partial references.
- `REQ-05` Cover an empty `.github/` tree and existing custom templates with fixtures.

### Non-Scope

- `NS-01` Making GitHub a required Memory Bank core dependency or changing non-GitHub adoption.
- `NS-02` Selecting an adapter command, flags, source layout, managed-file list, GitHub artifact contents, or ownership algorithm without an approved adapter contract.
- `NS-03` Changing requirements outside issue #6, including generic Memory Bank owner documents or unrelated GitHub automation.

### Constraints / Assumptions

- `ASM-01` Issue #6 is the only authoritative product input for this package; it names artifact categories but not their exact filenames, contents or public CLI interface.
- `ASM-02` The existing CLI exposes `init` and `update` ownership flows with `--dry-run`; its ownership model names `managed`, `adapted`, `user-owned` and `generated` classes. These are discovery facts, not an approved reuse decision for the adapter.
- `CON-01` User-owned `.github/` templates must not be overwritten.
- `CON-02` GitHub state must remain separate from canonical Memory Bank owner documents.
- `DEC-01` Accepted in `design.md`: use `memory-bank-cli github init|update`, embedded assets and marker-managed ownership.
- `DEC-02` Accepted: use the existing `standard` profile vocabulary; the feature's minimum is unit/CLI coverage, full Go suite, vet, navigation audit and green PR CI.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature changes a CLI/configuration contract, manages GitHub workflow/template artifacts, requires ownership/conflict semantics and has rollout/backout implications. | `design.md` |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- | --- |
| `design.md` | selected | Required explicit solution boundary and ownership semantics. | `design.md` |
| `implementation-plan.md` | selected | Execution and test sequencing. | `implementation-plan.md` |
| feature-local contracts / fixtures | omitted | Inline contract and Go fixture-like tests remain compact. | `design.md` |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `standard` | CLI and local filesystem contract change; no remote GitHub API, release publication or deployment. Full regression, vet, navigation audit and PR CI are required. | none |

## Verify

### Acceptance Scenarios

- `SC-01` A repository without `.github/` runs `memory-bank-cli github init` and receives the five managed assets; non-GitHub commands are unchanged.
- `SC-02` A repository with an existing unmarked custom GitHub template runs init/update and the file is preserved with a reported decision.
- `SC-03` A clean managed asset is modified inside its marker block; update reports conflict and applies no mutations.
- `SC-04` `--dry-run` reports planned creates/updates/conflicts and does not write.

### Exit Criteria

- `EC-01` The optional adapter installs only via `memory-bank-cli github`, preserves user-owned templates and reports decisions.
- `EC-02` Forms/PR/guidance cover the issue-required flow, evidence and closure distinctions.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01`–`REQ-03` | `CON-01`–`CON-02`, `DEC-01` | `EC-01`, `SC-01`–`SC-04` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-04`–`REQ-05` | `ASM-01`, `DEC-01` | `EC-02`, `SC-01`–`SC-02` | `CHK-01`, `CHK-03` | `EVID-01`, `EVID-03` |

### Checks

| Check | Covers | Procedure | Expected result | Evidence |
| --- | --- | --- | --- | --- |
| `CHK-01` | `SC-01`–`SC-04` | `go test ./internal/githubadapter` | all ownership and fixture scenarios pass | `EVID-01` |
| `CHK-02` | `SC-01`, `SC-04` | `go test ./internal/cli` | command/JSON contract passes | `EVID-02` |
| `CHK-03` | all `REQ-*` | `go test ./... && go vet ./... && go run ./cmd/memory-bank-cli lint --repo-root .` plus PR CI | local and CI checks pass | `EVID-03` |

### Evidence

- `EVID-01` Targeted adapter test output.
- `EVID-02` Targeted CLI test output.
- `EVID-03` Full local check output and linked PR CI runs.
