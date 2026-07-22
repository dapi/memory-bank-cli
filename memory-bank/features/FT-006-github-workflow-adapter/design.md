---
title: "FT-006: Opt-in GitHub Workflow Adapter Design"
doc_kind: feature
doc_function: canonical
purpose: "Выбранное feature-local решение для opt-in GitHub adapter: explicit CLI boundary, marker ownership и встроенные guidance assets."
derived_from:
  - brief.md
  - ../../engineering/architecture.md
  - "https://github.com/dapi/memory-bank/issues/40"
status: active
audience: humans_and_agents
must_not_define:
  - ft_006_scope
  - ft_006_acceptance_criteria
  - implementation_sequence
---

# FT-006: Design

## Selected Design

- `SOL-01` Add the explicit opt-in boundary `mb-cli github init|update`; do not alter existing generic `init` or `update` semantics.
- `SOL-02` Bundle the five GitHub assets in the executable: Small Change and Feature issue forms, PR template, validation configuration guidance and agent guidance. No GitHub API or runtime dependency is introduced.
- `SOL-03` Treat an unmarked existing destination as user-owned and preserve it. A managed block records its own content digest in paired markers; a changed or malformed managed block is a conflict and prevents any write. `--dry-run` returns the full plan without writing.

The upstream issue permits an opt-in command/config and says unmarked existing templates are user-owned. FPF B.5 applied: a distinct subcommand is the smallest hypothesis that keeps the generic ownership flow unchanged; its predicted isolation and safety behavior are covered by `CHK-01`–`CHK-03` before it is accepted.

## C4 Applicability Decision

`C4 required: no`. The change remains inside the single existing CLI container and adds no process, service, external API, storage or runtime topology. A component responsibility map is sufficient.

| Component | Responsibility | Connector |
| --- | --- | --- |
| `internal/cli` | Parse `github init|update`, resolve repo root, render text/JSON report and exit code. | In-process call to adapter. |
| `internal/githubadapter` | Plan marker-owned GitHub assets; apply only conflict-free plans. | Local filesystem below resolved repository root. |
| `.github/` | Optional downstream GitHub workflow surface. | Filesystem; no GitHub API call. |

## Architecture Coverage Decision

| Concern | Decision |
| --- | --- |
| Components | covered by `SOL-01`–`SOL-03` and map above |
| Connectors | covered: in-process call and local filesystem only |
| Configuration | covered: static embedded asset set and CLI flags |
| Behavioral semantics | covered by ownership state below |
| Quality/evolution | covered by marker digest, dry-run, unit/CLI tests |

## Ownership Contract

| `CTR-*` | State | Result |
| --- | --- | --- |
| `CTR-01` | Destination absent | create complete marker-managed asset |
| `CTR-02` | Destination exists without adapter marker | preserve as user-owned |
| `CTR-03` | One clean matching marker block | preserve if current; update if bundled content changes |
| `CTR-04` | Marker malformed or recorded digest differs from managed body | conflict; apply no mutations |

## Invariants, Failures, Rollback

- `INV-01` No command outside `mb-cli github` installs adapter assets.
- `INV-02` A user-owned unmarked `.github/` file is never overwritten.
- `FM-01` A symlink in an adapter destination path is rejected.
- `FM-02` Any conflict prevents all planned mutations.
- `RB-01` No delete operation exists. Backout is removal of adapter-managed files/blocks by the repository owner; a later run does not recreate an unmarked user-owned replacement without explicit confirmation.

## Design Verification

| Analysis | Required | Method / evidence |
| --- | --- | --- |
| Contract/behavior | yes | unit tests for create, dry-run, preservation, drift conflict and CLI JSON |
| Security/path safety | yes | unit test rejects symlink destination |
| C4/runtime topology | no | no new runtime boundary |
| Load/performance | no | bounded local files only; no new remote call |

## Design Pack

No delegated contract, diagram or ADR is required: the interface is compact and feature-local.
