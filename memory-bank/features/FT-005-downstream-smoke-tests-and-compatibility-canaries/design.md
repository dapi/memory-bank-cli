---
title: "FT-005: Downstream Smoke Tests and Compatibility Canaries Design"
doc_kind: feature
doc_function: canonical
purpose: "Selected solution for FT-005: isolated downstream fixture topology, stable/canary input policy, release-asset integrity and phase attribution."
derived_from:
  - brief.md
  - ../../engineering/testing-policy.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_005_scope
  - ft_005_acceptance_criteria
  - implementation_sequence
---

# FT-005: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `CTR-*`, `INV-*`, `FM-*`, design verification |

## Selected Design

- `SOL-01` Stable CI creates new temporary source and downstream Git repositories, clones this repository at `v1.0.1` (`b0d8ca47cdb3315df4b755a704d01ffb139754e7`), installs the CLI with documented `go install ...@v1.0.1`, and passes the clone, version and SHA to `init`/`update`. This forward-only pin replaces the poisoned `v1.0.0` module version.
- `SOL-02` Canary runs on schedule and `workflow_dispatch`; it accepts separate CLI/template refs, resolves each to a SHA before work, installs the CLI through `go install ...@<resolved-cli-sha>`, and clones the template at its resolved SHA. It reports requested refs, resolved SHAs, tool versions and phase result. Defaults are `main`/`main`.
- `SOL-03` Go installation remains the fixture's CLI path. When the selected release contains both `checksums.txt` and binary assets, a separate packaging phase downloads every binary asset and validates its SHA-256 entry.
- `SOL-04` Every run assigns one terminal boundary: `packaging` (install/assets/checksum), `template` (clone/source/provenance), `cli` (commands after inputs are established), or `external-tooling` (Go/Git/network/runner before a feature-owned phase). It reports the boundary, not an unproven root cause.

## C4 Applicability

`C4 required: no`: this adds CI/subprocess/filesystem boundaries, not a new deployed runtime component.

| Component | Responsibility | Connector |
| --- | --- | --- |
| workflow | selects lane and read-only environment | runs fixture |
| fixture | provisions repos, resolves inputs and reports phases | Go, Git and CLI subprocesses |
| template clone | explicit CLI source | local filesystem |
| release assets | conditional packaging evidence | HTTPS and SHA-256 |

## Architecture Coverage Decision

| Concern | Decision |
| --- | --- |
| Components | covered by `SOL-01`–`SOL-04` |
| Connectors | Go install, Git clone, subprocess, filesystem and conditional HTTPS covered |
| Configuration | stable fixed pair, canary refs and read-only permissions covered |
| Behavioral semantics | `CTR-01`–`CTR-03` cover provenance, integrity and classification |
| Quality/evolution | isolation, idempotence, provenance and checksums covered |

## Contracts, Invariants and Failures

| ID | Contract / invariant |
| --- | --- |
| `CTR-01` | Every run records requested CLI/template refs and resolved immutable commits before `init`; stable records its fixed pair. |
| `CTR-02` | Asset validation runs only when `checksums.txt` and a selected binary both exist; each downloaded binary must match its named SHA-256. |
| `CTR-03` | A report has exactly one terminal boundary classification. |
| `INV-01` | Stable never uses the triggering checkout as downstream repo or template source. |
| `INV-02` | Repeat-update no-diff runs after a minimal adaptation and user-owned file exist. |
| `INV-03` | Stable workflow uses `contents: read`, no write token or secret. |
| `FM-01` | Ref resolution/install/checksum failure stops the run and preserves inputs/report; no substitute ref is allowed. |
| `FM-02` | Lifecycle failure retains logs/diff and reports command boundary. |
| `RB-01` | All fixture repositories are temporary; backout removes additive fixture/workflow files. |

## Design Verification

| Analysis | Required | Method / evidence |
| --- | --- | --- |
| Contract compatibility | yes | fixture proves source flags, preservation and second-update no diff |
| Failure propagation | yes | phase report has one classification for each failure boundary |
| Security boundary | yes | workflow review proves temporary directories and read-only permissions |
| Concurrency/ordering | no | every job owns a temporary workspace |
| Capacity/latency | no | bounded one-repository fixture; no target is stated |
| Migration/evolution | yes | fixed stable and reported resolved canary inputs prevent silent drift |
