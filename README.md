# memory-bank-cli
CLI for installing, updating, validating, and diagnosing Memory Bank templates

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
