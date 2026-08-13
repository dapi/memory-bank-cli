package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dapi/memory-bank-cli/internal/analyzegraph"
)

func TestAnalyzeGraphCommandSupportsJSONAndMarkdown(t *testing.T) {
	handoff := filepath.Join("..", "analyzegraph", "testdata", "strong-coaccess.json")

	var jsonOutput bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := Run([]string{"analyze-graph", "--handoff", handoff, "--json"}, "test", &jsonOutput, &stderr); exitCode != 0 {
		t.Fatalf("JSON command exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	var report analyzegraph.Report
	if err := json.Unmarshal(jsonOutput.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Findings) == 0 || len(report.Recommendations) == 0 || len(report.Evidence) == 0 {
		t.Fatalf("JSON output did not distinguish report sections: %#v", report)
	}

	var markdownOutput bytes.Buffer
	if exitCode := Run([]string{"analyze-graph", "--handoff", handoff}, "test", &markdownOutput, &stderr); exitCode != 0 {
		t.Fatalf("Markdown command exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	for _, section := range []string{"## Recommendations", "## Findings", "## Evidence"} {
		if !strings.Contains(markdownOutput.String(), section) {
			t.Fatalf("markdown output missing %q:\n%s", section, markdownOutput.String())
		}
	}
}

func TestAnalyzeGraphCommandRequiresHandoff(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if exitCode := Run([]string{"analyze-graph"}, "test", &stdout, &stderr); exitCode != exitUsage {
		t.Fatalf("exit code = %d, want %d; stderr = %s", exitCode, exitUsage, stderr.String())
	}
	if !strings.Contains(stderr.String(), "--handoff is required") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
