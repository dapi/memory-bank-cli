---
title: "FT-014: Source Payload Path Translation Design"
doc_kind: feature
doc_function: canonical
purpose: "Feature-local solution for translating the renamed template-source payload into the unchanged downstream namespace."
derived_from:
  - brief.md
  - ../../engineering/architecture.md
source_refs:
  - ../../../internal/ownership/source.go
  - ../../../internal/ownership/update.go
  - ../../../internal/ownership/classify.go
  - ../../../internal/ownership/lock.go
  - ../../../internal/doctor/doctor.go
  - ../../../internal/cli/cli.go
  - ../../../internal/lint/audit.go
  - ../../../internal/agentinstructions/block.go
  - ../../../.github/workflows/release.yml
status: active
audience: humans_and_agents
must_not_define:
  - ft_014_scope
  - ft_014_acceptance_criteria
  - ft_014_evidence_contract
  - implementation_sequence
---

# FT-014: Source Payload Path Translation Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `ALT-*`, `TRD-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` |
| `decision-log.md` | Derived provenance ledger | FPF rationale and links to canonical owners; no requirements or selected solution |

## Context

The current ownership pipeline uses one spelling, `memory-bank/`, for two roles: selecting files inside a trusted source checkout and naming files inside the target repository. Issue #14 and coordinated issue #63 split those roles. The selected design introduces one explicit translation boundary while leaving the downstream ownership domain unchanged.

## Grounding / Current Implementation Facts

The facts below are observations of the current checkout, not target-state requirements. Their grounding mode is `observed`: each claim points to a reconstructible code or test location. The final column records only the design consequence supported by that observation.

| Observed fact | Grounded by | Design consequence |
| --- | --- | --- |
| `ownership.run` verifies the pinned source before reading, uses `readGitSource` for production, verifies the source again, and only then calls `buildPlan`. | `internal/ownership/update.go`: `run`, `readGitSource`, `buildPlan` | The safe translation seam is inside the source readers, before `buildPlan`; source pinning and the second verification remain unchanged. |
| `verifySourceCheckout`, `verifySourcePayload`, `readSource` and `readGitSource` all select literal `memory-bank`, and both readers currently emit those Git/filesystem paths directly as payload-map keys. | `internal/ownership/source.go`: `verifySourceCheckout`, `verifySourcePayload`; `internal/ownership/update.go`: `readSource`, `readGitSource` | All four source-selection/read surfaces need one exact-root selector; during rollout it accepts one of the two known source roots and both readers emit translated downstream keys consistently. |
| `buildPlan` rejects `memory-bank/.lock`, while `Classify` and lock decoding accept/interpret only `memory-bank/*` paths. | `internal/ownership/update.go`: `buildPlan`; `internal/ownership/classify.go`: `Classify`; `internal/ownership/lock.go`: `readLockSnapshot` | Planner, classification and lock namespaces stay downstream-only; the translated map must preserve their current input contract. |
| The generated managed agent block routes downstream repositories to `memory-bank/README.md`, `memory-bank/dna/README.md` and `memory-bank/flows/routing.md`. | `internal/agentinstructions/block.go`: `CurrentBlock` | Init/update-generated downstream agent guidance remains unchanged. This does not govern the template repository's own `AGENTS.md`. |
| Doctor currently normalizes `ScopeRoot` before resolving `Profile`, passes that scope to lint/governance, and separately hard-codes `memory-bank/README.md` in `checkIdentityAndDrift`. | `internal/doctor/doctor.go`: `Run`, `detectProfile`, `checkIdentityAndDrift`; `internal/cli/cli.go`: `runDoctor` | Doctor needs one resolved effective scope used by navigation, governance and the expected agent-entrypoint link; omitted scope must remain distinguishable from an explicit override. |
| Lint has no profile concept: it accepts any safe repository-relative `ScopeRoot`, derives `<scope>/README.md`, and audits that tree. The CLI already exposes `--scope-root`. | `internal/lint/audit.go`: `NormalizeScopeRoot`, `Run`; `internal/cli/cli.go`: `runLint` | Template lint can use the existing explicit scope; adding profile detection to lint is unnecessary. |
| The release workflow validates tests, vet and GoReleaser before an environment-protected job creates/verifies the requested tag and publishes. | `.github/workflows/release.yml`: `validate`, `release` jobs | FT-014 reuses the existing release boundary; it does not redesign release topology or select an unstated version. |

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-01` | `C3 compact view` | Responsibility changes across CLI, ownership and doctor/lint components inside the single CLI container; no new deployable is introduced. | component/connector table below |

| Component | Responsibility after FT-014 | Connector / direction |
| --- | --- | --- |
| `internal/ownership` source reader/verifier | validate one committed rollout source root and translate source-relative payload paths | Git/filesystem exactly one of `memory-bank/*` or `memory-bank-template/*` → in-process payload map keyed by `memory-bank/*` |
| ownership planner/lock/transaction | classify and apply only downstream paths | translated `memory-bank/*` map → target filesystem and `memory-bank/.lock` |
| `internal/doctor` + CLI | resolve template/downstream profile, effective scope and expected repository-local agent route | marker/profile/explicit scope → effective scope; `<effective-scope>/README.md` is the diagnostic route |
| `internal/lint` + template CI invocation | audit the normalized scope supplied by CLI | explicit template `--scope-root memory-bank-template`; downstream default `memory-bank` |
| `internal/agentinstructions` | generate the downstream managed routing block used by init/update | static downstream routes remain under `memory-bank/`; no template-source block is generated |
| release workflow | publish an approved compatible version after validation | exact commit → protected approval → immutable release tag |

## Architecture Coverage Decision

| Aspect | Status | Canonical owner / refs | Supporting view / artifact | Coverage note |
| --- | --- | --- | --- | --- |
| Components / responsibilities | covered | `SOL-01`–`SOL-04`, `C4-01` | compact C3 table | Source selection/translation, downstream planning, diagnostics and release responsibilities are separated. |
| Connectors / interactions | covered | `CTR-01`–`CTR-03` | compact C3 table | Git tree/filesystem reads, in-process namespace translation and profile-to-scope binding are explicit. |
| Configuration / topology | covered | `SOL-03`, `SOL-04`, `SD-02`, `SD-03` | none | Single binary topology remains; marker/profile and explicit lint scope select the source view. |
| Behavioral semantics | covered | `INV-01`–`INV-05`, `FM-01`–`FM-05` | none | Translation, rejection, defaulting and stop behavior are stated. |
| Quality / evolution concerns | covered | `TRD-01`–`TRD-03`, `RB-01`–`RB-04` | decision log | Bounded rollout compatibility, downstream regressions and release-before-rename ordering bound migration risk. |

## Selected Solution

- `SOL-01` Introduce distinct internal source-root and downstream-root concepts. During the coordinated rollout source inspection, dirty/ignored checks and pinned Git-tree reads select exactly one root: legacy `memory-bank/` or target `memory-bank-template/`; a source containing neither or both fails. Ownership classification, decisions, locks, mutations and init/update-generated downstream agent links continue to use `memory-bank/`.
- `SOL-02` Translate each accepted source entry by stripping exactly the selected source-root prefix and prefixing the unchanged relative suffix with `memory-bank/` before the payload map reaches the planner. Apply the same mapping to Git-backed production reads and filesystem-backed test reads.
- `SOL-03` Resolve doctor profile before choosing an implicit scope: auto/template uses `memory-bank-template` for a marker-valid template checkout, downstream uses `memory-bank`; an explicit `--scope-root` remains authoritative. Use that same effective scope for navigation, governance and the expected repository-local agent link `<effective-scope>/README.md`; keep downstream managed-block drift inspection conditional on an existing downstream lock.
- `SOL-04` Keep lint's existing public contract: downstream default remains `memory-bank`, while template CI/navigation invokes the existing explicit `--scope-root memory-bank-template`. Do not add a second lint profile detector.
- `SOL-05` Update CLI source guidance, architecture/use-case documentation and regression fixtures to name both contexts explicitly; publish through the existing validation and protected release workflow.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Globally rename `memory-bank/` constants and test paths to `memory-bank-template/` | Leaks a source-only name into downstream lock, classification, destination and navigation contracts, violating `REQ-03` and `CON-01`. |
| `ALT-02` | Switch immediately to the new source root | Release-before-rename ordering would make the newly released CLI unable to read the still-current official template. |
| `ALT-05` | Indefinite dual-root discovery/fallback | Issue #14 excludes indefinite dual-path support; unbounded fallback obscures retirement ownership. The selected bounded selector instead accepts one known root at a time and rejects a duplicate source. |
| `ALT-03` | Keep payload-map keys in source namespace and translate during every downstream operation | Spreads the bridge across classification, plan, lock and transaction code, increasing the chance that a source path escapes into downstream state. |
| `ALT-04` | Add `--profile` and auto-detection to lint | The existing `--scope-root` already gives template CI an explicit source target; a new public selector is not required by the acceptance outcome. |

## Trade-offs

| Trade-off ID | Decision | Benefit | Cost / Risk |
| --- | --- | --- | --- |
| `TRD-01` | Translate once at the source-reader boundary | Existing planner, ownership and lock semantics remain in one downstream namespace. | Source diagnostics must preserve enough context to report source-root errors clearly. |
| `TRD-02` | Accept either exact known source root only for the transition release | The release works with the official template before and after #63 changes its root, without changing the downstream namespace. | A separately reviewed removal release is required after #63; both roots in one source are rejected to avoid ambiguity. |
| `TRD-03` | Use explicit lint scope in template CI, but profile-derived default in doctor | Avoids inventing a new lint public contract while satisfying both repository contexts. | Template CI must pass the source scope explicitly for lint. |

## Accepted Local Decisions

- `SD-01` Internal terminology is `target source payload root` = `memory-bank-template/`, `legacy source payload root` = `memory-bank/` (transition input only), `downstream payload root` = `memory-bank/`, and `payload-relative path` = the suffix below the selected source root.
- `SD-02` Source root matching is exact and case-sensitive, consistent with Git tree paths. During the transition release only the two named roots are recognized, exactly one must be present, and no alternative spelling or discovery fallback is accepted. A later separately reviewed release removes the legacy root.
- `SD-03` An explicitly supplied normalized doctor/lint scope overrides defaults. Only an omitted doctor scope is derived from the resolved profile; lint continues to use its existing explicit scope mechanism. Doctor validates the repository-local agent entrypoint against the resulting scope, while `internal/agentinstructions.CurrentBlock` remains the unchanged downstream projection.
- `SD-04` The release version remains a protected workflow input selected by the maintainer. FT-014 records and verifies the selected immutable tag but does not choose an unstated version.

## Contracts

| Contract ID | Connector / direction | Roles and sync boundary | Guarantees / failure / evolution semantics |
| --- | --- | --- | --- |
| `CTR-01` | pinned Git/filesystem source `{memory-bank|memory-bank-template}/<suffix>` → ownership payload `memory-bank/<suffix>` | CLI initiates synchronous local read; source checkout provides committed blobs | select exactly one exact prefix then translate it; regular files and supported Git modes retain contents/modes; neither/both roots, dirty/ignored/unsupported source fails before planning. |
| `CTR-02` | translated payload → planner/lock/target filesystem | source reader provides downstream-keyed payload; planner consumes synchronously | keys are canonical downstream paths; source cannot supply `memory-bank/.lock`; existing ownership/conflict/atomicity semantics remain authoritative. |
| `CTR-03` | marker/profile/explicit scope → effective doctor/lint root and repository-local agent route | CLI preserves whether scope was explicit; doctor detects source identity; lint consumes normalized scope | explicit scope wins; implicit template doctor scope is `memory-bank-template`; implicit downstream scope and lint default are `memory-bank`; doctor expects `<effective-scope>/README.md` in the configured agent file; invalid scope fails without mutation. |

## Invariants

- `INV-01` No downstream ownership decision, lock entry, target mutation or init/update-generated managed agent route contains the source-only `memory-bank-template/` prefix. Template-repository diagnostics are the intentional exception: their local agent entrypoint routes to `memory-bank-template/README.md`.
- `INV-02` Every accepted source payload file maps one-to-one to the same suffix under downstream `memory-bank/`; contents and supported executable mode are preserved.
- `INV-03` `memory-bank/.lock`, ownership classes, conflict verdicts and transaction atomicity retain their existing semantics.
- `INV-04` Template marker detection still yields downstream when a downstream lock exists; profile identity and effective navigation scope cannot disagree when scope is implicit.
- `INV-05` Release publication occurs only after the exact candidate passes required validation and protected-environment approval.

## Failure Modes

- `FM-01` The pinned commit contains neither recognized source root or contains both roots. Fail source verification before a plan is produced; the latter is ambiguous.
- `FM-02` Source status/tree checks accidentally examine a non-selected root or the whole checkout but miss ignored/unsupported content under the selected root. Targeted fixtures must fail the change until both checks follow the selected root.
- `FM-03` Translation is omitted or applied twice, producing source-prefixed or `memory-bank/memory-bank-template/...` downstream keys. Validate every key at the boundary and assert reports/locks in regressions.
- `FM-04` Doctor resolves the template profile but audits downstream scope or expects the downstream agent route, or lint template CI uses the downstream default. Profile/scope/agent-route tests and the explicit template lint command must fail the candidate.
- `FM-05` Validation, approval or post-release install fails. Stop before publication where possible; never repoint an immutable tag, and do not unblock issue #63 without successful release evidence.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | Implement and validate translation | source/downstream contracts and targeted fixtures agree | revert unpublished code/docs; existing released CLI remains unchanged |
| `RB-02` | Publish compatible CLI release | all local/CI checks pass on exact commit and protected release approval is granted | before tag: stop and correct candidate; after tag: never repoint, publish a separately approved correction if needed |
| `RB-03` | Template-side rename handoff | clean-install evidence proves the released CLI supports FT-014 | keep issue #63 blocked and retain its old source tree until compatible release evidence exists |
| `RB-04` | Legacy-root retirement handoff | #63 has changed the template root and the transition release/compatibility matrix is recorded | create and publish a separately reviewed removal release; until it is published, legacy support remains deliberately bounded but active |

### Transition Compatibility Matrix

| CLI generation | Source checkout state | Expected result |
| --- | --- | --- |
| Existing release | legacy-only `memory-bank/` | Existing behavior; accepted. |
| Transition release | legacy-only `memory-bank/` | Accepted and translated to downstream `memory-bank/<suffix>`. |
| Transition release | new-only `memory-bank-template/` | Accepted and translated to downstream `memory-bank/<suffix>`. |
| Transition release | neither root or both roots | Rejected before planning; neither is missing and both is ambiguous. |
| Later strict release | new-only `memory-bank-template/` | Accepted; legacy support is removed only by that separately reviewed release. |

## Design Verification

| Analysis | Required | Reason / risk | Method | Result / evidence |
| --- | --- | --- | --- | --- |
| Contract compatibility | yes | Source path and template-local routing change while downstream contract must not. | Boundary inventory plus trace from source keys through planner/lock and from doctor profile through effective scope/agent route. | Grounding table, `SOL-01`–`SOL-04`, `CTR-01`–`CTR-03`, `INV-01`–`INV-04`; verified by `CHK-01`–`CHK-04`. |
| State / transition completeness | yes | Release-before-template-merge ordering has safe/unsafe states. | Walk through `RB-01`–`RB-04` and issue coordination gates. | The transition release accepts old/new single roots; failure retains the old template state; strict retirement is a later release. |
| Failure propagation | yes | Wrong-root or translation errors could mutate downstream paths. | Failure-mode review at pre-plan and pre-release boundaries. | `FM-01`–`FM-05` fail before mutation/publication or preserve the coordination block. |
| Concurrency / ordering | no | No new parallel writer or asynchronous runtime path; existing atomic transaction remains unchanged. | Existing transaction model inspection. | Not applicable beyond existing ownership transaction tests. |
| Security boundaries | yes | Source checkout trust, symlinks and path escape remain safety boundaries. | Verify exact prefix, Git mode and canonical downstream-key rules. | `CON-02`, `CTR-01`, `CTR-02`, `NEG-02`, `NEG-03` preserve current rejection rules. |
| Capacity / latency | no | The same bounded local payload is read once; no remote runtime or new scale dimension is introduced. | Change-surface review. | Not applicable. |
| Migration / evolution safety | yes | CLI release must precede removal of the old source root. | Compatibility matrix and rollout walk-through. | Transition CLI accepts new-only and legacy-only source; neither/both fail; a later strict release removes legacy after #63 coordination. |

## ADR / External Design Dependencies

`none`: the translation is feature-local and follows explicit cross-repository issue contracts. No reusable architecture policy is introduced.

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure / rollout refs |
| --- | --- | --- | --- |
| `REQ-01`, `REQ-02` | `SOL-01`, `SOL-02`, `C4-01`, `SD-01`, `SD-02` | `CTR-01`, `INV-01`, `INV-02` | `FM-01`–`FM-03`, `RB-01` |
| `REQ-03` | `SOL-02`, `TRD-01` | `CTR-02`, `INV-01`–`INV-03` | `FM-03`, `RB-01` |
| `REQ-04` | `SOL-03`, `SOL-04`, `SD-03`, `TRD-03` | `CTR-03`, `INV-04` | `FM-04` |
| `REQ-05`, `REQ-07` | `SOL-05` | `INV-01`–`INV-04` | `FM-01`–`FM-04`, `RB-01` |
| `REQ-06` | `SOL-05`, `SD-04` | `INV-05` | `FM-05`, `RB-02`–`RB-04` |
