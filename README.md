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

Install a released version with Go:

```sh
go install github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@vX.Y.Z
```

The first release under the `memory-bank-cli` executable identity will be
`v1.0.0`. After it is published and installed, run:

```sh
memory-bank-cli --version
```

## Upgrade

Install the desired newer semantic version with the same command:

```sh
go install github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@vX.Y.Z
```

## Breaking release change

`memory-bank-cli` is the only supported executable. No compatibility binary,
alias, or alternative installation path is provided.
