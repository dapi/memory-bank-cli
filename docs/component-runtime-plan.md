# Component and document runtime plan

This is CLI #62's implementation plan for W2, owned by memory-bank-cli. It imports
[CTR-01](https://github.com/dapi/memory-bank/blob/feat/141-component-adoption/docs/components.md)
and the parent issue's acceptance contract. The template feature owns payload declarations,
base templates and flow wrappers; this file owns CLI implementation sequencing only.

Status: candidate. Implementation waits for a clean review of CTR-01/ADR-002, this plan, and
the bridge checkpoint. Bridge candidate commit: a0811c4 (full SHA recorded with binary evidence
once independent review completes). No release or live downstream migration is included.

## Grounding and boundaries

Grounded CLI baseline ac7101c307e65566787bdb32a1bdad40b9a8b995 plus W1 a0811c4:

- ownership/source.go verifies pinned regular Git blobs; W1's source_format.go classifies them.
- ownership/update.go owns run, buildPlan, mutation preconditions and applyAtomicallyPinned.
- ownership/lock.go reads schema 0/1 and validates ownership/digests/modes.
- ownership/resolution_plan.go revalidates sources in PlanPull and ApplyResolutionPlan.
- doctor/governance.go currently applies metadata and feature lifecycle checks type-wide.
- cli/cli.go dispatches init/pull/doctor/lint and keeps update reserved for the executable.
- agentinstructions/block.go preserves bytes outside the managed instruction block.

Add a pure internal/contracts package for manifest/rule decoding and frozen validation semantics.
It must not import ownership, doctor or CLI. Ownership uses contracts during transaction preflight;
doctor/lint use the same validator, avoiding divergent activation rules. The existing legacy
validator remains available for legacy installations. Schema-2 installations use explicit adoption.

## Steps

1. **Composition** — internal/contracts/manifest.go and ownership/components.go validate the whole
   source inventory, resolve preset/adapter dependencies, filter payload and compose README/AGENTS.
   Extend Lock/Options and strict lock decoding for schema 2 without upgrading ordinary legacy
   source pulls. Unknown fields/paths/capabilities, missing inventory, cycles and removals fail
   before writes. Component-aware PlanPull and ApplyResolutionPlan are required in W2:
   both use the same composed transaction plan and validate selection, state and current source.
   Apply regenerates the entire plan before writes, so a saved pre-bridge plan receives no
   exemption. A matching-identity old-format plan against a component source is a direct
   negative fixture. Resolution planning is complete only after successful component preview,
   application, stale-file/lock/source rejection and no-mutation tests pass.
2. **Frozen contracts** — internal/contracts/rules.go and engine.go decode base types and immutable
   bundles. Each bundle embeds DNA, base and extension rules, plus the exact engine ID/digest.
   The trusted implementation embeds its immutable behavior artifact and a positive/negative
   corpus. Published bundles and engine behavior are never modified in place. A live DNA/base
   change cannot affect an adopted document's verdict. Missing or changed required bundles fail
   before payload mutation. Extensions cannot weaken their embedded base requirements.
3. **Adoption integrity** — ownership/adoption.go owns the registry snapshot, identity/path/type
   reconciliation and checksum binding to lock. Scan regular Markdown only under memory-bank/ for orphan projection
   fields, excluding .repo, dna, flows, templates, document-types, prompts and declared managed
   template assets. Do not inspect unrelated Markdown outside that root. Reject unsafe
   symlinks/aliases rather than following them. Recorded targets must resolve exactly once. Base documents have no record or marker.
   Schema-2 full/legacy always have a registry, even when empty; schema-0/1 installations
   remain on the old validator and have no registry requirement before explicit migration.
   Missing/corrupt schema-2 state is an error, never
   an invitation to recreate it. Read/validate both source and installed contracts.
4. **Document operations** — ownership/documents.go and cli/documents.go implement create, adopt,
   transition and move using the same handle-relative transaction engine. Report a dry-run plan;
   check applicable old/new gates and explicit evidence references; commit document, registry,
   history and lock together. Same adoption/move is idempotent. Unsupported detach/delete or
   transitions reject before writes. Base creation in full does not adopt implicitly. Legacy-flow
   creation explicitly resolves a per-document compatibility ID rather than adding to a selector.
5. **Legacy migration** — ownership/component_migration.go verifies the source-specific supported
   legacy map, ownership and local drift; freezes the prior document identities in selector
   snapshots; and previews the creation/validation semantics change. Apply requires both explicit
   --migrate-components and the current --migration-plan-digest. The digest binds old lock, source,
   resolution map, observed document bytes/modes and proposed mutations; deterministic snapshot
   identities/history ensure repeated preview is stable. Unsupported versions and incomplete,
   incompatible or ambiguous owner maps conflict. Migration preserves existing invalid legacy
   verdicts, while identity/integrity constraints still must hold. Selector transition atomically
   adds an exclusion plus a new per-document record, with no implicit precedence.
6. **Validation entrypoints** — doctor and lint check component state, dependencies, base documents,
   adoption and navigation; core/docs do not require absent Flows. Source-profile projection is
   explicitly distinguished from a downstream with missing lock. Preflight validates the resulting
   tree, including derived_from, Markdown paths, embedded frontmatter and priming manifest paths.
   Existing project-owned content is preserved by pull; scaffold ownership transfers on creation.
   Migration alone allows the same pre-existing legacy validation findings: compare multisets
   of stable identity, finding code, rule ID and subject before/after under the frozen engine.
   Any added or removed legacy finding, identity/integrity/path failure or new navigation violation blocks apply.
   Normal validation still reports the preserved errors. Test invalid legacy migration succeeds
   while an added violation fails without writes; ordinary pulls get no blanket exemption.

Selection precedence is CTR-01's contract: fresh init without flags chooses legacy; a
flagless schema-2 pull preserves the locked preset/components/adapters exactly. An explicit
preset resolves together with retained/new adapters and dependencies, then rejects removal
of any installed component. Adapter flags are additions, never replacement. No selection
flag opts a schema-0/1 installation into component migration. With a legacy source, any
component-selection flag is rejected before writes; it is never silently ignored. With a
component source and schema-0/1 lock, selection flags without explicit migration consent
also reject before writes. The flagless case rejects identically: any schema-0/1 lock plus
a component source requires --migrate-components, including unattended pull with no preset
or adapter flags. It never implicitly changes validation semantics. A direct flagless
fixture checks byte/mode/lock preservation. Direct negative fixtures cover both source formats and preserve
every downstream byte/mode and lock. Tests repeat flagless pulls for
every preset and adapter variant and assert unchanged selection and no automatic Flows.

Migration preview is `pull --migrate-components --dry-run --json` (with explicit source
inputs and optional --migration-resolution FILE). It writes no downstream state and returns
migration_plan_digest plus the exact proposed changes/semantics. Apply passes that digest
back as `pull --migrate-components --migration-plan-digest DIGEST` with the same source and
resolution input. Apply regenerates the preview and rejects changed observations or a stale
digest before writes. Neither unattended mode nor --preset legacy replaces this consent.

CLI flags: init/pull --preset NAME, repeated --adapter NAME; pull --migrate-components,
--migration-plan-digest DIGEST and --migration-resolution FILE. Document commands use --type,
--path, --contract, --to, --id (required for move), --dry-run and repeatable --evidence REF as applicable. An explicit
--legacy-flow chooses the installation's pinned compatibility contract. No adapter removal,
uninstall, contract composition, automatic adoption, arbitrary code execution or global service
is introduced. Exact serialized fields and encoding rules are owned by the shared
[CTR-01 wire format](https://github.com/dapi/memory-bank/blob/feat/141-component-adoption/docs/component-wire-format.md).
Go types and producer/consumer fixtures implement it; semantic changes return to design review.

## Verification and failure boundaries

Go contract/transaction fixtures cover every parent acceptance class: preset/default/adapter
matrix; repeated init/pull; docs-to-full and scaffold preservation; malicious paths and symlinks;
base vs adopted feature; registry/marker/identity/type/path tampering; frozen bundle/DNA/base/engine
drift and missing historical bundle; migration opt-in and stale digest; ambiguous moved documents
with valid/invalid owner maps; legacy selector exclusion plus rollback; fresh and migrated legacy
creation; unsupported transitions; failure during staged writes and concurrent changed lock.
Fixtures retain a byte/mode snapshot of the old tree and an external sentinel for negative paths.

Run `env -u GOROOT go test ./...`, `env -u GOROOT go vet ./...`, hermetic ownership E2E, and a
real component binary against the exact template candidate commit. Keep a separately built bridge
binary at the reviewed W1 commit; its real-binary fixture must reject the component candidate.
A pre-bridge binary is tested through the template's minimum-capability entrypoint. Direct
pre-bridge execution on a component source remains explicitly unsupported by the parent issue.

Before each implementation step, read its grounded owner files; update this plan when the exact
surface changes. Independent code-converge review uses a clean author commit, explicit baseline
and --max-cycles 0 so the run cannot fix, checkpoint or publish reviewed changes. Author fixes and
commits findings, then re-runs the review. Separate final code and simplification passes must be
clean. CLI and template PRs retain an explicit bridge-first release dependency; publication tags
are assigned by the release owner after review, not invented as already available binaries.

## Re-evaluated execution boundary

After five artifact review iterations, the implementation keeps one imported wire owner,
a bounded memory-bank document scan and an explicit digest-producing preview command.
The source-format bridge remains a separate delivery checkpoint; this plan does not
advertise component capability until its complete operation matrix is implemented.
Producer/consumer fixtures must cover CTR-01 selector grouping/IDs, context-root derivation,
canonical registry bytes and exact legacy finding multiset equality. Shared contract review
and this execution-plan review are separate gates; neither is assumed complete here.
