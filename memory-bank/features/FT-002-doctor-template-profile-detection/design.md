---
title: "FT-002: Design"
doc_kind: feature
doc_function: canonical
purpose: "Feature-local solution for marker-based `doctor --profile auto` detection."
derived_from:
  - brief.md
  - ../../engineering/architecture.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_002_scope
  - ft_002_acceptance_criteria
  - ft_002_evidence_contract
  - implementation_sequence
---

# FT-002: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `ALT-*`, `TRD-*`, `C4-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` |

## Context

`internal/doctor` is a local, read-only filesystem diagnostic. `brief.md` `DEC-01` fixes a cross-repository source-identity contract while issue `dapi/memory-bank#52` owns the marker's creation in the source repository.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-00` | not required | No runtime node, external service, trust boundary, or topology changes; this is a local function and root-file contract. | none |

## Architecture Coverage Decision

| Aspect | Status | Canonical owner / refs | Supporting view / artifact | Reason if N/A / coverage note |
| --- | --- | --- | --- | --- |
| Components / responsibilities | covered | `SOL-01`, `SD-01` | none | `internal/doctor` detects; template repo owns marker presence. |
| Connectors / interactions | covered | `CTR-01` | none | Root filesystem read from CLI to repository marker. |
| Configuration / topology | covered | `SOL-01`, `CTR-01` | none | Absolute repo root binds to fixed relative marker path. |
| Behavioral semantics | covered | `SOL-02`, `INV-01`, `FM-01` | none | Lock precedence, exact marker validation, and downstream fallback are explicit. |
| Quality / evolution concerns | covered | `TRD-01`, `RB-01` | none | Contract is minimal, versioned, and safely falls back to existing downstream behavior. |

## Selected Solution

- `SOL-01` Replace the `tools/go.mod` heuristic with a regular-file read at repository-root `.memory-bank-template`; classify as template only if it contains the single UTF-8 line `memory-bank-template-v1`, terminated by LF or CRLF.
- `SOL-02` Preserve precedence: a present `memory-bank/.lock` classifies downstream before marker evaluation; absent or invalid marker classifies downstream.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Continue matching `tools/go.mod` | Issue #2 explicitly removes that dependency. |
| `ALT-02` | Parse structured marker metadata | The classification is Boolean; no required metadata justifies an additional parser/schema contract. |
| `ALT-03` | Use a file inside `memory-bank/` | Issue #2 explicitly requires the marker outside the copied payload. |

## Trade-offs

| Trade-off ID | Decision | Benefit | Cost / Risk |
| --- | --- | --- | --- |
| `TRD-01` | Exact fixed logical line with `v1`, accepting LF or CRLF, rather than flexible parsing | Minimal surface, deterministic lookalike resistance, portable Git checkouts, and an explicit future evolution seam | Any future format must intentionally add a compatible version rule. |

## Accepted Local Decisions

- `SD-01` Keep the marker constant and detector private to `internal/doctor`; no public CLI flag or report schema changes are needed.

## Contracts

| Contract ID | Connector / direction | Roles and sync boundary | Guarantees / failure / evolution semantics |
| --- | --- | --- | --- |
| `CTR-01` | filesystem read: `memory-bank-cli doctor` -> `<repo-root>/.memory-bank-template` | CLI initiator, repository file provider; synchronous local read | Regular root file only; the single line `memory-bank-template-v1` with LF or CRLF termination means template. Missing, unreadable, non-regular, or different content means downstream. `memory-bank/.lock` has precedence. Future formats require a new accepted version rule. |

## Invariants

- `INV-01` Explicit `template` and `downstream` profiles bypass auto detection unchanged.
- `INV-02` A similarly named file, or a marker path with non-exact content, never creates template classification.

## Failure Modes

- `FM-01` If the marker is absent or malformed, auto profile reports downstream and retains the existing missing-lock diagnostic; it must not silently treat the repository as template.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | Land CLI detector and fixture contract; template repository adds marker through issue #52 before removing `tools/` | focused and full Go regression suites pass; marker contract is documented | Restore the previous CLI release if regression is found; do not remove `tools/` until a released CLI recognizes the marker. |

## Design Verification

| Analysis | Required | Reason / risk | Method | Result / evidence |
| --- | --- | --- | --- |
| Contract compatibility | yes | Marker contract spans CLI and template repository | fixture review against `CTR-01` | template/no-Go-module and downstream/lookalike tests in `CHK-01`–`CHK-03` |
| State / transition completeness | no | Classification is a precedence-ordered Boolean decision, not a lifecycle | scenario walk-through | `SOL-02` and four `SC-*` cover branches |
| Failure propagation | yes | Wrong classification changes missing-lock diagnostic | malformed/missing marker fixture tests | `FM-01`, `CHK-02` |
| Concurrency / ordering | no | One synchronous read; no shared mutable state | design review | N/A |
| Security boundaries | no | No auth, secret, or trust boundary changes | design review | N/A |
| Capacity / latency | no | One small local root-file read on an existing command | design review | N/A |
| Migration / evolution safety | yes | Template removes the old Go module only after released CLI support | rollout review and issue #52 dependency | `RB-01`, `CHK-04` |

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure / rollout refs |
| --- | --- | --- | --- |
| `REQ-01` | `SOL-01`, `TRD-01`, `C4-00`, `SD-01` | `CTR-01`, `INV-02` | `FM-01`, `RB-01` |
| `REQ-02` | `SOL-01`, `SOL-02` | `CTR-01`, `INV-01` | `FM-01`, `RB-01` |
| `REQ-03` | `SOL-02` | `INV-02` | `FM-01` |
| `REQ-04` | `SOL-02` | `INV-01` | none |
| `REQ-05` | `SOL-01` | `CTR-01` | `RB-01` |
