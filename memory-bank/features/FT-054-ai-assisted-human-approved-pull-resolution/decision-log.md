---
title: "FT-054: Decision Log"
doc_kind: feature
doc_function: derived
purpose: "Decision provenance for the selected FT-054 trusted-local workflow."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
---

# FT-054: Decision Log

## Ownership

`brief.md` owns requirements and verification. `design.md` owns the selected
solution. This ledger records why those decisions were accepted.

## Decisions and Open Questions

| ID | Status | Decision | Evidence / rationale | Consequence |
| --- | --- | --- | --- | --- |
| `DEC-01` | accepted | Use a versioned full-scope JSON plan and regenerate every deterministic field before apply. | Issue #54 requires durable review, stale/tamper rejection and atomic apply; current planner and transaction already expose the required local facts. | Only selected actions and reviewed merge payload are editable overlays. |
| `DEC-02` | accepted | Treat a two-sided adapted conflict as unresolved until `keep-local`, `take-upstream` or an eligible reviewed merge is explicitly selected. | Project-specific documents can encode intent that automation cannot infer. | Ordinary pull remains conservative; reviewed apply rejects incomplete plans. |
| `DEC-03` | accepted | Use a trusted-local authorization model rather than cryptographic reviewer identity. | The user approved this boundary on 2026-08-15. The CLI can prove state/result integrity but cannot distinguish a human from another process on the same account without an external authority. | No signing keys, reviewer registry or credential lifecycle; review remains visible in Git/user workflow. |
| `DEC-04` | accepted | Recover historical base from `.lock.template.source_ref` and verify the Git blob against `BaseDigest`/`BaseMode`. | The immutable source identity and base digest already exist; the upstream fetch contains reachable history. | No lock-history sidecar or protected receipt registry. Missing/mismatched history disables merge only. |
| `DEC-05` | accepted | Use deterministic non-overlapping line merge and include its exact result in the reviewed plan. | This allows mechanical assistance without claiming semantic correctness. | Apply recomputes exact bytes/mode; overlap keeps merge unavailable. |
| `DEC-06` | pending | Refresh issue #54 with the final implementation/PR/release carrier. | Feature Flow requires a current tracker route. | Complete during publication. |

## Open Questions

None block implementation. Exact JSON field names and bounded limits may be
selected during implementation while preserving `design.md` invariants.
