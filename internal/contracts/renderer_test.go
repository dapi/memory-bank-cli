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
