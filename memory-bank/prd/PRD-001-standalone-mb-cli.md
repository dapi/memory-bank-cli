---
title: "PRD-001: Standalone mb-cli"
doc_kind: prd
doc_function: canonical
purpose: "Фиксирует продуктовую инициативу самостоятельного CLI с единственной публичной командой mb-cli."
derived_from:
  - ../product/context.md
  - ../domain/rules.md
  - ../features/FT-001-migrate-cli-source-and-rename-executable/brief.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_verify_contract
---

# PRD-001: Standalone `mb-cli`

## Problem

The existing CLI source was historically located in the Memory Bank repository and exposed more than one executable identity. Consumers need one standalone Go module and one unambiguous command without losing the established `lint`, `doctor`, `init` and `update` behavior.

## Users and Jobs

| User | Job to be done | Current pain |
| --- | --- | --- |
| Repository maintainer | Run Memory Bank template installation, update and diagnostics from a standalone CLI. | A split source location and multiple executable identities make ownership and invocation ambiguous. |
| Contributor / automation | Continue to validate documentation and receive existing command/JSON outcomes. | A rename/migration can accidentally break scripts or audit contracts. |

## Goals

- `G-01` Provide the Go module `github.com/dapi/memory-bank-cli` and one public executable identity, `mb-cli`.
- `G-02` Preserve supported command semantics, exit codes and versioned JSON contracts for `lint`, `doctor`, `init` and `update`.
- `G-03` Ensure configured release builds do not emit an old executable identity.

## Non-Goals

- `NG-01` Do not redesign `doctor --profile auto` source-template detection.
- `NG-02` Do not promise a release publication, tag, installation guide, aliases or compatibility wrappers.
- `NG-03` Do not change the semantics of the supported commands as part of this initiative.

## Product Scope

### In scope

- Standalone command identity and module ownership for the existing CLI capability.
- A single user-facing executable name in help, diagnostics, examples and build configuration.
- Continued access to existing template ownership, update, audit and diagnosis capabilities.

### Out of scope

- New product capabilities, hosted services or UI.
- A compatibility path for removed executable identities.
- Release publication and installation evidence.

## Business Rules

- `mb-cli` is the only public executable identity.
- The `memory-bank/` path remains domain payload data and is not itself an executable identity.
- Existing command/JSON and exit-code contracts are preserved.

## Success Metrics

| Metric | Baseline | Target | Measurement |
| --- | --- | --- | --- |
| Module location | CLI under upstream `tools/` module | standalone `github.com/dapi/memory-bank-cli` module | inspect `go.mod` and source tree |
| Executable identities | legacy identities before initiative | only `mb-cli` configured/built | source/release configuration inspection |
| Supported command contract | existing commands | preserved `lint`, `doctor`, `init`, `update` contracts | regression suite |

## Risks and Open Questions

- `RISK-01`: A name migration can mistakenly alter the `memory-bank/` payload path or contract behavior.
- `RISK-02`: Release configuration can retain an old artifact despite source cleanup.
- `OQ-01`: What are the approved tag/publication/installation procedures and ownership for the deferred release work?
- `OQ-02`: Is the FT-001 `in_progress` lifecycle status still current now that this checkout contains the migrated implementation? The repository has no completion evidence update.

## Downstream Features

| Feature | Why it exists | Status |
| --- | --- | --- |
| [FT-001](../features/FT-001-migrate-cli-source-and-rename-executable/README.md) | Migrate source and make `mb-cli` the only public executable. | in_progress according to its brief |
