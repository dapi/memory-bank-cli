package contracts

import (
	"github.com/dapi/memory-bank-cli/internal/agentinstructions"
	"strings"
)

func ReadmeBlock(m Manifest, s Installation) []byte {
	version := 1
	if s.RendererVersion != nil {
		version = *s.RendererVersion
	}
	lines := []string{agentinstructions.StartMarker, "## Installed components", ""}
	add := func(label, target, annotation string) {
		line := "- [" + label + "](" + target + ")"
		if version >= 2 {
			line += " — " + annotation + "."
		}
		lines = append(lines, line)
	}
	add("DNA", "dna/README.md", "governance baseline")
	if s.Has("documents") {
		add("Document types", "document-types/README.md", "base document contracts")
		annotation := "project-owned draft templates"
		if version >= 3 {
			annotation = "managed templates for project-owned drafts"
		}
		add("Templates", "templates/README.md", annotation)
		for _, name := range []string{"product", "domain", "engineering", "ops", "adr", "prd", "use-cases", "features", "research", "epics"} {
			if f, ok := m.Files["memory-bank/"+name+"/README.md"]; ok && s.Has(f.Component) {
				add(name, name+"/README.md", "project documents")
			}
		}
	}
	if s.Has("flows") {
		add("Flows", "flows/README.md", "optional process contracts")
	}
	return []byte(strings.Join(append(lines, agentinstructions.EndMarker), "\n") + "\n")
}
func AgentBlock(s Installation) []byte {
	b := strings.Replace(string(agentinstructions.CurrentBlock), "BLOCK VERSION: 3", "BLOCK VERSION: 4", 1)
	if !s.Has("flows") {
		b = strings.Replace(b, "memory-bank/README.md, memory-bank/dna/README.md, and memory-bank/flows/routing.md.", "memory-bank/README.md and memory-bank/dna/README.md.", 1)
	}
	return []byte(b)
}
func CurrentRenderer() *int { v := 3; return &v }
