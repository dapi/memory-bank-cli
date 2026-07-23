---
title: "FT-003: Design"
doc_kind: feature
doc_function: canonical
purpose: "Feature-local CI, release and publication solution for the first stable `mb-cli` distribution."
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

Issue #3 requires a stable, Go-installable `v1.0.0` release. The repository already contains a single-binary GoReleaser configuration, but no CI workflow or released tag. The solution must prove validation before GitHub Release publication and isolate external effects behind a human gate.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-01` | `required (C2 compact view)` | The feature introduces GitHub Actions automation and a GitHub Release/Go module distribution boundary. | topology below |

| Container / boundary | Responsibility | Connector |
| --- | --- | --- |
| GitHub Actions validation workflow | runs tests, vet and release-build validation | repository source → Go toolchain / GoReleaser |
| GitHub Actions release workflow | repeats required validation, then invokes GoReleaser publication after approval | protected tag → GitHub Release |
| GitHub Release / Go module proxy | exposes tag and release assets to consumers | `v1.0.0` → `go install` |
| User environment | installs and invokes the published command | `go install` → `mb-cli` |

## Architecture Coverage Decision

| Aspect | Status | Canonical owner / refs | Supporting view / artifact | Coverage note |
| --- | --- | --- | --- | --- |
| Components / responsibilities | covered | `SOL-01`, `SOL-02` | C2 compact view | Separates validation from external publication. |
| Connectors / interactions | covered | `CTR-01` | release workflow | Tag-to-release and Go-install interaction are explicit. |
| Configuration / topology | covered | `SOL-01`, `SOL-02`, `C4-01` | workflow files and `.goreleaser.yml` | Uses existing one-binary release configuration. |
| Behavioral semantics | covered | `INV-01`, `INV-02`, `FM-01` | verify checks | Publication requires validation; only `mb-cli` is distributed. |
| Quality / evolution concerns | covered | `TRD-01`, `RB-01`, `RB-02` | decision log and approval gate | GitHub Release publication is auditable and can stop safely before that release boundary. |

## Selected Solution

- `SOL-01` Add a validation workflow for repository changes that runs `go test -count=1 -race ./...`, `go vet ./...`, GoReleaser configuration validation and a clean snapshot release build.
- `SOL-02` Add a tag-driven release workflow that repeats the required validation and invokes the existing GoReleaser GitHub-release configuration only after its validation job succeeds and the protected GitHub `release` environment has received required-reviewer approval.
- `SOL-03` Publish the first release from semantic-version tag `v1.0.0`; add repository install/upgrade instructions for the exact Go command in issue #3 and release notes declaring `memory-bank` breaking removal and `memory-bank-lint` removal. When GoReleaser has not injected a linker version, resolve `--version` from Go build information so module-installed binaries report their tagged version.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Publish manually without CI workflow | Fails `REQ-01` and cannot demonstrate validation before GitHub Release publication. |
| `ALT-02` | Publish a snapshot/non-semantic tag | Cannot satisfy the explicitly required stable `v1.0.0` Go install contract. |
| `ALT-03` | Reintroduce old executable artifacts | Explicitly forbidden by `REQ-04` and issue #3. |

## Trade-offs

| Trade-off ID | Decision | Benefit | Cost / Risk |
| --- | --- | --- | --- |
| `TRD-01` | Repeat validation in the release workflow instead of trusting an earlier CI run | The release event has direct, auditable proof that publication was gated. | Release takes longer and duplicates compute. |
| `TRD-02` | Keep existing GoReleaser destinations unchanged | Does not silently alter a declared distribution surface. | A required configured credential can block publication and must be escalated. |

## Accepted Local Decisions

- `SD-01` A semantic tag is the release trigger; this feature's first permitted public tag is `v1.0.0`.
- `SD-02` A snapshot release build validates artifact production without claiming publication evidence.
- `SD-03` The release workflow must use repository-held credentials only through GitHub secrets; secret values are never documented or emitted as evidence.

## Contracts

| Contract ID | Connector / direction | Roles and sync boundary | Guarantees / failure / evolution semantics |
| --- | --- | --- | --- |
| `CTR-01` | approved `v1.0.0` tag → GitHub Actions → GoReleaser/GitHub Release → Go module consumer | asynchronous workflow then synchronous consumer install | the pushed tag exposes the immutable Go module version; GitHub Release publication begins only after workflow validation and `AG-01`; published module path and command use `mb-cli`. |

## Invariants

- `INV-01` No publication job can run before its required validation job succeeds.
- `INV-02` Published executable artifacts, install documentation and release notes expose `mb-cli` only; payload-path mentions are not executable identities.
- `INV-03` A public GitHub Release is never published by unattended workflow execution: the GitHub `release` environment requires `AG-01` approval before its publication job can run. Creating the tag remains a separately authorized maintainer action under repository permissions.
- `INV-04` A binary installed from a tagged Go module reports that module tag through `mb-cli --version`; local development builds may report `dev`.

## Failure Modes

- `FM-01` Tests, vet, GoReleaser validation or snapshot build fail. Mitigation: before a tag push, do not tag; after tag-triggered validation fails, do not publish a GitHub Release or repoint the tag. Treat the tagged module version as public and use a new semantic version for any corrected release after human authorization.
- `FM-02` A required GitHub or existing distribution credential is absent or rejected. Mitigation: stop before tag/release and request credential/configuration direction; do not weaken or bypass the configured destination.
- `FM-03` Release assets or documentation expose an old executable identity. Mitigation: fail `CHK-03`/`CHK-05`, correct the source configuration/docs and publish only after a new approved validation run.
- `FM-04` A Go-installed binary retains the local `dev` fallback because no linker flag is applied. Mitigation: fall back to `debug.ReadBuildInfo` and cover linker, tagged-module and local-build cases with unit tests.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | Validate candidate | workflows and docs are committed; local/CI validation is green | revert unpublished workflow/config/doc changes |
| `RB-02` | Public `v1.0.0` release | a maintainer records authorization before the tag push; after tag-triggered validation succeeds, `AG-01` is approved and required credentials are confirmed | do not push without tag authorization; after an authorized push, stop before the release publication job if validation, credentials or `AG-01` fail. The tag is already a public module version: preserve it, do not repoint it, and use a new semantic version for any corrected release. |
| `RB-03` | Post-publication verification | GitHub release and Go install are observable | preserve evidence and escalate any release correction; never rewrite history as a local rollback |

## Design Verification

| Verification ID | Required | Method | Result / evidence |
| --- | --- | --- | --- |
| `DV-01` | yes | inspect workflow dependency and run a snapshot build | `EVID-01`, `EVID-02` |
| `DV-02` | yes | inspect the tag/release asset inventory and clean-cache Go install | `EVID-03`, `EVID-04` |
| `DV-03` | yes | inspect repository and release documentation for names/install paths | `EVID-05` |
