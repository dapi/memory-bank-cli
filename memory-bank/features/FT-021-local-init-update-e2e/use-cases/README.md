---
title: "FT-021: Feature Use Cases"
doc_kind: feature-support
doc_function: reference
purpose: "Derived review projection of FT-021 scenario groups; canonical acceptance remains in brief.md."
derived_from:
  - ../brief.md
  - ../design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_021_scope
  - ft_021_acceptance_criteria
  - canonical_checks
  - implementation_sequence
---

# FT-021: Feature Use Cases

Canonical `SC-*`, `CHK-*` and `EVID-*` remain in [brief.md](../brief.md).

| ID | Scenario group | Issue scenarios | Primary refs |
| --- | --- | --- | --- |
| `FUC-H01` | Tagged init and clean update lifecycle | E2E-01–E2E-03 | `REQ-01`–`REQ-02`, `SC-01`, `CHK-01` |
| `FUC-ER01` | Repeat-init, conflicting local change, dry-run and invalid source | E2E-02, E2E-04, E2E-06, E2E-09 | `REQ-02`, `SC-02`, `INV-02` |
| `FUC-H02` | Downstream and template auto-profile discovery | E2E-07–E2E-08 | `REQ-02`, `SC-02` |
| `FUC-E01` | Local edit with unchanged upstream is preserved; equal new upstream content is accepted | E2E-11, E2E-13 | `REQ-03`, `SC-03` |
| `FUC-ER02` | Both-side edit, mode change, symlink and local deletion fail safely | E2E-12, E2E-14–E2E-16 | `REQ-03`, `SC-03`, `INV-02` |
| `FUC-H03` | Upstream deletion, both-side deletion and rename are reconciled safely | E2E-17, E2E-19, E2E-20 clean branch | `REQ-03`, `SC-03` |
| `FUC-ER03` | Local edit before upstream deletion/rename conflicts | E2E-18, E2E-20 conflict branch | `REQ-03`, `SC-03`, `INV-02` |
| `FUC-ER04` | New-path and file/directory collisions do not overwrite user state | E2E-21–E2E-22 | `REQ-03`, `SC-04`, `INV-02` |
| `FUC-H04` | Unmanaged unrelated files survive successful update | E2E-05, E2E-23 | `REQ-02`–`REQ-03`, `SC-04` |
| `FUC-ER05` | One conflict aborts all changes and is reproducible; manual upstream resolution then succeeds | E2E-24–E2E-27 | `REQ-03`, `SC-04`, `INV-02` |
| `FUC-H05` | Local required merge gate, pre-publish candidate gate and non-blocking canary | E2E-10 + CI/canary requirements | `REQ-04`–`REQ-06`, `SC-05`–`SC-07` |

## Derived Test Case Candidates

| Test Case ID | Covers | Preconditions | Expected result | Automation candidate |
| --- | --- | --- | --- | --- |
| `TC-01` | `FUC-H01`–`FUC-H04`, `CHK-01` | Built binary and fresh local bare remote. | Required successful outcomes and lock assertions. | automated |
| `TC-02` | `FUC-ER01`–`FUC-ER05`, `CHK-01` | Recorded downstream tree and lock before invocation. | Expected failure leaves both snapshots unchanged. | automated |
| `TC-03` | `FUC-H05`, `CHK-02`–`CHK-04` | CI/release workflow definitions and admin access for binding. | Gate/dependency/trigger semantics match canonical checks. | automated plus `AG-01` |
