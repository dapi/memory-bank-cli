---
title: "UC-003: Audit Memory Bank Documentation"
doc_kind: use_case
doc_function: canonical
purpose: "Canonical owner устойчивого lint and doctor audit scenario."
derived_from:
  - ../product/customers.md
  - ../domain/rules.md
  - ../ops/config.md
status: active
audience: humans_and_agents
---

# UC-003: Audit Memory Bank Documentation

**Primary actor:** contributor or automation.

## Trigger and Preconditions

The actor wants to inspect a repository-local documentation tree. Lint defaults to `memory-bank`; doctor derives `memory-bank-template` for a detected template source and `memory-bank` for a downstream repository unless a valid repository-relative scope is supplied explicitly.

## Main Flow

1. The actor invokes `memory-bank-cli lint` for navigation integrity or `memory-bank-cli doctor` for broader diagnosis.
2. The CLI resolves the repository, profile and effective normalized scope. Template CI invokes lint with explicit `--scope-root memory-bank-template`.
3. It emits text or JSON findings and summary.

## Outcomes

Lint reports Markdown navigation integrity. Doctor reports adoption, governance, managed drift and navigation without mutating the worktree. Error-level findings produce a failure result.
