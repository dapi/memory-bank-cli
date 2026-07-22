---
title: memory-bank-cli Domain Model
doc_kind: domain
doc_function: canonical
purpose: "Canonical owner domain entities and relationships."
derived_from:
  - glossary.md
status: active
---

# Domain Model

```text
Template source --contains--> Memory Bank payload --adopted into--> Downstream repository
Downstream repository --has--> Ownership lock --identifies--> Template (version, source ref)
Ownership lock --tracks--> File (path, ownership, digests, mode)
Update plan --decides--> File mutation / conflict
Doctor and lint --emit--> Versioned report --contains--> Findings
```

`Template`, `Lock`, `File`, `Decision` and `Report` are explicit Go data types. A file is classified as managed, adapted, user-owned or generated; the lock makes that classification and content provenance durable across update operations.
