package contracts

import "testing"

func TestDraftCopyRejectsRelativeReferences(t *testing.T) {
	for _, body := range []string{"[source](../source.md)", "[source]: ../source.md", "![image](image.png)", "<a href='relative.md'>link</a>", "[link]:\n  relative.md"} {
		d, e := ParseDocument("drafts/input.md", []byte("---\nstatus: draft\n---\n"+body))
		if e != nil {
			t.Fatal(e)
		}
		if e = ValidateDraftCopy(d, "memory-bank/features/FT-1/brief.md"); e == nil {
			t.Fatalf("relative reference accepted: %s", body)
		}
	}
	for _, body := range []string{"[source](/memory-bank/README.md)", "[site](https://example.org)", "[self](#heading)", "plain draft", "```\n[example](relative.md)\n```"} {
		d, e := ParseDocument("drafts/input.md", []byte("---\nstatus: draft\n---\n"+body))
		if e != nil {
			t.Fatal(e)
		}
		if e = ValidateDraftCopy(d, "memory-bank/features/FT-1/brief.md"); e != nil {
			t.Fatalf("independent draft rejected: %s: %v", body, e)
		}
	}
}

func TestDraftGrammarVectors(t *testing.T) {
	for _, tc := range []struct {
		text  string
		valid bool
	}{
		{`[x](/memory-bank/README.md "Index")`, true},
		{`![x](https://example.org/x.png)`, true},
		{`[x]: </memory-bank/README.md>`, true},
		{`[x](https://example.org/a(b))`, false},
		{`[x](https://example.org/a&amp;b)`, false},
		{"`[example](relative.md)`", true},
		{"<!-- [example](relative.md) -->", true},
		{"---\nstatus: draft\nderived_from: [{path: /memory-bank/README.md, fit: exact}]\n---\n", true},
		{"---\nstatus: draft\nderived_from: &dep [/memory-bank/README.md]\n---\n", false},
		{"---\nstatus: draft\nderived_from: {path: [relative.md]}\n---\n", false},
	} {
		d, err := ParseDocument("drafts/input.md", []byte(tc.text))
		if err != nil {
			t.Fatal(err)
		}
		err = ValidateDraftCopy(d, "memory-bank/features/FT-1/brief.md")
		if (err == nil) != tc.valid {
			t.Fatalf("%q valid=%v: %v", tc.text, tc.valid, err)
		}
	}
}
