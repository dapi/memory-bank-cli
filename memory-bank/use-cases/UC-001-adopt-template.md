---
title: "UC-001: Adopt a Memory Bank Template"
doc_kind: use_case
doc_function: canonical
purpose: "Canonical owner устойчивого сценария initial template adoption."
derived_from:
  - ../product/customers.md
  - ../domain/rules.md
  - ../ops/config.md
status: active
audience: humans_and_agents
---

# UC-001: Adopt a Memory Bank Template

**Primary actor:** repository maintainer.

## Trigger and Preconditions

The maintainer needs to adopt a template. A separate clean Git source checkout, its matching full commit SHA, a template version, and a target repository are available.

## Main Flow

1. The actor runs `mb-cli init` with source, template version and source ref (optionally using `--dry-run`).
2. The CLI validates source/repository boundaries and pins the source.
3. It builds ownership decisions for template payload and configured agent instruction file.
4. On a non-dry successful run, it applies the plan atomically and records `memory-bank/.lock`.

## Outcomes

The repository has an adopted Memory Bank and lock, or receives an error/conflict without a successful partial update. Exact option spelling and output fields are owned by code.
