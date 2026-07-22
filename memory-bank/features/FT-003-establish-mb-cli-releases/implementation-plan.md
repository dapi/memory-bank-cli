---
title: "FT-003: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for accepted FT-003 CI, release, documentation and approved-publication design."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_003_scope
  - ft_003_selected_solution
  - ft_003_acceptance_criteria
---

# FT-003: Implementation Plan

## Preconditions

- `PRE-01` The target commit is approved for the first release candidate and local Go checks pass.
- `PRE-02` A maintainer confirms access to the repository release permission and any credential required by the existing `.goreleaser.yml`; no secret value is stored in the repository.
- `PRE-03` `AG-01` is approved before creating/pushing `v1.0.0` or allowing external publication.

## Design Realization Mapping

| Canonical solution refs | Owner | Realization target | Steps | Checks | Evidence |
| --- | --- | --- | --- | --- | --- |
| `SOL-01`, `SD-02`, `INV-01`, `FM-01` | `design.md` | validation workflow | `STEP-01`, `STEP-02` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `SOL-02`, `CTR-01`, `SD-01`, `SD-03`, `FM-02` | `design.md` | tag-driven release workflow | `STEP-02`, `STEP-04` | `CHK-02`, `CHK-03` | `EVID-02`, `EVID-03` |
| `SOL-03`, `INV-02`, `INV-04`, `FM-03`, `FM-04`, `RB-01`–`RB-03` | `design.md` | version resolution, documentation, release notes, tag/release verification | `STEP-03`, `STEP-04`, `STEP-05` | `CHK-03`–`CHK-05` | `EVID-03`–`EVID-05` |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01`, `SOL-01` | validated CI workflow and snapshot build evidence | either | `PRE-01` |
| `WS-2` | `REQ-03`, `REQ-04`, `SOL-03` | install/upgrade docs and release-note source | either | none |
| `WS-3` | `REQ-02`, `REQ-04`, `SOL-02` | approved `v1.0.0` GitHub release and external evidence | human/either | `WS-1`, `WS-2`, `PRE-02`, `PRE-03` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | candidate is validated and ready for `v1.0.0` | `STEP-04` | Tag/release creation and credentialed external publication are irreversible public effects. | required reviewer approves the GitHub `release` environment deployment; deployment record in `EVID-03` |

## Order of Work

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | either | `REQ-01`, `SOL-01` | add validation workflow | `.github/workflows/`, `.goreleaser.yml` only if validation requires no semantic change | workflow definitions | `CHK-01` | `EVID-01` | run Go tests, vet, GoReleaser check and clean snapshot release build | `PRE-01` | none | GoReleaser config requires an unrecorded external change |
| `STEP-02` | either | `REQ-01`, `REQ-02`, `SOL-02`, `INV-01` | add release workflow dependency on validation | `.github/workflows/` | tag-driven release workflow | `CHK-02` | `EVID-02` | inspect job graph and run a non-publishing validation path | `STEP-01` | none | publication could bypass failed validation |
| `STEP-03` | either | `REQ-03`, `REQ-04`, `SOL-03`, `INV-02`, `INV-04` | make Go-installed version observable; add install/upgrade docs and breaking release notes | `cmd/mb-cli`, unit tests, root README and release-note source/config | version resolution and reviewable documentation | `CHK-04`, `CHK-05` | `EVID-04`, `EVID-05` | run version-resolution tests; review exact install command and prohibited identity surface | none | none | a tagged module cannot report its version or docs contradict the issue |
| `STEP-04` | human/either | `REQ-02`, `SOL-02`, `SD-01`, `SD-03` | create approved `v1.0.0` tag and publish | GitHub tag/release and configured secrets | public release | `CHK-03` | `EVID-03` | approve `AG-01`, trigger tag workflow, inspect release/assets | `STEP-02`, `STEP-03`, `PRE-02`, `PRE-03` | `AG-01` | credential missing/rejected or workflow fails |
| `STEP-05` | either | `REQ-02`–`REQ-04`, `RB-03` | independently verify consumer installation and documentation | clean Go cache, published release, docs | final evidence | `CHK-04`, `CHK-05` | `EVID-04`, `EVID-05` | execute exact Go install then smoke command; review docs/release notes | `STEP-04` | none | install or identity checks fail |

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `SOL-01`, `DV-01` | validation workflow has successful snapshot-build evidence | `EVID-01` |
| `CP-02` | `STEP-02`, `INV-01` | publish job is structurally dependent on validation | `EVID-02` |
| `CP-03` | `STEP-04`, `AG-01`, `RB-02` | human approval and release prerequisites are recorded before external effect | `EVID-03` |
| `CP-04` | `STEP-05`, `EC-02`–`EC-04` | release, clean-cache install and documentation evidence agree | `EVID-03`–`EVID-05` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Candidate checks fail | invalid release could be published | keep publication dependent on validation | any failed test/vet/build |
| `ER-02` | Existing distribution credential is unavailable | release cannot complete | stop before tag/release and obtain human direction | credential validation fails |
| `ER-03` | Tag or asset names expose removed identities | breaking-release contract is violated | inspect names before and after release | `CHK-03` or `CHK-05` fails |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `FM-01`, `ER-01` | validation fails | do not tag or publish; correct candidate | committed but unpublished candidate |
| `STOP-02` | `FM-02`, `ER-02`, `CON-01` | credential/approval absent or rejected | stop external workflow and request maintainer direction | validated, unpublished candidate |
| `STOP-03` | `FM-03`, `ER-03` | old identity is present in artifact/docs | do not publish or claim completion; correct source | validated state before tag |
