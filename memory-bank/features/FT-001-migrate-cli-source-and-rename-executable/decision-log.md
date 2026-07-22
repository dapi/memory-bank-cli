---
title: "FT-001: Decision Log"
doc_kind: feature-support
doc_function: derived
purpose: "Audit ledger for FT-001 decisions and FPF reasoning. It records rationale and links to canonical brief/design owners without creating a second owner."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
must_not_define:
  - requirements
  - selected_solution
  - implementation_sequence
---

# FT-001: Decision Log

## Ownership

This log records decision provenance. `brief.md` owns problem-space facts and `design.md` owns selected solution decisions; a conflict is resolved by updating those canonical documents first, then this ledger.

## Decisions

| ID | Status | Decision | Facts considered | FPF reasoning | Canonical owner |
| --- | --- | --- | --- | --- | --- |
| `DEC-01` | accepted | Route issue #1 through Feature Flow as one delivery-unit. | Issue #1 changes the public CLI identity and module; issue #51 lists it as one CLI-repository work item, while issues #2 and #3 own distinct later outcomes. | Bounded-context decomposition separates migration/rename, detection, and release. Public CLI change fails Small Change predicates; one independently verifiable unit does not require a new Epic. | [brief.md](brief.md) `REQ-01`–`REQ-04`, `NS-01`–`NS-02` |
| `DEC-02` | accepted | Require a design document and use `release-deployment` validation profile. | Module path, executable identity, import history and release-build config change; no publication occurs in this issue. | Distinguish design work from execution: public contract and artifact configuration require explicit solution reasoning. Profile selection follows the strongest applicable build/release trigger, without conflating it with actual release execution. | [brief.md](brief.md) Design Requirement and Validation Profile decisions; [design.md](design.md) |
| `DEC-03` | accepted | Fix the import baseline at source SHA `0957f3c495f2c0518c8a81448694cf0e231d3209` and preserve the filtered `tools/` lineage. | The observed `main` SHA contains the complete `tools/` tree; relevant history includes `b164c03` and `0039b3e`; issue #1 requires useful history where practical. | Evidence/provenance reasoning: a fixed observable source state makes the migration reproducible. Filtering only the bounded `tools/` context preserves relevant lineage while excluding unrelated template history. | [brief.md](brief.md) `ASM-01`, `CON-01`; [design.md](design.md) `SOL-01`, `SD-01` |
| `DEC-04` | accepted | Preserve `memory-bank/` payload paths, but remove old executable identities everywhere in the migrated CLI surface. | Issue #1 bans old executable compatibility; source code also uses `memory-bank/` as the managed documentation payload path. | Object-of-talk distinction: executable identity and payload directory are different concepts. Replacing both would change behavior beyond the issue. | [brief.md](brief.md) `CON-03`; [design.md](design.md) `TRD-02`, `INV-03` |
| `DEC-05` | accepted | Hand off template detection to #2 and tagging/publication/installation documentation to #3. | Issues #2 and #3 explicitly own those scopes; issue #51 specifies their delivery order; target has no tag/release today. | Boundary discipline: do not solve adjacent open work inside FT-001. The handoff keeps acceptance evidence honest and prevents unsupported claims. | [brief.md](brief.md) `NS-01`, `NS-02`, `CON-02`; [design.md](design.md) `SD-03` |
| `DEC-06` | accepted | Treat historical mentions of old executable names in this feature package as governance evidence, not user-facing product documentation. | Issue #1 prohibits old-name references in shipped source/test/documentation, while the feature package must retain the issue-derived removal decision and review evidence. | Context-of-meaning distinction resolves the apparent conflict: a delivery record explains a breaking removal; it does not expose a supported command to CLI users. Product-surface searches exclude this package and generated evidence, but include all migrated docs. | [brief.md](brief.md) `EC-03`, `CHK-03` |
| `DEC-07` | implemented | Imported the source through a disposable clone filtered with `git filter-branch --subdirectory-filter tools`, then merged the filtered history into this feature branch. | The source SHA is `0957f3c495f2c0518c8a81448694cf0e231d3209`; the filtered branch retains the CLI lineage rooted at the original `tools/` commits. | The procedure realizes `DEC-03` while keeping the target's initial commit and unrelated template history outside the imported tree. | [design.md](design.md) `SOL-01`, `SD-02`; [implementation-plan.md](implementation-plan.md) `STEP-01` |
