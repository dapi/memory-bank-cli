# memory-bank-cli
CLI for installing, updating, validating, and diagnosing Memory Bank templates

`init` and `update` treat every tracked regular file below an upstream
`template/` directory as canonical payload. `template/memory-bank/**` installs
to `memory-bank/**`; every other path retains its repository-relative suffix.
Dotfiles and executable files are included, while symlinks are rejected.
Existing locks from the legacy payload roots are migrated conservatively:
unchanged files adopt canonical ownership, while local customization is
preserved for explicit resolution.

## Analyse an execution handoff

Inspect an Execution Handoff without changing the handoff or Memory Bank:

```sh
memory-bank-cli analyze-graph \
  --handoff .memory-bank/handoffs/FT-042.json
```

The default Markdown report separates recommendations, findings, and typed
evidence. Add `--json` for the versioned structured report. Recommendations
retain the contributing node IDs, relation types, and source references;
unresolved or weakly sourced relations are reported for review.

## Resolve update collisions interactively

`update` preserves user-owned files by default. To decide each managed-file
collision interactively, run it from a terminal with `--ask`:

```sh
memory-bank-cli update --ask
```

Choose `keep` to retain the local file and its ownership, or `overwrite` to
replace it (including its executable mode) from the source template. All
answers are collected before changes are applied. `--ask --dry-run` shows the
resolved plan without changing files; `--ask` is rejected when standard input
is not a terminal.

## Publish managed changes upstream

From a downstream Git repository with a clean upstream checkout at `memory-bank/.repo`, preview the managed changes that can be proposed upstream:

```sh
memory-bank-cli push --dry-run
```

Without `--dry-run`, `push` creates a fresh upstream branch, publishes every
changed path recorded as `managed` in the ownership lock back below
`template/`, pushes the branch and creates a GitHub PR. It never pushes the
upstream default branch directly. Non-managed paths, including project
artifacts, lock/state and `.repo`, are reported as exclusions.

## Install

Install the latest released version with Go:

```sh
go install github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@latest
```

For a reproducible install, replace `latest` with a tag such as `v1.4.0`.
After installation, run:

```sh
memory-bank-cli --version
```

See [CHANGELOG.md](CHANGELOG.md) for release notes.

## Repair a missing ownership lock

`doctor` is read-only by default. If it reports `template.identity_missing`,
preview the same safe adoption plan used by `init`:

```sh
memory-bank-cli doctor --fix --dry-run
```

Rerun without `--dry-run` to create `memory-bank/.lock`, then commit that lock
before running `update`. The repair fetches `main` from `memory-bank/.repo`'s
clean `origin` or the default upstream. To use a specific trusted checkout,
pass `--source`, `--template-version`, and `--source-ref` together. The repair
never replaces an existing lock.

## Build an execution handoff

Build a deterministic, read-only projection from a task document and explicit
execution evidence. Markdown is written by default; add `--json` for the
machine-readable projection.

```sh
memory-bank-cli handoff build \
  --from features/FT-042/implementation-plan.md \
  --git-range main..HEAD \
  --test-report reports/test-results.json \
  --out .memory-bank/handoffs/FT-042.md
```

The command never edits source documents. Missing documents, reports, broken
links, and invalid Git ranges are included in the output as unresolved sources
and cause a non-zero exit status.

## Upgrade

Install the desired newer semantic version with the same command:

```sh
go install github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@vX.Y.Z
```

## Breaking release change

`memory-bank-cli` is the only supported executable. No compatibility binary,
alias, or alternative installation path is provided.
