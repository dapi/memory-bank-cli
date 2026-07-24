---
title: memory-bank-cli Domain Rules
doc_kind: domain
doc_function: canonical
purpose: "Canonical owner безопасных ownership, source и navigation rules."
derived_from:
  - model.md
status: active
---

# Domain Rules

- `init`, `update`, and `doctor --fix` resolve a clean upstream checkout when
  source flags are omitted. An explicit `--source`, `--template-version`, and
  `--source-ref` trio overrides that resolution and must be supplied together.
- Template source must be a clean Git checkout and the supplied full ref must match its HEAD/payload; source and downstream roots must not overlap.
- `memory-bank/.lock` is reserved and cannot be supplied as template content.
- User-owned files are not overwritten or deleted by update. Conflicts prevent successful application.
- Updates stage the complete plan and apply atomically; interrupted/failed mutations are rolled back by the implementation.
- Destination paths must remain inside the pinned repository root; symlink/topology changes are guarded by implementation checks.
- `doctor` is read-only. `lint` and `doctor` return a non-zero result when their report contains error-level outcome.
- Markdown navigation is audited within a normalized repository-relative scope; escaping repository scope is rejected.

Implementation-level mechanics and exact messages remain owned by code and tests.
