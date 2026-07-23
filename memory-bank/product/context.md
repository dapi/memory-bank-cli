---
title: memory-bank-cli Product Context
doc_kind: product
doc_function: canonical
purpose: "Canonical owner product problem, capabilities и границ memory-bank-cli."
derived_from:
  - ../README.md
source_refs:
  - ../go.mod
  - ../cmd/memory-bank-cli/main.go
  - ../internal/cli/cli.go
status: active
audience: humans_and_agents
---

# Product Context

## Problem

Проекту, который принимает шаблон Memory Bank, нужны безопасная установка и обновление шаблона, а также проверяемая диагностика состояния документации. Ручное копирование не даёт versioned ownership, не различает локальную адаптацию и upstream-managed content и не предоставляет единый audit navigation/governance.

## Product

`memory-bank-cli` — единственная публичная команда проекта. Она предоставляет четыре подтверждённые capability:

- `init` — принимает clean Git checkout шаблона и создаёт/adopts `memory-bank/` вместе с ownership lock;
- `update` — строит и атомарно применяет безопасный update plan по существующему lock;
- `doctor` — read-only диагностирует adoption, governance, managed drift и navigation;
- `lint` — проверяет markdown navigation integrity в Memory Bank scope.

Подробные контракты терминов и правил принадлежат [domain/README.md](../domain/README.md), а реализация — [engineering/architecture.md](../engineering/architecture.md).

## Boundaries

Продукт управляет repository-local Memory Bank content и одним repository-relative agent instruction file. Он не является хостингом документации, редактором Markdown или release service. Source template обязан быть Git checkout; публикация релизов и install instructions не подтверждены источниками этого репозитория.

## Open Questions

- Кто является первичным внешним customer segment и какие каналы распространения CLI приняты?
- Какие production/business metrics, support process и compatibility policy существуют?
