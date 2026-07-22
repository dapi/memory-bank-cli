---
title: memory-bank-cli Metrics
doc_kind: product
doc_function: canonical
purpose: "Canonical owner известных измеримых product и quality signals."
derived_from:
  - context.md
  - ../engineering/testing-policy.md
status: active
---

# Metrics

Источники не задают business KPI, baseline, target или telemetry. Поэтому они не утверждаются как факты.

| Signal | What it indicates | Evidence available |
| --- | --- | --- |
| Command exit status | Success, usage failure или detected audit/conflict failure. | CLI implementation and regression tests |
| Versioned JSON report fields | Stability of automation-facing report contracts. | `internal/cli` and `internal/doctor` tests |
| Go test and vet result | Regression/static quality gate for code changes. | FT-001 verification contract |

## Open Questions

- Какие adoption, update-success, lint/doctor finding или release metrics должны собираться и кем?
