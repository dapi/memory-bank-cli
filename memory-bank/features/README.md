---
title: Feature Packages Index
doc_kind: feature
doc_function: index
purpose: Навигация по instantiated feature packages. Читать, чтобы найти существующую delivery-единицу или понять, где создавать новую.
derived_from:
  - ../dna/governance.md
  - ../flows/feature.md
  - ../flows/feature-artifact-catalog.md
status: active
audience: humans_and_agents
---

# Feature Packages Index

Каталог `memory-bank/features/` хранит instantiated feature packages вида `FT-XXX/`.

## Instantiated Features

- [FT-001: Migrate CLI Source and Rename Executable](FT-001-migrate-cli-source-and-rename-executable/README.md) — active delivery package for standalone migration and sole `memory-bank-cli` executable identity; its brief is still marked `in_progress`.
- [FT-002: Doctor Template-Profile Detection](FT-002-doctor-template-profile-detection/README.md) — draft package for issue #2; blocked pending the explicit template-source marker contract and validation-profile confirmation.
- [FT-003: Establish memory-bank-cli Releases](FT-003-establish-memory-bank-cli-releases/README.md) — planned release-deployment package for CI, the approved `v1.0.0` publication and Go installation documentation.
- [FT-005: Downstream Smoke Tests and Compatibility Canaries](FT-005-downstream-smoke-tests-and-compatibility-canaries/README.md) — active delivery package for issue #5's blocking stable smoke and scheduled/manual compatibility canary.
- [FT-006: Opt-in GitHub Workflow Adapter](FT-006-github-workflow-adapter/README.md) — active delivery package for issue #6's opt-in, marker-owned GitHub workflow adapter.

## Rules

- Каждый package создается по правилам из [`../flows/feature.md`](../flows/feature.md).
- Optional problem, solution, execution и review artifacts выбираются по [`../flows/feature-artifact-catalog.md`](../flows/feature-artifact-catalog.md); каталог является меню, а не checklist.
- Bootstrap package начинается с `README.md` и `brief.md`; после `Problem Ready` в него добавляется `design.md`, если `brief.md` фиксирует `Design required: yes`; `implementation-plan.md` появляется после готовности нужных upstream owners.
- Для bootstrap и downstream-документов используй шаблоны из [`../flows/templates/feature/`](../flows/templates/feature/).
- Если работа требует roadmap, risk register и нескольких delivery subissues, сначала создай или обнови epic package в [`../epics/README.md`](../epics/README.md).
- По умолчанию feature ссылается на общий product context из [`../product/context.md`](../product/context.md), а при изменении предметных правил также на соответствующие документы из [`../domain/README.md`](../domain/README.md).
- Если feature реализует или существенно меняет устойчивый сценарий проекта, она должна ссылаться на соответствующий `UC-*` из [`../use-cases/README.md`](../use-cases/README.md).
- В шаблонном репозитории этот каталог может быть пустым. Это нормально.

## Naming

- Базовый формат: `FT-XXX/`
- Вместо `XXX` используй идентификатор, принятый в проекте: issue id, ticket id или другой стабильный ключ
- Один package = одна delivery-единица
