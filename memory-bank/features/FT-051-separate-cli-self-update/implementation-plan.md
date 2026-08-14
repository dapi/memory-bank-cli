---
title: "FT-051: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Grounded execution plan for FT-051 without redefining canonical problem or solution facts."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_051_scope
  - ft_051_selected_design
  - ft_051_acceptance_criteria
---

# FT-051: Implementation Plan

## Grounding / Current State

| Path | Role | Implication |
| --- | --- | --- |
| `internal/cli/cli.go` | top-level dispatch and current ownership `update` command | route ownership behavior through `pull` and self-update through `update` |
| `.goreleaser.yml` | existing raw platform assets and `checksums.txt` | use asset names without changing release configuration |
| `cmd/memory-bank-cli/main.go` | version output | staged asset must report its release tag |
| `scripts/e2e-init-update.sh`, `scripts/downstream-smoke.sh` | public synchronization command regression | migrate invocations to `pull` |

## Test Strategy

| Surface | Refs | Automated coverage | Required suites |
| --- | --- | --- | --- |
| pull command | `REQ-01`, `CTR-01` | existing CLI and E2E synchronization tests with renamed invocation | `go test ./...`; `scripts/e2e-init-update.sh` |
| self-update | `REQ-02`, `CTR-02`, `INV-*`, `FM-*` | local HTTP release fixture: current, verified update, checksum/rename/platform failures | `go test ./internal/selfupdate` |
| user docs | `REQ-03`–`REQ-04` | lint plus review of help/README | `memory-bank-cli lint --scope-root memory-bank` |

## Preconditions

| ID | Ref | State |
| --- | --- | --- |
| `PRE-01` | `SOL-02` | GoReleaser continues to publish matching raw binaries and `checksums.txt`. |

## Steps

| ID | Implements | Work | Verifies |
| --- | --- | --- | --- |
| `STEP-01` | `SOL-01`, `CTR-01` | Rename CLI synchronization route, user messages, scripts and regressions to `pull`. | `CHK-01`, `EVID-01` |
| `STEP-02` | `SOL-02`–`SOL-03`, `CTR-02`, `INV-*`, `FM-*` | Add deterministic release lookup, verification, staging and replacement service plus tests. | `CHK-02`, `EVID-02` |
| `STEP-03` | `REQ-03`–`REQ-05`, `RB-01` | Update user-facing docs, lint, run relevant tests and preserve evidence. | `CHK-03`–`CHK-04`, `EVID-03`–`EVID-04` |

## Stop Conditions

| ID | Trigger | Action |
| --- | --- | --- |
| `STOP-01` | release assets/checksum manifest do not match the selected contract | stop; update `design.md` before changing release behavior |
| `STOP-02` | staged verification or rename fails | retain current executable and report the failure; do not merge a partial updater |
