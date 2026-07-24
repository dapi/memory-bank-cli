---
title: "FT-026: Ignore Project-Local Source Root Design"
doc_kind: feature
doc_function: canonical
purpose: "Selected source-root precedence and regression design for FT-026."
derived_from:
  - brief.md
  - ../../flows/feature.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_026_scope
  - ft_026_acceptance_criteria
  - implementation_sequence
---

# FT-026: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `ALT-*`, `C4-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` |

## Context

Git and filesystem discovery currently enumerate the same three roots and delegate to a cardinality-only selector. Target plus local copy is therefore rejected like two legacy roots. Issue #26 makes target authoritative when present and keeps legacy behavior only when absent.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-00` | `not required` | No deployable/API/storage/component topology changes; a local selector rule changes. | none |

## Architecture Coverage Decision

| Aspect | Status | Canonical owner / refs | Supporting view | Coverage note |
| --- | --- | --- | --- | --- |
| Components / responsibilities | covered | `SOL-01` | none | Both discovery paths feed one selector; init/update consume its result. |
| Connectors / interactions | covered | `CTR-01` | none | Selection precedes pinned verification and downstream translation. |
| Configuration / topology | covered | `SOL-01`, `SOL-02` | none | Recognized paths and target-present branch are explicit. |
| Behavioral semantics | covered | `INV-01`–`INV-02`, `FM-01`–`FM-02` | none | Target-present and target-absent outcomes are bounded. |
| Quality / evolution concerns | covered | `CON-01`–`CON-02`, `RB-01` | none | Pinned safety and legacy compatibility remain testable. |

## Selected Solution

- `SOL-01` Make `template/memory-bank` authoritative: when present, both Git-tree and filesystem discovery return it before evaluating legacy candidates. A project-local `memory-bank/.lock` is therefore outside the candidate set in this branch.
- `SOL-02` When target is absent, retain current bounded legacy selection over `memory-bank` and `memory-bank-template`: accept exactly one, reject neither or multiple.
- `SOL-03` Add selector and init/update regressions with distinguishable target/local contents and local `.lock`; assert only target content reaches downstream paths.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Reject any multiple recognized roots | Fails issue #26's dual-root acceptance criterion. |
| `ALT-02` | Ignore `memory-bank/` only if `.lock` exists while still comparing other legacy roots with target | Issue says target presence selects it and restricts legacy compatibility to target-absent checkouts; lock inspection is not needed for target precedence. |
| `ALT-03` | Rename installed downstream root | Contradicts `REQ-03` and exceeds source-selection scope. |

## Contracts

| Contract ID | Connector / direction | Roles and boundary | Guarantees / failure / evolution semantics |
| --- | --- | --- | --- |
| `CTR-01` | source discovery → selector → pinned verifier/planner | Discovery observes checkout paths; selector returns source path; planner translates it. | Target-present selects target; target-absent uses legacy rule; source bytes stay pinned and downstream namespace stays `memory-bank`. |

## Invariants

- `INV-01` Target presence yields it as the only selected source payload, regardless of project-local `memory-bank/`.
- `INV-02` Pinned-source validation and downstream namespace do not change.

## Failure Modes

- `FM-01` Precedence applies to only Git or filesystem discovery; dual-path regression fails.
- `FM-02` Change alters pinning or downstream namespace; existing validation/output assertions fail.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | Merge selector/test change after all checks pass. | Target and legacy regressions green. | Revert feature commit; no data migration is needed. |

## Design Verification

| Analysis | Required | Reason / risk | Method | Result / evidence |
| --- | --- | --- | --- | --- |
| Contract compatibility | yes | Source-root/downstream paths are CLI contract. | Target-present and target-absent matrices. | `CTR-01`, `CHK-01`–`CHK-02` |
| State / transition completeness | yes | Init and update must share semantics. | Dual-root init/update tests. | `SOL-03`, `CHK-02` |
| Failure propagation | yes | Legacy missing/multiple and unsafe source failures remain. | Existing matrix plus relevant suite. | `SOL-02`, `FM-*`, `CHK-01`, `CHK-03` |
| Concurrency / ordering | no | No writer/lock protocol change. | N/A | N/A |
| Security boundaries | no | Existing source validation remains unchanged. | N/A | `INV-02`, `CHK-03` |
| Capacity / latency | no | No performance target/change stated. | N/A | N/A |
| Migration / evolution safety | yes | Legacy behavior must be bounded to target-absent checkout. | Compatibility matrix. | `SOL-02`, `CHK-01` |

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure / rollout refs |
| --- | --- | --- | --- |
| `REQ-01`–`REQ-02` | `SOL-01` | `CTR-01`, `INV-01` | `FM-01`, `RB-01` |
| `REQ-03` | `SOL-03` | `CTR-01`, `INV-02` | `FM-01`–`FM-02`, `RB-01` |
| `REQ-04` | `SOL-02`–`SOL-03` | `CTR-01`, `INV-02` | `FM-01`, `RB-01` |
