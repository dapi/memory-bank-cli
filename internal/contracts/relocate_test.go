package contracts

import (
	"strings"
	"testing"
)

func TestBaseRelocationPreservesReferenceTargets(t *testing.T) {
	input := "---\r\nstatus: draft\r\nderived_from:\r\n  - '../dna/frontmatter.md' # owner\r\n  - {path: ../dna/lifecycle.md, fit: exact}\r\n---\r\n\r\n[DNA](../dna/README.md#baseline)\r\n[ref]: ../dna/README.md\r\n![asset](../assets/icon.png)\r\n```\r\n[example](../dna/README.md)\r\n```\r\n"
	got, err := RelocateBaseDocument([]byte(input), "memory-bank/templates/feature.md", "memory-bank/features/FT-1/brief.md")
	if err != nil {
		t.Fatal(err)
	}
	expected := strings.Replace(input, "'../dna/frontmatter.md'", `"../../dna/frontmatter.md"`, 1)
	expected = strings.Replace(expected, "path: ../dna/lifecycle.md", `path: "../../dna/lifecycle.md"`, 1)
	expected = strings.Replace(expected, "[DNA](../dna/README.md#baseline)", "[DNA](../../dna/README.md#baseline)", 1)
	expected = strings.Replace(expected, "[ref]: ../dna/README.md", "[ref]: ../../dna/README.md", 1)
	expected = strings.Replace(expected, "![asset](../assets/icon.png)", "![asset](../../assets/icon.png)", 1)
	if string(got) != expected {
		t.Fatalf("relocation changed unrelated bytes:\n%s", got)
	}
}

func TestBaseRelocationDecodesPercentEscapesOnce(t *testing.T) {
	input := []byte("---\nstatus: draft\n---\n[x](docs/My%20File.md?q=1#part)\n[y](docs/Literal%2520.md)\n")
	got, err := RelocateBaseDocument(input, "memory-bank/templates/feature.md", "memory-bank/features/FT-1/brief.md")
	if err != nil {
		t.Fatal(err)
	}
	want := strings.ReplaceAll(string(input), "](docs/", "](../../templates/docs/")
	if string(got) != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestBaseRelocationRejectsUnsupportedAbsoluteSchemes(t *testing.T) {
	for _, ref := range []string{"HTTPS://example.org", "tel:+123", "ftp://example.org", "custom:document"} {
		input := []byte("---\nstatus: draft\n---\n[x](" + ref + ")\n")
		if _, err := RelocateBaseDocument(input, "memory-bank/templates/feature.md", "memory-bank/features/FT-1/brief.md"); err == nil {
			t.Fatalf("absolute URI silently relocated: %s", ref)
		}
	}
}
