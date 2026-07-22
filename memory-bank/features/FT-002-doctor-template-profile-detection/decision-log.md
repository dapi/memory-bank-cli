---
title: "FT-002: Decision Log"
doc_kind: feature-support
doc_function: derived
purpose: "Audit ledger for FT-002 decisions and FPF reasoning; it links to canonical owners without becoming a second owner."
derived_from:
  - brief.md
status: active
audience: humans_and_agents
must_not_define:
  - requirements
  - selected_solution
  - implementation_sequence
---

# FT-002: Decision Log

## Ownership

`brief.md` owns problem-space facts, scope, validation decision, and acceptance. A later `design.md` will own the selected marker and detection solution. This log only preserves reasoning provenance; if there is a conflict, update the canonical owner first and then this ledger.

## Decisions

| ID | Status | Decision | Facts considered | FPF reasoning | Canonical owner |
| --- | --- | --- | --- | --- | --- |
| `DEC-01` | accepted | Use root marker `.memory-bank-template` with exact UTF-8 content `memory-bank-template-v1\n`; `dapi/memory-bank#52` owns adding and documenting it. | Issue #2 requires an explicit stable marker in `dapi/memory-bank`, outside copied `memory-bank/`, and rejects lookalike false positives. Issue #52 explicitly assigns adding the required marker to the template repository. Auto detection has only one Boolean question: source template or downstream. | FPF bounded contexts distinguish source identity from copied payload and downstream lock. Candidate A, fixed filename/content, is selected over structured metadata because it answers the Boolean question with fewer parser and schema dependencies. Candidate B, structured/versioned data, adds format degrees of freedom without an issue requirement. The `v1` token keeps a deliberate evolution seam. Filters: scope fit (root and outside payload) and collision resistance (exact path plus exact bytes; similarly named files do not qualify). | [brief.md](brief.md) `DEC-01`; [design.md](design.md) `SOL-01`, `CTR-01` |
| `DEC-02` | accepted | Route issue #2 as one designed delivery-unit. | Issue #2 has one observable outcome: auto-profile detection becomes independent of `tools/go.mod`; it names fixtures and documentation as acceptance. The change affects CLI profile-selection and a cross-repository config contract. | FPF bounded-context decomposition keeps detection separate from CLI migration (#1) and template cleanup (#52). Strict owner separation keeps brief, design, and execution facts distinct. | [brief.md](brief.md) Design Requirement and Artifact Routing decisions |
| `DEC-03` | accepted | Use the `standard` validation profile with Go test, vet, and contract-focused regression. | The brief template names `standard`; testing policy says ordinary code changes require Go test/vet and contract regression. This feature has no release, security, migration, or performance trigger. | FPF evidence grounding selects the least profile that covers every evidenced risk, avoiding an unneeded release-deployment profile. | [brief.md](brief.md) Validation Profile Decision |
| `DEC-04` | accepted | Add the source-issue backlink after this feature branch is first published; use the permanent GitHub URL of `brief.md`. | Feature Flow requires a tracker backlink; no remote branch currently exists, so a link now would be broken. | FPF evidence graph keeps the issue-to-brief edge durable: publish the carrier first, then record its immutable/reviewable URL. This is a sequencing decision, not a marker-contract blocker. | [brief.md](brief.md) source refs; [implementation-plan.md](implementation-plan.md) `STEP-05` |
