package ownership

import "testing"

func TestCanonicalPayloadPathRoundTripsWithoutPathRules(t *testing.T) {
	for _, sourceRelative := range []string{
		"memory-bank/dna/rule.md",
		".config/hidden",
		".github/workflows/check.yml",
		"AGENTS.md",
		"nested/new/path/tool",
	} {
		downstream := CanonicalDownstreamPath(sourceRelative)
		if downstream != sourceRelative {
			t.Fatalf("downstream path for %q = %q", sourceRelative, downstream)
		}
		if got, want := CanonicalTemplatePath(downstream), CanonicalTemplateRoot+"/"+sourceRelative; got != want {
			t.Fatalf("canonical round trip for %q = %q, want %q", sourceRelative, got, want)
		}
	}
}
