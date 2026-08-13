package analyzegraph

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeFixtures(t *testing.T) {
	tests := []struct {
		name         string
		fixture      string
		findingKinds []string
		wantTypes    []string
	}{
		{name: "strong co-access", fixture: "strong-coaccess.json", findingKinds: []string{"repeated_co_access"}, wantTypes: []string{"co_access", "derived_from"}},
		{name: "high fan-out", fixture: "high-fanout.json", findingKinds: []string{"high_fan_out"}, wantTypes: []string{"navigation"}},
		{name: "unresolved link", fixture: "unresolved-link.json", findingKinds: []string{"unresolved_link", "weak_evidence"}, wantTypes: []string{"derived_from", "depends_on"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := readFixture(t, test.fixture)
			for _, kind := range test.findingKinds {
				if !hasFinding(report, kind) {
					t.Fatalf("finding %q not present: %#v", kind, report.Findings)
				}
			}
			if len(report.Recommendations) != 1 {
				t.Fatalf("recommendations = %d, want 1", len(report.Recommendations))
			}
			for _, wantType := range test.wantTypes {
				if !contains(report.Recommendations[0].RelationTypes, wantType) && !containsFindingTypes(report, wantType) {
					t.Fatalf("relation type %q was not preserved", wantType)
				}
			}
			if len(report.Recommendations[0].ContributingNodes) == 0 || len(report.Recommendations[0].Sources) == 0 {
				t.Fatalf("recommendation lost contributing nodes or sources: %#v", report.Recommendations[0])
			}
		})
	}
}

func TestAnalyzePreservesTypedEvidenceInJSON(t *testing.T) {
	report := readFixture(t, "strong-coaccess.json")
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Report
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.FormatVersion != ReportFormatVersion {
		t.Fatalf("format_version = %d, want %d", decoded.FormatVersion, ReportFormatVersion)
	}
	if !contains(decoded.Recommendations[0].RelationTypes, "co_access") || !contains(decoded.Recommendations[0].RelationTypes, "derived_from") {
		t.Fatalf("typed relations collapsed or lost: %#v", decoded.Recommendations[0].RelationTypes)
	}
	if !contains(decoded.Recommendations[0].Sources, "handoff:FT-042") {
		t.Fatalf("relation source lost: %#v", decoded.Recommendations[0].Sources)
	}
}

func TestPrintMarkdownDistinguishesSectionsAndEvidence(t *testing.T) {
	report := readFixture(t, "strong-coaccess.json")
	var output bytes.Buffer
	PrintMarkdown(&output, report)
	for _, want := range []string{"## Recommendations", "## Findings", "## Evidence", "co_access", "handoff:FT-042", "Contributing nodes"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("markdown output missing %q:\n%s", want, output.String())
		}
	}
}

func TestAnalyzeNormalizesGraphEdgesAndEvidenceRelations(t *testing.T) {
	data := []byte(`{
		"kind": "execution_handoff",
		"graph": {
			"nodes": [
				{"id": "a", "type": "document", "source": "a.md"},
				{"id": "b", "type": "decision", "source": "b.md"}
			],
			"edges": [
				{"source": "a", "target": "b", "relation_type": "co_access", "evidence": "run-1.log"}
			]
		},
		"evidence": [
			{"from": "a", "to": "b", "relation_type": "co_access", "source": "run-2.log"}
		]
	}`)
	report, err := Analyze("handoff.json", data)
	if err != nil {
		t.Fatal(err)
	}
	if hasFinding(report, "unresolved_link") {
		t.Fatalf("graph edge aliases were not resolved: %#v", report.Findings)
	}
	if !hasFinding(report, "repeated_co_access") {
		t.Fatalf("repeated co-access finding missing: %#v", report.Findings)
	}
	if report.Evidence[2].Source != "run-1.log" || report.Evidence[3].Source != "run-2.log" {
		t.Fatalf("relation provenance was not preserved: %#v", report.Evidence)
	}
}

func readFixture(t *testing.T, name string) Report {
	t.Helper()
	path := filepath.Join("testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	report, err := Analyze(path, data)
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func hasFinding(report Report, kind string) bool {
	for _, finding := range report.Findings {
		if finding.Kind == kind {
			return true
		}
	}
	return false
}

func containsFindingTypes(report Report, want string) bool {
	for _, finding := range report.Findings {
		if contains(finding.RelationTypes, want) {
			return true
		}
	}
	return false
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
