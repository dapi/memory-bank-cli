---
title: "FT-003: Decision Log"
doc_kind: feature-support
doc_function: derived
purpose: "Audit ledger for FT-003 decisions and FPF reasoning; it links canonical brief/design owners without becoming a second owner."
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

# FT-003: Decision Log

## Ownership

This log records decision provenance. `brief.md` owns problem-space facts and `design.md` owns selected solution decisions; conflicts are resolved by updating those owners first, then this ledger.

## Decisions

| ID | Status | Decision | Facts considered | FPF reasoning | Canonical owner |
| --- | --- | --- | --- | --- | --- |
| `DEC-01` | accepted | Route issue #3 as a separate release-deployment feature, not as a continuation of FT-001. | Issue #3 has its own outcome (CI, tagged publication and docs); FT-001 explicitly hands these outcomes off to issue #3. | Bounded-context decomposition separates migration/identity work from release delivery; each has independently observable acceptance. | [brief.md](brief.md) `REQ-01`–`REQ-04`; FT-001 `NS-02` |
| `DEC-02` | accepted | Require a design document and the `release-deployment` profile. | The issue adds CI, an external GitHub release, semantic tagging and consumer installation. | A strict distinction between problem, solution and execution places release topology/approval in design and checks/evidence in brief. The strongest profile trigger is external release deployment. | [brief.md](brief.md) Design/Validation decisions; [design.md](design.md) |
| `DEC-03` | accepted | Use a validation workflow plus a tag-driven workflow that repeats validation before GoReleaser publication. | Issue #3 requires tests, static analysis and release builds before publication; `.goreleaser.yml` already defines a single GitHub release build. | Assurance reasoning: publication is a high-impact claim, so the evidence-producing validation must be a prerequisite of that same release execution rather than an assumed earlier event. | [design.md](design.md) `SOL-01`, `SOL-02`, `INV-01` |
| `DEC-04` | accepted | Use `v1.0.0` as the first release tag and the issue-specified Go install command as the consumer acceptance test. | Issue #3 explicitly names `v1.0.0` and `go install github.com/dapi/memory-bank-cli/cmd/mb-cli@vX.Y.Z`; the repository has no current tag/release. | Evidence/provenance reasoning binds a public tag to a reproducible consumer check instead of treating a local build as release proof. | [brief.md](brief.md) `ASM-01`, `CON-02`, `CHK-04`; [design.md](design.md) `SD-01` |
| `DEC-05` | accepted | Preserve existing GoReleaser destinations and treat unavailable configured credentials as a stop condition. | `.goreleaser.yml` declares the GitHub release plus a Homebrew cask requiring `HOMEBREW_TAP_GITHUB_TOKEN`; issue #3 does not authorize altering destinations. | Boundary discipline: missing operational evidence is not a licence to invent a new distribution policy or weaken security. The approved maintainer chooses how to satisfy or change that external contract. | [brief.md](brief.md) `ASM-02`, `NS-04`, `CON-01`; [design.md](design.md) `TRD-02`, `FM-02` |
| `DEC-06` | accepted | Separate maintainer authorization of the public tag push from the repository-enforced human gate for GitHub Release publication. | Issue #3 requires external publication; the tag-triggered workflow's `release` environment has a required reviewer and therefore blocks publication, while tag creation itself remains a repository-permission action. No source grants an agent authority to create either external effect or use credentials. | Role-method-work separation distinguishes the maintainer's tag authorization from the environment's publication gate. Gate/evidence distinction prevents attributing enforcement to the environment where it has none, while still preserving a verified candidate and a protected publication step. | [brief.md](brief.md) `CON-01`; [design.md](design.md) `INV-03`, `RB-02`; [implementation-plan.md](implementation-plan.md) `PRE-03`, `AG-01`, `STEP-04` |
| `DEC-07` | accepted | Resolve the version of Go-installed binaries from Go build information when no release linker version is present. | `go install ...@vX.Y.Z` does not apply GoReleaser linker flags, while Go build information records the selected main-module version; local builds record `(devel)`. | Evidence-chain reasoning requires the documented `--version` smoke check to observe the same tag that identifies the installed module. Linker injection remains authoritative for GoReleaser artifacts, and `dev` remains the honest fallback when no release version exists. | [brief.md](brief.md) `EC-03`, `CHK-04`; [design.md](design.md) `SOL-03`, `INV-04`, `FM-04` |
| `DEC-08` | implemented | Enforce `AG-01` through a GitHub `release` environment with a required reviewer, rather than relying on a workflow label or maintainer convention. | A job-level environment name alone does not create protection; GitHub environment API confirms the `release` environment has a `required_reviewers` rule. | Gate/evidence distinction: workflow dependency proves validation, while the protected environment supplies independent human authorization specifically before GitHub Release publication. It does not claim to govern the earlier tag-push action. | [design.md](design.md) `SOL-02`, `INV-03`; [implementation-plan.md](implementation-plan.md) `AG-01` |
