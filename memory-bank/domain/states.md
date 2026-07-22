---
title: memory-bank-cli States
doc_kind: domain
doc_function: canonical
purpose: "Canonical owner lifecycle states observable in adoption and audit flows."
derived_from:
  - model.md
  - rules.md
status: active
---

# States

## Template adoption/update

```text
unmanaged -> planned -> applied
                    -> conflict (no successful update)
                    -> failed (implementation rolls back)
```

`--dry-run` produces a plan without transitioning repository content to `applied`.

## File ownership

Each tracked file is one of `managed`, `adapted`, `user-owned` or `generated`. These are classes, not sequential states.

## Documentation audit

A lint/doctor report aggregates findings with `error`, `warning` or `info` severity. An error result causes the corresponding CLI command to fail; warnings do not by themselves establish a failure in the visible command contract.
