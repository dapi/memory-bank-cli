---
title: "FT-003: Establish mb-cli Releases"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для CI, первого stable Go release и install/upgrade документации standalone `mb-cli`."
derived_from:
  - ../../prd/PRD-001-standalone-mb-cli.md
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

# FT-003: Establish mb-cli Releases

## What

### Problem

The standalone module has a GoReleaser configuration for `mb-cli`, but no repository CI workflow, tag, GitHub release, installation/upgrade documentation, or release procedure is documented. Issue #3 requires the first stable release while preserving `mb-cli` as the only distributed executable identity.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Automated release validation | no `.github` workflow existed at the feature baseline | a repository workflow runs tests, vet and a release build before publication | inspect workflow and its successful run |
| `MET-02` | Stable Go module availability | no tags or releases | tagged `v1.0.0` is installable through the issue-specified `go install` command | execute install and version/help smoke check |
| `MET-03` | Public release identities | release config names only `mb-cli`, but nothing has been published | release assets and documentation expose only `mb-cli`; removed identities have no compatibility artifact | inspect release, assets and docs |

### Scope

- `REQ-01` Add repository automation that validates Go tests, static analysis and the configured release build before any publication.
- `REQ-02` Define and execute the first semantic-version release as tag `v1.0.0`, publishing `mb-cli` through the existing GitHub/GoReleaser release surface only after the required validation succeeds.
- `REQ-03` Add repository installation and upgrade documentation using `go install github.com/dapi/memory-bank-cli/cmd/mb-cli@vX.Y.Z` and document the intentional breaking rename/removal in the first release notes.
- `REQ-04` Ensure the release workflow, release assets, documentation and release notes contain no compatibility artifact or supported installation path for `memory-bank` or `memory-bank-lint`.

### Non-Scope

- `NS-01` Change the semantics, exit codes or versioned JSON contracts of `lint`, `doctor`, `init` or `update`.
- `NS-02` Restore an alias, wrapper binary or transition path for `memory-bank` or `memory-bank-lint`.
- `NS-03` Redesign `doctor --profile auto` detection, which belongs to issue #2.
- `NS-04` Change GoReleaser distribution destinations beyond what the existing configuration already declares; a missing required credential is a stop condition, not permission to redesign distribution.

### Constraints / Assumptions

- `ASM-01` Issue #3 is open and explicitly requires a tagged, Go-installable `v1.0.0`; the repository currently has no tag or GitHub release.
- `ASM-02` `.goreleaser.yml` is the existing release-build configuration: it has one `mb-cli` build, GitHub release settings and a Homebrew cask that requires `HOMEBREW_TAP_GITHUB_TOKEN`.
- `CON-01` Creating a public tag/release, using GitHub credentials and publishing external assets are irreversible external effects; they require an explicit human approval gate.
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

- `EC-01` Automation runs Go tests, `go vet ./...`, GoReleaser configuration validation and a release build before publication.
- `EC-02` GitHub contains the stable `v1.0.0` release and its assets contain only `mb-cli` executable names.
- `EC-03` `go install github.com/dapi/memory-bank-cli/cmd/mb-cli@v1.0.0` succeeds from a clean Go module cache and the installed command responds as `mb-cli`.
- `EC-04` Repository documentation and `v1.0.0` release notes state the intentional breaking rename and removal, without presenting a compatibility installation path.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-02`, `CON-01` | `EC-01`, `SC-01` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-02` | `ASM-01`, `CON-01`, `CON-02` | `EC-02`, `EC-03`, `SC-02` | `CHK-03`, `CHK-04` | `EVID-03`, `EVID-04` |
| `REQ-03` | `ASM-01`, `CON-02` | `EC-03`, `EC-04`, `SC-02` | `CHK-04`, `CHK-05` | `EVID-04`, `EVID-05` |
| `REQ-04` | `ASM-02`, `CON-03` | `EC-02`, `EC-04`, `SC-03` | `CHK-03`, `CHK-05` | `EVID-03`, `EVID-05` |

### Acceptance Scenarios

- `SC-01` A contributor opens a release candidate change; automation completes tests, vet and release-build validation before it can be published.
- `SC-02` After approved publication of `v1.0.0`, a user installs the module with the issue-specified Go command and invokes `mb-cli`.
- `SC-03` A user reads the repository/release documentation and can identify the breaking rename without being offered a removed executable as an installable compatibility path.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Inspect the validation and release workflow definitions; run the validation workflow or its equivalent commands. | tests, vet, release-config validation and snapshot release build succeed before publish step. | `artifacts/ft-003/verify/chk-01/` |
| `CHK-02` | `EC-01`, `SC-01` | Inspect the release workflow job dependency and run record. | publish job cannot run until validation job succeeds. | `artifacts/ft-003/verify/chk-02/` |
| `CHK-03` | `EC-02`, `REQ-04`, `SC-03` | Inspect tag `v1.0.0`, GitHub release and asset names. | release exists; assets have only `mb-cli` executable identity. | `artifacts/ft-003/verify/chk-03/` |
| `CHK-04` | `EC-03`, `SC-02` | From a clean Go module cache run `go install github.com/dapi/memory-bank-cli/cmd/mb-cli@v1.0.0`, then run the installed binary. | install exits 0 and the invoked command identifies as `mb-cli`. | `artifacts/ft-003/verify/chk-04/` |
| `CHK-05` | `EC-04`, `REQ-04`, `SC-03` | Review repository install/upgrade docs and generated `v1.0.0` release notes. | required install command and breaking-change statement exist; no compatibility install path exists. | `artifacts/ft-003/verify/chk-05/` |

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | validation workflow run URL/log | CI | `artifacts/ft-003/verify/chk-01/` | `CHK-01` |
| `EVID-02` | release workflow dependency/config inspection | reviewer/CI | `artifacts/ft-003/verify/chk-02/` | `CHK-02` |
| `EVID-03` | release/tag URL and asset inventory | release maintainer | `artifacts/ft-003/verify/chk-03/` | `CHK-03` |
| `EVID-04` | clean-cache install command and command smoke output | release maintainer/CI | `artifacts/ft-003/verify/chk-04/` | `CHK-04` |
| `EVID-05` | documentation and release-note review output | reviewer | `artifacts/ft-003/verify/chk-05/` | `CHK-05` |
