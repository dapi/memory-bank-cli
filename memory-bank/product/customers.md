---
title: memory-bank-cli Users and Jobs
doc_kind: product
doc_function: canonical
purpose: "Canonical owner подтверждённых пользователей и jobs to be done."
derived_from:
  - context.md
status: active
---

# Users and Jobs

| Actor | Job to be done | Confirmed interface |
| --- | --- | --- |
| Repository maintainer | Принять или установить версионированный шаблон Memory Bank в repository. | `memory-bank-cli init` |
| Repository maintainer | Безопасно обновить ранее adopted template, сохраняя локальную ownership classification. | `memory-bank-cli update` |
| Contributor / automation | Проверить navigation и governance документации до принятия изменений. | `memory-bank-cli lint`, `memory-bank-cli doctor` |

Это operational roles, выведенные из command help и тестируемых flows. Персоны, purchasing users и их приоритеты в источниках не описаны.
