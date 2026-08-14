---
title: "FT-051: Separate CLI Self-Update Design"
doc_kind: feature
doc_function: canonical
purpose: "Selected command split and verified release-backed self-update design for FT-051."
derived_from:
  - brief.md
  - ../../flows/feature.md
  - ../../../.goreleaser.yml
status: active
audience: humans_and_agents
must_not_define:
  - ft_051_scope
  - ft_051_acceptance_criteria
  - implementation_sequence
---

# FT-051: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` |

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- |
| `C4-00` | not required | One executable gains a release HTTP/filesystem connector; no deployable or internal component topology changes. | none |

## Selected Solution

- `SOL-01` Move the current ownership synchronization dispatch unchanged to `pull`; reserve top-level `update` for self-update.
- `SOL-02` `update` resolves the latest non-prerelease GitHub Release for `dapi/memory-bank-cli`, selects the existing matching `memory-bank-cli-<os>-<arch>` raw asset and `checksums.txt`, and only upgrades a strictly older installed SemVer version.
- `SOL-03` On macOS/Linux amd64/arm64, verify the asset checksum, stage it beside the resolved running executable, verify staged `--version` exactly matches the release tag, then atomically rename it into place. On Windows, report the matching release asset and require manual replacement.

## Contracts and Invariants

| Contract ID | Connector / direction | Guarantees |
| --- | --- | --- |
| `CTR-01` | user → `pull` → ownership engine | Existing synchronization flags, plans, conflicts and outcomes are preserved under `pull`. |
| `CTR-02` | `update` → GitHub Release API/assets → running executable | Only a newer stable compatible release with a valid `checksums.txt` entry and staged version may replace the invoked executable. |

- `INV-01` Equal or newer installed versions exit successfully without download or replacement.
- `INV-02` Checksum, staged-version, permission, download, platform or rename failure leaves the current executable unchanged.
- `INV-03` The updater follows the resolved executable rather than guessing another PATH entry.

## Failure Modes / Rollback

| ID | Condition | Response |
| --- | --- | --- |
| `FM-01` | release metadata, matching asset, checksum or staged version is invalid | fail clearly; retain existing executable |
| `FM-02` | unsupported platform or replacement permission failure | fail clearly; Windows gives manual asset guidance; retain existing executable |
| `RB-01` | regression after merge | revert the feature commit; no data migration exists |

## Design Verification

| Analysis | Required | Method / evidence |
| --- | --- | --- |
| Contract compatibility | yes | command routing and pull regressions (`CHK-01`) |
| Failure propagation | yes | deterministic HTTP/checksum/version/rename/platform tests (`CHK-02`) |
| Security boundary | yes | checksum and staged-version checks (`INV-02`, `CHK-02`) |
| Migration safety | yes | README and help audit including `update` → `pull` guidance (`CHK-03`) |
| Concurrency / capacity | no | no shared writer or performance-sensitive path changes |

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure / rollout refs |
| --- | --- | --- | --- |
| `REQ-01` | `SOL-01` | `CTR-01` | `RB-01` |
| `REQ-02` | `SOL-02`–`SOL-03` | `CTR-02`, `INV-01`–`INV-03` | `FM-01`–`FM-02`, `RB-01` |
| `REQ-03`–`REQ-04` | `SOL-01`, `SOL-03` | `CTR-01`–`CTR-02` | `FM-02` |
| `REQ-05` | `SOL-01`–`SOL-03` | `INV-01`–`INV-02` | `FM-01`–`FM-02` |
