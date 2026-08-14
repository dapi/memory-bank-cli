---
title: memory-bank-cli Events
doc_kind: domain
doc_function: canonical
purpose: "Canonical owner observable command and audit events."
derived_from:
  - states.md
  - ../use-cases/README.md
status: active
---

# Events

| Event | Producer | Observable result |
| --- | --- | --- |
| Template source pinned | `init` / `pull` | Source checkout/ref validation succeeds before planning. |
| Ownership plan produced | `init` / `pull` | Text or JSON decisions; dry run changes no files. |
| Update applied | `init` / `pull` | Atomic mutation completes and lock is written/updated. |
| Update conflict detected | `init` / `pull` | Conflict report and failure result. |
| CLI update completed | `update` | The installed CLI is replaced with a verified newer compatible release, or reports that it is already current. |
| Documentation audited | `lint` | Navigation report and exit result. |
| Repository diagnosed | `doctor` | Governance/adoption/drift/navigation findings and summary. |

No external event transport, queue or persistence beyond repository files is evidenced.
