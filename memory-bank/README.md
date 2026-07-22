---
title: memory-bank-cli Documentation Index
doc_kind: project
doc_function: index
purpose: "Корневая навигация по подтверждённым знаниям проекта memory-bank-cli."
derived_from:
  - dna/principles.md
  - ../README.md
status: active
audience: humans_and_agents
---

# memory-bank-cli Documentation Index

`memory-bank-cli` — Go CLI для установки, обновления, проверки и диагностики шаблонов Memory Bank. Корневое описание продукта находится в [product/context.md](product/context.md); этот индекс ведёт к единственному owner каждого устойчивого факта.

## Разделы

- [Product](product/README.md) — продуктовая цель, пользователи, outcomes, позиционирование и известные неопределённости.
- [Domain](domain/README.md) — язык шаблонов, ownership lock, диагностики и правила безопасного обновления.
- [PRD](prd/README.md) — первая продуктовая инициатива: самостоятельный CLI `mb-cli`.
- [Use cases](use-cases/README.md) — устойчивые flows установки, обновления и аудита документации.
- [Engineering](engineering/README.md) — Go-архитектура, package boundaries, quality attributes и тестирование.
- [Operations](ops/README.md) — локальная разработка, конфигурация и ограниченная информация о сборке/release.
- [Features](features/README.md) — delivery package FT-001, сохранившийся из репозитория.
- [ADRs](adr/README.md) — реестр; отдельных документированных ADR пока нет.
- [DNA](dna/README.md) — правила владения документацией и её lifecycle.
- [Flows](flows/README.md) — reusable workflows и templates Memory Bank.
- [Epics](epics/README.md) — место для крупных инициатив; instantiated epics отсутствуют.
- [Prompts](prompts/README.md) — reusable prompt templates.

## Source Basis

Адаптация опирается на корневой [README](../README.md), `go.mod`, `.goreleaser.yml`, исходники и тесты в `cmd/` и `internal/`, а также существующий [FT-001](features/FT-001-migrate-cli-source-and-rename-executable/README.md). В репозитории нет project-level `docs/`, `AGENTS.md` или `CLAUDE.md`; `AGENTS.md` в testdata — fixture, а не инструкция проекта.
