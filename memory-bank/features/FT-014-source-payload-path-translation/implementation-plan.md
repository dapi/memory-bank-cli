---
title: "FT-014: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Grounded execution plan for FT-014 source/downstream translation, regressions, documentation and approved release handoff."
derived_from:
  - brief.md
  - design.md
source_refs:
  - ../../../internal/ownership/source.go
  - ../../../internal/ownership/source_test.go
  - ../../../internal/ownership/update.go
  - ../../../internal/ownership/update_test.go
  - ../../../internal/ownership/transaction_test.go
  - ../../../internal/ownership/classify.go
  - ../../../internal/ownership/lock.go
  - ../../../internal/ownership/types.go
  - ../../../internal/doctor/doctor.go
  - ../../../internal/doctor/doctor_test.go
  - ../../../internal/doctor/testdata/template/AGENTS.md
  - ../../../internal/doctor/testdata/template/memory-bank/README.md
  - ../../../internal/doctor/testdata/downstream/AGENTS.md
  - ../../../internal/doctor/testdata/downstream/memory-bank/README.md
  - ../../../internal/cli/cli.go
  - ../../../internal/cli/cli_test.go
  - ../../../internal/lint/audit.go
  - ../../../internal/agentinstructions/block.go
  - ../../../internal/agentinstructions/block_test.go
  - ../../../.github/workflows/release.yml
status: active
audience: humans_and_agents
must_not_define:
  - ft_014_scope
  - ft_014_selected_design
  - ft_014_acceptance_criteria
  - ft_014_blocker_state
  - ft_014_validation_profile
---

# FT-014: Implementation Plan

## Цель текущего плана

Realize the accepted one-way source/downstream path bridge, prove compatibility and safety across ownership and diagnostics, update affected stable use cases, then publish an approved release that can unblock `dapi/memory-bank#63`.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `brief.md` | canonical problem/validation/verify owner | `REQ-*`, `SC-*`, `NEG-*`, `CHK-*`, `EVID-*` | update `brief.md` first |
| `design.md` | canonical solution owner | `SOL-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` | update `design.md` first |
| `../../engineering/architecture.md` | current package boundaries and doctor profile contract | CLI/ownership/doctor/lint responsibilities | reconcile architecture owner before acceptance |
| `../../use-cases/UC-001-adopt-template.md`–`UC-003-audit-documentation.md` | stable scenario owners | adopt/update/audit triggers and outcomes | update use-case owner when delivered semantics change |
| `https://github.com/dapi/memory-bank/issues/63` | coordination owner | new source root, no duplicate, release-before-merge gate | stop and reconcile if upstream issue changes |

## Current State / Reference Points

| Path / symbol | Observed current behavior | Existing evidence / pattern | Required realization |
| --- | --- | --- | --- |
| `internal/ownership/source.go`: `verifySourceCheckout` | Verifies checkout root/HEAD/clean status, then checks ignored content with pathspec `memory-bank`. | `TestSourceRefMustMatchCleanGitCheckout`; hidden-worktree and CRLF tests prove pinned-object behavior. | Select exactly one recognized source root before applying the pathspec; retain checkout/ref and double-verification safety. |
| `internal/ownership/source.go`: `verifySourcePayload` | Runs `git ls-tree ... -- memory-bank`, accepts blob modes `100644`/`100755`, rejects empty/unsupported tree. | `TestPinnedSourceExecutableModeIsInstalled`. | Select exactly one of legacy `memory-bank` or target `memory-bank-template`, reject neither/both, and preserve all type/mode rejection behavior. |
| `internal/ownership/update.go`: `run` | Production calls `readGitSource`; injected-verifier tests call `readSource`; both feed `buildPlan` after a second source verification. | `opts`/`initialize` helpers in `update_test.go`; production Git fixtures in `source_test.go`. | Keep both reader implementations congruent and translate before their common `buildPlan` handoff. |
| `internal/ownership/update.go`: `readSource`, `readGitSource` | Walk/read `memory-bank` and use the source path directly as the payload-map key. | Filesystem-backed planner tests and Git-backed source tests exercise different branches. | Read the selected root (`memory-bank` or `memory-bank-template`) and emit `memory-bank/<suffix>` for both branches; preserve bytes, digest and mode. |
| `internal/ownership/update.go`: `buildPlan` | Rejects `memory-bank/.lock`, classifies payload keys, compares old lock entries and feeds atomic mutations. | `TestInitRejectsReservedLockPathInTemplate`, idempotence/conflict/delete/rollback suites. | Do not add source-prefix handling here; use it as the downstream-boundary oracle. |
| `internal/ownership/classify.go`: `Classify`; `lock.go`: `readLockSnapshot`; `types.go`: `LockFileName` | Classification prefixes and decoded lock paths are restricted to `memory-bank/*`; lock is `memory-bank/.lock`. | classification, schema upgrade and lock/symlink tests. | No semantic change; add assertions that translated reports and lock keys remain in this namespace. |
| `internal/agentinstructions/block.go`: `CurrentBlock` | Generated managed block contains only downstream `memory-bank/*` routes. | `internal/agentinstructions/block_test.go`; CLI alternative-agent update test. | Keep unchanged and prove init/update still generate the downstream projection. |
| `internal/doctor/doctor.go`: `Run`, `detectProfile` | `Run` normalizes scope before resolving auto profile; lock wins over exact `.memory-bank-template` marker. | `TestProfileDetectionUsesExactMarkerContract`, malformed-marker and CRLF tests. | Preserve detector rules; derive omitted scope after resolved profile and use one effective scope across navigation/governance. |
| `internal/doctor/doctor.go`: `checkIdentityAndDrift` | Hard-codes `memory-bank/README.md` for every profile; downstream managed-block drift inspection already runs only when a lock exists. | template/downstream doctor fixtures and `TestDoctorDoesNotMutateWorktree`. | Expect `<effective-scope>/README.md` for repository-local routing; retain downstream-only managed-block drift inspection and read-only behavior. |
| `internal/cli/cli.go`: `runDoctor`, `runLint`, `runOwnership` | Doctor/lint flags both default to `memory-bank`; lint already accepts explicit scope; source help says checkout contains `memory-bank/`. | CLI JSON, scope rejection, dry-run and alternative agent target tests. | Preserve whether doctor scope was explicitly supplied, keep lint default, update source help, and leave command/report contracts unchanged. |
| `internal/lint/audit.go`: `NormalizeScopeRoot`, `Run` | Accepts safe repository-relative scope and derives navigation from that scope. | `TestConfiguredEntrypointPrefersScope`, traversal rejection and golden report tests. | Reuse unchanged for explicit template lint and doctor effective scope. |
| `internal/doctor/testdata/{template,downstream}` | Both fixtures currently store their docs under `memory-bank/`; template `AGENTS.md` also routes there. | `TestProfilesUseSeparateFixturesAndProduceCleanReports`. | Rename only template payload/route to `memory-bank-template`; leave downstream fixture under `memory-bank` and assert both clean. |
| `.github/workflows/release.yml`: `validate`, `release` | Validate runs race tests/vet/GoReleaser; protected `release` job validates tag uniqueness and publishes. | existing `release` environment and v1.0.0 workflow pattern | Reuse without weakening dependency/approval; supply a new approved unused tag only at release time. |
| `README.md`, `memory-bank/engineering/architecture.md`, `UC-001`–`UC-003` | Source checkout is not path-specific in use cases; architecture and audit defaults describe current downstream root. | canonical project documentation | Add explicit source/downstream terminology while keeping downstream user navigation on `memory-bank/`. |

## Grounding Verdict

| Claim type | Verdict | Consequence for execution |
| --- | --- | --- |
| Observed | Source selection is duplicated across verifier, Git reader and filesystem reader, but downstream semantics converge at `buildPlan`. | `STEP-01` and `STEP-02` must change all source selectors/readers together and must not teach the planner about source paths. |
| Observed | Doctor profile, scope and agent-route expectations are currently separate; lint already supports arbitrary explicit scope. | `STEP-03` must create one effective doctor scope and dynamic local route, while lint remains structurally unchanged. |
| Observed | Downstream managed agent guidance, classification, lock validation and transactions are already isolated behind `memory-bank/*`. | These are regression oracles, not rename targets. |
| Inferential, backed by issues #14/#63 | Template-local navigation must move to `memory-bank-template/*`, while installed navigation remains `memory-bank/*`. | Template and downstream fixtures must intentionally diverge; a mechanical global test rename is invalid. |
| Unresolved | none | The root matrix is specified by the coordinated rollout: accept exactly one known root and reject neither/both. No discovery spike is required before `STEP-01`; contradictory runtime evidence activates `STOP-01`. |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| source verification/read | `REQ-01`, `REQ-02`, `SC-01`, `SC-07`, `NEG-01`, `NEG-02`, `CTR-01` | `TestSourceRefMustMatchCleanGitCheckout`, hidden-worktree, CRLF and executable-mode cases use old source root | add new-root-only and legacy-root-only fixtures; assert identical source→downstream keys; reject neither-root and both-root fixtures | `go test ./internal/ownership -run 'Test(Source|Pinned|CleanCheckout)'`; then full package | repository Validate job | none | none |
| init/update/lock/transactions | `REQ-03`, `SC-02`, `SC-03`, `NEG-03`, `CTR-02`, `INV-01`–`INV-03` | idempotence, reserved lock, adapted/managed conflict, deletion, transaction rollback and symlink suites | change only source-side fixture writes; retain downstream assertions and add report/lock namespace checks | `go test ./internal/ownership -run 'Test(CleanUpdate|InitRejectsReserved|Adapted|Managed|Removed|Transaction|Lock)'`; then full package | repository Validate job | none | none |
| doctor/lint/CLI | `REQ-04`, `SC-04`, `SC-05`, `CTR-03`, `INV-04` | exact marker/profile tests, separate fixtures, lint scope tests, CLI doctor/source dry-run tests | split template/downstream fixture roots and AGENTS routes; cover omitted vs explicit doctor scope, lock precedence, explicit template lint, unchanged downstream managed block/help/report | `go test ./internal/doctor ./internal/lint ./internal/cli` | repository Validate job | none | none |
| documentation/navigation | `REQ-01`, `REQ-07`, `EC-06` | project use cases and repo lint | terminology/use-case updates plus navigation audit | `CHK-04`; repository lint | repository Validate job | semantic review is not fully executable; reviewer checks no downstream route uses source root | `AG-02` |
| full regression | `REQ-05`, `EC-04` | canonical race/vet contract | full suite after targeted changes | `CHK-05` | repository Validate job | none | none |
| release/handoff | `REQ-06`, `SC-06`, `NEG-04`, `INV-05`, `RB-02`–`RB-04` | protected manual release workflow | exact-commit validation, approval, clean install/version, #63 handoff and separately reviewed legacy-removal reference | `CHK-06`–`CHK-08` | Validate + Publish release jobs | public publication and external template CI handoff require human/repository action | `AG-01` |

## Open Questions / Ambiguities

`none`: release tag selection is deliberately performed by the authorized maintainer through the protected workflow and recorded as evidence; it does not change the accepted source/downstream contract.

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Go version from `go.mod`, Git CLI, clean worktree; tests may create isolated temporary Git repositories | `STEP-01`–`STEP-05` | fixtures cannot commit/resolve HEAD or Go suites cannot run |
| test | targeted package tests precede `go test -count=1 -race ./...`, `go vet ./...` and repository lint | `CHK-01`–`CHK-05` | partial package pass is not acceptance evidence |
| access / network / secrets | no secrets for implementation; release requires GitHub workflow permission and repository-held secrets through protected environment | `STEP-06`, `STEP-07` | release validation/approval/credential fails; activate stop condition |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `design.md` `SOL-01`–`SOL-05` | design active and issue #63 still specifies the no-duplicate new source root | `STEP-01`–`STEP-05` | yes |
| `PRE-02` | `CON-03`, `INV-05`, `RB-02` | exact implementation commit passes required validation before release approval | `STEP-06` | no; blocks publication |
| `PRE-03` | `CON-04`, `SD-04` | authorized maintainer supplies an unused approved release version to the protected workflow | `STEP-06` | no; blocks publication |

## Design Realization Mapping

| Canonical solution refs | Owner | Realization target | Steps | Checks | Evidence |
| --- | --- | --- | --- | --- | --- |
| `SOL-01`, `SOL-02`, `C4-01`, `SD-01`, `SD-02`, `CTR-01`, `INV-01`, `INV-02`, `FM-01`–`FM-03` | `design.md` | `internal/ownership/source.go`, source readers and targeted tests | `STEP-01`, `STEP-02` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `CTR-02`, `INV-03`, `RB-01` | `design.md` | planner/lock/transaction regression surface | `STEP-02` | `CHK-02`, `CHK-05` | `EVID-02`, `EVID-05` |
| `SOL-03`, `SOL-04`, `SD-03`, `CTR-03`, `INV-04`, `FM-04` | `design.md` | `doctor.Run` effective scope, `checkIdentityAndDrift` local route, CLI explicit-scope detection, diverged doctor fixtures and lint invocation/tests | `STEP-03` | `CHK-03` | `EVID-03` |
| `SOL-05`, `TRD-03` | `design.md` | CLI help, architecture and `UC-001`–`UC-003` | `STEP-04` | `CHK-04` | `EVID-04` |
| `SD-04`, `INV-05`, `FM-05`, `RB-02`–`RB-04` | `design.md` | release workflow run, immutable tag, #63 handoff and legacy-retirement handoff | `STEP-06`–`STEP-08` | `CHK-06`–`CHK-08` | `EVID-06`–`EVID-08` |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01`–`REQ-03`, `SOL-01`, `SOL-02` | bounded rollout source reader/translation with unchanged downstream ownership behavior | either | `PRE-01` |
| `WS-2` | `REQ-04`, `SOL-03`, `SOL-04` | correct template/downstream diagnostics and navigation scopes | either | `PRE-01` |
| `WS-3` | `REQ-05`, `REQ-07`, `SOL-05` | complete regression and documentation evidence | either | `WS-1`, `WS-2` |
| `WS-4` | `REQ-06`, `RB-02`–`RB-04` | approved compatible release, template-side handoff and legacy-retirement handoff | human/either | `WS-3`, `PRE-02`, `PRE-03` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | validated candidate is ready for a new release tag and GitHub Release | `STEP-06`, `WS-4` | immutable public tag and credentialed publication are external effects | required reviewer approves protected `release` environment; `EVID-06` |
| `AG-02` | implementation docs/use cases are ready for acceptance | documentation portion of `STEP-04` | human semantic review confirms no source-only route is presented to downstream users | PR review record in `EVID-04` |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent/either | `REQ-01`, `REQ-02`, `SOL-01`, `CTR-01` | define exact source/downstream constants and a selector that accepts exactly one recognized source root | `verifySourceCheckout`, `verifySourcePayload`, source Git fixtures | verified legacy/new single-root boundary with unchanged pinning/mode rules | `CHK-01` | `EVID-01` | run targeted tests for legacy-only, new-only, neither and both roots | `PRE-01` | none | issue #63 contract changes or root selection cannot remain unambiguous |
| `STEP-02` | agent/either | `REQ-02`, `REQ-03`, `SOL-02`, `CTR-02` | translate once in both source readers and preserve planner/lock/transaction behavior | `readSource`, `readGitSource`, `buildPlan` boundary, ownership fixtures/tests | congruent downstream-keyed payload from each accepted source-root branch | `CHK-02` | `EVID-02` | targeted init/update/transaction tests for both source roots; inspect decision paths and decoded lock keys | `STEP-01` | none | readers disagree, any source-prefixed key reaches planner, or safety verdict changes without owner update |
| `STEP-03` | agent/either | `REQ-04`, `SOL-03`, `SOL-04`, `CTR-03` | derive one effective doctor scope and local agent route; exercise explicit template lint scope | `doctor.Run`, `checkIdentityAndDrift`, `runDoctor`, doctor fixtures/tests, CLI/lint tests | template `memory-bank-template` and downstream `memory-bank` profile/scope/route coverage | `CHK-03` | `EVID-03` | run doctor/lint/CLI suites; assert template and downstream clean reports, explicit override, lock precedence and unchanged downstream managed block | `PRE-01` | none | satisfying issue requires a new public lint contract or downstream managed-block semantics change |
| `STEP-04` | agent/either | `REQ-01`, `REQ-07`, `SOL-05` | update source help, architecture and stable use cases | CLI help, `README.md`, architecture, `UC-001`–`UC-003` | consistent terminology/navigation | `CHK-04` | `EVID-04` | repository lint plus semantic diff review | `STEP-01`, `STEP-03` | `AG-02` for acceptance | downstream instructions would need source-only path |
| `STEP-05` | agent/either | `REQ-05`, all implemented refs | run complete local validation and reconcile evidence | all changed code/docs | release candidate | `CHK-05` | `EVID-05` | race suite, vet, repository lint | `STEP-02`–`STEP-04` | none | targeted and full-suite verdicts disagree |
| `STEP-06` | human/either | `REQ-06`, `INV-05`, `RB-02` | validate exact commit and publish approved release | GitHub release workflow/environment | immutable compatible release | `CHK-06` | `EVID-06` | dispatch with approved unused version; validate; approve `AG-01`; inspect tag/release | `STEP-05`, `PRE-02`, `PRE-03` | `AG-01` | validation, approval, credentials or tag uniqueness fail |
| `STEP-07` | human/either | `REQ-06`, `RB-03` | prove consumer availability and hand version to issue #63/template CI | clean Go install, GitHub issues/CI | release handoff | `CHK-07` | `EVID-07` | clean install, `--version`, link handoff evidence | `STEP-06` | none | install/version fails or template rename starts before evidence |
| `STEP-08` | human/either | `REQ-06`, `EC-07`, `RB-04` | record the bounded compatibility matrix and route the separately reviewed legacy-root removal after #63 | issue #63 handoff and follow-up delivery record | explicit retirement handoff, without inventing a version | `CHK-08` | `EVID-08` | record owner/link and post-#63 removal condition; do not claim automatic removal | `STEP-07` | none | no owner or follow-up record exists when #63 is ready to merge |

## Parallelizable Work

- `PAR-01` After `STEP-01` establishes constants/bridge, doctor/lint work in `STEP-03` can proceed in parallel with most init/update fixture work in `STEP-02`.
- `PAR-02` Documentation drafting may parallel code work, but final `STEP-04` acceptance waits for actual CLI semantics.
- `PAR-03` Release and handoff are strictly sequential after all implementation and verification work.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `STEP-02`, `CTR-01`, `CTR-02` | each accepted single source root passes into an exclusively downstream-keyed planner without safety regression | `EVID-01`, `EVID-02` |
| `CP-02` | `STEP-03`, `CTR-03`, `INV-04` | template and downstream profile/scope matrices pass | `EVID-03` |
| `CP-03` | `STEP-04`, `STEP-05`, `AG-02` | docs, use cases, targeted and full local checks agree | `EVID-04`, `EVID-05` |
| `CP-04` | `STEP-06`, `AG-01`, `INV-05` | exact commit validation and approval precede immutable release | `EVID-06` |
| `CP-05` | `STEP-07`, `RB-03` | clean install succeeds and issue #63 records the compatible version before merge | `EVID-07` |
| `CP-06` | `STEP-08`, `RB-04` | #63 handoff records bounded compatibility and the separately reviewed legacy-removal release route | `EVID-08` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | mechanical fixture rename also changes downstream expected paths | tests stop proving the preserved contract | separate source setup paths from target/assertion paths; review lock/report values | downstream assertion contains `memory-bank-template/` |
| `ER-02` | dirty/ignored Git pathspec remains on old source root | unsafe new-root content can evade validation | explicit negative Git fixtures for new root | `NEG-02` unexpectedly passes source validation |
| `ER-03` | CLI cannot distinguish omitted scope from explicit `memory-bank`, or doctor derives scope but keeps the hard-coded downstream agent route | template default overrides user intent or template doctor reports a false routing error | test flag-presence handling plus profile/scope/agent-route matrix | navigation root and expected agent link do not share the same effective scope |
| `ER-04` | release occurs without usable consumer artifact | issue #63 merges against unavailable support | require clean-install handoff before unblocking | `CHK-07` absent/fails |
| `ER-05` | transition compatibility is left in place without a recorded retirement action | an excluded indefinite dual-root behavior persists | require `STEP-08`/`EVID-08` before declaring the rollout handoff complete | #63 is ready to merge but no removal route is recorded |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `FM-01`–`FM-04`, `ER-01`–`ER-03` | source/downstream invariant or regression fails | stop implementation progression; update canonical design only if issue-backed semantics must change | unpublished branch; existing released CLI/template layout |
| `STOP-02` | `CON-03`, `INV-05`, `NEG-04` | validation or `AG-01` fails | do not create/continue publication; correct candidate or obtain approval | validated or unvalidated untagged candidate |
| `STOP-03` | `FM-05`, `RB-02` | defect found after tag creation | never repoint tag; keep issue #63 blocked and request approved corrective release direction | immutable release retained; template old source root retained |
| `STOP-04` | `RB-03`, `ER-04` | template rename is about to merge without `EVID-07` | notify coordination owner and keep dependency gate closed | template repository remains on old payload root |
| `STOP-05` | `RB-04`, `ER-05` | no separately reviewed removal route is recorded after the #63 rename is coordinated | stop closure of the rollout handoff and obtain an owner/follow-up record | released transition CLI continues to support either single root |

## Plan-local Evidence

No separate plan-local evidence is required. All planned evidence is part of the canonical `brief.md` evidence contract.

## Готово для приемки

- All workstreams and checkpoints are complete.
- `CHK-01`–`CHK-08` have concrete `EVID-01`–`EVID-08` carriers.
- Required local and CI suites are green; `AG-01` and `AG-02` have review records.
- No source-only path appears in downstream decisions, locks, targets or user navigation.
- Template issue #63 records the released compatible version and remains blocked until that evidence exists.
- The #63 handoff records the separately reviewed legacy-root removal release route.
- Final acceptance follows `brief.md#verify`.
