---
title: memory-bank-cli Local Development
doc_kind: ops
doc_function: canonical
purpose: "Canonical owner local development commands confirmed by module and FT-001."
derived_from:
  - ../engineering/architecture.md
  - ../engineering/testing-policy.md
status: active
---

# Local Development

Prerequisites confirmed by `go.mod` and the CLI implementation: Go 1.21, Git (for source-template operations), and a local checkout of this repository.

```sh
go test -count=1 -race ./...
go vet ./...
go run ./cmd/memory-bank-cli --help
go run ./cmd/memory-bank-cli lint --repo-root .
```

`init` and `update` additionally require a separate clean template checkout, `--template-version`, and `--source-ref` matching its HEAD. No `.env`, database, service dependency or container workflow is documented.
