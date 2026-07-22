---
title: "FT-001: Design"
doc_kind: feature
doc_function: canonical
purpose: "Feature-local решение для history-preserving переноса Go CLI и замены его единственной публичной исполняемой идентичности на `mb-cli`."
derived_from:
  - brief.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_001_scope
  - ft_001_acceptance_criteria
  - ft_001_evidence_contract
  - implementation_sequence
---

# FT-001: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `ALT-*`, `TRD-*`, `SD-*`, `INV-*`, `FM-*`, `RB-*` |
| `decision-log.md` | Derived decision ledger | decision rationale and links to canonical owners; no new requirements or solution |

## Context

The source snapshot in `brief.md` packages the CLI below `tools/` with module path `github.com/dapi/memory-bank/tools`. It has two entrypoints and a GoReleaser configuration for `memory-bank` and `legacy compatibility executable`. The target repository must retain the useful `tools/` history but expose only the renamed primary executable while leaving issue #2's source-template detection and issue #3's release publication out of scope.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-00` | `not required` | The feature relocates one local CLI module and does not select a new runtime topology, external connector or deployable component. | none |

## Architecture Coverage Decision

| Aspect | Status | Canonical owner / refs | Supporting view / artifact | Reason if N/A / coverage note |
| --- | --- | --- | --- | --- |
| Components / responsibilities | covered | `design.md` `SOL-01`, `SOL-02` | source package layout | Defines retained internal packages and the single entrypoint boundary. |
| Connectors / interactions | N/A | `brief.md` `NS-01` | none | No new integration is designed; issue #2 owns its changed detection mechanism. |
| Configuration / topology | covered | `design.md` `SOL-03` | GoReleaser configuration | Release-build configuration changes from two old binaries to one new binary. |
| Behavioral semantics | covered | `design.md` `INV-01`, `INV-02`, `FM-01` | migrated tests | Existing command behavior is retained while its public executable prefix changes. |
| Quality / evolution concerns | covered | `brief.md` `CON-01`, `CON-02`; `design.md` `TRD-01`, `RB-01` | Git provenance and issue handoff | Preserves auditability and avoids claiming future release evidence. |

## Selected Solution

- `SOL-01` Create a history-filtered import of the `tools/` subtree from source SHA `0957f3c495f2c0518c8a81448694cf0e231d3209`, removing the `tools/` path prefix in the imported branch. Merge that import into the target history, retaining the target's initial repository commit. This preserves the meaningful CLI lineage while making its files native to the standalone repository.
- `SOL-02` Change internal imports and `go.mod` to `github.com/dapi/memory-bank-cli`; rename `cmd/memory-bank` to `cmd/mb-cli`; route all primary CLI strings, help, errors, version output and examples through `mb-cli`.
- `SOL-03` Delete the compatibility entrypoint and `RunLint` compatibility path, tests dedicated to it, and every `legacy compatibility executable` release-build/archive/package configuration. Configure the remaining release build to emit only `mb-cli`; do not perform publication in this feature.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Copy the current files without imported history | Fails `REQ-01`'s requirement to preserve relevant history where practical. |
| `ALT-02` | Keep `memory-bank` or `legacy compatibility executable` wrapper binaries | Explicitly forbidden by issue #1 and `NS-04`. |
| `ALT-03` | Redesign `doctor --profile auto` during import | Explicitly owned by issue #2 and excluded by `NS-01`. |

## Trade-offs

| Trade-off ID | Decision | Benefit | Cost / Risk |
| --- | --- | --- | --- |
| `TRD-01` | Preserve filtered `tools/` history rather than original repository-wide topology | Reviewers can trace the CLI's evolution without retaining unrelated template history. | Imported commit IDs may differ from source IDs; provenance must record the source SHA and filtering command. |
| `TRD-02` | Treat old executable spelling as identity only, not as a blanket text replacement | Removes prohibited executable identities without corrupting legitimate `memory-bank/` payload paths. | Search review must classify matches rather than relying on an unsafe global rename. |

## Accepted Local Decisions

- `SD-01` The source snapshot is fixed to the SHA recorded in `brief.md`; a newer source revision requires updating the brief, decision log and verification baseline before import.
- `SD-02` The history import is performed in a disposable clone and inspected before it is merged into the target. This keeps source and target histories recoverable if the filtered tree is wrong.
- `SD-03` Release configuration is changed only far enough to make `mb-cli` the sole build artifact. Tagging, publishing, CI release automation and installation documentation remain issue #3 work.

## Contracts

| Contract ID | Connector / direction | Roles and sync boundary | Guarantees / failure / evolution semantics |
| --- | --- | --- | --- |
| `CTR-01` | user or automation → `mb-cli` process | synchronous CLI invocation | `lint`, `doctor`, `init`, `update`, exit codes and versioned JSON contracts are preserved; only executable/usage identity changes. |

## Invariants

- `INV-01` `lint`, `doctor`, `init` and `update` retain semantics, exit codes and versioned JSON contracts.
- `INV-02` No executable, wrapper, archive or release-build target named `memory-bank` or `legacy compatibility executable` remains.
- `INV-03` `memory-bank/` is permitted only where it denotes the documentation payload or an existing data path, never an executable identity.

## Failure Modes

- `FM-01` An indiscriminate rename changes payload paths or behavior. Mitigation: classify all residual `memory-bank` matches and run retained regression tests.
- `FM-02` A history filter imports a wrong tree or omits relevant lineage. Mitigation: inspect filtered tree and log source SHA, command and resulting import commit before merge.
- `FM-03` Legacy release configuration emits an old artifact despite source cleanup. Mitigation: inspect config and build output with `CHK-04`.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | History import and rename on feature branch | filtered tree inspection confirms only former `tools/` content | abandon the unmerged import branch; target `main` is unchanged |
| `RB-02` | Handoff to release work | all FT-001 checks and review pass | do not tag/publish; retain the verified feature branch for issue #3 |

## Design Verification

| Verification ID | Required | Method | Result / evidence |
| --- | --- | --- | --- |
| `DV-01` | yes | inspect filtered tree and import provenance against `ASM-01` and `CON-01` | `EVID-05` before merge |
| `DV-02` | yes | run retained contract tests and full race/vet checks | `EVID-01`, `EVID-02`, `EVID-04` |
| `DV-03` | yes | inspect old-name search and configured/build output | `EVID-03` |
