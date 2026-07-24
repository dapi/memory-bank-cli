---
title: "FT-021: Local Init/Update E2E"
doc_kind: feature
doc_function: canonical
purpose: "Canonical problem, scope, validation profile and verify contract for local end-to-end coverage of init/update and profiles."
derived_from:
  - ../../flows/feature.md
  - ../../engineering/testing-policy.md
  - ../../engineering/validation-profiles.md
  - "https://github.com/dapi/memory-bank-cli/issues/21"
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - selected_solution
  - implementation_sequence
---

# FT-021: Local Init/Update E2E

## What

### Problem

Issue #21 requires reproducible end-to-end evidence for the built `memory-bank-cli`: real Git URL/ref/SHA handling, `.lock` semantics, update safety, profile detection, CI merge gating, release-binary validation and a non-blocking external canary. Current downstream smoke uses networked public inputs and does not cover the issue's local bare-remote scenarios.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Required local E2E coverage | No issue-21 local E2E suite exists. | E2E-01–E2E-09 and E2E-11–E2E-27 pass without GitHub/network; E2E-10 validates the pre-publish release binary. | Versioned scenario suite, CI logs and retained reports. |
| `MET-02` | Merge gate | `main` has neither branch protection nor rulesets. | The named local E2E job is required for PR merge. | Repository ruleset/protection evidence. |

### Scope

- `REQ-01` Add a hermetic shell-driven E2E suite that builds the current checkout's binary, creates clean temporary Git repositories and a local bare `template.git`, and publishes `v1.0.0` and `v1.1.0` template versions.
- `REQ-02` Cover init, repeat-init, clean update, dry-run, source failures, downstream/template doctor profiles, Git provenance and `.lock` assertions (`E2E-01`–`E2E-09`).
- `REQ-03` Cover all required managed-file, deletion/rename, path-collision, atomicity and repeatability outcomes (`E2E-11`–`E2E-27`).
- `REQ-04` Run the local suite in a dedicated CI job for pull requests and pushes to `main`, and make its check required for merge.
- `REQ-05` Before publishing a release, exercise the release candidate binary with init and clean update (`E2E-10`).
- `REQ-06` Preserve or extend the scheduled/manual, non-blocking canary against real `dapi/memory-bank`; failure must leave actionable signal/evidence without blocking PRs.

### Non-Scope

- `NS-01` Change `init`, `update`, lock format, ownership semantics or introduce `update --check`, `--force`, or acceptance of a local version as a new lock base.
- `NS-02` Require GitHub, public network or credentials for the local blocking E2E suite.
- `NS-03` Redefine the external canary's compatibility policy beyond the issue's init/clean-update/doctor scope.

### Constraints / Assumptions

- `ASM-01` Issue #21 supplies the mandatory E2E inventory; its comment makes E2E-11–E2E-27 mandatory and labels the three future UX options non-required.
- `ASM-02` `init`/`update` already require a clean local template checkout plus `--source`, `--template-version` and full `--source-ref`.
- `ASM-03` Existing `scripts/downstream-smoke.sh` uses public GitHub, Go module installation and release APIs, so it cannot be the required hermetic local suite unchanged.
- `CON-01` Local scenario inputs must be generated in temporary directories and use only local Git paths after the test starts.
- `CON-02` Atomic conflict/source-error checks must prove both managed paths and `.lock` remain unchanged.
- `CON-03` The current `main` branch has no GitHub branch protection and no rulesets; repository-admin configuration is therefore needed to make any CI check required.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature adds test/CI/release topology, local Git boundaries, safety semantics and an external administrative merge gate. | `design.md` |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- | --- |
| `decision-log.md` | selected | Selected fixture and gate decisions need auditable provenance. | `decision-log.md` |
| `design.md` | selected | CI/release topology and safety contracts require solution reasoning. | `design.md` |
| `use-cases/README.md` | selected | 26 mandatory scenarios need review-friendly grouping without moving acceptance out of this brief. | `use-cases/README.md` |
| Separate C4/contract diagram | omitted | The selected design has no new runtime or external contract; compact tables are sufficient. | none |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `release-deployment` | The feature changes the release workflow and requires a pre-publish binary gate; the existing profile assigns this class to `release-deployment`. | none |

## Verify

### Exit Criteria

- `EC-01` A no-network local E2E suite proves the complete mandatory scenario inventory against a built binary and two local template tags.
- `EC-02` The dedicated E2E CI job runs on every PR and push to `main` and is required for merge.
- `EC-03` Release publication is blocked unless its candidate binary passes `E2E-10`; the external scheduled/manual canary remains non-blocking and retains an actionable result.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01`–`REQ-03` | `ASM-01`–`ASM-03`, `CON-01`–`CON-02` | `EC-01`, `SC-01`–`SC-04` | `CHK-01` | `EVID-01` |
| `REQ-04` | `CON-03` | `EC-02`, `SC-05` | `CHK-02` | `EVID-02` |
| `REQ-05` | `CON-02` | `EC-03`, `SC-06` | `CHK-03` | `EVID-03` |
| `REQ-06` | `ASM-03` | `EC-03`, `SC-07` | `CHK-04` | `EVID-04` |

### Acceptance Scenarios

- `SC-01` E2E-01–E2E-03 prove tagged init, repeat-init safety and clean update/create/delete with a valid lock.
- `SC-02` E2E-04–E2E-09 prove atomic conflict/dry-run/source-error behavior and downstream/template doctor profile detection.
- `SC-03` E2E-11–E2E-20 prove local-edit, mode, symlink, deletion and rename behavior, including the distinct unchanged-upstream, both-changed and equal-to-upstream cases.
- `SC-04` E2E-21–E2E-27 prove collision protection, user-owned preservation, all-or-nothing application, repeatable conflict and manual resolution in favor of upstream.
- `SC-05` The named local E2E job passes on PR and `main` push and its required-check policy is observable.
- `SC-06` A release candidate binary passes init and clean update before publication.
- `SC-07` The external canary runs only on schedule/manual dispatch and reports the result without becoming a PR merge dependency.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01`–`SC-04` | Run the local suite with network disabled after initial runner setup. | Every mandatory scenario passes; failed conflict/source cases preserve recorded paths and lock bytes. | `artifacts/ft-021/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-05` | Inspect/run the PR/main workflow and repository protection/ruleset. | Named job is green on both triggers and required for merge. | `artifacts/ft-021/verify/chk-02/` |
| `CHK-03` | `EC-03`, `SC-06` | Run release workflow through snapshot candidate and E2E-10 before publish step. | Candidate passes before a publish action is reachable. | `artifacts/ft-021/verify/chk-03/` |
| `CHK-04` | `EC-03`, `SC-07` | Inspect/dispatch canary. | It has only schedule/manual triggers and retains report/log on failure. | `artifacts/ft-021/verify/chk-04/` |

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Per-scenario stdout/stderr, filesystem/lock assertions and summary | local E2E runner | `artifacts/ft-021/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Workflow run and ruleset/protection configuration reference | CI + repository admin | `artifacts/ft-021/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Release-candidate E2E log before publish | release CI | `artifacts/ft-021/verify/chk-03/` | `CHK-03` |
| `EVID-04` | Canary run/report or failure signal | canary CI | `artifacts/ft-021/verify/chk-04/` | `CHK-04` |
