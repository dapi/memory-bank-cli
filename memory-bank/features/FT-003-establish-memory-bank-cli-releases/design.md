---
title: "FT-003: Design"
doc_kind: feature
doc_function: canonical
purpose: "Feature-local CI, release and publication solution for the first stable `memory-bank-cli` distribution."
derived_from:
  - brief.md
  - ../../ops/release.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_003_scope
  - ft_003_acceptance_criteria
  - ft_003_evidence_contract
  - implementation_sequence
---

# FT-003: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `ALT-*`, `TRD-*`, `C4-*`, `SD-*`, `INV-*`, `FM-*`, `RB-*` |
| `decision-log.md` | Derived decision ledger | decision rationale and links to canonical owners; no new requirements or solution |

## Context

The current executable rename requires a stable, Go-installable `v1.0.0`
release. The repository already contains a single-binary GoReleaser
configuration and a validation workflow. The solution validates the exact
release commit before it creates the public tag or GitHub Release, then
isolates those external effects behind a human gate.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-01` | `required (C2 compact view)` | The feature introduces GitHub Actions automation and a GitHub Release/Go module distribution boundary. | topology below |

| Container / boundary | Responsibility | Connector |
| --- | --- | --- |
| GitHub Actions validation workflow | runs tests, vet and release-build validation | repository source → Go toolchain / GoReleaser |
| GitHub Actions release workflow | repeats required validation, then invokes GoReleaser publication after approval | protected tag → GitHub Release |
| GitHub Release / Go module proxy | exposes tag and release assets to consumers | `v1.0.0` → `go install` |
| User environment | installs and invokes the published command | `go install` → `memory-bank-cli` |

## Architecture Coverage Decision

| Aspect | Status | Canonical owner / refs | Supporting view / artifact | Coverage note |
| --- | --- | --- | --- | --- |
| Components / responsibilities | covered | `SOL-01`, `SOL-02` | C2 compact view | Separates validation from external publication. |
| Connectors / interactions | covered | `CTR-01` | release workflow | Tag-to-release and Go-install interaction are explicit. |
| Configuration / topology | covered | `SOL-01`, `SOL-02`, `C4-01` | workflow files and `.goreleaser.yml` | Uses existing one-binary release configuration. |
| Behavioral semantics | covered | `INV-01`, `INV-02`, `FM-01` | verify checks | Publication requires validation; only `memory-bank-cli` is distributed. |
| Quality / evolution concerns | covered | `TRD-01`, `RB-01`, `RB-02` | decision log and approval gate | GitHub Release publication is auditable and can stop safely before that release boundary. |

## Selected Solution

- `SOL-01` Add a validation workflow for repository changes that runs `go test -count=1 -race ./...`, `go vet ./...`, GoReleaser configuration validation and a clean snapshot release build.
- `SOL-02` Add a `workflow_dispatch` release workflow that validates the selected `main` commit, then—only after its validation job succeeds and the protected GitHub `release` environment has received required-reviewer approval—creates `v1.0.0` on that exact commit and invokes the existing GoReleaser GitHub-release configuration.
- `SOL-03` Publish the first release under the current executable identity from semantic-version tag `v1.0.0`; add repository install/upgrade instructions for the exact Go command and release notes declaring the breaking identity change. When GoReleaser has not injected a linker version, resolve `--version` from Go build information so module-installed binaries report their tagged version.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Publish manually without CI workflow | Fails `REQ-01` and cannot demonstrate validation before the public tag or GitHub Release. |
| `ALT-02` | Publish a snapshot/non-semantic tag | Cannot satisfy the explicitly required stable `v1.0.0` Go install contract. |
| `ALT-03` | Reintroduce old executable artifacts | Explicitly forbidden by `REQ-04` and issue #3. |

## Trade-offs

| Trade-off ID | Decision | Benefit | Cost / Risk |
| --- | --- | --- | --- |
| `TRD-01` | Validate in the manually dispatched release workflow instead of trusting an earlier CI run | The release event has direct, auditable proof that the exact tagged commit passed before publication. | Release takes longer and duplicates compute. |
| `TRD-02` | Keep existing GoReleaser destinations unchanged | Does not silently alter a declared distribution surface. | A required configured credential can block publication and must be escalated. |

## Accepted Local Decisions

- `SD-01` A manually dispatched release run on `main` is the release trigger; after its validation and `AG-01`, it creates this feature's first permitted public tag, `v1.0.0`, on the validated commit.
- `SD-02` A snapshot release build validates artifact production without claiming publication evidence.
- `SD-03` The release workflow must use repository-held credentials only through GitHub secrets; secret values are never documented or emitted as evidence.

## Contracts

| Contract ID | Connector / direction | Roles and sync boundary | Guarantees / failure / evolution semantics |
| --- | --- | --- | --- |
| `CTR-01` | manual release run on `main` → validation → `AG-01` → `v1.0.0` tag / GoReleaser GitHub Release → Go module consumer | asynchronous workflow then synchronous consumer install | validation and approval precede the immutable tag and GitHub Release; the tag points to the validated commit; published module path and command use `memory-bank-cli`. |

## Invariants

- `INV-01` No tag or GitHub Release can be created before its required validation job succeeds on the exact release commit.
- `INV-02` Published executable artifacts, install documentation and release notes expose `memory-bank-cli` only; payload-path mentions are not executable identities.
- `INV-03` A public tag or GitHub Release is never created by unattended workflow execution: the GitHub `release` environment requires `AG-01` approval before the job that creates both can run.
- `INV-04` A binary installed from a tagged Go module reports that module tag through `memory-bank-cli --version`; local development builds may report `dev`.

## Failure Modes

- `FM-01` Tests, vet, GoReleaser validation or snapshot build fail. Mitigation: the release job does not run, so no tag or GitHub Release is created; correct the candidate and restart validation.
- `FM-02` A required GitHub or existing distribution credential is absent or rejected. Mitigation: stop before tag/release and request credential/configuration direction; do not weaken or bypass the configured destination.
- `FM-03` Release assets or documentation expose an old executable identity. Mitigation: fail `CHK-03`/`CHK-05`; before tag creation, correct the source and rerun validation. If this is discovered only after `v1.0.0` is public, FT-003 is blocked pending human approval to change its required version/acceptance; do not repoint the tag or treat a new version as satisfying this feature.
- `FM-04` A Go-installed binary retains the local `dev` fallback because no linker flag is applied. Mitigation: fall back to `debug.ReadBuildInfo` and cover linker, tagged-module and local-build cases with unit tests.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | Validate candidate | workflows and docs are committed; local/CI validation is green | revert unpublished workflow/config/doc changes |
| `RB-02` | Public `v1.0.0` release | manual validation succeeds on the selected `main` commit; `AG-01` is approved and required credentials are confirmed | do not create the tag or GitHub Release if validation, credentials or `AG-01` fail. If `v1.0.0` is created and a later acceptance defect is discovered, preserve the tag and mark FT-003 blocked pending human-approved change to its required version/acceptance. |
| `RB-03` | Post-publication verification | GitHub release and Go install are observable | preserve evidence and escalate any release correction; never rewrite history as a local rollback |

## Design Verification

| Verification ID | Required | Method | Result / evidence |
| --- | --- | --- | --- |
| `DV-01` | yes | inspect workflow dependency and run a snapshot build | `EVID-01`, `EVID-02` |
| `DV-02` | yes | inspect the tag/release asset inventory and clean-cache Go install | `EVID-03`, `EVID-04` |
| `DV-03` | yes | inspect repository and release documentation for names/install paths | `EVID-05` |
