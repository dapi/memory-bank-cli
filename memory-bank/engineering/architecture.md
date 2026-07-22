---
title: memory-bank-cli Architecture
doc_kind: engineering
doc_function: canonical
purpose: "Canonical owner архитектуры standalone Go CLI и package boundaries."
derived_from:
  - ../product/context.md
  - ../dna/governance.md
source_refs:
  - ../../go.mod
  - ../../cmd/mb-cli/main.go
  - ../../internal/cli/cli.go
status: active
---

# Architecture

## Runtime shape

One executable, `cmd/mb-cli`, imports `internal/cli` and exits with its result. The module path is `github.com/dapi/memory-bank-cli`; the module declares Go 1.21 and depends on `golang.org/x/sys` and `gopkg.in/yaml.v3`.

```text
cmd/mb-cli -> internal/cli
internal/cli -> internal/{ownership, doctor, lint, repository}
internal/doctor -> ownership + lint/governance inspection
internal/ownership -> Git source + local filesystem + memory-bank/.lock
```

## Module boundaries

| Package | Responsibility |
| --- | --- |
| `internal/cli` | Command dispatch, flags, usage, JSON/text output and exit codes. |
| `internal/ownership` | Template source validation, ownership classes, lock, plan and transactional apply. |
| `internal/doctor` | Profile-driven read-only findings for adoption, governance, drift and navigation. |
| `internal/lint` | Markdown parsing, navigation audit and reports. |
| `internal/repository` | Explicit/nearest-Git repository-root resolution. |
| `internal/agentinstructions` | Managed block planning for one agent instruction file. |

## Quality attributes

- Safety: source/repository pinning, clean-ref verification, symlink-aware destination handling and atomic/rollback update paths.
- Contract stability: CLI output supports text and versioned JSON reports; tests assert public fields and exit behavior.
- Portability: Unix and Windows secure-path variants exist.
- Local operation: the CLI operates on filesystem and Git checkout inputs; no remote runtime is required by the visible implementation.

## Doctor template-profile marker

For `doctor --profile auto`, a repository is the template source only when it has a regular root file named `.memory-bank-template` whose complete UTF-8 content is exactly `memory-bank-template-v1` followed by one newline. `memory-bank/.lock` takes precedence and always selects downstream; a missing, unreadable, non-regular, differently named, or differently formatted marker also selects downstream. Explicit `--profile template` and `--profile downstream` bypass this detector.

The marker is source-repository metadata, never part of the copied `memory-bank/` payload. The `dapi/memory-bank` template repository owns adding and documenting it through issue #52.

`memory-bank/` denotes payload data, not an executable name.
