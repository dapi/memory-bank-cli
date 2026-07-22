---
title: "FT-001: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for the accepted FT-001 migration design; it sequences history import, rename, compatibility removal and verification without changing canonical requirements."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_001_scope
  - ft_001_selected_solution
  - ft_001_acceptance_criteria
---

# FT-001: Implementation Plan

## Preconditions

- `PRE-01` Confirm source SHA `0957f3c495f2c0518c8a81448694cf0e231d3209` is available and record any intentional update through `SD-01`.
- `PRE-02` Work on the FT-001 feature branch; do not perform release/tag/publication actions owned by issue #3.

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- |
| `WS-1` | `REQ-01`, `SOL-01`, `SD-01`, `SD-02` | reviewed history-preserving source import | either | `PRE-01` |
| `WS-2` | `REQ-01`, `REQ-02`, `REQ-03`, `SOL-02`, `SOL-03`, `INV-01`–`INV-03` | standalone `mb-cli` code and tests | either | `WS-1` |
| `WS-3` | `REQ-04`, `SOL-03` | sole `mb-cli` release-build configuration, no publication | either | `WS-2` |
| `WS-4` | all requirements | recorded verification and issue #3 handoff | either | `WS-2`, `WS-3` |

## Traceability

| Canonical refs | Owner | Implementation target | Verifies | Evidence |
| --- | --- | --- | --- | --- |
| `REQ-01`, `SOL-01`, `SD-01`, `SD-02` | brief/design | imported tree and Git provenance | `CHK-06` | `EVID-05` |
| `REQ-02`, `SOL-02`, `SOL-03`, `INV-02`, `INV-03` | brief/design | entrypoint, CLI strings, compatibility deletion | `CHK-03`, `CHK-04` | `EVID-03` |
| `REQ-03`, `CTR-01`, `INV-01` | brief/design | internal packages and migrated tests | `CHK-01`, `CHK-05` | `EVID-01`, `EVID-04` |
| `REQ-04`, `SD-03` | brief/design | release-build configuration and handoff note | `CHK-04`, `CHK-06` | `EVID-03`, `EVID-05` |

## Order of Work

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | either | `SOL-01`, `SD-01`, `SD-02` | prepare filtered import from the fixed source snapshot | disposable source clone, target feature branch | inspected filtered import branch and provenance note | `CHK-06` | `EVID-05` | filter `tools/` to repository root; inspect tree and `git log` before merge | `PRE-01` | none | source SHA differs or filtering does not produce only CLI content |
| `STEP-02` | either | `REQ-01`, `SOL-02` | merge reviewed import and establish standalone module/entrypoint | `go.mod`, internal imports, `cmd/mb-cli` | compilable standalone source tree | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` | run Go checks after rename | `STEP-01` | none | import conflict changes target README or retained source is incomplete |
| `STEP-03` | either | `REQ-02`, `REQ-03`, `SOL-02`, `SOL-03`, `INV-01`–`INV-03` | rename user-facing identity and remove compatibility paths without altering payload paths | CLI code/tests, docs/examples within migrated CLI, old command directory | only `mb-cli` command surface and preserved contract tests | `CHK-03`, `CHK-05` | `EVID-03`, `EVID-04` | classify name matches; run contract tests | `STEP-02` | none | a required compatibility removal changes command/JSON/exit-code contract |
| `STEP-04` | either | `REQ-04`, `SOL-03`, `SD-03` | configure one release-build artifact, but do not publish | release config | `mb-cli` only build configuration | `CHK-04` | `EVID-03` | inspect config and build output | `STEP-03` | none | configuration requires a tag, external credential or publication action |
| `STEP-05` | either | all requirements | execute final profile checks and record release handoff | repository and `artifacts/` evidence | FT-001 evidence and issue #3 dependency note | `CHK-01`–`CHK-06` | `EVID-01`–`EVID-05` | execute brief Verify; review evidence independently | `STEP-04` | human approval only for any release/publication action in issue #3 | any required check fails or an out-of-scope issue #2/#3 change becomes necessary |

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `SOL-01`, `DV-01` | filtered import is reviewed before merge | `EVID-05` |
| `CP-02` | `STEP-03`, `INV-01`–`INV-03` | old compatibility paths are gone and command contracts pass | `EVID-03`, `EVID-04` |
| `CP-03` | `STEP-05`, `EC-01`–`EC-05` | all executable checks pass; future release evidence is explicitly handed off | `EVID-01`–`EVID-05` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Source advances after snapshot | imported behavior is not the reviewed baseline | stop and update `ASM-01`, `SD-01` and test baseline before continuing | requested source SHA is unavailable or a newer revision is selected |
| `ER-02` | Old-name search treats payload paths as executable compatibility | accidental semantic change to CLI data handling | apply `CON-03` and `INV-03`; review every retained match | rename changes a `memory-bank/` payload path |
| `ER-03` | Release work leaks into this feature | scope and approval boundary breach | stop at `RB-02` and hand off to issue #3 | tag, publish, credentials or release CI are required |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `FM-02`, `ER-01` | import tree/provenance is wrong or source snapshot changes | do not merge; update facts or request a new decision | target branch before `STEP-01` |
| `STOP-02` | `FM-01`, `INV-01` | rename alters preserved CLI contract | revert rename changes and diagnose with regression tests | imported source before identity change |
| `STOP-03` | `NS-01`, `NS-02`, `ER-03` | issue #2 or #3 work is necessary to pass a local step | stop local step and create/record handoff | verified FT-001 branch without out-of-scope action |
