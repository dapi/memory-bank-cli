---
title: "FT-006: Feature Review Report"
doc_kind: feature-support
doc_function: derived
purpose: "Records the bounded review-improve result and human gate for FT-006 without becoming a canonical owner."
derived_from:
  - brief.md
  - decision-log.md
  - ../../flows/feature.md
status: active
audience: humans_and_agents
must_not_define:
  - requirements
  - selected_solution
  - implementation_sequence
---

# FT-006: Feature Review Report

## Cycle 1

### Review summary

The bootstrap package faithfully maps all five explicit issue acceptance areas to `REQ-01`–`REQ-05` and keeps solution/execution artifacts absent, as Feature Flow requires before upstream owners are ready. The package is internally consistent: `README.md`, `brief.md` and this report all identify `Draft Feature`; no decision log entry overrides the brief.

### Findings

| Severity | Finding | Evidence | Required action |
| --- | --- | --- | --- |
| `critical` | The issue does not define the adapter's public contract or managed GitHub path set, so selected design, acceptance scenarios and non-destructive collision behaviour cannot be specified unambiguously. | Issue #6 names artifact categories and safety outcomes only; [brief.md](brief.md) `DEC-01`; [decision-log.md](decision-log.md) `DEC-01`. | Human decision on the adapter contract. |
| `important` | The repository has no applicable, documented validation-profile catalogue/minimum evidence contract for this feature; the issue's fixture requirement alone does not select one. | [engineering validation profiles](../../engineering/validation-profiles.md); [brief.md](brief.md) `DEC-02`; [decision-log.md](decision-log.md) `DEC-02`. | Human approval of a validation profile/evidence minimum. |

No `minor` findings were acted on.

### Open questions closed through FPF

None. FPF B.5 was applied to distinguish evidence-backed facts from untested solution hypotheses. It establishes that neither `DEC-01` nor `DEC-02` can be closed from the current documents without inventing a material design or verification decision. The resulting rationale is recorded in [decision-log.md](decision-log.md).

### Changes made

- Created the FT-006 bootstrap package: `README.md`, canonical `brief.md`, and `decision-log.md`.
- Recorded the issue-derived requirements, non-scope, material blockers and explicit absence of downstream artifacts.
- Saved this review report.

### Human gate

**Yes — stop auto-improvement here.**

| Question | Available facts | Options | Risk of a wrong choice | Needed from a human |
| --- | --- | --- | --- | --- |
| What is the explicit adapter contract? | Issue #6 requires opt-in installation/update, four artifact categories, no overwrite of user-owned templates and decision reporting. Existing CLI has `init`/`update` and dry-run, but no adapter command/interface. | (1) a separate GitHub adapter command; (2) an explicit adapter mode of an existing ownership flow; (3) another contract with named invocation, asset source, paths and collision semantics. | The wrong interface or path set can break compatibility, manage the wrong files or violate the non-destructive promise. | Approve one contract, including command/flags or discovery mechanism, adapter asset source, exact managed paths, and treatment of pre-existing paths. |
| What validation profile and evidence minimum applies? | Issue #6 requires fixtures for empty and custom `.github/` trees. Project validation document defines no matching profile catalogue. | (1) approve a named existing/project profile with its obligations; (2) define an FT-006 minimum (local/CI suites, fixture evidence and review/approval); (3) provide another governed validation policy. | Under-validation can ship destructive template handling; over/incorrect validation can create an ungoverned delivery gate. | Name or approve the profile and its minimum required checks/evidence, including CI expectation if any. |

This was the initial bootstrap finding. The source issue #40 was subsequently read and supplied the missing marker-ownership, form-field and closeout facts; cycle 2 supersedes this gate.

## Cycle 2

### Review summary

The package is now consistent with issue #6 and its migrated source issue #40. `brief.md` owns requirements, validation and acceptance; `design.md` owns the selected explicit command and marker contract; `implementation-plan.md` maps every design fact to code and checks. The implementation supplies the required opt-in boundary, five guidance artifacts, ownership-aware create/preserve/update/conflict decisions, dry-run, and representative empty/custom-template coverage.

### Findings

No `critical` or `important` findings remain. Minor: the adapter supplies validation configuration guidance rather than a GitHub Actions workflow, because no released/pinned CLI version is available in this feature and the project doctor explicitly warns against a moving `@latest` installation. The issue permits validation workflow/configuration; this omission is intentional and does not block the feature.

### Open questions closed through FPF

- `DEC-01`: closed as `mb-cli github init|update` with embedded marker-managed assets; recorded in [design.md](design.md) and [decision-log.md](decision-log.md).
- `DEC-02`: closed as `standard` validation; recorded in [brief.md](brief.md) and [decision-log.md](decision-log.md).

### Changes made

- Added `internal/githubadapter` with safe marker reconciliation, user-owned preservation, drift conflicts, dry-run and symlink rejection.
- Added the `mb-cli github init|update` CLI surface and JSON/text reports.
- Added Go tests for empty-tree install, custom template preservation, managed drift, dry-run, symlink rejection and CLI JSON.
- Added design and implementation-plan owners and reconciled the feature brief/decision log.

### Human gate

No. The migrated source issue explicitly supplies the facts needed to select and verify the feature-local design.
