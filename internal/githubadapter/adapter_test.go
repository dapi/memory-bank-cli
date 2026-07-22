package githubadapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func decision(t *testing.T, report Report, path string) Decision {
	t.Helper()
	for _, item := range report.Decisions {
		if item.Path == path {
			return item
		}
	}
	t.Fatalf("missing decision for %s: %#v", path, report)
	return Decision{}
}

func TestInitCreatesOptInAdapterAndUpdateIsIdempotent(t *testing.T) {
	repo := t.TempDir()
	report, err := Run(Options{RepoRoot: repo})
	if err != nil || !report.Applied || len(report.Decisions) != 5 {
		t.Fatalf("init failed: report=%#v err=%v", report, err)
	}
	path := filepath.Join(repo, ".github", "ISSUE_TEMPLATE", "memory-bank-feature.yml")
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), "Expected outcome") || !strings.Contains(string(data), "# MB-CLI GITHUB ADAPTER START") {
		t.Fatalf("feature form was not installed: %q, %v", data, err)
	}
	var form yaml.Node
	if err := yaml.Unmarshal(data, &form); err != nil {
		t.Fatalf("installed feature form is not valid YAML: %v", err)
	}
	report, err = Run(Options{RepoRoot: repo})
	if err != nil || report.Applied || report.ConflictCount != 0 || decision(t, report, ".github/pull_request_template.md").Action != Preserve {
		t.Fatalf("repeated update was not idempotent: report=%#v err=%v", report, err)
	}
}

func TestMalformedReversedMarkerIsAConflict(t *testing.T) {
	item := defaultAssets()[0]
	// The end marker appears before the terminator of this deliberately broken start line.
	malformed := "# MB-CLI GITHUB ADAPTER START: " + item.id + " sha256:missing" + markerSyntax(item).end + item.content
	_, action, _ := reconcile(item, malformed)
	if action != Conflict {
		t.Fatalf("malformed marker action=%q, want conflict", action)
	}
}

func TestDryRunDoesNotWrite(t *testing.T) {
	repo := t.TempDir()
	report, err := Run(Options{RepoRoot: repo, DryRun: true})
	if err != nil || !report.DryRun || report.Applied || decision(t, report, ".github/pull_request_template.md").Action != Create {
		t.Fatalf("unexpected dry run: report=%#v err=%v", report, err)
	}
	if _, err := os.Stat(filepath.Join(repo, ".github")); !os.IsNotExist(err) {
		t.Fatalf("dry run wrote GitHub tree: %v", err)
	}
}

func TestApplyFailureRollsBackAllManagedAssets(t *testing.T) {
	repo := t.TempDir()
	writes := 0
	report, err := Run(Options{RepoRoot: repo, writeFile: func(path string, data []byte) error {
		writes++
		if writes == 3 {
			return os.ErrPermission
		}
		return atomicWriteFile(path, data)
	}})
	if err == nil || report.Applied || writes != 3 {
		t.Fatalf("expected third write to fail atomically: report=%#v err=%v writes=%d", report, err, writes)
	}
	if _, statErr := os.Stat(filepath.Join(repo, ".github")); !os.IsNotExist(statErr) {
		t.Fatalf("failed apply left adapter files or directories: %v", statErr)
	}
}

func TestExistingCustomTemplatesArePreserved(t *testing.T) {
	repo := t.TempDir()
	path := filepath.Join(repo, ".github", "pull_request_template.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# Our custom PR template\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := Run(Options{RepoRoot: repo})
	if err != nil || report.ConflictCount != 0 || decision(t, report, ".github/pull_request_template.md").Action != Preserve {
		t.Fatalf("custom template was not preserved: report=%#v err=%v", report, err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "# Our custom PR template\n" {
		t.Fatalf("custom template changed: %q", data)
	}
}

func TestEditedManagedBlockConflictsWithoutWriting(t *testing.T) {
	repo := t.TempDir()
	if _, err := Run(Options{RepoRoot: repo}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repo, ".github", "pull_request_template.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(data), "## What changed", "## Custom changed", 1)
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := Run(Options{RepoRoot: repo})
	if err != nil || report.Applied || report.ConflictCount != 1 || decision(t, report, ".github/pull_request_template.md").Action != Conflict {
		t.Fatalf("managed edit did not conflict: report=%#v err=%v", report, err)
	}
	current, _ := os.ReadFile(path)
	if string(current) != edited {
		t.Fatal("conflict overwrote user edit")
	}
}

func TestSymlinkDestinationIsRejected(t *testing.T) {
	repo := t.TempDir()
	if err := os.Symlink(t.TempDir(), filepath.Join(repo, ".github")); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(Options{RepoRoot: repo}); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected unsafe symlink error, got %v", err)
	}
}
