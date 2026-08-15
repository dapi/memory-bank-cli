---
title: "FT-054: AI-Assisted, Human-Approved Pull Resolution"
doc_kind: feature
doc_function: canonical
purpose: "Canonical scope and verification contract for reviewable pull resolution plans."
derived_from:
  - ../../flows/feature.md
  - https://github.com/dapi/memory-bank-cli/issues/54
status: active
delivery_status: in_progress
audience: humans_and_agents
---

# FT-054: AI-Assisted, Human-Approved Pull Resolution

## What

### Problem

`pull` stops conservatively when both an upstream template and a downstream
`adapted` document changed. The stop is safe, but it also prevents unrelated
managed updates from applying and gives an agent no durable way to prepare a
reviewable merge. Re-running `pull` repeats the same unapplied plan.

### Outcome

`pull` can emit a complete, non-mutating resolution plan. An agent may prepare
candidate decisions and mechanical merge results, a human reviews the plan,
and the CLI applies the reviewed result only while every recorded input still
matches. The default `pull` remains conservative.

### Scope

- `REQ-01` Add a versioned JSON resolution plan containing the target source,
  lock identity and, for every affected path, ownership, base/local/upstream
  identity, deterministic proposal, reason, allowed actions and whether a human
  decision is required.
- `REQ-02` Let a reviewer explicitly select `keep-local`, `take-upstream`, or
  `apply-reviewed-merge` for a two-sided adapted conflict. An unresolved or
  unavailable action blocks apply.
- `REQ-03` Recover a candidate historical base from the immutable source ref
  already stored in `.lock`, verify its bytes and executable mode against the
  lock entry, and offer a deterministic non-overlapping three-way merge only
  when that verification succeeds.
- `REQ-04` Recompute every deterministic plan field at apply time and reject a
  stale, malformed or altered plan before mutation.
- `REQ-05` Apply the selected resolutions, all deterministic safe changes and
  the new `.lock` through the existing atomic ownership transaction.
- `REQ-06` Keep ordinary `pull` and `pull --ask` backward compatible; the CLI
  does not call an LLM or infer a semantic document decision.
- `REQ-07` Keep user-owned content that disappeared upstream and detach its
  obsolete lock entry during reviewed full-plan apply.
- `REQ-08` Document the trusted-local review model, commands, plan editing
  boundary and repeat-pull no-op outcome.

### Non-Scope

- `NS-01` Automatically select a semantic outcome for a two-sided conflict.
- `NS-02` Prove cryptographically that a biological human, rather than a local
  process, edited the plan.
- `NS-03` Add a protected registry, compatibility sidecar or separate
  credential lifecycle.
- `NS-04` Send document contents to an external provider, create a GitHub PR or
  replace Git review.
- `NS-05` Support partial/bounded plan application in the first public format.

### Constraints / Assumptions

- `ASM-01` `.lock.template.source_ref` is an immutable Git commit and each
  adapted entry records the expected historical base digest and mode.
- `ASM-02` The resolved upstream checkout contains the reachable history of
  `main`; a missing historical commit or blob makes mechanical merge unavailable
  rather than guessed.
- `ASM-03` The current ownership transaction already provides rollback-safe
  payload-plus-lock mutation.
- `CON-01` Plan review is a trusted-local workflow. The CLI proves state and
  result integrity by strict decoding, deterministic regeneration and exact
  comparison; human authorization is an operational review act, not a new
  cryptographic identity boundary.
- `CON-02` The exact reviewed merge bytes and mode are included in the plan and
  recomputed before apply.
- `CON-03` Plan generation and every rejected apply are read-only.

## Blocking Decisions

None. On 2026-08-15 the user approved the trusted-local authorization model and
Git-backed historical-base recovery described in `design.md`. This resolves the
former `BD-01` and `BD-02` without a sidecar or protected registry.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature changes a public CLI/file-format contract and adds deterministic merge and stale-plan validation. | `design.md` |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- | --- |
| `decision-log.md` | selected | Records the trusted-local and Git-backed-base decisions. | `decision-log.md` |
| `design.md` | selected | Owns plan schema, merge and transaction semantics. | `design.md` |
| `implementation-plan.md` | selected | The accepted design crosses CLI, ownership, Git and transaction code. | `implementation-plan.md` |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `standard` | Public CLI/file contract with Go unit, hermetic E2E, vet and release-CI validation; no production data mutation. | none |

## Verify

### Exit Criteria

- `EC-01` Planning makes no downstream or lock mutation and exposes every
  affected path and required human decision.
- `EC-02` A complete reviewed plan applies atomically only against its exact
  recorded source, lock and local state.
- `EC-03` Every two-sided adapted conflict remains unresolved until an explicit
  currently allowed action is selected.
- `EC-04` A verified non-overlapping merge writes exactly the reviewed bytes and
  mode; unavailable or overlapping history cannot select merge.
- `EC-05` Stale, malformed or altered plans and injected transaction failures
  leave payload and lock unchanged.
- `EC-06` A successful full-plan apply advances `.lock`; the next ordinary
  `pull` reports no changes.
- `EC-07` Help, README and the stable update use case describe the delivered
  agent/human boundary.

### Acceptance Scenarios

- `SC-01` `pull --plan FILE` writes a deterministic JSON plan without changing
  downstream files or `.lock`.
- `SC-02` A plan with any unresolved required decision is rejected before
  mutation.
- `SC-03` `keep-local` retains local bytes/mode and records the new upstream as
  the adapted base, preventing the same conflict on the next pull.
- `SC-04` `take-upstream` writes upstream bytes/mode and adopts canonical
  managed ownership during legacy-to-canonical migration.
- `SC-05` `apply-reviewed-merge` is offered only when the old source blob
  matches `.lock`; its exact deterministic result is embedded, reviewed and
  recomputed at apply.
- `SC-06` If historical Git data is missing, mismatched, non-textual,
  overlapping or mode-ambiguous, planning keeps non-merge choices available and
  marks reviewed merge unavailable.
- `SC-07` Managed updates and selected conflict resolutions commit together.
- `SC-08` A changed source, `.lock`, local file, upstream file or reviewed merge
  causes stale/tamper rejection with no partial mutation.
- `SC-09` A user-owned file removed upstream stays untouched and its obsolete
  lock entry is detached in the same successful reviewed transaction.
- `SC-10` Kirasa's three current canonical-migration conflicts produce clean
  merge candidates, apply with the managed updates, and leave a second pull at
  no-op.

### Negative Coverage

- `NEG-01` Reject unknown plan fields, versions, actions, duplicate paths and
  non-normalized or reserved paths.
- `NEG-02` Reject a selected merge whose encoded bytes, digest, mode, algorithm
  or recomputed result differs.
- `NEG-03` Reject plan application when agent-instruction maintenance would add
  an unreviewed mutation outside the plan.
- `NEG-04` Never use plan-supplied base bytes or an unverified digest-only base.

### Checks

| Check ID | Covers | How to check | Expected result |
| --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01`, `NEG-01` | Plan codec and CLI read-only tests. | Complete strict plan, no mutation. |
| `CHK-02` | `EC-02`–`EC-04`, `SC-02`–`SC-07`, `NEG-02`, `NEG-04` | Ownership and hermetic Git-history E2E tests. | Exact allowed outcomes and merge eligibility. |
| `CHK-03` | `EC-05`, `SC-08`, `NEG-03` | Stale/tamper/race/injected-failure regressions. | Rejection before mutation or complete rollback. |
| `CHK-04` | `EC-06`, `SC-09`–`SC-10` | Full-plan apply followed by ordinary pull. | First apply succeeds; second pull is no-op. |
| `CHK-05` | `EC-07` | README/help/use-case review, `go test ./...`, `go vet ./...`. | Public contract and implementation agree. |

### Evidence Contract

| Evidence ID | Artifact | Producer | Reused by checks |
| --- | --- | --- | --- |
| `EVID-01` | Plan codec/read-only test output | Go test runner | `CHK-01` |
| `EVID-02` | Historical-base and resolution-action test output | Go test runner | `CHK-02` |
| `EVID-03` | Stale/tamper/rollback test output | Go test runner | `CHK-03` |
| `EVID-04` | Repeat-pull and Kirasa-shaped regression output | Go test runner | `CHK-04` |
| `EVID-05` | Documentation audit and repository validation output | Implementer / CI | `CHK-05` |
