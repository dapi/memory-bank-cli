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
| `DEC-01` | accepted | Route release delivery as a separate feature, not as a continuation of FT-001. | Release validation, tagged publication and installation documentation have their own externally observable outcome. | Bounded-context decomposition separates migration/identity work from release delivery; each has independently observable acceptance. | [brief.md](brief.md) `REQ-01`–`REQ-04`; FT-001 `NS-02` |
| `DEC-02` | accepted | Require a design document and the `release-deployment` profile. | The issue adds CI, an external GitHub release, semantic tagging and consumer installation. | A strict distinction between problem, solution and execution places release topology/approval in design and checks/evidence in brief. The strongest profile trigger is external release deployment. | [brief.md](brief.md) Design/Validation decisions; [design.md](design.md) |
| `DEC-03` | accepted | Use a manually dispatched release workflow that validates the exact `main` commit before its approval-gated job creates the tag or invokes GoReleaser. | The release contract requires tests, static analysis and release builds before publication; a pushed tag makes the Go module publicly installable; `.goreleaser.yml` defines a single GitHub release build. | Assurance reasoning places evidence-producing validation before the first public effect, not merely before a later GitHub Release. | [design.md](design.md) `SOL-01`, `SOL-02`, `CTR-01`, `INV-01` |
| `DEC-04` | accepted | Use `v1.0.0` as the first release under the current executable identity and the documented Go install command as the consumer acceptance test. | The executable rename is a breaking change and `go install github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@vX.Y.Z` is the supported installation surface. | Evidence/provenance reasoning binds a public tag to a reproducible consumer check instead of treating a local build as release proof. | [brief.md](brief.md) `ASM-01`, `CON-02`, `CHK-04`; [design.md](design.md) `SD-01` |
| `DEC-05` | accepted | Preserve existing GoReleaser destinations and treat unavailable configured credentials as a stop condition. | `.goreleaser.yml` declares the GitHub release plus a Homebrew cask requiring `HOMEBREW_TAP_GITHUB_TOKEN`; the rename does not authorize altering destinations. | Boundary discipline: missing operational evidence is not a licence to invent a new distribution policy or weaken security. The approved maintainer chooses how to satisfy or change that external contract. | [brief.md](brief.md) `ASM-02`, `NS-04`, `CON-01`; [design.md](design.md) `TRD-02`, `FM-02` |
| `DEC-06` | accepted | Use the repository-enforced `AG-01` gate for the release job that creates both the public tag and GitHub Release. | External publication is irreversible; the release job runs only after validation and the GitHub `release` environment has a required reviewer. | Role-method-work separation puts the human authorization directly before the irreversible work, while assurance evidence keeps validation before it. | [brief.md](brief.md) `CON-01`; [design.md](design.md) `INV-03`, `RB-02`; [implementation-plan.md](implementation-plan.md) `PRE-03`, `AG-01`, `STEP-04` |
| `DEC-07` | accepted | Resolve the version of Go-installed binaries from Go build information when no release linker version is present. | `go install ...@vX.Y.Z` does not apply GoReleaser linker flags, while Go build information records the selected main-module version; local builds record `(devel)`. | Evidence-chain reasoning requires the documented `--version` smoke check to observe the same tag that identifies the installed module. Linker injection remains authoritative for GoReleaser artifacts, and `dev` remains the honest fallback when no release version exists. | [brief.md](brief.md) `EC-03`, `CHK-04`; [design.md](design.md) `SOL-03`, `INV-04`, `FM-04` |
| `DEC-08` | implemented | Enforce `AG-01` through a GitHub `release` environment with a required reviewer, rather than relying on a workflow label or maintainer convention. | A job-level environment name alone does not create protection; GitHub environment API confirms the `release` environment has a `required_reviewers` rule. | Gate/evidence distinction: workflow dependency proves validation, while the protected environment supplies independent human authorization before the job creates the tag or GitHub Release. | [design.md](design.md) `SOL-02`, `INV-03`; [implementation-plan.md](implementation-plan.md) `AG-01` |
| `DEC-09` | accepted | Treat a pushed semantic tag as an immutable public Go module version; never repoint it, and block FT-003 if a later defect prevents its required `v1.0.0` acceptance. | `CON-02` requires consumers to install the tag through `go install`; Go proxies may cache the tagged module version; `EC-02` and `EC-03` specifically require `v1.0.0`. | Boundary and evidence reasoning keeps the immutable tag fact separate from the feature's required outcome. A new version cannot silently substitute for the specified acceptance; a post-tag acceptance defect therefore requires human-approved change control. | [brief.md](brief.md) `CON-01`, `REQ-02`, `EC-02`, `EC-03`; [design.md](design.md) `FM-03`, `RB-02`; [implementation-plan.md](implementation-plan.md) `STOP-03`, `STOP-04` |
