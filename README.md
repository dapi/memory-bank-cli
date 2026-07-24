# memory-bank-cli
CLI for installing, updating, validating, and diagnosing Memory Bank templates

## Publish managed changes upstream

From a downstream Git repository with a clean upstream checkout at `memory-bank/.repo`, preview the managed changes that can be proposed upstream:

```sh
memory-bank-cli push --dry-run
```

Without `--dry-run`, `push` creates a fresh upstream branch, commits only changed `managed` Memory Bank paths, pushes it and creates a GitHub PR. It never pushes the upstream default branch directly. Non-managed paths, including project artifacts, lock/state and `.repo`, are reported as exclusions.

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

## Upgrade

Install the desired newer semantic version with the same command:

```sh
go install github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@vX.Y.Z
```

## Breaking release change

`memory-bank-cli` is the only supported executable. No compatibility binary,
alias, or alternative installation path is provided.
