package handoff

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildFeatureFlowHandoffIsDeterministic(t *testing.T) {
	repositoryRoot := copyFixture(t, filepath.Join("testdata", "feature-flow"))
	runGitTestCommand(t, repositoryRoot, "init", "--quiet")
	runGitTestCommand(t, repositoryRoot, "add", "--all")
	runGitTestCommand(t, repositoryRoot, "-c", "user.name=Memory Bank Tests", "-c", "user.email=tests@example.invalid", "commit", "--quiet", "-m", "fixture")
	if err := os.WriteFile(filepath.Join(repositoryRoot, "src.go"), []byte("package fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitTestCommand(t, repositoryRoot, "add", "src.go")
	runGitTestCommand(t, repositoryRoot, "-c", "user.name=Memory Bank Tests", "-c", "user.email=tests@example.invalid", "commit", "--quiet", "-m", "implement feature")

	planBefore, err := os.ReadFile(filepath.Join(repositoryRoot, "features/FT-042/implementation-plan.md"))
	if err != nil {
		t.Fatal(err)
	}
	options := Options{
		RepoRoot:    repositoryRoot,
		From:        "features/FT-042/implementation-plan.md",
		GitRange:    "HEAD~1..HEAD",
		TestReports: []string{"reports/test-results.json"},
	}
	report, err := Build(options)
	if err != nil {
		t.Fatal(err)
	}
	planAfter, err := os.ReadFile(filepath.Join(repositoryRoot, "features/FT-042/implementation-plan.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(planBefore, planAfter) {
		t.Fatal("handoff build changed the source document")
	}
	if report.StartingDocument != "features/FT-042/implementation-plan.md" || len(report.DeclaredPriming.Dependencies) != 2 {
		t.Fatalf("unexpected document projection: %#v", report.DeclaredPriming)
	}
	if len(report.ObservedExecution.Commits) != 1 || len(report.ObservedExecution.ChangedFiles) != 1 {
		t.Fatalf("unexpected git projection: %#v", report.ObservedExecution)
	}
	if report.ObservedExecution.Commits[0].Subject != "implement feature" || report.ObservedExecution.ChangedFiles[0].Path != "src.go" {
		t.Fatalf("unexpected commit evidence: %#v", report.ObservedExecution.Commits[0])
	}
	if report.ObservedExecution.ChangedFiles[0].Source.Ref != report.ObservedExecution.Commits[0].Hash {
		t.Fatalf("changed file lost its commit source: %#v", report.ObservedExecution.ChangedFiles[0])
	}
	if len(report.ObservedExecution.Verification) != 1 || report.ObservedExecution.Verification[0].Source.Ref != "reports/test-results.json" {
		t.Fatalf("unexpected verification projection: %#v", report.ObservedExecution.Verification)
	}
	for _, item := range report.Items {
		if item.Source.Ref == "" {
			t.Fatalf("item has no source: %#v", item)
		}
		if item.Context != "declared_priming" && item.Context != "observed_execution" {
			t.Fatalf("item has invalid context: %#v", item)
		}
	}
	firstJSON, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := Build(options)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(repeated)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) || RenderMarkdown(report) != RenderMarkdown(repeated) {
		t.Fatal("handoff output is not deterministic")
	}
	markdown := RenderMarkdown(report)
	for _, expected := range []string{"Declared priming context", "Observed execution context", "DEC-01", "HEAD~1..HEAD", "reports/test-results.json", "passed"} {
		if !strings.Contains(markdown, expected) {
			t.Fatalf("markdown omitted %q:\n%s", expected, markdown)
		}
	}
}

func TestBuildReportsUnresolvedEvidence(t *testing.T) {
	repositoryRoot := copyFixture(t, filepath.Join("testdata", "feature-flow"))
	report, err := Build(Options{
		RepoRoot:    repositoryRoot,
		From:        "features/FT-042/implementation-plan.md",
		TestReports: []string{"reports/missing.json"},
	})
	if err == nil || !strings.Contains(err.Error(), "unresolved evidence") {
		t.Fatalf("expected unresolved evidence error, got %v", err)
	}
	if len(report.Unresolved) != 1 || report.Unresolved[0].Reference != "reports/missing.json" {
		t.Fatalf("unexpected unresolved report: %#v", report.Unresolved)
	}
	if !strings.Contains(RenderMarkdown(report), "reports/missing.json") {
		t.Fatal("markdown omitted unresolved source")
	}
}

func copyFixture(t *testing.T, source string) string {
	t.Helper()
	destination := t.TempDir()
	entries, err := os.ReadDir(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		copyFixtureEntry(t, filepath.Join(source, entry.Name()), filepath.Join(destination, entry.Name()))
	}
	return destination
}

func copyFixtureEntry(t *testing.T, source, destination string) {
	t.Helper()
	info, err := os.Stat(source)
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() {
		if err := os.MkdirAll(destination, 0o755); err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(source)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			copyFixtureEntry(t, filepath.Join(source, entry.Name()), filepath.Join(destination, entry.Name()))
		}
		return
	}
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, data, info.Mode().Perm()); err != nil {
		t.Fatal(err)
	}
}

func runGitTestCommand(t *testing.T, root string, arguments ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, arguments...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", arguments, err, output)
	}
}
