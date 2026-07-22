---
title: memory-bank-cli Glossary
doc_kind: domain
doc_function: canonical
purpose: "Canonical owner терминов предметной области CLI."
derived_from:
  - ../product/context.md
status: active
---

# Glossary

| Term | Meaning |
| --- | --- |
| Memory Bank | Documentation tree managed or audited by the CLI; default scope is `memory-bank`. |
| Template source | Clean Git checkout containing a `memory-bank/` payload, pinned to an expected full commit SHA. |
| Downstream repository | Repository into which a template is adopted or updated. |
| Ownership lock | `memory-bank/.lock` record containing template identity and per-file ownership/digests. |
| Managed | File class whose content is controlled by the template. |
| Adapted | File class with an upstream base and downstream customization, requiring conflict-aware updates. |
| User-owned | File class never overwritten or deleted by template update. |
| Generated | File class regenerated from the template. |
| Update plan | Complete proposed set of ownership decisions/mutations before application. |
| Doctor | Read-only diagnosis of adoption, governance, managed drift and navigation. |
| Lint | Audit of Markdown navigation integrity for a selected scope. |
