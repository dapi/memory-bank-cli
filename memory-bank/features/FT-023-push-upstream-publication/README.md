---
title: "FT-023: Push Upstream Publication"
doc_kind: feature
doc_function: index
purpose: "Bootstrap-safe навигация по feature package issue #23. Сначала читать canonical brief; downstream routes добавляются только после готовности их owner-документов."
derived_from:
  - ../../flows/feature.md
  - brief.md
status: active
audience: humans_and_agents
---

# FT-023: Push Upstream Publication

## О разделе

Пакет ведёт [issue #23](https://github.com/dapi/memory-bank-cli/issues/23): безопасная публикация пригодных для upstream изменений из downstream Memory Bank через PR. `brief.md` владеет problem space и verify, `design.md` — accepted solution, а `implementation-plan.md` — grounded execution.

## Аннотированный индекс

- [brief.md](brief.md) — canonical problem space, validation-profile decision и verify contract.
- [design.md](design.md) — selected publication, transaction and Git/GitHub boundary design.
- [implementation-plan.md](implementation-plan.md) — archived execution and
  validation record; all checkpoints are closed.
- [decision-log.md](decision-log.md) — provenance и FPF-разбор blocking decisions; не владеет требованиями или selected solution.
- [feature-review-report.md](feature-review-report.md) — результат bounded review–improve.
