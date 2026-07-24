---
title: memory-bank-cli Configuration
doc_kind: ops
doc_function: canonical
purpose: "Canonical owner user-configurable CLI inputs and constraints."
derived_from:
  - development.md
  - ../domain/rules.md
status: active
---

# Configuration

| Area | Confirmed configuration |
| --- | --- |
| Repository selection | `--repo-root`; otherwise nearest Git root/current directory is resolved. |
| Documentation scope | `lint` and `doctor` use `--scope-root` (default `memory-bank`) and `--max-depth` (default 3). |
| Output | `lint`, `doctor`, `init`, `update` support `--json`; lint also has `--version`. |
| Template mutation | `init` requires `--source`, `--template-version`, `--source-ref`. `update` accepts the same reproducible override trio; without it, it fetches `main` from `memory-bank/.repo`'s clean `origin` or `https://github.com/dapi/memory-bank.git`, records its immutable SHA, and accepts `--dry-run` and `--agent-file`. |
| Doctor | `--profile` supports `auto`, `template`, `downstream`; `--agent-file` selects one checked instruction file. |

There is no confirmed environment-variable, secret, remote endpoint or persistent service configuration.
