# memory-bank-cli
CLI for installing, updating, validating, and diagnosing Memory Bank templates

`init` and `update` treat every tracked regular file below an upstream
`template/` directory as canonical payload. `template/memory-bank/**` installs
to `memory-bank/**`; every other path retains its repository-relative suffix.
Dotfiles and executable files are included, while symlinks are rejected.
Existing locks from the legacy payload roots are migrated conservatively:
unchanged files adopt canonical ownership, while local customization is
preserved for explicit resolution.

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

## Upgrade

Install the desired newer semantic version with the same command:

```sh
go install github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@vX.Y.Z
```

## Breaking release change

`memory-bank-cli` is the only supported executable. No compatibility binary,
alias, or alternative installation path is provided.
