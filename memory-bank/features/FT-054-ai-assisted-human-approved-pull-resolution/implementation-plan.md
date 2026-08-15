---
title: "FT-054: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution plan for the accepted trusted-local resolution workflow."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_054_scope
  - ft_054_selected_design
  - ft_054_acceptance_criteria
  - ft_054_validation_profile
---

# FT-054: Implementation Plan

## Goal

Deliver the opt-in plan/apply workflow from `brief.md` and `design.md`, reusing
the existing ownership planner, source verifier and atomic transaction.

## Current State / Reference Points

| Area | Current fact | Implication |
| --- | --- | --- |
| `internal/cli/cli.go` | `pull` supports `--ask`, `--dry-run` and JSON. | Add mutually exclusive plan/apply modes without changing default behavior. |
| `internal/ownership/update.go` | Planner derives decisions and one transaction commits payload plus lock. | Extend planning/resolution; do not add a second writer. |
| `internal/ownership/lock.go` | Lock is strict v1 and records source/base identity. | Keep schema unchanged and use it for Git base verification. |
| `internal/cli/upstream.go` | Temporary checkout fetches reachable `main` history. | Resolve old lock source ref from the same verified Git database. |
| Existing tests | Cover dry-run, ask, source provenance and rollback. | Extend their hermetic fixture patterns. |

## Preconditions

- `PRE-01` Trusted-local and Git-backed-base decisions are accepted in
  `decision-log.md`.
- `PRE-02` Working branch starts from released `v2.0.1` with a clean tree.
- `PRE-03` No plan apply is exposed before stale/tamper and rollback tests pass.

## Steps

| Step | Implements | Touchpoints | Verify |
| --- | --- | --- | --- |
| `STEP-01` | Plan structs, strict bounded codec and read-only projection. | `internal/ownership/types.go`, new resolution-plan module/tests. | `CHK-01` |
| `STEP-02` | Historical source-ref/path/blob lookup and deterministic line merge. | ownership source/merge modules and Git fixtures. | `CHK-02` |
| `STEP-03` | Adapted/managed resolution actions, keep-and-detach and atomic final state. | planner/update transaction and regressions. | `CHK-02`–`CHK-04` |
| `STEP-04` | Public `--plan` / `--apply-plan` CLI, strict file IO and help. | `internal/cli/cli.go`, CLI tests. | `CHK-01`, `CHK-03` |
| `STEP-05` | Kirasa-shaped canonical migration and repeat-pull E2E. | hermetic CLI/ownership tests. | `CHK-04` |
| `STEP-06` | README, UC-002, changelog and evidence alignment. | docs and feature package. | `CHK-05` |
| `STEP-07` | Full validation, PR/main integration, semantic release and exact-tag install verification. | CI/release workflow. | Repository release gate |

## Test Strategy

- Unit: strict codec, normalized paths, historical path mapping, blob
  digest/mode verification, merge hunk overlap and mode resolution.
- Ownership: each selected action, unresolved/unavailable rejection,
  keep-and-detach, stale lock/local/source and injected transaction failure.
- CLI E2E: plan file creation, reviewed apply, malformed carrier, default pull
  compatibility and second-pull no-op.
- Kirasa-shaped regression: three legacy adapted files with non-overlapping
  upstream/local additions plus unrelated managed updates.
- Repository: `go test ./...`, `go vet ./...`, documentation checks and required
  GitHub Actions release validation.

## Checkpoints / Stop Conditions

| Checkpoint | Pass criterion |
| --- | --- |
| `CP-01` | Planning is complete and read-only. |
| `CP-02` | Merge is available only from a verified Git base and exact recomputation. |
| `CP-03` | Every rejection leaves payload and lock unchanged. |
| `CP-04` | Successful apply followed by ordinary pull is a no-op. |
| `CP-05` | Public docs and released binary agree. |

Stop and update `design.md` if implementation would require a lock schema
change, persistent sidecar, external authorization service, automatic semantic
decision or non-atomic payload/lock writer.
