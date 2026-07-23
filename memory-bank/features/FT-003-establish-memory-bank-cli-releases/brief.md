---
title: "FT-003: Establish memory-bank-cli Releases"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для CI, первого stable Go release и install/upgrade документации standalone `memory-bank-cli`."
derived_from:
  - ../../prd/PRD-001-standalone-memory-bank-cli.md
  - ../../ops/release.md
source_refs:
  - "https://github.com/dapi/memory-bank-cli/issues/3"
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - selected_solution
  - implementation_sequence
---

# FT-003: Establish memory-bank-cli Releases

## What

### Problem

The standalone module has release automation and a prior public module
version, but the current `memory-bank-cli` executable identity has not been
published. The rename requires a new major release while preserving
`memory-bank-cli` as the only distributed executable identity.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Automated release validation | validation workflow exists | the workflow runs tests, vet and a release build before tag and GitHub Release publication | inspect workflow and its successful run |
| `MET-02` | Stable Go module availability | no released version contains the current entrypoint | tagged `v1.0.0` is installable through the documented `go install` command | execute install and version/help smoke check |
| `MET-03` | Public release identities | current identity is configured but unpublished | release assets and documentation expose only `memory-bank-cli` | inspect release, assets and docs |

### Scope

- `REQ-01` Maintain repository automation that validates Go tests, static analysis and the configured release build before publication of the release tag or GitHub Release.
- `REQ-02` Define and execute the first release under the current executable identity as tag `v1.0.0`, publishing `memory-bank-cli` through the existing GitHub/GoReleaser release surface only after the required validation succeeds.
- `REQ-03` Provide installation and upgrade documentation using `go install github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@vX.Y.Z` and document the breaking executable-identity change in the release notes.
- `REQ-04` Ensure the release workflow, release assets, documentation and release notes contain no compatibility artifact or supported installation path for `memory-bank` or `memory-bank-lint`.

### Non-Scope

- `NS-01` Change the semantics, exit codes or versioned JSON contracts of `lint`, `doctor`, `init` or `update`.
- `NS-02` Restore an alias, wrapper binary or transition path for `memory-bank` or `memory-bank-lint`.
- `NS-03` Redesign `doctor --profile auto` detection, which belongs to issue #2.
- `NS-04` Change GoReleaser distribution destinations beyond what the existing configuration already declares; a missing required credential is a stop condition, not permission to redesign distribution.

### Constraints / Assumptions

- `ASM-01` The current executable rename requires a tagged, Go-installable `v1.0.0`; existing public versions do not contain `cmd/memory-bank-cli`.
- `ASM-02` `.goreleaser.yml` is the existing release-build configuration: it has one `memory-bank-cli` build, GitHub release settings and a Homebrew cask that requires `HOMEBREW_TAP_GITHUB_TOKEN`.
- `CON-01` Creating a public tag, using GitHub credentials and publishing external assets are irreversible external effects. A pushed semantic tag makes that Go module version publicly retrievable by `go install` and may be cached by Go proxies; it must not be repointed. The protected release workflow validates the exact `main` commit before its approval-gated job creates the tag and GitHub Release.
- `CON-02` The release must be independently installable with the exact Go command from issue #3; local/snapshot builds cannot prove that criterion.
- `CON-03` `memory-bank/` remains payload terminology; the prohibition in `REQ-04` applies to executable identities, release artifacts and installation paths.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature adds a CI/release topology, external publication boundary, semantic-version contract and rollback/approval behaviour. | [design.md](design.md) |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- | --- |
| Feature-local decision log | selected | Decisions and human approval boundary require auditable FPF rationale. | [decision-log.md](decision-log.md); canonical facts remain in brief/design |
| Separate release runbook | omitted | Current procedure is compact and has one feature consumer; rollout/backout are owned in `design.md`. | none |
| C4 diagram | omitted | The design contains a compact C2 topology table; a standalone diagram would not improve the review boundary. | `design.md` `C4-01` |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `release-deployment` | The feature creates CI/release automation and performs an external tagged publication. | none |

## Verify

### Exit Criteria

- `EC-01` Automation runs Go tests, `go vet ./...`, GoReleaser configuration validation and a release build on the exact commit before it is tagged or published as a GitHub Release.
- `EC-02` GitHub contains the stable `v1.0.0` release and its assets contain only `memory-bank-cli` executable names.
- `EC-03` `go install github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@v1.0.0` succeeds from a clean Go module cache and the installed command reports `memory-bank-cli v1.0.0`.
- `EC-04` Repository documentation and `v1.0.0` release notes state the intentional breaking rename and removal, without presenting a compatibility installation path.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-02`, `CON-01` | `EC-01`, `SC-01` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-02` | `ASM-01`, `CON-01`, `CON-02` | `EC-02`, `EC-03`, `SC-02` | `CHK-03`, `CHK-04`, `CHK-06` | `EVID-03`, `EVID-04`, `EVID-06` |
| `REQ-03` | `ASM-01`, `CON-02` | `EC-03`, `EC-04`, `SC-02` | `CHK-04`, `CHK-05` | `EVID-04`, `EVID-05` |
| `REQ-04` | `ASM-02`, `CON-03` | `EC-02`, `EC-04`, `SC-03` | `CHK-03`, `CHK-05` | `EVID-03`, `EVID-05` |

### Acceptance Scenarios

- `SC-01` A maintainer starts the `v1.0.0` release from `main`; automation completes tests, vet and release-build validation on that exact commit before the approval-gated job creates its tag or GitHub Release.
- `SC-02` After approved publication of `v1.0.0`, a user installs the module with the issue-specified Go command and invokes `memory-bank-cli`.
- `SC-03` A user reads the repository/release documentation and can identify the breaking rename without being offered a removed executable as an installable compatibility path.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Inspect the validation and release workflow definitions; run the validation workflow or its equivalent commands. | tests, vet, release-config validation and snapshot release build succeed before publish step. | `artifacts/ft-003/verify/chk-01/` |
| `CHK-02` | `EC-01`, `SC-01` | Inspect the release workflow job dependency and run record. | approval-gated tag/release job cannot run until validation succeeds on its exact workflow commit. | `artifacts/ft-003/verify/chk-02/` |
| `CHK-03` | `EC-02`, `REQ-04`, `SC-03` | Inspect tag `v1.0.0`, GitHub release and asset names. | release exists; assets have only `memory-bank-cli` executable identity. | `artifacts/ft-003/verify/chk-03/` |
| `CHK-04` | `EC-03`, `SC-02` | From a clean Go module cache run `go install github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@v1.0.0`, then run `memory-bank-cli --version`. | install exits 0 and the command prints `memory-bank-cli v1.0.0`. | `artifacts/ft-003/verify/chk-04/` |
| `CHK-05` | `EC-04`, `REQ-04`, `SC-03` | Review repository install/upgrade docs and generated `v1.0.0` release notes. | required install command and breaking-change statement exist; no compatibility install path exists. | `artifacts/ft-003/verify/chk-05/` |
| `CHK-06` | `REQ-02`, `AG-01`, `CP-03` | Inspect the manual release run: validation succeeds on the selected `main` commit, then the GitHub `release` environment approves the job that creates the tag and publishes the release. | the approval record and workflow commit SHA precede tag creation; neither is replaced by a release URL or asset inventory. | `artifacts/ft-003/verify/chk-06/` |

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | validation workflow run URL/log | CI | `artifacts/ft-003/verify/chk-01/` | `CHK-01` |
| `EVID-02` | release workflow dependency/config inspection | reviewer/CI | `artifacts/ft-003/verify/chk-02/` | `CHK-02` |
| `EVID-03` | release/tag URL and asset inventory | release maintainer | `artifacts/ft-003/verify/chk-03/` | `CHK-03` |
| `EVID-04` | clean-cache install command and command smoke output | release maintainer/CI | `artifacts/ft-003/verify/chk-04/` | `CHK-04` |
| `EVID-05` | documentation and release-note review output | reviewer | `artifacts/ft-003/verify/chk-05/` | `CHK-05` |
| `EVID-06` | exact validated workflow-commit record and GitHub `release` environment deployment-approval record | CI/GitHub | `artifacts/ft-003/verify/chk-06/` | `CHK-06` |
