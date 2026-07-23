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

1. The actor runs `memory-bank-cli update`, optionally first with `--dry-run`.
2. The CLI validates the source and reads the existing lock.
3. It classifies changes and produces decisions for managed, adapted, user-owned and generated content.
4. If the plan has no conflict, mutations and lock changes are applied atomically.

## Outcomes

User-owned content is not overwritten/deleted. A conflict yields a failure result instead of a successful update; a dry run leaves files unchanged.
