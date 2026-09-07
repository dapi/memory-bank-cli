package contracts

import (
	"encoding/json"
	"reflect"
	"testing"
)

func fixture() (Manifest, map[string][]byte) {
	const legacy = "f1f04de843aef45a2425d4a7351d577bbf89e940"
	inventory := map[string][]byte{ManifestPath: []byte("manifest"), "memory-bank/README.md": []byte("index"), "memory-bank/dna/rules.json": []byte("rules"), "memory-bank/document-types/feature.json": []byte("type"), "memory-bank/flows/contracts/feature.json": []byte("bundle"), ".codex/agents/a.toml": []byte("adapter")}
	m := Manifest{SchemaVersion: 1, Capabilities: []string{"adoption/v1", "components/v1"}, DNAContract: "memory-bank/dna/rules.json", Components: map[string]Component{"dna": {Dependencies: []string{}}, "documents": {Dependencies: []string{"dna"}}, "flows": {Dependencies: []string{"dna", "documents"}}, "codex": {Dependencies: []string{"flows"}, Adapter: true, Legacy: true}}, Presets: map[string][]string{"core": {"dna"}, "docs": {"dna", "documents"}, "full": {"dna", "documents", "flows"}, "legacy": {"codex", "dna", "documents", "flows"}}, Files: map[string]File{}, DocumentTypes: map[string]string{"feature": "memory-bank/document-types/feature.json"}, Contracts: map[string]BundleRef{"feature/v1": {"memory-bank/flows/contracts/feature.json", Digest([]byte("bundle"))}}, LegacySources: map[string]Compatibility{legacy: {"legacy-f1f04de/v1", map[string]string{"feature": "feature/v1"}}}, LegacyDefaultSourceRef: legacy}
	for p := range inventory {
		m.Files[p] = File{"dna", "managed"}
	}
	m.Files["memory-bank/document-types/feature.json"] = File{"documents", "managed"}
	m.Files["memory-bank/flows/contracts/feature.json"] = File{"flows", "managed"}
	m.Files[".codex/agents/a.toml"] = File{"codex", "managed"}
	return m, inventory
}
func TestPresetMatrix(t *testing.T) {
	m, inv := fixture()
	data, _ := json.Marshal(m)
	parsed, err := ReadManifest(data, inv)
	if err != nil {
		t.Fatal(err)
	}
	m = parsed
	for _, tc := range []struct {
		preset   string
		adapters []string
		want     []string
	}{{"core", nil, []string{"dna"}}, {"docs", nil, []string{"dna", "documents"}}, {"full", nil, []string{"dna", "documents", "flows"}}, {"", nil, []string{"codex", "dna", "documents", "flows"}}, {"full", []string{"codex"}, []string{"codex", "dna", "documents", "flows"}}, {"core", []string{"codex"}, []string{"codex", "dna", "documents", "flows"}}} {
		s, err := m.Select(tc.preset, tc.adapters, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(sorted(append(append([]string{}, s.Components...), s.Adapters...)), tc.want) {
			t.Fatalf("%s: %#v", tc.preset, s)
		}
		again, err := m.Select("", nil, &s)
		if err != nil || !reflect.DeepEqual(again, s) {
			t.Fatalf("selection changed: %#v %v", again, err)
		}
	}
	docs, _ := m.Select("docs", nil, nil)
	if _, err := m.Select("core", nil, &docs); err == nil {
		t.Fatal("accepted downgrade")
	}
	if full, err := m.Select("full", nil, &docs); err != nil || !full.Has("flows") {
		t.Fatal("upgrade failed")
	}
	withAdapter, _ := m.Select("full", []string{"codex"}, nil)
	if s, err := m.Select("core", nil, &withAdapter); err != nil || !s.Has("flows") {
		t.Fatalf("compared before retained adapter closure: %v", err)
	}
}
func TestInventoryRejections(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*Manifest, map[string][]byte)
	}{
		{"unknown file", func(m *Manifest, i map[string][]byte) { i["extra"] = nil }},
		{"cycle", func(m *Manifest, i map[string][]byte) {
			m.Components["codex"] = Component{Dependencies: []string{"codex"}, Adapter: true, Legacy: true}
		}},
		{"manifest ownership", func(m *Manifest, i map[string][]byte) { m.Files[ManifestPath] = File{"documents", "managed"} }},
		{"bundle tamper", func(m *Manifest, i map[string][]byte) {
			i["memory-bank/flows/contracts/feature.json"] = []byte("changed")
		}},
		{"legacy omitted adapter", func(m *Manifest, i map[string][]byte) { m.Presets["legacy"] = []string{"dna", "documents", "flows"} }},
		{"unknown component", func(m *Manifest, i map[string][]byte) { m.Files["memory-bank/README.md"] = File{"unknown", "managed"} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, i := fixture()
			tc.edit(&m, i)
			data, _ := json.Marshal(m)
			if _, err := ReadManifest(data, i); err == nil {
				t.Fatal("accepted invalid manifest")
			}
		})
	}
}
func TestPortablePaths(t *testing.T) {
	for _, paths := range [][]string{{"x/Foo.md", "x/foo.md"}, {"x/é.md", "x/e\u0301.md"}, {"x/straße.md", "x/STRASSE.md"}, {"A/x.md", "a/y.md"}, {"../outside"}, {"x/.GIT/config"}, {"/absolute"}, {"x\\y"}} {
		if err := CheckPortable(paths); err == nil {
			t.Fatalf("accepted %q", paths)
		}
	}
	if err := CheckPortable([]string{"x/a.md", "x/b.md"}); err != nil {
		t.Fatal(err)
	}
}
func TestStrictJSON(t *testing.T) {
	for _, s := range []string{`{"schema_version":1,"schema_version":1}`, `{"Schema_Version":1}`, `{"schema_version":1,"unknown":false}`, `{"schema_version":"1"}`, `null`, `{} {}`, `{"schema_version":null}`} {
		var m Manifest
		if err := Decode([]byte(s), &m); err == nil {
			t.Fatalf("accepted %s", s)
		}
	}
}
func TestCanonicalStructsAndFrames(t *testing.T) {
	type v struct {
		Z string `json:"z"`
		A int    `json:"a"`
	}
	data, err := Canonical(v{"<\n", 1})
	if err != nil || string(data) != `{"a":1,"z":"\u003c\n"}` {
		t.Fatalf("%s %v", data, err)
	}
	if FramedID("d", "ab", "c") == FramedID("d", "a", "bc") {
		t.Fatal("ambiguous frames")
	}
}

func TestUnicodeConformance(t *testing.T) {
	for input, want := range map[string]string{"Foo.md": "foo.md", "e\u0301.md": "é.md", "É.md": "é.md", "Straße.md": "strasse.md", "STRASSE.md": "strasse.md", "K.md": "k.md", "Σ.md": "σ.md", "ς.md": "σ.md"} {
		if got := PortableKey(input); got != want {
			t.Fatalf("%q: %q != %q", input, got, want)
		}
	}
}
func TestMissingRequiredField(t *testing.T) {
	var c Component
	if err := Decode([]byte(`{"dependencies":[],"adapter":false}`), &c); err == nil {
		t.Fatal("missing legacy field was accepted")
	}
}
