---
title: "FT-021: Decision Log"
doc_kind: feature-support
doc_function: derived
purpose: "Audit ledger for FPF-backed FT-021 decisions; canonical facts remain in brief.md and design.md."
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

# FT-021: Decision Log

## Ownership

`brief.md` owns scope and verification. `design.md` owns selected solution. This log records the evidence and FPF reasoning; if it differs from an owner, update the owner first.

| ID | Status | Record | Facts considered | FPF reasoning | Canonical owner |
| --- | --- | --- | --- | --- | --- |
| `DEC-01` | accepted | Use a built-current-checkout binary plus fresh local bare/source/downstream Git repositories and two tagged template versions for the required local suite. | Issue requires a built binary, clean temporary Git repositories, local `template.git`, real URLs/tags/SHAs and no network. Existing smoke script depends on public GitHub/Go/release APIs. | FPF bounded-context separation distinguishes hermetic acceptance evidence from the external consumer canary. Of the candidates (reuse smoke, internal tests, local Git fixture), only local Git fixture satisfies every stated observable constraint without inventing a network exception. | [design.md](design.md) `SOL-01`, `CTR-01` |
| `DEC-02` | accepted | Isolate and name every mandatory scenario; assert pre/post tree and `.lock` equality for expected failures. | Issue lists E2E-01–E2E-09 and comment makes E2E-11–E2E-27 mandatory; it requires atomic conflicts and no source-error residue. | FPF B.5 reasoning cycle: a safety claim is credible only if its predicted unchanged state is observed. Scenario isolation prevents one result from becoming hidden setup for another. | [design.md](design.md) `SOL-02`, `INV-02`, `FM-02` |
| `DEC-03` | accepted | Treat E2E-04 and E2E-11 as distinct conditions: both-side change conflicts, while a local change with unchanged upstream is preserved. | E2E-04 says update must fully refuse a local managed-file change in the base update case; the later mandatory E2E-11 explicitly narrows a different condition: upstream did not change that file. E2E-12 separately requires conflict when both changed. | FPF strict distinction prevents collapsing different state transitions into one rule. The later scenario set supplies the discriminating condition; no behavior is inferred beyond those named cases. | [brief.md](brief.md) `SC-02`–`SC-03`; [use-cases/README.md](use-cases/README.md) `FUC-E01`, `FUC-ER02` |
| `DEC-04` | accepted | Validate the release candidate binary in `validate` immediately after snapshot build; the separate `release` job depends on `validate`. Retain canary as schedule/manual non-blocking. | Issue requires E2E-10 before publication and separately asks a non-blocking real-template canary. Current release workflow builds snapshot in `validate`, then starts `release` only after `validate`; existing canary has schedule/manual triggers. | FPF boundary/ordering analysis selects the actual snapshot-to-publish seam: it validates the deliverable without inventing a published prerelease. Separating lanes preserves the issue's different trust and availability assumptions. | [design.md](design.md) `SOL-04`, `INV-03` |
| `DEC-05` | accepted | Make the local job required by repository protection or a ruleset after its workflow job name is stable; track the external administrator action explicitly. | Issue requires a required check. Read-only repository inspection found `main` unprotected and no rulesets; workflow YAML alone cannot set repository merge policy. | FPF role-method distinction: CI produces a status, while repository governance consumes it as a merge gate. Branch protection is selected as the current-state-compatible default; an equivalent ruleset is acceptable only if it requires the same stable job. | [design.md](design.md) `SOL-03`, `FM-03`; [implementation-plan.md](implementation-plan.md) `AG-01` |
