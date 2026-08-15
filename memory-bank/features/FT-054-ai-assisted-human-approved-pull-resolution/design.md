---
title: "FT-054: Design"
doc_kind: feature
doc_function: canonical
purpose: "Selected trusted-local plan/apply and Git-backed merge design for FT-054."
derived_from:
  - brief.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_054_scope
  - ft_054_acceptance_criteria
  - ft_054_evidence_contract
  - implementation_sequence
---

# FT-054: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` |

## Context

The CLI already has a conservative ownership planner, strict `.lock` decoder,
pinned Git source verification and an atomic payload-plus-lock transaction. The
selected solution adds a review carrier around those capabilities. It does not
add a second writer, external authorization service or compatibility registry.

The user approved this trusted-local model on 2026-08-15: software validates
that the plan still describes the exact current state and deterministic result;
the surrounding human/agent workflow owns who reviewed the semantic choice.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-01` | `C3 compact view` | Planning, Git history, strict decoding and atomic apply collaborate inside one CLI process. | Component table below |

### Component View

| Component | Responsibility | Connector |
| --- | --- | --- |
| Pull orchestrator | Select ordinary pull, plan generation or reviewed apply. | CLI flags -> planner/apply coordinator. |
| Resolution planner | Derive complete affected-path context and allowed actions without mutation. | Pinned source + `.lock` + local payload -> JSON plan. |
| Historical base reader | Resolve the old payload root/path at `.lock.template.source_ref`, read the blob and verify digest/mode. | Git object database -> verified base or unavailable reason. |
| Mechanical merger | Produce one deterministic line merge or declare overlap/unavailability. | Verified base + local + upstream -> candidate bytes/mode. |
| Strict plan codec | Bound and decode one versioned full plan; reject unknown fields and invalid paths/actions. | JSON file -> typed plan. |
| Apply coordinator | Regenerate deterministic context, validate selected overlays and stage final ownership mutations. | Reviewed plan + current state -> existing transaction. |
| Ownership transaction | Commit payload and `.lock` together or roll back both. | Staged mutations -> filesystem. |

## Architecture Coverage Decision

| Aspect | Status | Coverage note |
| --- | --- | --- |
| Components | covered | Each planner, Git, merge, codec and transaction responsibility is isolated. |
| Connectors | covered | All connectors are synchronous local reads or the existing atomic filesystem transaction. |
| Configuration | covered | Existing source selection and remote configuration are reused; no new credential/config surface. |
| Behavioral semantics | covered | Plan generation, resolution actions, regeneration and commit are explicit below. |
| Quality/evolution | covered | Versioned strict plan format, immutable source refs and default-pull compatibility bound evolution. |

## Selected Solution

- `SOL-01` Add `pull --plan FILE`. It resolves the same target source as ordinary
  pull, builds the complete full-scope ownership plan and writes a versioned
  JSON carrier. Every entry records path, ownership, base/local/upstream
  identity, deterministic proposal/reason, allowed actions and selected action.
  Planning does not run the transaction writer.
- `SOL-02` Add `pull --apply-plan FILE`. It strictly decodes the carrier,
  resolves the current target source, regenerates the full deterministic plan,
  compares all non-decision fields exactly, validates each selected action and
  then invokes the existing ownership transaction. Selected actions and the
  exact reviewed merge payload are the only reviewer-editable overlays.
- `SOL-03` Recover historical base bytes from the Git commit already stored in
  `.lock.template.source_ref`. Select the payload root at that commit using the
  existing canonical/legacy root rules, map the downstream path back into that
  tree, read only the Git blob, and require its SHA-256 and mode to match the
  lock entry. A missing commit/path or mismatch makes merge unavailable.
- `SOL-04` Implement `mbc-diff3-lines-v1`: accept UTF-8 text, preserve exact
  line bytes, compute base-to-local and base-to-upstream replacement hunks and
  reject overlapping changes unless both sides produce identical replacement
  bytes. Non-overlapping changes are applied in base order. Mode resolution is
  available when at most one side changed mode from base, or both selected the
  same mode.
- `SOL-05` For a reviewed conflict: `keep-local` retains local bytes/mode and
  advances an adapted base to current upstream; `take-upstream` writes upstream
  bytes/mode and adopts current source ownership; `apply-reviewed-merge` writes
  the recomputed candidate, retains `adapted`, and advances base to upstream.
  Merge is unavailable when either side is absent.
- `SOL-06` Full reviewed apply includes every affected path. Deterministic
  managed changes apply with resolutions. A user-owned path removed upstream is
  kept and detached. Agent-instruction mutation must already be current because
  it is outside the source/lock plan identity.
- `SOL-07` The plan carrier is trusted-local, not an authorization credential.
  Git review, an interactive user instruction or an equivalent surrounding
  process records approval. CLI guarantees deterministic context and atomic
  consequences, not reviewer identity.

### Plan Shape

The format is strict JSON with `format_version`, target `template`, prior
`base_template`, serialized `lock_digest`, and sorted `entries`. An entry has
present/absent identities for base/local/upstream, canonical reason/action,
allowed/selected resolution actions and an optional merge candidate containing
algorithm, base64 result, digest and mode. Paths are normalized repository
relative slash paths and may not target `.git` or `.lock`.

### State / Action Contract

| State | Allowed result | Resulting ownership/base |
| --- | --- | --- |
| Deterministic managed change | proposed action | Existing planner semantics. |
| Two-sided adapted or canonical-migration conflict, both present | keep, take, merge when eligible | `adapted` with new upstream base; source ownership for take; `adapted` with new upstream base for merge. |
| Managed local drift with changed upstream | keep or take | `adapted` with upstream base for keep; `managed` for take. |
| Local present, upstream absent conflict | keep or take; no merge | Keep preserves and detaches; take deletes and detaches. |
| Local absent, upstream present conflict | take only; no merge | Lock v1 cannot retain a reviewed local-deletion baseline, so take restores upstream and advances ownership. |
| User-owned path removed upstream | keep-and-detach | Local content unchanged; lock entry removed. |

## Alternatives Considered

| Alternative | Decision | Reason |
| --- | --- | --- |
| Automatic semantic merge | rejected | Transfers project-specific decisions to automation. |
| Digest-only/manual `.lock` surgery | rejected | Cannot prove base bytes and creates unsupported state. |
| Protected sidecar plus external receipt registry | rejected | Solves reviewer-identity and replay threats outside the approved trusted-local model while adding a second atomicity domain. |
| Store historical snapshots in `.lock` | rejected | Breaks strict v1 rollback compatibility and duplicates immutable Git data. |
| Restore base from old Git `source_ref` | selected | Existing lock identity and Git blob verification provide the required merge base without new persistent state. |
| Interactive-only `--ask` extension | rejected as sole interface | Cannot be committed, reviewed asynchronously or prepared by an agent. |

## Accepted Local Decisions

- `SD-01` Plan format v1 is full-scope and strict; partial scope is deferred.
- `SD-02` Reviewer-selected actions are plan overlays validated against allowed
  actions; they are not cryptographic authorization.
- `SD-03` Historical Git objects are authoritative only after digest/mode match
  with `.lock`; otherwise merge is unavailable.
- `SD-04` Candidate result bytes are base64-encoded and recomputed during apply.
- `SD-05` Keep ordinary pull conservative and reuse its source resolver.

## Contracts

| Contract | Parties | Failure behavior |
| --- | --- | --- |
| `CTR-01` JSON resolution plan | CLI -> agent/reviewer -> CLI | Strict decode; unknown/duplicate/invalid/oversized input rejects before planning mutations. |
| `CTR-02` Historical base lookup | `.lock` + Git source -> merger | Missing or mismatched data disables merge and never falls back to plan-supplied bytes. |
| `CTR-03` Reviewed apply | Plan + regenerated state -> ownership transaction | Any deterministic mismatch or unresolved entry rejects before staging. |

## Invariants

- `INV-01` Plan generation never changes payload, agent instructions or `.lock`.
- `INV-02` Apply mutates only after complete strict decode, state regeneration
  and exact deterministic comparison.
- `INV-03` Every affected path appears exactly once and every required human
  decision has one currently allowed selected action.
- `INV-04` Merge result bytes and executable mode equal the deterministic
  recomputation from verified Git base, current local and current upstream.
- `INV-05` Payload resolutions and the resulting `.lock` commit together.
- `INV-06` Ordinary pull behavior does not consume or imply a reviewed plan.
- `INV-07` A successful full apply advances enough ownership/base state that an
  immediate ordinary pull against the same source is a no-op.

## Failure Modes

- `FM-01` Invalid schema/action/path/base64/UTF-8 -> reject without mutation.
- `FM-02` Old source commit/blob absent or digest/mode mismatch -> merge
  unavailable; keep/take remain available when representable.
- `FM-03` Overlapping content or incompatible mode changes -> merge unavailable.
- `FM-04` Source, lock, local or upstream changed after review -> stale-plan
  rejection before staging.
- `FM-05` Transaction write failure -> existing rollback restores payload and
  lock.
- `FM-06` Agent instructions require an unplanned change -> reject reviewed
  apply and require ordinary pull first.

## Rollout / Backout

- `RB-01` Ship as an opt-in minor release. Existing locks and default commands
  remain valid for older binaries. Backout is reinstalling `v2.0.1`; no new
  persistent control file exists.

## Design Verification

| Analysis | Required | Method | Result |
| --- | --- | --- | --- |
| Contract compatibility | yes | CLI/codec and default-pull regression matrix | Covered by versioned opt-in flags and strict tests. |
| State completeness | yes | State/action table and present/absent tests | Covered; merge restricted to verified present triples. |
| Failure propagation | yes | Stale/tamper/injected transaction tests | All failures precede staging or roll back. |
| Concurrency/ordering | yes | Existing transaction preconditions plus final regeneration | Cooperating transaction remains atomic; observed external drift rejects. |
| Security boundaries | yes | Trusted-local threat-model review | CLI validates integrity but does not claim reviewer identity. |
| Capacity/latency | yes | Bounded plan decode and line-merge limits | Plan and per-file merge inputs receive explicit limits. |
| Migration/evolution | yes | Legacy/canonical source-ref fixtures and older-binary backout | No lock schema or persistent sidecar change. |
