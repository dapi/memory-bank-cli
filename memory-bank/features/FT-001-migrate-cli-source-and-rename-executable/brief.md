---
title: "FT-001: Migrate CLI Source and Rename Executable"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для переноса Go CLI в standalone-репозиторий и смены единственного публичного имени исполняемого файла на `memory-bank-cli`."
derived_from:
  - ../../product/context.md
  - ../../domain/rules.md
source_refs:
  - "https://github.com/dapi/memory-bank-cli/issues/1"
  - "https://github.com/dapi/memory-bank/issues/51"
  - "https://github.com/dapi/memory-bank-cli/issues/2"
  - "https://github.com/dapi/memory-bank-cli/issues/3"
  - "https://github.com/dapi/memory-bank-cli/issues/4"
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_solution
  - implementation_sequence
---

# FT-001: Migrate CLI Source and Rename Executable

## What

### Problem

Go CLI находится в `dapi/memory-bank/tools/`, а новый репозиторий `dapi/memory-bank-cli` содержит только начальный README. Публичное имя текущей основной команды — `memory-bank`; также существует отдельный legacy compatibility binary. Issue #1 требует отдельный модуль и единственную публичную команду `memory-bank-cli`, без compatibility path.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Go module location and import path | `dapi/memory-bank/tools`, module `github.com/dapi/memory-bank/tools` | target repository has module `github.com/dapi/memory-bank-cli` | inspect `go.mod` and run Go checks |
| `MET-02` | Public executable identities | `memory-bank` plus `legacy compatibility executable` | `memory-bank-cli` only; neither old binary is built or released | source/test/docs search and release-build configuration inspection |
| `MET-03` | Supported command behaviour | source CLI provides `lint`, `doctor`, `init`, `update` | same commands, exit codes and versioned JSON contracts under `memory-bank-cli` | migrated regression suite and acceptance commands |

### Scope

- `REQ-01` Import the source rooted at `dapi/memory-bank/tools/` into this repository with relevant Git history, change the Go module path to `github.com/dapi/memory-bank-cli`, and move the primary entrypoint to `cmd/memory-bank-cli`.
- `REQ-02` Make `memory-bank-cli` the only user-facing executable identity: rename primary usage, diagnostics, examples and release-build configuration; remove the legacy compatibility command directory, its tests, documentation and every compatibility code path.
- `REQ-03` Preserve the semantics, exit codes and versioned JSON contracts of `lint`, `doctor`, `init` and `update`; retain their regression coverage, including the read-only `doctor` contract tracked by issue #4.
- `REQ-04` Configure the migrated repository so it builds only `memory-bank-cli`; hand off tagged release publication and install-documentation evidence to issue #3.

### Non-Scope

- `NS-01` Replace `doctor --profile auto` source-template detection. That contract and its marker belong to issue #2.
- `NS-02` Create CI, publish a release/tag, or write installation/upgrade documentation. These belong to issue #3; this feature only leaves a release configuration that cannot build old binaries.
- `NS-03` Change the semantics, exit codes or versioned JSON contracts of `lint`, `doctor`, `init` or `update`.
- `NS-04` Provide aliases, wrapper binaries, transition support or compatibility documentation for `memory-bank` or `legacy compatibility executable`.

### Constraints / Assumptions

- `ASM-01` The source snapshot is commit `0957f3c495f2c0518c8a81448694cf0e231d3209` on `dapi/memory-bank` `main`, observed on 2026-07-22. The snapshot contains `tools/go.mod`, the primary command, a separate legacy compatibility command, internal packages and `.goreleaser.yml`.
- `CON-01` Issue #1 requires useful history where practical; source history relevant to `tools/` includes the introduction of the CLI (`b164c03`) and its relocation into `tools/` (`0039b3e`).
- `CON-02` The target currently has no tags or releases. Therefore `go install ...@vX.Y.Z` cannot be evidenced by this feature before the release work in issue #3.
- `CON-03` References to the *Memory Bank payload* (for example a `memory-bank/` directory) are domain data, not old executable identities; only CLI identity strings and executable/release artifacts are renamed or removed.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The change alters a public CLI identity, Go module import path, history-import procedure and release-build surface while preserving multiple contracts. | [design.md](design.md) |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- | --- |
| Feature-local decision log | selected | User requires an auditable record of FPF resolutions and migration decisions. | [decision-log.md](decision-log.md); canonical facts remain in brief/design |
| Separate API/interaction contract | omitted | No new integration protocol is introduced; existing JSON contracts remain in the migrated code and tests. | none |
| C4 diagram | omitted | One local Go module is moved; no new runtime component, external system or topology is selected by this feature. | C4 N/A in design |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `release-deployment` | `REQ-04` changes build/release-artifact configuration. No release is executed here; issue #3 owns publication and its approvals. | none |

## Verify

### Exit Criteria

- `EC-01` `go test -count=1 -race ./...` and `go vet ./...` pass in the target repository.
- `EC-02` The target module and primary entrypoint are `github.com/dapi/memory-bank-cli` and `cmd/memory-bank-cli`.
- `EC-03` No migrated source, test, user-facing product documentation or release-build configuration retains the legacy compatibility executable; no `memory-bank` compatibility binary is configured. Historical terms remain permitted only in this feature package as governance evidence of the breaking migration.
- `EC-04` The migrated regression suite demonstrates preserved `lint`, `doctor`, `init` and `update` semantics, exit codes and versioned JSON contracts under `memory-bank-cli`.
- `EC-05` The history import and source snapshot are recorded in Git; release-tag installation is handed off to issue #3 rather than claimed as evidence here.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `CON-01` | `EC-01`, `EC-02`, `EC-05` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-02` | `CON-03` | `EC-02`, `EC-03` | `CHK-03`, `CHK-04` | `EVID-03` |
| `REQ-03` | issue #1, issue #4 | `EC-01`, `EC-04` | `CHK-01`, `CHK-05` | `EVID-01`, `EVID-04` |
| `REQ-04` | `CON-02` | `EC-03`, `EC-05` | `CHK-04`, `CHK-06` | `EVID-03`, `EVID-05` |

### Acceptance Scenarios

- `SC-01` A contributor runs the migrated test suite and static analysis in the standalone repository; all checks pass.
- `SC-02` A user invokes each supported command as `memory-bank-cli lint`, `memory-bank-cli doctor`, `memory-bank-cli init` and `memory-bank-cli update`; the migrated regression suite observes the preserved command contracts.
- `SC-03` A release maintainer builds the configured project and obtains only an `memory-bank-cli` executable; the release work in issue #3 can subsequently tag and install it.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-04`, `SC-01`, `SC-02` | `go test -count=1 -race ./...` | exit 0 | `artifacts/ft-001/verify/chk-01/` |
| `CHK-02` | `EC-01`, `EC-02` | `go vet ./...` and inspect `go.mod`, `cmd/memory-bank-cli` | exit 0; required module and entrypoint exist | `artifacts/ft-001/verify/chk-02/` |
| `CHK-03` | `EC-03` | Search the migrated source, tests, user-facing product docs and release config for the removed compatibility executable, excluding this feature package and generated evidence. | no matches in the checked product surface | `artifacts/ft-001/verify/chk-03/` |
| `CHK-04` | `EC-03`, `SC-03` | inspect release-build configuration and build outputs | only `memory-bank-cli` is configured/built; no old compatibility binary | `artifacts/ft-001/verify/chk-04/` |
| `CHK-05` | `EC-04`, `SC-02` | run migrated CLI contract tests, including JSON and exit-code assertions | tests pass under `memory-bank-cli` naming | `artifacts/ft-001/verify/chk-05/` |
| `CHK-06` | `EC-05` | inspect import commit(s), source SHA and handoff reference | source snapshot and history strategy are traceable; issue #3 dependency recorded | `artifacts/ft-001/verify/chk-06/` |

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | race-test output | implementer/CI | `artifacts/ft-001/verify/chk-01/` | `CHK-01` |
| `EVID-02` | vet output and module/entrypoint inspection | implementer/CI | `artifacts/ft-001/verify/chk-02/` | `CHK-02` |
| `EVID-03` | old-name search output and release-build config inspection | implementer/reviewer | `artifacts/ft-001/verify/chk-03/`, `chk-04/` | `CHK-03`, `CHK-04` |
| `EVID-04` | CLI contract-test output | implementer/CI | `artifacts/ft-001/verify/chk-05/` | `CHK-05` |
| `EVID-05` | import provenance and issue #3 handoff note | implementer/reviewer | `artifacts/ft-001/verify/chk-06/` | `CHK-06` |
