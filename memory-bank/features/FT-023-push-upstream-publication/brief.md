---
title: "FT-023: Push Upstream Publication"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для публикации upstream-пригодных изменений через PR. Фиксирует problem space, scope, blocking decisions и verify без принятия solution или execution decisions."
derived_from:
  - ../../flows/feature.md
  - ../../engineering/validation-profiles.md
  - ../../use-cases/UC-004-publish-managed-changes-upstream.md
  - "https://github.com/dapi/memory-bank-cli/issues/23"
status: active
delivery_status: done
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
  - upstream_selection_algorithm
  - transaction_and_cleanup_protocol
---

# FT-023: Push Upstream Publication

## What

### Problem

После отладки общих flow, процессов и шаблонов в downstream-проекте их нужно перенести в upstream-шаблон. Сейчас для этого требуется вручную работать с отдельным checkout и GitHub PR. Issue #23 запрашивает `memory-bank-cli push`, которая из `memory-bank/` текущего Git-репозитория безопасно создаёт upstream PR, не публикуя project-specific artifacts или локальное состояние.

### Outcome

Issue #23 задаёт качественные acceptance criteria, но не задаёт baseline или числовой target; они не выводятся из package.

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Выполнение явных acceptance criteria issue #23 | not stated | Все criteria продемонстрированы | Canonical scenarios, checks и evidence после принятия solution contract |

### Scope

- `REQ-01` Expose `memory-bank-cli push` and document it in CLI help and README; its source is `memory-bank/` of the current Git repository and the target checkout is `memory-bank/.repo`.
- `REQ-02` Before mutation, validate Git-repository context, safe and valid `memory-bank/.repo` path, upstream remote/identity, clean checkout and absence of unresolved conflicts; report actionable diagnostics and stop on unsafe, dirty, invalid or ambiguous state.
- `REQ-03` Select only changes suitable for canonical upstream Memory Bank and exclude project-dependent artifacts, lock/state and `.repo` contents.
- `REQ-04` By default create a separate upstream branch, apply the selected changes, commit, push and create a GitHub PR in the upstream described by `.repo`; direct push to the default branch is prohibited by default.
- `REQ-05` Support `--dry-run` that displays the plan and exclusions without changing checkout, GitHub or remotes; cover successful, dry-run and key failure scenarios with tests.

### Non-Scope

- `NS-01` Automatically merge the upstream PR.
- `NS-02` Directly push to the upstream default branch without a separate explicit safe mechanism.
- `NS-03` Transfer arbitrary project-dependent files.

### Constraints / Assumptions

- `ASM-01` Issue #23 is the authoritative product input for this delivery unit; it names required outcomes but does not define an inclusion classification algorithm, branch/PR metadata contract, or recovery protocol after a partial external operation.
- `ASM-02` Current `internal/ownership.Classify` classifies `memory-bank/dna/`, `flows/` and `prompts/` as `managed`; selected project sections as `adapted`; `features/`, `adr/`, `prd/`, `epics/` and `use-cases/` contents as `user-owned`; and unknown Memory Bank paths as `user-owned`. This is a discovery fact, not an approved publish-selection policy.
- `CON-01` Any ambiguous selection, unsafe/dirty checkout, invalid remote or unresolved conflict must stop the operation before irreversible action.
- `CON-02` The standard upstream is `dapi/memory-bank`, but `.repo` may identify a different upstream; `dapi/memory-bank` must not be the only supported upstream.
- `CON-03` `--dry-run` must not create a branch, commit, push or PR, and must not mutate the local upstream checkout.
- Accepted FPF decisions are owned by `design.md` (`SD-01`–`SD-03`); their provenance is retained in `decision-log.md`.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature adds a CLI and Git/GitHub integration contract, crosses filesystem, remote and GitHub boundaries, requires safety/failure/rollback semantics and explicit selection trade-offs. | `design.md` after blocking decisions close |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- | --- |
| `decision-log.md` | selected | FPF provenance is needed to distinguish issue facts from unresolved material solution decisions. | `decision-log.md`; canonical owners remain `brief.md` / future `design.md` |
| `design.md` | selected | CLI, local filesystem, Git remote and GitHub PR boundaries require explicit selected solution and failure semantics. | `design.md` |
| `implementation-plan.md` | deferred | This documentation task does not authorize implementation; create it only when the feature enters execution. | `implementation-plan.md` |
| feature-local contract / sequence diagram / ADR | omitted | The compact design contains one feature-local contract and its C1 context; no reusable/cross-feature decision or separate temporal view is yet justified. | none |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `standard` | CLI/integration change with safety-critical failure behavior but no release or deployment. The feature-local minimum is targeted unit/integration coverage, full Go suite, `go vet`, navigation audit and approved live GitHub PR evidence. | none; selected by user-directed FPF decision |

## Verify

### Exit Criteria

- `EC-01` A valid downstream repository can obtain an upstream PR containing only the approved publishable change set, without direct default-branch push.
- `EC-02` Unsafe, dirty, invalid, conflicting or ambiguous conditions produce actionable diagnostics and no irreversible action.
- `EC-03` `--dry-run` presents the planned inclusions/exclusions and produces no local, remote or GitHub mutation.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `CON-02`, `DEC-03` | `EC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CON-01`, `DEC-02`, `DEC-03` | `EC-02` | `CHK-02` | `EVID-02` |
| `REQ-03` | `ASM-02`, `DEC-01`, `DEC-03` | `EC-01`, `EC-02` | `CHK-03` | `EVID-03` |
| `REQ-04` | `CON-02`, `DEC-02`, `DEC-03` | `EC-01`, `EC-02` | `CHK-04` | `EVID-04` |
| `REQ-05` | `CON-03`, `DEC-02`, `DEC-03` | `EC-03` | `CHK-05` | `EVID-05` |

### Acceptance Scenarios

- `SC-01` A downstream Git repository whose safe, clean `memory-bank/.repo` identifies `dapi/memory-bank` runs `push`; the approved change set is proposed in a separate upstream branch and PR, while the default branch is not directly pushed.
- `SC-02` A safe, clean `.repo` identifying a non-default upstream runs `push`; the PR targets that identified upstream rather than requiring `dapi/memory-bank`.
- `SC-03` A candidate set containing project-dependent files, lock/state or `.repo` content reports them as excluded; ambiguous classification stops before irreversible action.
- `SC-04` An unsafe/dirty `.repo`, invalid upstream remote/identity or unresolved conflict stops with a reason and corrective next step, without partial application.
- `SC-05` `push --dry-run` displays planned inclusions/exclusions and performs no checkout, GitHub or remote mutation.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `SC-01`, `SC-02` | Targeted CLI/Git integration suite plus approved live GitHub PR procedure | branch/PR workflow and non-default-upstream behavior pass | `EVID-01` |
| `CHK-02` | `SC-04` | Targeted safety/failure suite | each preflight failure is diagnostic and leaves no accepted mutation | `EVID-02` |
| `CHK-03` | `SC-03` | Selection fixtures | only managed paths are included; excluded or non-classifiable paths are reported or stop as specified | `EVID-03` |
| `CHK-04` | `SC-01`, `SC-04` | Transaction/recovery suite plus failure-injection review | remote/local compensation follows the accepted protocol | `EVID-04` |
| `CHK-05` | `SC-05` | Dry-run mutation-observation suite | plan is shown with no mutation | `EVID-05` |

### Evidence

- `EVID-01` **PASS** — `TestRunCreatesBranchCopiesManagedFileAndReturnsPR`,
  default/custom-upstream contract tests, and the controlled live
  [dapi/memory-bank#76](https://github.com/dapi/memory-bank/pull/76). The live
  PR targeted `main` through a fresh branch and contained one managed path.
- `EVID-02` **PASS** — `TestRejectsDirtyUpstreamCheckout`,
  `TestSafeCheckoutRejectsSymlinkedMemoryBankParent`,
  `TestSafeCheckoutRejectsSymlinkedRepo`,
  `TestCleanReportsUpstreamConflictBeforeDirtyState`, invalid-remote coverage,
  and the complete unmerged-status table pass in
  [Validate run 30116321734](https://github.com/dapi/memory-bank-cli/actions/runs/30116321734).
- `EVID-03` **PASS** — `TestDryRunIncludesOnlyManagedPaths` proves managed
  inclusion and adapted exclusion; the live carrier included only
  `memory-bank/dna/push-live-validation.md`.
- `EVID-04` **PASS** — `TestRunCompensatesRemoteBranchWhenPRCreationFails`
  covers injected PR failure and compensation. Live PR #76 left upstream
  `main` at `9a4463cf8ee06860a0a7238cb283337df0ae496e`, was closed without
  merge, and its temporary branch was deleted.
- `EVID-05` **PASS** — targeted dry-run coverage asserts unchanged upstream
  HEAD, branch and status; the live dry run also left the remote default SHA
  unchanged.

The full required CI profile passed in the
[completion PR #34](https://github.com/dapi/memory-bank-cli/pull/34):
[Validate](https://github.com/dapi/memory-bank-cli/actions/runs/30116321734),
[local E2E](https://github.com/dapi/memory-bank-cli/actions/runs/30116321739),
and [stable downstream smoke](https://github.com/dapi/memory-bank-cli/actions/runs/30116321679).
