package contracts

import (
	"bytes"
	"reflect"
	"testing"
)

func document(t *testing.T, p, s string) Document {
	t.Helper()
	d, e := ParseDocument(p, []byte(s))
	if e != nil {
		t.Fatal(e)
	}
	return d
}
func TestProjectionPreservesUnrelatedBytes(t *testing.T) {
	before := []byte("---\r\n# owner comment\r\nstatus: 'draft' # retain\r\ncustom: |\r\n  document_id: body data\r\n---\r\n\r\n# Title\r\n")
	updates := map[string]string{"document_type": "feature", "document_id": "doc-abc", "flow_contract": "feature/v1"}
	after, err := Project(before, updates)
	if err != nil {
		t.Fatal(err)
	}
	want := bytes.Replace(before, []byte("---\r\n\r\n#"), []byte("document_type: \"feature\"\r\ndocument_id: \"doc-abc\"\r\nflow_contract: \"feature/v1\"\r\n---\r\n\r\n#"), 1)
	if !bytes.Equal(after, want) {
		t.Fatalf("unexpected bytes: %q", after)
	}
	again, err := Project(after, updates)
	if err != nil || !bytes.Equal(again, after) {
		t.Fatalf("projection not idempotent: %v", err)
	}
	changed, err := Project(after, map[string]string{"flow_contract": "feature/v2"})
	if err != nil || !bytes.Equal(changed, bytes.Replace(after, []byte("feature/v1"), []byte("feature/v2"), 1)) {
		t.Fatalf("transition changed unrelated bytes: %v", err)
	}
}
func TestProjectionRejectsAmbiguousYAML(t *testing.T) {
	for _, raw := range []string{"---\nstatus: draft\nstatus: active\n---\n", "---\nstatus: draft\n", "---\ndocument_id: |\n  old\n---\n", "---\n'document_id': old\n---\n"} {
		if _, err := Project([]byte(raw), map[string]string{"document_id": "new"}); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
}
func TestRuleExtensionsCannotWeaken(t *testing.T) {
	parent := Rules{Fields: map[string][]string{"status": {"active", "draft"}}, Sections: []string{"Problem"}, ActiveRequiresUpstream: true}
	for _, child := range []Rules{{Fields: map[string][]string{"status": {}}}, {Fields: map[string][]string{"status": {"archived"}}}, {Fields: map[string][]string{"status": {"active", "archived", "draft"}}}} {
		if _, err := MergeRules(parent, child); err == nil {
			t.Fatal("accepted weakening")
		}
	}
	merged, err := MergeRules(parent, Rules{Fields: map[string][]string{"status": {"draft"}}, Sections: []string{"Verify"}})
	if err != nil || !merged.ActiveRequiresUpstream || !reflect.DeepEqual(merged.Sections, []string{"Problem", "Verify"}) {
		t.Fatalf("invalid strengthening: %#v %v", merged, err)
	}
}
func TestBaseAndFlowGateDiffer(t *testing.T) {
	d := document(t, "memory-bank/features/FT-1/brief.md", "---\nstatus: draft\ndocument_type: feature\n---\n# Feature\n## Problem\n")
	id := Identity{"", d.Path, "feature", "memory-bank/features/FT-1"}
	base := Rules{Fields: map[string][]string{"status": {"active", "archived", "draft"}}, Sections: []string{"Problem"}}
	if got := ValidateRules(d, base, id, nil); len(got) != 0 {
		t.Fatal(got)
	}
	flow, _ := MergeRules(base, Rules{Fields: map[string][]string{"delivery_status": {"planned"}}, Sections: []string{"Verify"}})
	if got := ValidateRules(d, flow, id, nil); len(got) != 2 {
		t.Fatalf("missing flow gates: %v", got)
	}
}
func TestLegacyFindingsSurviveProjectionAndPinnedDependencies(t *testing.T) {
	p := "memory-bank/features/FT-1/brief.md"
	d := document(t, p, "---\nstatus: active\nderived_from: ../../product/context.md\ndelivery_status: planned\n---\n# Feature\n")
	design := document(t, "memory-bank/features/FT-1/design.md", "---\nstatus: draft\nderived_from: brief.md\n---\n# Design\n")
	id := Identity{"doc-stable", p, "feature", "memory-bank/features/FT-1"}
	bundle, _ := LegacyBundle("legacy/f1f04de/feature/v1", "feature")
	ctx := map[string]Document{d.Path: d, design.Path: design}
	before := ValidateBundle(d, bundle, id, ctx)
	if len(before) != 1 || before[0].Code != "lifecycle.design_requirement_decision_invalid" {
		t.Fatalf("expected missing mandatory legacy decision section: %v", before)
	}
	raw, err := Project(d.Raw, map[string]string{"document_id": id.ID, "document_type": "feature"})
	if err != nil {
		t.Fatal(err)
	}
	after := document(t, p, string(raw))
	ctx[p] = after
	if got := ValidateBundle(after, bundle, id, ctx); !reflect.DeepEqual(got, before) {
		t.Fatalf("changed legacy verdict: %v / %v", before, got)
	}
	valid := document(t, p, string(raw)+"\n## Design Requirement Decision\nDesign required: yes\n")
	if got := ValidateBundle(valid, bundle, id, ctx); len(got) != 0 {
		t.Fatalf("valid legacy document failed: %v", got)
	}
	// Latest live rules are deliberately not supplied to ValidateBundle.
	stricter := Rules{Sections: []string{"New gate"}}
	if len(ValidateRules(valid, stricter, id, ctx)) == 0 || len(ValidateBundle(valid, bundle, id, ctx)) != 0 {
		t.Fatal("live rules leaked into frozen verdict")
	}
}
func TestHeadingsIgnoreCommentsAndFences(t *testing.T) {
	d := document(t, "memory-bank/a.md", "---\nstatus: draft\n---\n<!--\n## Hidden\n-->\n```markdown\n## Hidden too\n```\n## Real ##\n")
	if h := Headings(d); !reflect.DeepEqual(h, map[string]bool{"Real": true}) {
		t.Fatal(h)
	}
}
func TestWindowsAndUnicodePortableNames(t *testing.T) {
	for _, p := range []string{"x/CON.md", "x/LPT¹.txt", "x/COM9", "x/file:stream", "x/name.", "x/name ", "x/aux .txt", "x/a\x7fb"} {
		if ValidPath(p) {
			t.Fatalf("accepted %q", p)
		}
	}
	for input, want := range map[string]string{"Straße": "strasse", "Σ/ς/σ": "σ/σ/σ", "e\u0301.md": "é.md", "K": "k"} {
		if got := PortableKey(input); got != want {
			t.Fatalf("%q: %q", input, got)
		}
	}
}

func TestFenceCommentsDoNotHideLaterHeadings(t *testing.T) {
	d := document(t, "memory-bank/a.md", "---\nstatus: draft\n---\n```html\n<!-- unfinished literal comment\n```\n## C#\n## Real ###\n")
	if got := Headings(d); !reflect.DeepEqual(got, map[string]bool{"C#": true, "Real": true}) {
		t.Fatal(got)
	}
}

func TestLegacyDesignParserRetainsHistoricalCommentBehavior(t *testing.T) {
	d := document(t, "memory-bank/features/FT-1/brief.md", "---\nstatus: active\nderived_from: upstream.md\ndelivery_status: planned\n---\n<!--\n## Design Requirement Decision\nDesign required: yes\n-->\n")
	if _, valid := DesignDecision(d); valid {
		t.Fatal("modern engine treated a comment as a section")
	}
	if decision, valid := legacyDesignDecision(string(d.Raw)); !valid || decision != "yes" {
		t.Fatal("compatibility parser changed historical behavior")
	}
}
