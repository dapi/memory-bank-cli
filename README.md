# memory-bank-cli
CLI for installing, synchronizing, validating, and diagnosing Memory Bank templates

`init` and `pull` treat every tracked regular file below an upstream
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

## Resolve pull collisions interactively

`pull` preserves user-owned files by default. To decide each managed-file
collision interactively, run it from a terminal with `--ask`:

```sh
memory-bank-cli pull --ask
```

Choose `keep` to retain the local file and its ownership, or `overwrite` to
replace it (including its executable mode) from the source template. All
answers are collected before changes are applied. `pull --ask --dry-run` shows the
resolved plan without changing files; `--ask` is rejected when standard input
is not a terminal.

## Prepare and apply a reviewed pull plan

Ordinary `pull` automatically applies a merge only when it can prove all of
the following: the historical Git source blob matches the base in `.lock`, both
files are regular supported files, and local and upstream line edits do not
overlap. It preserves both independent changes and records the new upstream
base atomically.

Create a complete reviewable plan only when `pull` reports a remaining
conflict, or when you want an explicit audit record:

```sh
memory-bank-cli pull --plan memory-bank-pull-plan.json
```

The CLI writes the plan with owner-only permissions. Use `--plan -` to emit it
to standard output instead. An agent may analyze the entries and candidate
merges, but a reviewer must explicitly set `selected_action` for every entry
whose `requires_human_decision` is true. Available actions are:

- `keep-local` — retain the downstream bytes and advance its adapted base;
- `take-upstream` — write the upstream bytes and mode;
- `apply-reviewed-merge` — write the exact candidate from `merge`, only when
  the old Git source blob matches the base recorded in `.lock` and the line
  changes do not overlap.

Apply the reviewed plan with:

```sh
memory-bank-cli pull --apply-plan memory-bank-pull-plan.json
```

Apply resolves the current source again, strictly regenerates every
non-decision field, and rejects unresolved, altered or stale input before
mutation. Accepted resolutions, deterministic managed updates and `.lock`
commit atomically. The CLI never calls an LLM or treats a mechanical merge as
semantic approval: overlapping edits, missing historical source and explicit
ownership choices still require review.

Resolution plans may contain base64-encoded merged document content. Treat
them with the same privacy as the downstream repository and review them before
committing.

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

For a reproducible install, replace `latest` with an exact release tag.
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
before running `pull`. The repair fetches `main` from `memory-bank/.repo`'s
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

## Update the CLI

Update a released macOS or Linux binary in place:

```sh
memory-bank-cli update
```

The command resolves the latest stable GitHub Release, selects the binary for
the current supported platform, verifies it against the release
`checksums.txt`, verifies the staged binary version, then atomically replaces
the executable you invoked. If the installed version is current or newer, it
reports a successful no-op. Windows users should download
`memory-bank-cli-windows-amd64.exe` from the latest release and replace their
executable manually.

To synchronize installed Memory Bank content, use `memory-bank-cli pull`.
Existing automation and documentation using `memory-bank-cli update` for that
purpose must be migrated to `memory-bank-cli pull`.

## Breaking release change

`memory-bank-cli` is the only supported executable. No compatibility binary,
alias, or alternative installation path is provided.
