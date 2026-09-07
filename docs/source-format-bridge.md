# Source-format bridge

This is W1 of [CLI #62](https://github.com/dapi/memory-bank-cli/issues/62), supporting
[memory-bank #141](https://github.com/dapi/memory-bank/issues/141). Scope is source format
classification before any installation plan or mutation. Component installation and adoption
remain later units governed by the shared component design.

## Requirements and acceptance

- BR-01: A bridge accepts the pinned manifestless template commit
  `f1f04de843aef45a2425d4a7351d577bbf89e940`, and rejects an unknown manifestless source.
- BR-02: A declared source has `memory-bank-source.json` at its checkout root, schema 1,
  `payload_format: legacy/v1`, and `capabilities: [legacy/v1]`. The file must be a tracked
  regular Git blob at the selected commit. The declaration array is exactly [legacy/v1];
  source-format/v1 belongs to the CLI handshake, not this legacy declaration schema.
  Unknown/duplicate fields, trailing JSON,
  unsupported schema/format/capabilities, missing mandatory capabilities and a component
  manifest inside a declared legacy source are errors. The component marker is
  `<payload-root>/memory-bank/components.json` for template/, or `<payload-root>/components.json`
  for the legacy payload roots. The bridge cannot process a component source.
- BR-03: Rejection leaves every downstream file and lock unchanged, including dry-run,
  `pull --plan`, `pull --apply-plan` and doctor repair paths using ownership.Init.
- BR-04: `capabilities [--require CAP ...]` prints machine-readable supported capabilities
  and returns nonzero when a required capability is unavailable. The bridge advertises
  `source-format/v1` and `legacy/v1`; it does not advertise components or adoption.
- BR-05: The manifestless SHA explicitly listed in BR-01 retains its legacy installation
  behavior. Declared legacy fixture sources retain the existing ownership/transaction behavior.
  Other previously usable custom commits are intentionally unsupported by this bridge:
  owners must keep their previous CLI/source pair, or add a declaration and explicitly repin.
  No automatic repinning or rewriting of a source is performed.

The declaration is an explicit supported legacy format for custom sources, not an allowlist
bypass for manifestless source. The built-in SHA list is the only way to accept an undeclared
legacy tree. Future formats must change the declaration. Pre-bridge programs remain unable to
read this gate: their direct execution on component payload is unsupported; the later template
entrypoint checks capabilities before calling installer. This is the explicit compatibility
boundary accepted in issue #141: no guarantee is made for direct pre-bridge invocation on
a component source. BR-01–BR-05 apply to the bridge binary. Retrofitting old executables is
out of scope and cannot be achieved by changing a source manifest.

The capability wire response is exactly one JSON object on stdout followed by a newline:
`{"schema_version":1,"cli_version":"<build version>","capabilities":["source-format/v1","legacy/v1"],"unsupported":[]}`.
Arrays contain strings; capabilities have deterministic declared order. Repeated `--require CAP`
is allowed; unsupported contains the requested unavailable values in request order. Exit 0 means
all requirements are available, exit 1 means at least one is unavailable (same JSON response),
exit 2 is invalid command syntax (diagnostic on stderr). No prose is printed on stdout.
The future entrypoint relies on exit status of --require, not on parsing incidental output.

## Design and execution plan

Grounded revision: `ac7101c307e65566787bdb32a1bdad40b9a8b995`.
`internal/ownership/source.go#verifySourceCheckout` is called by both run and PlanPull and
already verifies clean checkout and pinned regular blobs. The complete repair call chain is
`cli.runDoctor --fix → ownership.Init → run → verifySourceCheckout`; it reaches the same gate.
Add an explicit doctor-repair rejection test asserting unchanged target files and no lock.
`cli.runOwnership --apply-plan → ApplyResolutionPlan → PlanPull → verifySourceCheckout`
revalidates the actual source before comparing a saved plan; the final
`ApplyResolutionPlan → Update → run → verifySourceCheckout` validates it again before writes.
Thus a plan produced by any earlier binary conveys no source-format exemption. The negative
apply-plan fixture supplies a matching source identity in an old-format plan against an
unsupported checkout and asserts a source-format error and byte-identical downstream state.
All source-consuming mutation routes listed in BR-03 therefore use the same gate.
Add sourceGate there after pinned
payload-root validation, using Git objects rather than mutable working-tree files.
`internal/ownership/source_format.go` owns declaration schema/classification and strict decoding.
`internal/cli/cli.go#Run` adds the capability command; no new mutating command is introduced.

The only alternatives are accepting every undeclared tree (violates BR-01) or refusing all
legacy trees (violates compatibility). Use the compiled supported SHA plus strict declarations.
No new state schema, transaction writer or target paths are introduced in W1.
The canonical-source canary exposed one existing bookkeeping bug: a preserve decision can
change ownership without a file mutation, so run must persist changed Files entries, not
only a changed file count. A regression fixture verifies persistence and the next no-op pull.
Git plumbing disables replacement objects so local replace refs cannot change a pinned
source commit's contents. A real Git fixture proved the previous substitution and now rejects it.

Test helpers in `internal/ownership/source_test.go#commitTestSource`,
`internal/cli/cli_test.go#commitCLISource` and `scripts/e2e-init-update.sh#setup_case` declare
synthetic legacy fixtures without disabling production gates. New source-format tests retain
explicit undeclared/invalid fixtures and verify no mutation. Any other fixture producer that
uses an actual Git source receives the same declaration; fixture changes are supporting BR-05.

## Validation and readiness

Validation profile: standard (CLI/source format contract). No live state or release mutation.
Before implementation: independent document review of this plan must be clean. After it:
1. Add strict source classification and capability reporting (BR-01/02/04).
2. Add positive/negative pinned Git fixtures and update existing fixture producers (BR-03/05).
3. Run `env -u GOROOT go test ./...`, `env -u GOROOT go vet ./...`, build the bridge and run hermetic E2E.
4. Independent code-converge review; fix findings and re-review. Record bridge commit and binary
   SHA-256 before adding component capabilities in the subsequent unit.

Baseline tests already pass with `env -u GOROOT go test ./...`; the inherited GOROOT mismatches
the PATH compiler, so use a consistent toolchain per command. Failed preflight preserves the
legacy tree. Code rollback is a revert before release. A bridge release is prepared as a
separate commit/PR; publishing it is outside the current implementation/PR task.
