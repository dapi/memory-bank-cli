---
title: "FT-021: Local Init/Update E2E Design"
doc_kind: feature
doc_function: canonical
purpose: "Selected local fixture, CI, release and canary design for FT-021."
derived_from:
  - brief.md
  - ../../engineering/testing-policy.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_021_scope
  - ft_021_acceptance_criteria
  - implementation_sequence
---

# FT-021: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `C4-00`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` |
| `use-cases/README.md` | Derived scenario projection | `FUC-*`, `TC-*`; no canonical acceptance |

## Context

The public-network `downstream-smoke.sh` validates a released consumer path, while issue #21 requires deterministic current-checkout-binary behavior with local Git provenance, update safety and a pre-publish binary gate. The solution must preserve the existing external canary rather than treating it as a substitute for local E2E.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-00` | `not required` | No runtime/deployable, API, storage or component responsibility changes; this adds test and CI orchestration only. | none |

## Architecture Coverage Decision

| Aspect | Status | Canonical owner / refs | Supporting view | Coverage note |
| --- | --- | --- | --- | --- |
| Components / responsibilities | covered | `SOL-01`–`SOL-04` | none | runner, local remote, workflows and canary have distinct roles. |
| Connectors / interactions | covered | `CTR-01`–`CTR-03` | none | local Git, CLI subprocess and CI dependency bindings are defined. |
| Configuration / topology | covered | `SOL-01`–`SOL-04` | none | temporary paths, workflow triggers and release ordering are explicit. |
| Behavioral semantics | covered | `INV-01`–`INV-03`, `FM-01`–`FM-03` | `use-cases/README.md` | success, conflict and source-failure outcomes are traced. |
| Quality / evolution concerns | covered | `CON-01`–`CON-03`, `RB-01` | none | hermeticity, atomicity and external administration are bounded. |

## Selected Solution

- `SOL-01` Add one shell E2E runner that builds the binary from the current checkout and, per scenario, creates fresh temporary bare/source/downstream Git repositories. It writes two fixture template commits tagged `v1.0.0` and `v1.1.0`, resolves their full SHAs locally, and invokes only the built executable.
- `SOL-02` Model each E2E-01–E2E-09 and E2E-11–E2E-27 as a named independent scenario in that runner. It snapshots the relevant downstream tree and `.lock` before expected failures, then compares them after the command; a scenario cannot depend on the state of another scenario.
- `SOL-03` Add a dedicated local-E2E CI job on pull requests and pushes to `main`. Configure GitHub repository protection or a ruleset to require that job; workflow YAML alone is not treated as proof of required-merge policy.
- `SOL-04` Reuse the same runner with an explicit binary-path input for release validation: in the existing `validate` job, run E2E-10 immediately after its snapshot build creates the candidate binary. The separate `release` job already depends on `validate`, so no tag/publish action is reachable first. Keep the external canary as schedule/manual only and make it emit retained failure evidence rather than a PR dependency.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Extend networked downstream smoke as the local E2E gate | It installs/clones/downloads from public services and cannot prove the no-network requirement. |
| `ALT-02` | Test ownership internals only | It bypasses built binary, Git URL/ref/SHA and CLI boundary required by the issue. |
| `ALT-03` | Make the canary the merge gate | The issue explicitly distinguishes external non-blocking canary from required local E2E. |

## Contracts

| Contract ID | Connector / direction | Roles and boundary | Guarantees / failure / evolution semantics |
| --- | --- | --- | --- |
| `CTR-01` | local filesystem Git URL: runner → `template.git` → CLI | runner creates/tags; CLI reads clean source clone synchronously | full tag/SHA passed to CLI; no GitHub/network after setup; bad ref/URL stops before mutation. |
| `CTR-02` | CLI subprocess: runner → built binary → downstream repo | one scenario owns one downstream repository | exit code, tree and lock are asserted; expected failure preserves snapshot. |
| `CTR-03` | CI dependency: local E2E job → repository merge/release decision | CI reports status; repository admin binds required check; release job waits for candidate E2E | canary has no PR dependency; missing required-check binding is an acceptance failure, not silently waived. |

## Invariants

- `INV-01` A local scenario never contacts GitHub or another network service after the runner begins; all template refs/URLs resolve to its temporary bare remote.
- `INV-02` Expected conflict and source-error scenarios make no partial downstream mutation and leave `.lock` byte-identical.
- `INV-03` Local E2E and external canary are separate lanes: only local E2E can be required for merge.

## Failure Modes

- `FM-01` Fixture setup/tag/SHA failure prevents its CLI invocation and fails only that scenario with retained diagnostic output.
- `FM-02` A CLI result that differs from the expected exit code, path snapshot or lock snapshot fails the scenario and retains those comparisons.
- `FM-03` Missing repository-admin required-check configuration blocks acceptance; code/workflow changes may be ready but cannot claim `EC-02`.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | Add runner and CI, then bind required job and release pre-publish dependency | Local suite passes and job name is stable. | Remove additive runner/workflow wiring; repository admin removes only the FT-021 required-check binding. |

## Design Verification

| Analysis | Required | Reason / risk | Method | Result / evidence |
| --- | --- | --- | --- | --- |
| Contract compatibility | yes | CLI flags, ref/SHA and lock behavior are public paths. | Scenario walkthrough against issue and existing CLI contract. | `CTR-01`–`CTR-02`, `CHK-01`. |
| State / transition completeness | yes | Update outcomes include preservation, conflict, deletion and resolution. | Map E2E-01–E2E-27 in use cases. | `use-cases/README.md`. |
| Failure propagation | yes | Atomicity/source errors must not create partial state. | Pre/post tree and lock snapshots. | `INV-02`, `CHK-01`. |
| Concurrency / ordering | no | Each scenario owns a fresh temporary workspace; no shared writer is introduced. | N/A. | N/A. |
| Security boundaries | yes | Required merge status is an external repository-admin control. | Inspect workflow permissions and protection/ruleset. | `CHK-02`, `AG-01`. |
| Capacity / latency | no | No performance target or production load change is stated. | N/A. | N/A. |
| Migration / evolution safety | yes | Existing release/canary behavior must stay distinct from the local lane. | Workflow trigger/dependency review. | `SOL-04`, `INV-03`, `CHK-03`–`CHK-04`. |

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure / rollout refs |
| --- | --- | --- | --- |
| `REQ-01`–`REQ-03` | `SOL-01`, `SOL-02` | `CTR-01`, `CTR-02`, `INV-01`–`INV-02` | `FM-01`–`FM-02`, `RB-01` |
| `REQ-04` | `SOL-03` | `CTR-03`, `INV-03` | `FM-03`, `RB-01` |
| `REQ-05`–`REQ-06` | `SOL-04` | `CTR-03`, `INV-03` | `FM-03`, `RB-01` |
