---
title: memory-bank-cli Context Map
doc_kind: domain
doc_function: canonical
purpose: "Canonical owner bounded contexts and boundary contracts."
derived_from:
  - model.md
status: active
---

# Context Map

| Bounded context | Responsibility | Boundary / collaborator |
| --- | --- | --- |
| CLI orchestration | Parse commands/flags, select operation, render text/JSON and map outcomes to exit codes. | `internal/cli` calls other internal packages. |
| Template ownership | Pin source/repository, classify content, plan/update lock and safely mutate files. | Git checkout and filesystem. |
| Documentation audit | Parse Markdown/navigation and report structural findings. | Repository-local Memory Bank scope. |
| Governance diagnosis | Combine profile, lock/adoption, governance, drift and navigation findings. | Ownership and lint concepts; read-only output. |

All contexts execute in one local Go process; no network service contract is confirmed.
