---
title: "UC-004: Publish Managed Changes Upstream"
doc_kind: use_case
doc_function: canonical
purpose: "Canonical owner of the stable managed-change upstream publication scenario."
derived_from:
  - ../flows/use-case.md
  - ../product/context.md
  - ../domain/rules.md
status: active
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_test_matrix
---

# UC-004: Publish Managed Changes Upstream

## Goal

Propose reusable downstream Memory Bank changes to the configured upstream
repository without publishing project-specific state or writing directly to
the upstream default branch.

## Primary Actor

Repository maintainer.

## Trigger

The maintainer has validated reusable changes in a downstream Memory Bank and
wants to propose them upstream.

## Preconditions

- The current directory or `--repo-root` identifies a Git repository.
- `memory-bank/.repo` is a real, clean, conflict-free checkout with an
  accessible GitHub `origin` and a resolvable default branch.
- The downstream repository contains changed managed Memory Bank paths.

## Main Flow

1. The actor runs `memory-bank-cli push --dry-run` and reviews every inclusion
   and exclusion.
2. The actor runs `memory-bank-cli push`.
3. The CLI revalidates the checkout and upstream identity, creates a fresh
   branch from the upstream default branch, copies only managed changes,
   commits and pushes that branch, and creates a GitHub pull request.
4. The CLI reports the decisions and created pull-request URL.

## Alternate Flows / Exceptions

- `ALT-01` A dry run reports the plan without changing the checkout, remotes or
  GitHub.
- `EX-01` Unsafe paths, dirty state, unresolved conflicts, an invalid remote,
  ambiguous paths or an empty managed set stop before publication with a
  corrective next step.
- `EX-02` A post-mutation failure restores the original local checkout and
  attempts bounded cleanup of only the command-created remote branch; any
  residual state is reported.

## Postconditions

- On success, an upstream PR targets the default branch from a dedicated
  non-default branch and contains only managed Memory Bank paths.
- On dry-run or preflight failure, no working-tree files, upstream branch or
  commit, remote branch, or GitHub PR is created. A non-dry validation may
  refresh the local `origin/<default-branch>` tracking ref.
- The upstream default branch is never a direct push target.

## Business Rules

- `BR-01` Only paths classified as `managed` are publishable by this flow.
- `BR-02` Project-specific, adapted, generated, lock/state, `.repo` and unknown
  paths are excluded.
- `BR-03` Publication through a dedicated PR is mandatory; automatic merge and
  direct default-branch push are outside this flow.

## Traceability

| Upstream / Downstream | References |
| --- | --- |
| PRD | none |
| Features | `FT-023` |
| ADR | none |
| Runbooks / Ops | none |
