---
title: "UC-002: Safely Update a Memory Bank Template"
doc_kind: use_case
doc_function: canonical
purpose: "Canonical owner устойчивого сценария conflict-aware template update."
derived_from:
  - UC-001-adopt-template.md
  - ../domain/rules.md
  - ../ops/config.md
status: active
audience: humans_and_agents
---

# UC-002: Safely Update a Memory Bank Template

**Primary actor:** repository maintainer.

## Trigger and Preconditions

An adopted repository has an ownership lock; a clean, pinned template source and metadata are supplied.

## Main Flow

1. The actor runs `memory-bank-cli pull`, optionally first with `--dry-run`. During the coordinated payload rename, the pinned source contains exactly one recognized root—legacy `memory-bank/`, legacy `memory-bank-template/`, or target `template/`—while the existing locked destination stays repository-relative.
2. The CLI validates the source and reads the existing lock.
3. It classifies changes and produces decisions for managed, adapted, user-owned and generated content.
4. If the plan has no conflict, mutations and lock changes are applied atomically.

## Reviewed Conflict Flow

1. When ordinary pull reports a two-sided adapted conflict, the actor runs
   `memory-bank-cli pull --plan <file>`.
2. The CLI emits the complete affected-path inventory without changing payload
   or `.lock`. For an eligible mechanical merge it verifies the old template
   blob from `.lock.template.source_ref` and includes the exact candidate.
3. An agent may prepare an analysis, but the reviewer explicitly selects an
   allowed action for every required decision.
4. The actor runs `memory-bank-cli pull --apply-plan <file>`.
5. The CLI regenerates deterministic context, rejects stale or altered input,
   and atomically commits all accepted resolutions, safe updates and `.lock`.
6. Repeating ordinary pull against the same source is a no-op.

## Outcomes

User-owned content is not overwritten or deleted. Ordinary conflicts still
yield a failure. A reviewed complete plan can resolve adapted conflicts without
partial application; any rejected plan and every dry run leave files unchanged.
