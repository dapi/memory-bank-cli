---
title: "FT-014: Source Payload Path Translation"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для чтения переименованного template-source payload при сохранении downstream Memory Bank contract."
derived_from:
  - ../../flows/feature.md
  - ../../engineering/architecture.md
  - ../../engineering/testing-policy.md
  - ../../use-cases/UC-001-adopt-template.md
  - ../../use-cases/UC-002-update-template.md
  - ../../use-cases/UC-003-audit-documentation.md
  - "https://github.com/dapi/memory-bank-cli/issues/14"
  - "https://github.com/dapi/memory-bank/issues/63"
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
  - source_to_downstream_translation_algorithm
---

# FT-014: Source Payload Path Translation

## What

### Problem

The coordinated template change in `dapi/memory-bank#63` renames only the upstream payload directory from `memory-bank/` to `memory-bank-template/`. The CLI currently inspects and reads `memory-bank/` in the source checkout and also uses `memory-bank/` as the downstream destination, lock namespace, classification namespace and navigation default. Without an explicit source/downstream distinction, the template rename either makes init/update reject the source or leaks the source-only name into downstream repositories.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Source payload compatibility | CLI requires source `memory-bank/` | The transition release accepts a clean pinned checkout containing exactly one recognized root: legacy `memory-bank/` or new `memory-bank-template/`; both map to the same downstream paths | source-verification and init regression tests |
| `MET-02` | Downstream contract stability | installed path, lock and defaults are `memory-bank/` | init/update results, lock entries, agent routes and downstream lint/doctor defaults remain `memory-bank/` | end-to-end init/update/doctor/navigation tests |
| `MET-03` | Coordination readiness | template rename is blocked on compatible CLI release | an approved released CLI version with FT-014 support is available before issue #63 merges | release URL/tag and clean install/smoke evidence |

### Scope

- `REQ-01` Define `memory-bank-template/` as the target source payload root, `memory-bank/` as the unchanged downstream payload root, and the legacy source root as a bounded rollout compatibility input in CLI implementation and user-facing source option guidance.
- `REQ-02` During the coordinated rollout, validate the clean pinned Git source containing exactly one recognized payload root (`memory-bank/` or `memory-bank-template/`) and translate each source payload path to the corresponding downstream `memory-bank/` path before ownership classification, planning, lock generation or mutation.
- `REQ-03` Preserve init/update safety and managed-file semantics while clean init creates `memory-bank/` and update reconciles an existing locked `memory-bank/` from the renamed source tree.
- `REQ-04` Make template-source diagnostics and navigation validate `memory-bank-template/`, while downstream lint/doctor defaults, lock inspection and agent-routing checks remain `memory-bank/`.
- `REQ-05` Add regression coverage for source verification, init, update, doctor template-profile detection and source/downstream navigation.
- `REQ-06` Publish an approved CLI release containing the support and make its version available to template CI before `dapi/memory-bank#63` merges.
- `REQ-07` Update the stable adopt, update and audit use-case documentation when delivered behavior changes, without telling downstream users to navigate `memory-bank-template/`.

### Non-Scope

- `NS-01` Rename the `memory-bank-cli` executable, public subcommands or downstream `memory-bank/` directory.
- `NS-02` Retain `memory-bank/` as an alternative upstream payload root indefinitely or introduce open-ended dual-path source discovery. The bounded compatibility input in `REQ-02` ends in a separately released removal change after the #63 rename is coordinated.
- `NS-03` Change ownership classes, lock schema, managed-file semantics, downstream lock location or agent instruction route except for the source-to-downstream path translation required by this feature.
- `NS-04` Perform the template-repository rename or its template-local documentation/CI changes; those remain owned by `dapi/memory-bank#63`.
- `NS-05` Select or publish a release tag without the existing protected release-environment approval.

### Constraints / Assumptions

- `ASM-01` Issue #14 and coordinated issue #63 are authoritative for the two path meanings and merge ordering. Issue #63 requires the upstream repository to contain `memory-bank-template/` and no duplicate `memory-bank/` payload.
- `ASM-02` The current CLI already accepts an explicit lint `--scope-root` and doctor `--profile auto|template|downstream`; only doctor currently has a profile detector.
- `CON-01` All paths stored in ownership decisions and `memory-bank/.lock` remain downstream paths beginning with `memory-bank/`.
- `CON-02` Source verification remains bound to a clean Git checkout whose HEAD equals the supplied full commit SHA; symlink and unsupported Git-entry rejection remain in force.
- `CON-03` The compatible CLI release is a dependency gate for the template-side merge. Public tag creation and release publication are externally effective and require human approval.
- `CON-04` No exact release version is specified by either issue. The approved workflow input and resulting immutable tag are evidence; this package must not invent a version number.
- `CON-05` The release-before-rename ordering means a released CLI must continue to read the current legacy-only template until #63 changes the template. The later removal of that compatibility is a separately reviewed release; no automatic online detection of #63 state is assumed.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The change crosses Git-tree, filesystem, ownership/lock, CLI diagnostic and release boundaries; it requires explicit translation, compatibility, failure and rollout decisions. | [design.md](design.md) |

## Artifact Routing Decision

| Artifact | Decision | Trigger / reason | Route / owner |
| --- | --- | --- | --- |
| `design.md` | selected | Source/downstream namespace bridge, diagnostics and release ordering require a solution owner. | [design.md](design.md) |
| `implementation-plan.md` | selected | Multiple packages, regression surfaces and an approval-gated release require grounded sequencing. | [implementation-plan.md](implementation-plan.md) |
| `decision-log.md` | selected | The user requested FPF closure and auditable reasoning for blocking questions. | [decision-log.md](decision-log.md) |
| Separate contract/C4/sequence artifact | omitted | The connector and topology are compact and fully reviewable in `design.md`. | none |

## Validation Profile Decision

| Profile | Triggers / rationale | Downgrade approval |
| --- | --- | --- |
| `release-deployment` | The delivery unit changes compatibility-sensitive CLI/filesystem behavior and explicitly publishes a release used by another repository's CI. Targeted regressions, full race suite, vet, navigation checks, release workflow validation and post-release install evidence are required. | none |

## Verify

### Exit Criteria

- `EC-01` During rollout, a clean pinned source containing exactly one recognized root—legacy `memory-bank/` or new `memory-bank-template/`—is accepted; source status and Git-tree verification are scoped to that selected root, while neither-root and both-root sources fail.
- `EC-02` Clean init and locked update translate source files to downstream `memory-bank/` paths without changing ownership, lock schema, conflict, atomicity or agent-routing behavior.
- `EC-03` Template-source doctor/navigation inspect `memory-bank-template/`; downstream lint/doctor without a scope override continue to inspect `memory-bank/`.
- `EC-04` Automated regression coverage includes init, update, source verification, template-profile detection and navigation in both contexts; full Go race tests and vet pass.
- `EC-05` An approved released CLI version containing the change is installable and recorded for template CI before the template-side rename merges.
- `EC-06` Canonical use cases distinguish the source checkout root from the downstream installed root and retain downstream navigation terminology.
- `EC-07` The #63 handoff records that the released CLI accepts both roots during rollout and records the required separately reviewed legacy-removal release after the template rename is coordinated.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01`, `REQ-02` | `ASM-01`, `CON-01`, `CON-02`, `CON-05` | `EC-01`, `SC-01`, `SC-07`, `NEG-01`, `NEG-02` | `CHK-01`, `CHK-05` | `EVID-01`, `EVID-05` |
| `REQ-03` | `CON-01`, `CON-02` | `EC-02`, `SC-02`, `SC-03`, `NEG-03` | `CHK-02`, `CHK-05` | `EVID-02`, `EVID-05` |
| `REQ-04` | `ASM-02`, `CON-01` | `EC-03`, `SC-04`, `SC-05` | `CHK-03`, `CHK-05` | `EVID-03`, `EVID-05` |
| `REQ-05` | `CON-01`, `CON-02` | `EC-04`, `SC-01`–`SC-05` | `CHK-01`–`CHK-03`, `CHK-05` | `EVID-01`–`EVID-03`, `EVID-05` |
| `REQ-07` | `CON-01` | `EC-06`, `SC-02`–`SC-05` | `CHK-04` | `EVID-04` |
| `REQ-06` | `CON-03`–`CON-05` | `EC-05`, `EC-07`, `SC-06`, `NEG-04` | `CHK-06`–`CHK-08` | `EVID-06`–`EVID-08` |

### Acceptance Scenarios

- `SC-01` A maintainer supplies a clean pinned template checkout whose only payload root is `memory-bank-template/`; the transition CLI accepts the committed regular files and supported modes.
- `SC-07` Before #63 renames the template, a maintainer supplies a clean pinned template checkout whose only payload root is legacy `memory-bank/`; the same transition CLI accepts it and emits the same downstream `memory-bank/<suffix>` payload keys as `SC-01`.
- `SC-02` Clean init from that checkout creates downstream `memory-bank/README.md`, downstream ownership paths and `memory-bank/.lock`; it does not create `memory-bank-template/` in the target.
- `SC-03` Update from a newer pinned checkout safely reconciles existing downstream `memory-bank/` content with the current conflict, preservation, delete and atomic-apply semantics.
- `SC-04` In a marker-bearing template checkout, doctor auto/template diagnostics and source CI navigation validate `memory-bank-template/`.
- `SC-05` In an adopted downstream repository, lint and doctor defaults, lock inspection and agent guidance continue to use `memory-bank/`.
- `SC-06` After code and CI acceptance, a maintainer approves the protected release job; the resulting released version is clean-installed and supplied to template CI before issue #63 merges.

### Negative / Edge Scenarios

- `NEG-01` A pinned commit with neither recognized source root, or with both `memory-bank/` and `memory-bank-template/`, is rejected before planning; the latter is ambiguous rather than a supported duplicate state.
- `NEG-02` Dirty, ignored, symlink or unsupported entries under the source payload remain rejected and cannot be hidden by reading a different tree.
- `NEG-03` Translation cannot emit a path outside `memory-bank/`, overwrite `memory-bank/.lock` from source content or introduce `memory-bank-template/` into lock entries.
- `NEG-04` Failed validation or absent release approval stops before a new public tag/release is created.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01`, `SC-07`, `NEG-01`, `NEG-02` | Run targeted `internal/ownership` source-verification tests against Git fixtures using each recognized root, neither root and both roots. | either clean single source root is accepted and maps identically downstream; unsafe, missing and ambiguous cases fail. | `artifacts/ft-014/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-02`, `SC-03`, `NEG-03` | Run init/update/transaction regressions from new-root fixtures and inspect reports/locks/target paths. | translated downstream paths and existing safety verdicts are correct. | `artifacts/ft-014/verify/chk-02/` |
| `CHK-03` | `EC-03`, `SC-04`, `SC-05` | Run CLI/doctor/lint tests for marker-bearing template and locked downstream repositories with default and explicit scopes. | template source validates `memory-bank-template/`; downstream defaults remain `memory-bank/`. | `artifacts/ft-014/verify/chk-03/` |
| `CHK-04` | `EC-06` | Review and lint updated `UC-001`–`UC-003`, root help/source guidance and feature navigation. | source/downstream terms are explicit; no downstream instruction routes to `memory-bank-template/`. | `artifacts/ft-014/verify/chk-04/` |
| `CHK-05` | `EC-04` | Run `go test -count=1 -race ./...`, `go vet ./...` and `go run ./cmd/memory-bank-cli lint --repo-root .`. | all checks pass. | `artifacts/ft-014/verify/chk-05/` |
| `CHK-06` | `EC-05`, `SC-06`, `NEG-04` | Inspect the release workflow run, exact validated commit, protected-environment approval, tag and GitHub Release. | validation and approval precede publication; release contains FT-014. | `artifacts/ft-014/verify/chk-06/` |
| `CHK-07` | `EC-05`, `SC-06` | Clean-install the approved release tag, smoke `--version`, and record the version handed to template CI/issue #63. | installed version matches the immutable tag and is available before template merge. | `artifacts/ft-014/verify/chk-07/` |
| `CHK-08` | `EC-07` | Record the compatibility matrix and a link/owner for the separately reviewed legacy-removal release in the #63 handoff. | the rollout has an explicit bounded retirement action; no claim of automatic removal is made. | `artifacts/ft-014/verify/chk-08/` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-014/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-014/verify/chk-02/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-014/verify/chk-03/` |
| `CHK-04` | `EVID-04` | `artifacts/ft-014/verify/chk-04/` |
| `CHK-05` | `EVID-05` | `artifacts/ft-014/verify/chk-05/` |
| `CHK-06` | `EVID-06` | `artifacts/ft-014/verify/chk-06/` |
| `CHK-07` | `EVID-07` | `artifacts/ft-014/verify/chk-07/` |
| `CHK-08` | `EVID-08` | `artifacts/ft-014/verify/chk-08/` |

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | targeted source-verification test log | test runner | `artifacts/ft-014/verify/chk-01/` | `CHK-01` |
| `EVID-02` | init/update translation and safety test log | test runner | `artifacts/ft-014/verify/chk-02/` | `CHK-02` |
| `EVID-03` | doctor/lint/CLI profile and navigation test log | test runner | `artifacts/ft-014/verify/chk-03/` | `CHK-03` |
| `EVID-04` | documentation/navigation review result | reviewer | `artifacts/ft-014/verify/chk-04/` | `CHK-04` |
| `EVID-05` | full race suite, vet and repository lint output | test runner | `artifacts/ft-014/verify/chk-05/` | `CHK-05` |
| `EVID-06` | release run URL, exact commit, approval and tag/release inventory | CI/release maintainer | `artifacts/ft-014/verify/chk-06/` | `CHK-06` |
| `EVID-07` | clean-install/version output and template-side handoff link | release maintainer | `artifacts/ft-014/verify/chk-07/` | `CHK-07` |
| `EVID-08` | #63 handoff record with compatibility matrix and legacy-removal follow-up reference | release maintainer | `artifacts/ft-014/verify/chk-08/` | `CHK-08` |
