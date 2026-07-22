# memory-bank-cli
CLI for installing, updating, validating, and diagnosing Memory Bank templates

## Install

Install a released version with Go:

```sh
go install github.com/dapi/memory-bank-cli/cmd/mb-cli@vX.Y.Z
```

The first stable standalone release is `v1.0.0`. After installing it, run:

```sh
mb-cli --version
```

## Upgrade

Install the desired newer semantic version with the same command:

```sh
go install github.com/dapi/memory-bank-cli/cmd/mb-cli@vX.Y.Z
```

## Breaking release change

`mb-cli` is the only supported executable. The former `memory-bank` name is intentionally breaking, and `memory-bank-lint` has been removed. Neither has a compatibility binary, alias, or installation path.
