---
title: memory-bank-cli Vision
doc_kind: product
doc_function: canonical
purpose: "Canonical owner подтверждённого направления и non-goals продукта."
derived_from:
  - context.md
  - ../README.md
status: active
---

# Vision

Проект стремится сделать adoption и сопровождение Memory Bank проверяемыми CLI-операциями: состояние template provenance и ownership должно быть inspectable, а обновление не должно молча затирать локально адаптированный или user-owned content.

## Product Principles

- Безопасность обновления важнее безусловного применения: conflicts сообщаются, а mutations применяются атомарно.
- Диагностика отделена от изменения: `doctor` не должен менять worktree.
- Machine-readable reports versioned alongside text output.
- `memory-bank-cli` остаётся единственной публичной executable identity.

## Non-Goals

- Не подтверждены GUI, hosted service, cloud synchronization или collaborative editing.
- Не следует выводить из кода обещания backward compatibility, support SLA или release cadence.
