package contracts

import (
	"strings"
	"testing"
)

func registryFixture(t *testing.T) (Registry, Catalog, map[string]Document) {
	t.Helper()
	m, _ := fixture()
	legacyID := "legacy/f1f04de/feature/v1"
	legacy, _ := LegacyBundle(legacyID, "feature")
	next := legacy
	next.ID = "feature/v1"
	next.Legacy = false
	oldHash, newHash := Digest([]byte("legacy")), Digest([]byte("new"))
	m.Contracts = map[string]BundleRef{legacyID: {"legacy.json", oldHash}, next.ID: {"new.json", newHash}}
	m.LegacySources[m.LegacyDefaultSourceRef] = Compatibility{"legacy-f1f04de/v1", map[string]string{"feature": legacyID}}
	c := Catalog{Manifest: m, Types: map[string]DocumentType{"feature": {Type: "feature"}}, Bundles: map[string]Bundle{legacyID: legacy, next.ID: next}}
	first := Identity{"doc-" + strings.Repeat("1", 64), "memory-bank/features/FT-1/brief.md", "feature", "memory-bank/features/FT-1"}
	second := Identity{"doc-" + strings.Repeat("2", 64), "memory-bank/features/FT-2/brief.md", "feature", "memory-bank/features/FT-2"}
	sel := Selector{SelectorID(m.LegacyDefaultSourceRef, "feature", legacyID, oldHash), m.LegacyDefaultSourceRef, "feature", legacyID, oldHash, []Identity{first, second}, []string{}}
	r := EmptyRegistry()
	r.Selectors = []Selector{sel}
	docs := map[string]Document{}
	for _, id := range []Identity{first, second} {
		r.History = append(r.History, Event{"migrate", id.ID, id.Path, id.Path, "", legacyID, []string{}})
		raw, _ := Project([]byte("---\nstatus: draft\n---\n# Feature\n"), map[string]string{"document_type": "feature", "document_id": id.ID})
		docs[id.Path] = document(t, id.Path, string(raw))
	}
	return r, c, docs
}
func TestSelectorTransitionAndMoveHistory(t *testing.T) {
	r, c, docs := registryFixture(t)
	if active, err := ValidateRegistry(r, c, docs); err != nil || len(active) != 2 {
		t.Fatalf("snapshot: %v", err)
	}
	id := r.Selectors[0].Snapshot[0]
	old := r.Selectors[0].ContractID
	next := "feature/v1"
	r.Selectors[0].Exclusions = []string{id.ID}
	r.Records = []Record{{id.ID, id.Path, id.Type, id.ContextRoot, next, c.Manifest.Contracts[next].Digest}}
	r.History = append(r.History, Event{"transition", id.ID, id.Path, id.Path, old, next, []string{"review/1"}})
	raw, _ := Project(docs[id.Path].Raw, map[string]string{"flow_contract": next})
	docs[id.Path] = document(t, id.Path, string(raw))
	active, err := ValidateRegistry(r, c, docs)
	if err != nil || active[id.ID].SelectorID != "" {
		t.Fatalf("transition: %v", err)
	}
	moved := id.ContextRoot + "/renamed.md"
	r.Records[0].Path = moved
	r.History = append(r.History, Event{"move", id.ID, id.Path, moved, next, next, []string{}})
	docs[moved] = document(t, moved, string(raw))
	delete(docs, id.Path)
	if active, err = ValidateRegistry(r, c, docs); err != nil || active[id.ID].Identity.Path != moved {
		t.Fatalf("move: %v", err)
	}
	r.Selectors[0].Exclusions = []string{}
	if _, err = ValidateRegistry(r, c, docs); err == nil {
		t.Fatal("accepted selector plus record without exclusion")
	}
}
func TestRegistryTamperingFails(t *testing.T) {
	for _, name := range []string{"missing target", "removed id", "orphan", "changed type", "history", "selector id", "missing bundle"} {
		t.Run(name, func(t *testing.T) {
			r, c, docs := registryFixture(t)
			id := r.Selectors[0].Snapshot[0]
			switch name {
			case "missing target":
				delete(docs, id.Path)
			case "removed id":
				d := docs[id.Path]
				delete(d.Fields, "document_id")
				docs[id.Path] = d
			case "orphan":
				docs["memory-bank/features/FT-1/copy.md"] = docs[id.Path]
			case "changed type":
				d := docs[id.Path]
				d.Fields["document_type"] = "adr"
				docs[id.Path] = d
			case "history":
				r.History[0].ToPath = "memory-bank/features/FT-1/other.md"
			case "selector id":
				r.Selectors[0].ID = "sel-tampered"
			case "missing bundle":
				delete(c.Bundles, r.Selectors[0].ContractID)
			}
			if _, err := ValidateRegistry(r, c, docs); err == nil {
				t.Fatal("accepted tampered state")
			}
		})
	}
}
func TestCanonicalRegistryAndBaseDocument(t *testing.T) {
	r, c, docs := registryFixture(t)
	base := document(t, "memory-bank/features/FT-3/brief.md", "---\nstatus: draft\ndocument_type: feature\n---\n# Base\n")
	docs[base.Path] = base
	active, err := ValidateRegistry(r, c, docs)
	if err != nil || len(active) != 2 {
		t.Fatalf("base joined snapshot: %v", err)
	}
	data, err := RegistryBytes(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ReadRegistry(data); err != nil {
		t.Fatal(err)
	}
	if _, err = ReadRegistry(append([]byte(" "), data...)); err == nil {
		t.Fatal("accepted noncanonical registry")
	}
}
