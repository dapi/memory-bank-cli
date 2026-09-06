# Component installation delivery

Tracker: [CLI #62](https://github.com/dapi/memory-bank-cli/issues/62).
Source: [memory-bank #141](https://github.com/dapi/memory-bank/issues/141).
CLI owner: memory-bank-cli. The shared protocol belongs to the
[template contract](https://github.com/dapi/memory-bank/blob/feat/141-component-adoption/docs/components.md).
The [FT-141 brief](https://github.com/dapi/memory-bank/blob/feat/141-component-adoption/memory-bank/features/FT-141/brief.md)
owns imported requirements and acceptance; this file owns CLI realization and evidence routing.

Grounded CLI revision: `ac7101c307e65566787bdb32a1bdad40b9a8b995`.
`internal/ownership/update.go` provides a handle-relative payload/agent/lock transaction.
`internal/ownership/source.go` verifies pinned blobs but lacks a component source gate.
`internal/doctor/governance.go` applies feature lifecycle checks type-wide.
Baseline `env -u GOROOT go test ./...` passes all packages on 2026-09-07.

## Gates and realization

CLI-01 uses the independently reviewed [bridge plan](source-format-bridge.md). CLI-02/03 wait for the shared component design gate. CLI code, tests
and evidence stay here. The slice is independently verifiable with synthetic source fixtures;
the template PR supplies the final cross-repository source integration.

| Step | Imported requirement | CLI surface | Verification |
| --- | --- | --- | --- |
| CLI-01 | REQ-06/08 | ownership/source_format.go; source verification; capabilities command | Real bridge/pre-bridge binary fixtures; incompatible sources cannot mutate |
| CLI-02 | REQ-01/02/04 | ownership/components.go; Lock/Options/run and resolution planning | Preset/adapter/ownership, no-op, stale-plan and rollback tests |
| CLI-03 | REQ-03/04/05 | ownership/documents.go; validator; doctor/lint; document commands | Creation/adoption, tampering, immutable rules, selectors and legacy migration |

## Acceptance and evidence

Imported SC-01…06 and NEG-01…05 are mandatory CLI checks; SC-07/08 additionally use the actual
template payload. Concrete test names, commands and CI results are recorded in the PR as they
are delivered. Required local checks: `env -u GOROOT go test ./...`, `env -u GOROOT go vet ./...`, existing hermetic
init/pull E2E and component fixtures. Required GitHub Actions must be green on the reviewed
commit. Independent code-converge implementation and simplify review must have no actionable
findings. Unexecuted checks are not evidence.

## Rollout

Prepare the bridge commit and record its actual binary identity before component support.
Component source must wait for supporting CLI. No live downstream mutation, merge or release
publication belongs to this task. Related PRs state the required release order explicitly.
