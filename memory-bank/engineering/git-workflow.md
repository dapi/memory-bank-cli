---
title: memory-bank-cli Git Workflow
doc_kind: engineering
doc_function: convention
purpose: "Canonical owner подтверждённых Git workflow facts."
derived_from:
  - architecture.md
  - ../features/FT-001-migrate-cli-source-and-rename-executable/brief.md
status: active
---

# Git Workflow

The product treats a template source as a clean Git checkout and requires a full commit SHA matching its HEAD. FT-001 records a history-preserving import of the upstream `tools/` subtree into this standalone repository.

The repository sources do not define branch naming, review requirements, commit format, merge strategy or CI enforcement. Do not infer these conventions.
