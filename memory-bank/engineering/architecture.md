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
  - ../../cmd/memory-bank-cli/main.go
  - ../../internal/cli/cli.go
status: active
---

# Architecture

## Runtime shape

One executable, `cmd/memory-bank-cli`, imports `internal/cli` and exits with its result. The module path is `github.com/dapi/memory-bank-cli`; the module declares Go 1.21 and depends on `golang.org/x/sys` and `gopkg.in/yaml.v3`.

```text
cmd/memory-bank-cli -> internal/cli
internal/cli -> internal/{ownership, doctor, lint, push, repository}
internal/doctor -> ownership + lint/governance inspection
internal/ownership -> Git source + local filesystem + memory-bank/.lock
internal/push -> ownership payload-path model + local/upstream Git + GitHub PR
```

## Module boundaries

| Package | Responsibility |
| --- | --- |
| `internal/cli` | Command dispatch, flags, usage, JSON/text output and exit codes. |
| `internal/ownership` | Template source validation, ownership classes, lock, plan and transactional apply. |
| `internal/doctor` | Profile-driven read-only findings for adoption, governance, drift and navigation. |
| `internal/lint` | Markdown parsing, navigation audit and reports. |
| `internal/push` | Lock-backed selection and compensating upstream branch/PR publication. |
| `internal/repository` | Explicit/nearest-Git repository-root resolution. |
| `internal/agentinstructions` | Managed block planning for one agent instruction file. |

## Quality attributes

- Safety: source/repository pinning, clean-ref verification, symlink-aware destination handling and atomic/rollback update paths.
- Contract stability: CLI output supports text and versioned JSON reports; tests assert public fields and exit behavior.
- Portability: Unix and Windows secure-path variants exist.
- Bounded external effects: init/update/doctor/lint remain local; `push`
  explicitly crosses the configured upstream Git and GitHub boundaries through
  a new branch and PR.

## Template payload and doctor profile detection

For `doctor --profile auto`, a repository is the template source when it has a
real root directory `template/` and no `memory-bank/.lock`. The lock takes
precedence and always selects downstream; a missing, non-directory or symlinked
source root also selects downstream. Explicit `--profile template` and
`--profile downstream` bypass this detector. The implicit documentation audit
scope for the template profile remains `template/memory-bank/`.

The canonical source payload is every tracked regular Git file below
`template/`. Its deterministic downstream path is the same relative suffix
after stripping `template/`, so `template/memory-bank/**` remains
`memory-bank/**` while dotfiles and arbitrary new nested paths require no code
change. The inverse mapping prefixes a lock-managed downstream path with
`template/` for `push`. Both directions share the ownership payload-path model.

Legacy source roots `memory-bank-template/` and `memory-bank/` remain migration
inputs and map below downstream `memory-bank/`. Canonical `template/` takes
precedence when it coexists with a project-local legacy tree. Ownership locks
may cover safe repository-relative paths outside `memory-bank/`; Git metadata,
the lock itself, symlinks and non-regular Git entries are rejected.

`memory-bank/` denotes payload data, not an executable name.
