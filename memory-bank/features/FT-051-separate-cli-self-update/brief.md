---
title: "FT-051: Separate CLI Self-Update"
doc_kind: feature
doc_function: canonical
purpose: "Canonical problem, scope, blockers, validation profile and verify contract for separating CLI self-update from Memory Bank synchronization."
derived_from:
  - ../../flows/feature.md
  - ../../engineering/validation-profiles.md
  - "https://github.com/dapi/memory-bank-cli/issues/51"
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_solution
  - implementation_sequence
---

# FT-051: Separate CLI Self-Update

## What

### Problem

Issue #51 identifies an ambiguous command responsibility: the existing `memory-bank-cli update` synchronizes Memory Bank content, while the requested public contract assigns `update` to updating the installed CLI itself and assigns Memory Bank synchronization to `memory-bank-cli pull`. Current user documentation, CLI help, E2E script, and command tests use `update` for synchronization.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Public command responsibility | `update` performs Memory Bank synchronization. | `pull` performs that existing synchronization behavior; `update` performs an installed-CLI update and reports its outcome. | Command-path, error-path, documentation, and relevant automated-test evidence. |

### Scope

- `REQ-01` Rename the current Memory Bank synchronization command from `update` to `pull` without changing its synchronization behavior.
- `REQ-02` Provide `memory-bank-cli update` as a CLI self-update command that updates the installed CLI version and clearly reports its outcome.
- `REQ-03` Update command help, README, examples, shell completion, and other user-facing documentation so `update` is not ambiguous as Memory Bank synchronization.
- `REQ-04` Provide migration guidance telling users to replace `update` with `pull` for Memory Bank synchronization.
- `REQ-05` Add or update automated tests for both command paths, including error handling.

### Non-Scope

- `NS-01` Change the existing Memory Bank synchronization semantics, ownership lock format, or downstream payload layout beyond moving its public command name to `pull`.
- `NS-03` Publish a release or change release artifacts as part of this documentation package.

### Constraints / Assumptions

- `ASM-01` Issue #51 is the sole product request and has no comments or attached implementation decision.
- `ASM-02` Current CLI help advertises `update` as “Safely update a template using its ownership lock”; `internal/cli` dispatches current synchronization behavior through that command.
- `ASM-03` Current README documents `go install ...@latest` and a manual versioned `go install ...@vX.Y.Z` upgrade; GoReleaser publishes platform-specific binaries and checksums.
- `CON-01` The requested self-update must not silently replace an executable through an unverified, incompatible, or unsupported installation channel.
- `CON-02` The command rename is a public CLI contract change, so help, documentation, examples, completion, automated command tests and error messages must agree.

## Blocking Decisions

`none`: the FPF review adopted the existing self-update delivery pattern from `dapi/code-converge` and `dapi/start-issue`, grounded in this repository's matching GoReleaser release assets. The accepted solution is owned by [design.md](design.md).

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature changes a public CLI contract and introduces a binary-update/security/operational boundary. | `design.md` |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- |
| `decision-log.md` | selected | The blocking decision needs a fact-bounded FPF record and human gate. | `decision-log.md` |
| `design.md` | selected | The feature changes CLI and update-delivery contracts. | `design.md` |
| `implementation-plan.md` | selected | Execution must refine the selected command and release-update contracts. | `implementation-plan.md` |
| Separate contract / diagram / use-case | omitted | One CLI/release boundary is sufficiently covered by compact design tables. | none |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `standard` | Current project policy confirms Go test/vet and contract-focused regressions for this code change; the selected updater reuses existing release assets and does not publish a release. | none |

## Verify

### Exit Criteria

- `EC-01` `memory-bank-cli pull` performs the existing Memory Bank synchronization behavior.
- `EC-02` `memory-bank-cli update` updates the installed CLI version and clearly reports the outcome according to the accepted delivery contract.
- `EC-03` User-facing documentation contains no ambiguous use of `update` for Memory Bank synchronization and includes the required migration guidance.
- `EC-04` Relevant automated tests, including both command error paths, pass.

### Traceability Matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-02`, `CON-02` | `EC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `ASM-01`, `ASM-03`, `CON-01` | `EC-02` | `CHK-02` | `EVID-02` |
| `REQ-03`–`REQ-04` | `ASM-02`–`ASM-03`, `CON-02` | `EC-03` | `CHK-03` | `EVID-03` |
| `REQ-05` | `CON-01`–`CON-02` | `EC-04` | `CHK-01`–`CHK-04` | `EVID-01`–`EVID-04` |

### Acceptance Scenarios

- `SC-01` A user invokes `pull` with a valid current synchronization setup and receives the existing synchronization outcome.
- `SC-02` A user invokes `update` from an installation covered by the accepted self-update contract and receives a clear successful-update or already-current outcome.
- `SC-03` A user invokes either command in a supported error condition and receives the error behavior defined by its accepted contract without conflating CLI update with Memory Bank synchronization.
- `SC-04` A user follows current documentation and migration guidance and uses `pull`, not `update`, for Memory Bank synchronization.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01`, `SC-03` | Run focused command and synchronization regressions after the rename. | `pull` retains the current synchronization behavior and its relevant error handling. | `artifacts/ft-051/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-02`–`SC-03` | Run deterministic self-update success, no-update and error-path tests defined by the accepted delivery contract. | `update` follows the accepted mechanism and reports the correct outcome. | `artifacts/ft-051/verify/chk-02/` |
| `CHK-03` | `EC-03`, `SC-04` | Audit README, CLI help, examples, completion and other user-facing surfaces. | No user-facing synchronization instruction calls the command `update`; migration guidance is present. | `artifacts/ft-051/verify/chk-03/` |
| `CHK-04` | `EC-04` | Run applicable Go test and vet suites. | Required automated suites pass. | `artifacts/ft-051/verify/chk-04/` |

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Pull command/synchronization regression output | Go test runner | `artifacts/ft-051/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Self-update contract regression output | Go test runner | `artifacts/ft-051/verify/chk-02/` | `CHK-02` |
| `EVID-03` | User-facing documentation and help audit | Implementer / reviewer | `artifacts/ft-051/verify/chk-03/` | `CHK-03` |
| `EVID-04` | Applicable Go test and vet output | Go toolchain | `artifacts/ft-051/verify/chk-04/` | `CHK-04` |
