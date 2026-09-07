package contracts

import (
	"github.com/dapi/memory-bank-cli/internal/agentinstructions"
	"strings"
	"testing"
)

func TestRendererVersionsAndOutsideBytes(t *testing.T) {
	m := Manifest{Files: map[string]File{"memory-bank/features/README.md": {Component: "documents"}}}
	s := Installation{Components: []string{"dna", "documents"}, RendererVersion: CurrentRenderer()}
	block := ReadmeBlock(m, s)
	if strings.Contains(string(block), "Flows") || !strings.Contains(string(block), " — project documents.") {
		t.Fatal(string(block))
	}
	original := []byte("owned\r\n\r\n" + string(block) + "after\r\n")
	plan := agentinstructions.BuildPlanWithBlock(original, block)
	if plan.Status != agentinstructions.Current {
		t.Fatal(plan.Status)
	}
	drift := []byte(strings.Replace(string(original), " — project documents.", "", 1))
	if agentinstructions.BuildPlanWithBlock(drift, block).Status != agentinstructions.Outdated {
		t.Fatal("v2 annotation drift accepted")
	}
	s.RendererVersion = nil
	v1 := ReadmeBlock(m, s)
	legacy := []byte("owned\r\n\r\n" + string(v1) + "after\r\n")
	if agentinstructions.BuildPlanWithBlock(legacy, v1).Status != agentinstructions.Current {
		t.Fatal("v1 rejected")
	}
	upgraded := agentinstructions.BuildPlanWithBlock(legacy, block).Data
	if string(upgraded) != string(original) {
		t.Fatal("renderer changed outside bytes")
	}
	if strings.Contains(string(AgentBlock(s)), "flows/routing") {
		t.Fatal("Documents requires Flows")
	}
}

func TestRendererThreePreservesHistoricalBlocks(t *testing.T) {
	m := Manifest{Files: map[string]File{}}
	for _, documents := range []bool{false, true} {
		components := []string{"dna"}
		if documents {
			components = append(components, "documents")
		}
		two, three := 2, 3
		s := Installation{Components: components, RendererVersion: &two}
		v2 := ReadmeBlock(m, s)
		s.RendererVersion = &three
		v3 := ReadmeBlock(m, s)
		if documents {
			if !strings.Contains(string(v2), "— project-owned draft templates.") || !strings.Contains(string(v3), "— managed templates for project-owned drafts.") {
				t.Fatal("canonical annotations missing")
			}
			expected := strings.Replace(string(v2), "— project-owned draft templates.", "— managed templates for project-owned drafts.", 1)
			if expected != string(v3) {
				t.Fatal("v3 changes more than the Templates annotation")
			}
			if agentinstructions.BuildPlanWithBlock(v2, v3).Status != agentinstructions.Outdated || agentinstructions.BuildPlanWithBlock(v3, v2).Status != agentinstructions.Outdated {
				t.Fatal("cross-version annotation drift accepted")
			}
		} else if string(v2) != string(v3) {
			t.Fatal("core renderer bytes changed")
		}
	}
}
