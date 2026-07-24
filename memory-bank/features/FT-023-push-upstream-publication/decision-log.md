---
title: "FT-023: Decision Log"
doc_kind: feature-support
doc_function: derived
purpose: "Audit ledger for FT-023 provenance and FPF reasoning. It links to canonical owners and does not define requirements, selected solution or implementation sequence."
derived_from:
  - brief.md
  - design.md
  - ../../flows/feature.md
status: active
audience: humans_and_agents
must_not_define:
  - requirements
  - selected_solution
  - implementation_sequence
---

# FT-023: Decision Log

## Ownership

`brief.md` owns problem-space facts, validation decision and verify. `design.md` owns accepted feature-local solution facts. This ledger records the evidence and FPF reasoning only; if it conflicts with a canonical owner, update that owner first and then this log.

## Decisions and Open Questions

| ID | Status | Record | Facts considered | FPF reasoning | Canonical owner |
| --- | --- | --- | --- | --- | --- |
| `DEC-01` | accepted — `SD-01` | Include only changed paths whose current ownership class is exactly `managed`; report every other path as excluded with its class. A path that cannot be safely normalized below `memory-bank/`, or whose class cannot be obtained, is an ambiguity and stops before mutation. | Issue #23 requires upstream-suitable changes only and exclusions. `internal/ownership/classify.go` makes `managed` the explicit template boundary (`dna/`, `flows/`, `prompts/`) and marks project-facing/unknown content as adapted or user-owned. | FPF B.5 abduction: managed-only is the smallest new policy following the only explicit template boundary. Deduction: it cannot publish adapted, user-owned, generated, `.lock` or `.repo` content; it may omit desired local content, which is safer and visible. Induction: `CHK-03` must prove fixtures and ambiguity stop behavior. | [design.md](design.md) `SD-01`, `SOL-01`, `CTR-01` |
| `DEC-02` | accepted — `SD-02` | Use a compensating transaction: preflight first; mutate only a new upstream branch; after a later failure restore original local branch/HEAD and attempt to delete only the new remote branch. If compensation fails, return failure with exact residual state and recovery command; never alter the default branch. | Issue #23 requires branch → commit → push → PR and no partial application. Local Git, remote Git and GitHub are distinct external systems; no existing atomic connector defines their joint commit. | FPF A.7 separates protocol from actual work and A.10 forbids claiming compensation before it occurs. FPF B.5 selects the smallest reversible protocol: bounded new branch, deterministic local restoration and best-effort remote compensation. `CHK-04` must inject failures and evidence recovery. | [design.md](design.md) `SD-02`, `SOL-02`, `CTR-02`, `INV-02`, `FM-02`, `RB-01` |
| `DEC-03` | accepted — validation profile | Select `standard`: targeted unit/integration tests, full `go test ./...`, `go vet ./...`, navigation audit and approved live PR evidence. | The validation document records ordinary Go test/vet checks; issue #23 requires success, dry-run and key failure tests. The feature changes a CLI/integration contract but does not publish a release or deploy a service. | FPF B.5: this is a bounded evidence hypothesis. Deduction maps behavior to `CHK-01`–`CHK-05`; induction requires concrete evidence and approved live PR carrier before closure. | [brief.md](brief.md) Validation Profile; future `implementation-plan.md` when execution is authorized |

## FPF Closure Record

The user directed FPF-based selection after the initial gate. Each accepted decision remains an abductive design claim until its stated deductive predictions and `CHK-*` evidence are completed; it does not claim implementation or runtime success.
