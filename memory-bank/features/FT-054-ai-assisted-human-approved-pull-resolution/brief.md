---
title: "FT-054: AI-Assisted, Human-Approved Pull Resolution"
doc_kind: feature
doc_function: canonical
purpose: "Canonical problem, scope, validation profile and verify contract for reviewable pull resolution plans."
derived_from:
  - ../../flows/feature.md
  - ../../engineering/validation-profiles.md
  - ../../use-cases/UC-002-update-template.md
  - "https://github.com/dapi/memory-bank-cli/issues/54"
status: active
delivery_status: blocked_pending_human_approval
audience: humans_and_agents
must_not_define:
  - selected_solution
  - implementation_sequence
---

# FT-054: AI-Assisted, Human-Approved Pull Resolution

## What

### Problem

`pull` stops conservatively on ownership conflicts and `pull --ask` collects one-off terminal answers, but neither leaves an agent a durable, reviewable resolution plan. Projects with adapted Memory Bank documents therefore cannot prepare a proposed update without either an interactive terminal or an unsafe transfer of ownership choices to automation.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Pull conflict resolution | Conflicts stop `pull`; `--ask` resolves only interactively. | A non-mutating plan exposes every decision, and a reviewed complete plan applies atomically only when its recorded inputs still match. | Hermetic CLI/ownership E2E and plan-tamper/staleness regressions. |

### Scope

- `REQ-01` Provide a machine-readable, non-mutating `pull` plan containing path, ownership, base identity, local/upstream state, safe proposed action, human-decision requirement, and concise conflict context. All non-decision review context must be deterministically regenerated and compared by a trusted CLI before apply; only the selected action, its mechanically derived decision state, and separately verified human authorization are a decision overlay. A full-scope plan exposes every affected path; a declared bounded explicit scope may represent a selected subset under the selected design's exact coverage rules.
- `REQ-02` Define a versioned, reviewable resolution-plan file with independently verifiable, tamper-resistant human authorization covering the complete reviewed plan; it records accepted actions and the inputs and result needed to review any eligible non-overlapping mechanical merge. Mechanical merge is available only when the selected design can establish its required compatible, verifiable historical state; otherwise the plan reports that action unavailable while preserving existing non-merge behavior.
- `REQ-03` Apply a resolution plan only after validating its lock, source identity, local payload, template payload and lock-bound historical-state sidecar under a cooperating-writer serialization boundary; perform final revalidation and apply all accepted actions, `.lock` and that sidecar atomically without changing the existing `.lock` wire format.
- `REQ-04` Require an explicit human-selected currently selectable action for every two-sided `adapted` change: `keep-local` or `take-upstream`, plus `apply-reviewed-merge` only when its verification and resource preconditions are met; an unresolved path blocks apply.
- `REQ-05` Preserve deterministic safe actions, including managed no-drift update, user-owned missing-template `keep-and-detach`, one-sided adapted behavior, and an eligible reviewed non-overlapping mechanical merge.
- `REQ-06` Keep default `pull` conservative and backward-compatible; document the agent/human boundary and `keep-and-detach` policy.
- `REQ-07` Add hermetic E2E coverage for the issue acceptance matrix, including executable modes and all-or-nothing failure.

### Non-Scope

- `NS-01` Automatically infer a semantic merge or resolve a two-sided adapted conflict.
- `NS-02` Delete a user-owned file merely because it disappeared upstream.
- `NS-03` Send document contents to an external AI provider, create a GitHub PR, or replace Git review.

### Constraints / Assumptions

- `ASM-01` Issue #54 is the sole product request. Its existing comment is a routing/provenance record and adds no product requirements or attached implementation decision.
- `ASM-02` Current `pull --dry-run --json` already derives a non-mutating ownership report; `pull --ask` collects all terminal answers before mutation.
- `ASM-03` Current ownership planning validates an existing `.lock` and writes destination payload plus lock through an atomic transaction; lock entries record ownership, digests and modes.
- `CON-01` A plan must be bound to the reviewed source, lock and payload identities and to independently verifiable, tamper-resistant human authorization over the complete reviewed plan. A trusted CLI must regenerate and compare all non-decision review context, so stale, tampered or agent-misleading input fails before mutation.
- `CON-02` The CLI must remain deterministic and must never call an LLM or treat an AI proposal as approval.
- `CON-03` The public `pull` contract and the stable update use case must remain conservative when no plan is applied.

## Blocking Decisions

`BD-01` **Human approval required before Plan Ready or implementation.** Select and approve the human-authorization carrier and verification format, canonicalization procedure, authorized-reviewer trust policy/configuration, credential lifecycle and revocation model, and reviewer authorization workflow. Until then, FT-054 may describe the required authorization outcome but must not claim a deployable approval mechanism. Exact flag spelling and JSON field names remain implementation details once `BD-01` is resolved.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature changes a public CLI/file-format contract and adds a trust boundary, state validation and atomic rollback semantics. | `design.md` |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- | --- |
| `decision-log.md` | selected | Plan representation and approval boundary require fact-bounded FPF provenance. | `decision-log.md` |
| `design.md` | selected | CLI, plan-file, ownership and atomicity contracts require explicit solution reasoning. | `design.md` |
| `implementation-plan.md` | selected | Execution crosses CLI, ownership, lock, transaction, documentation and E2E surfaces. | `implementation-plan.md` |
| Separate contract / sequence diagram | omitted | The one CLI-to-local-file contract is reviewable in compact design tables. | none |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `standard` | Current project policy calls for Go test/vet and contract-focused regressions; the change is local, has no release/publication action, and its higher integrity risk is covered by targeted hermetic E2E. | none |

## Verify

### Exit Criteria

- `EC-01` Planning is read-only and reports enough context for an agent or reviewer without deciding ambiguous adapted paths.
- `EC-02` A human-authorized plan containing only deterministic actions applies atomically to its matching serialized pre-state and exactly covers either the full affected-path set or its declared authorized scope. Re-executing it after its apply changes a bound identity is rejected as stale; a no-op plan whose bound identities remain unchanged is idempotent. Executing it on separate identical pre-state fixtures produces the same result.
- `EC-03` Every two-sided adapted conflict remains unapplied until an explicit allowed human action is recorded; each selected action gives its specified content and mode result.
- `EC-04` Stale, malformed, tampered or misleading-context plans, or human-authorization material, fail before mutating any downstream file, `.lock` or lock-history sidecar; a competing writer that uses the update guard cannot interleave a change between final validation and commit. Non-cooperating writers are outside this mutual-exclusion guarantee and may cause a later stale-state rejection if their change is observed before commit.
- `EC-05` Default `pull` behavior and documentation preserve the conservative agent/human boundary.

### Traceability Matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01`–`REQ-02` | `ASM-01`–`ASM-02`, `CON-01`–`CON-02` | `EC-01`–`EC-02`, `SC-01`–`SC-02` | `CHK-01`, `CHK-02`, `CHK-04` | `EVID-01`, `EVID-02`, `EVID-04` |
| `REQ-03` | `ASM-03`, `CON-01` | `EC-02`, `EC-04`, `SC-02`, `SC-05` | `CHK-02`, `CHK-04` | `EVID-02`, `EVID-04` |
| `REQ-04`–`REQ-05` | `ASM-02`–`ASM-03`, `CON-01`–`CON-02` | `EC-03`, `SC-03`–`SC-04h` | `CHK-03` | `EVID-03` |
| `REQ-03`, `REQ-07` | `CON-01` | `EC-03`–`EC-04`, `SC-03`–`SC-05` | `CHK-03`, `CHK-04` | `EVID-03`, `EVID-04` |
| `REQ-06` | `CON-02`–`CON-03` | `EC-05`, `SC-06` | `CHK-05` | `EVID-05` |

### Acceptance Scenarios

- `SC-01` An agent reads a JSON pull plan for a changed template without filesystem or lock mutation and can identify each required human decision.
- `SC-02` A reviewer applies a complete human-authorized deterministic plan against unchanged serialized inputs and receives its specified file and lock updates as one committed result; if that commit changes a bound identity, replaying the plan is rejected as stale, while a no-op plan whose bound identities remain unchanged is idempotent. Separate identical pre-state fixtures produce the same committed result.
- `SC-03` A plan with an unresolved path, or whose actions do not exactly cover the full affected-path set for implicit scope or the declared selected affected paths for explicit scope, is rejected without mutation.
- `SC-04` After a human records each allowed action, apply produces the recorded local, upstream, or reviewed-merge content and executable mode.
- `SC-04a` For an `adapted` path with only local drift, apply keeps the local content and mode and retains `adapted` ownership.
- `SC-04b` For an `adapted` path with only upstream drift, apply takes the upstream content and mode and retains `adapted` ownership.
- `SC-04b1` For an `adapted` path deleted locally while upstream is unchanged, apply deterministically keeps the local deletion and retains the current present upstream base and its `SOL-05` historical-base state; when upstream alone is deleted, it deterministically deletes the local path and retains an absent upstream-base tombstone with no bytes or mode.
- `SC-04b2` For a two-sided `adapted` path where exactly one side is absent, a human-authorized `keep-local` or `take-upstream` produces respectively the local-side or upstream-side outcome, including deletion when the chosen side is absent; ownership remains `adapted` and the resulting upstream base is a present verified snapshot or unavailable historical-base state according to `SOL-05`, or an absent tombstone. `apply-reviewed-merge` is unavailable. When both sides are absent, either human-authorized non-merge action deletes the path and establishes the absent tombstone.
- `SC-04b3` On a subsequent pull with an `adapted` path's verified absent tombstone as its recorded upstream base, a local-present/upstream-absent state deterministically preserves the local bytes and executable mode and retains the tombstone; a local-absent/upstream-present state deterministically writes the upstream bytes and executable mode and replaces the tombstone with the resulting present verified snapshot or unavailable historical-base state according to `SOL-05`. Ownership remains `adapted` in both cases.
- `SC-04c` For a managed path with no local drift, apply retains `managed` ownership and atomically follows the upstream state: when upstream is present and changed, it writes the recorded upstream bytes and executable mode and records the resulting upstream identity and mode; when upstream is absent, it deletes the local path and records a verified absent upstream-base tombstone with no bytes or mode.
- `SC-04d` For a user-owned path missing from the template, apply `keep-and-detach` without changing its bytes or executable mode, removes that path's ownership entry from `.lock` and its lock-history sidecar record, and commits the unchanged payload with both state updates atomically.
- `SC-04e` A reviewed mechanical merge is available only when the selected design's historical-state verification preconditions are satisfied. Otherwise planning marks `apply-reviewed-merge` unavailable and apply rejects that action without mutation; a later clean baseline/update may establish eligibility.
- `SC-04f` At each selected-design resource bound, planning or apply either records mechanical merge unavailable or rejects without mutation as applicable; no bound permits truncation, and permitted non-merge actions remain available.
- `SC-04g` When a full plan cannot fit the selected-design plan limits, planning may produce an explicit normalized, sorted, unique scope selector for a requested subset of affected paths. Apply recomputes the full affected set, requires exactly one action for each selected affected path and none elsewhere, and leaves every unselected affected path and lock entry unchanged.
- `SC-04h` An oversized serialized plan carrier is rejected before decode, and carrier, action-count, nesting, path, identity and metadata boundary cases are rejected without mutation. Decoded byte-field accounting is distinct from encoded carrier accounting. Multi-path snapshot replacement at and just beyond the aggregate limit produces the same retained/unavailable entries on identical fixtures according to normalized path order.
- `SC-05` A changed lock, source ref, local payload, template payload, lock-history sidecar, malformed/tampered/misleading-context plan or human-authorization material, competing cooperating-writer attempt, or injected write failure leaves all downstream files, `.lock` and the sidecar unchanged or rejects the stale plan before commit. A non-cooperating writer is outside the serialization guarantee; if its mutation is observed by final validation, apply rejects before commit.
- `SC-06` A user following help and README understands that an AI may prepare a proposal but cannot approve it; ordinary `pull` stays conservative.

### Negative Coverage

- `NEG-01` Reject an unknown schema/action, a plan with unrecognized fields, or a path outside the payload boundary.
- `NEG-02` Reject a purported reviewed merge whose encoded result does not match the bytes and mode recomputed under the selected design's merge contract, including its content-collision and mode-conflict rules.
- `NEG-03` Reject a plan whose complete reviewed content is not covered by valid human-authorization material under the `BD-01`-approved trust policy, or whose regenerated non-decision review context differs from the authorized plan.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Run CLI and ownership plan serialization/read-only tests. | Planning has no mutation and exposes required identities/context. | `artifacts/ft-054/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-02` | Run hermetic deterministic plan/apply E2E on separate identical pre-state fixtures, then replay a state-changing plan on the changed fixture and a no-op plan on its unchanged fixture. | Complete human-authorized safe plan applies atomically and identically on fresh fixtures; replay of a state-changing plan is stale and rejected, while a no-op plan is idempotent. | `artifacts/ft-054/verify/chk-02/` |
| `CHK-03` | `EC-03`, `SC-03`–`SC-04h`, `NEG-02` | Run full and explicit-scope affected-path completeness; present- and tombstone-based one-sided and two-sided adapted actions; managed no-drift update including upstream deletion; user-owned missing-template detach; historical-state verification; legacy-ineligible-merge rejection; selected-design merge-contract E2E; resource-boundary E2E; and downgrade compatibility after creating historical state. Cover the selected algorithm's tie-breaking, collisions, line preservation, binary rejection, mode conflicts and mode preservation, plus deterministic multi-path aggregate snapshot allocation and every plan-field, aggregate decoded-plan, merge-input and aggregate mechanical-merge LCS-work limit at and just beyond its boundary. | Omitted or unresolved selected affected paths fail; scoped apply leaves unselected paths unchanged; each safe action produces its specified content, mode, ownership and lock state; tombstone-baseline one-sided changes and managed upstream deletion produce their specified deterministic payload and lock outcomes; an older binary can read the unchanged `.lock` and perform ordinary pull after new historical state is written; only an explicit allowed two-sided action meeting the selected design's verification preconditions yields its exact reviewed result; ineligible, binary, colliding or mode-conflicting inputs cannot mechanically merge; each resource-bound outcome is safe, untruncated and preserves allowed non-merge behavior. | `artifacts/ft-054/verify/chk-03/` |
| `CHK-04` | `EC-04`, `SC-04h`, `SC-05`, `NEG-01`, `NEG-03` | Run pre-decode carrier-size, structural/metadata/path limit, stale/tamper/misleading-context/human-authorization, cooperating-writer and injected-failure transaction regressions. | Invalid input is rejected without unbounded decode or mutation; a changed ownership/state label, proposed action, allowed-action set, reason/context or other non-decision field is rejected when it differs from trusted regeneration; a cooperating writer cannot interleave after final validation; interrupted apply leaves payload, lock and sidecar unchanged. | `artifacts/ft-054/verify/chk-04/` |
| `CHK-05` | `EC-05`, `SC-06` | Review help/README and run relevant Go test and vet suites. | Boundary and `keep-and-detach` policy agree; required suites pass. | `artifacts/ft-054/verify/chk-05/` |

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Read-only plan test output and sample JSON | Go test runner | `artifacts/ft-054/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Deterministic apply E2E output | Go test runner | `artifacts/ft-054/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Affected-path completeness, adapted and deterministic-safe-action content/mode/ownership/lock, historical-state, merge-contract and resource-boundary E2E output | Go test runner | `artifacts/ft-054/verify/chk-03/` | `CHK-03` |
| `EVID-04` | Staleness/tamper and transaction-failure output | Go test runner | `artifacts/ft-054/verify/chk-04/` | `CHK-04` |
| `EVID-05` | Documentation audit plus applicable Go test/vet output | Implementer / Go toolchain | `artifacts/ft-054/verify/chk-05/` | `CHK-05` |
