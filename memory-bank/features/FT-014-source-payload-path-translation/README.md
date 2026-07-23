---
title: "FT-014: Source Payload Path Translation"
doc_kind: feature
doc_function: index
purpose: "Навигация по feature package issue #14. Сначала читать canonical brief, затем accepted design и execution plan."
derived_from:
  - ../../flows/feature.md
  - brief.md
status: active
audience: humans_and_agents
---

# FT-014: Source Payload Path Translation

## О разделе

Пакет ведёт [issue #14](https://github.com/dapi/memory-bank-cli/issues/14): CLI должен читать template-source payload из `memory-bank-template/`, но устанавливать, обновлять и диагностировать downstream payload как `memory-bank/`.

## Аннотированный индекс

- [brief.md](brief.md) — canonical problem space, scope, validation profile и verify contract.
- [design.md](design.md) — selected source/downstream translation, diagnostics semantics и release ordering.
- [implementation-plan.md](implementation-plan.md) — grounded execution sequencing, test strategy и evidence mapping.
- [decision-log.md](decision-log.md) — FPF rationale и provenance; canonical facts остаются в brief/design.
- [feature-review-report.md](feature-review-report.md) — bounded review-improve history и итоговый verdict.
