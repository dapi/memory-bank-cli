package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dapi/memory-bank-cli/internal/doctor"
	"github.com/dapi/memory-bank-cli/internal/lint"
)

func testRepository(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "lint", "testdata", "repository"))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func commitCLISource(t *testing.T, root, message string) string {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, ".git")); os.IsNotExist(err) {
		runCLIGit(t, root, "init", "--quiet")
	}
	runCLIGit(t, root, "add", "--all")
	runCLIGit(t, root, "-c", "user.name=Memory Bank Tests", "-c", "user.email=tests@example.invalid", "commit", "--quiet", "-m", message)
	return runCLIGit(t, root, "rev-parse", "HEAD")
}

func runCLIGit(t *testing.T, root string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, arguments...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", arguments, err, output)
	}
	return strings.TrimSpace(string(output))
}

func TestLintJSONReport(t *testing.T) {
	arguments := []string{"--repo-root", testRepository(t), "--max-depth", "1", "--json"}
	var primaryStdout, primaryStderr bytes.Buffer
	primaryExit := Run(append([]string{"lint"}, arguments...), "test", &primaryStdout, &primaryStderr)
	if primaryExit != 1 {
		t.Fatalf("unexpected exit=%d stderr=%q", primaryExit, primaryStderr.String())
	}
	var report lint.Report
	if err := json.Unmarshal(primaryStdout.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON report: %v", err)
	}
	if report.FormatVersion != 1 || report.Stats.BrokenLinkCount != 1 {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestRootHelpAndVersion(t *testing.T) {
	for _, test := range []struct {
		arguments []string
		want      string
	}{
		{arguments: []string{"--help"}, want: "Usage: memory-bank-cli <command>"},
		{arguments: []string{"--version"}, want: "memory-bank-cli v1.2.3\n"},
	} {
		var stdout, stderr bytes.Buffer
		if exitCode := Run(test.arguments, "v1.2.3", &stdout, &stderr); exitCode != 0 {
			t.Fatalf("unexpected exit code %d for %v: %s", exitCode, test.arguments, stderr.String())
		}
		if !strings.Contains(stdout.String(), test.want) {
			t.Fatalf("unexpected stdout for %v: %q", test.arguments, stdout.String())
		}
		if test.arguments[0] == "--help" && !strings.Contains(stdout.String(), "push    Publish locked canonical template changes upstream through a PR") {
			t.Fatalf("root help does not document push: %q", stdout.String())
		}
	}
}

func TestPushHelpDocumentsDryRun(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if exitCode := Run([]string{"push", "--help"}, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("help exit=%d stderr=%q", exitCode, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Usage: memory-bank-cli push") || !strings.Contains(stderr.String(), "without mutating checkout, remotes, or GitHub") {
		t.Fatalf("push help is incomplete: %q", stderr.String())
	}
}

func TestGitHubAdapterDryRunJSON(t *testing.T) {
	repo := t.TempDir()
	var stdout, stderr bytes.Buffer
	if exitCode := Run([]string{"github", "init", "--repo-root", repo, "--dry-run", "--json"}, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("unexpected exit=%d stderr=%q", exitCode, stderr.String())
	}
	var report struct {
		DryRun    bool `json:"dry_run"`
		Decisions []struct {
			Action string `json:"action"`
		} `json:"decisions"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil || !report.DryRun || len(report.Decisions) != 5 {
		t.Fatalf("unexpected report=%s err=%v", stdout.String(), err)
	}
	if _, err := os.Stat(filepath.Join(repo, ".github")); !os.IsNotExist(err) {
		t.Fatalf("dry run changed repository: %v", err)
	}
}

func TestGitHubAdapterHelpSucceeds(t *testing.T) {
	for _, command := range []string{"init", "update"} {
		t.Run(command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if exitCode := Run([]string{"github", command, "--help"}, "test", &stdout, &stderr); exitCode != 0 {
				t.Fatalf("help exit=%d stderr=%q", exitCode, stderr.String())
			}
			if !strings.Contains(stderr.String(), "-dry-run") {
				t.Fatalf("help output missing flags: %q", stderr.String())
			}
		})
	}
}

func TestRootRejectsMissingAndUnknownCommands(t *testing.T) {
	for _, arguments := range [][]string{nil, {"unknown"}} {
		var stdout, stderr bytes.Buffer
		if exitCode := Run(arguments, "test", &stdout, &stderr); exitCode != 2 {
			t.Fatalf("unexpected exit code %d for %v", exitCode, arguments)
		}
		if !strings.Contains(stderr.String(), "Usage: memory-bank-cli <command>") {
			t.Fatalf("unexpected stderr for %v: %q", arguments, stderr.String())
		}
	}
}

func TestLintRejectsNegativeDepth(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if exitCode := Run([]string{"lint", "--max-depth", "-1"}, "test", &stdout, &stderr); exitCode != 2 {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "greater than or equal to 0") {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}

func TestOwnershipDryRunJSONReportsPlanWithoutMutation(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "memory-bank", "dna"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "memory-bank", "dna", "rule.md"), []byte("rule\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sourceRef := commitCLISource(t, source, "initial source")
	arguments := []string{"init", "--repo-root", repo, "--source", source, "--template-version", "v1", "--source-ref", sourceRef, "--dry-run", "--json"}
	var stdout, stderr bytes.Buffer
	if exitCode := Run(arguments, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("unexpected exit %d: %s", exitCode, stderr.String())
	}
	var report struct {
		FormatVersion int  `json:"format_version"`
		DryRun        bool `json:"dry_run"`
		Decisions     []struct {
			Action string `json:"action"`
		} `json:"decisions"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("invalid report: %v\n%s", err, stdout.String())
	}
	if report.FormatVersion != 1 || !report.DryRun || len(report.Decisions) != 2 || report.Decisions[0].Action != "create" || report.Decisions[1].Action != "create" {
		t.Fatalf("unexpected report: %#v", report)
	}
	if _, err := os.Stat(filepath.Join(repo, "memory-bank", "dna", "rule.md")); !os.IsNotExist(err) {
		t.Fatalf("dry-run mutated repository: %v", err)
	}
}

func TestUpdateWithoutSourceUsesRepoUpstreamMain(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "memory-bank", "dna"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "memory-bank", "dna", "rule.md"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	initialRef := commitCLISource(t, source, "initial source")
	remote := filepath.Join(t.TempDir(), "upstream.git")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	runCLIGit(t, remote, "init", "--bare", "--quiet")
	runCLIGit(t, source, "branch", "-M", "main")
	runCLIGit(t, source, "remote", "add", "origin", remote)
	runCLIGit(t, source, "push", "--quiet", "origin", "main")
	runCLIGit(t, remote, "symbolic-ref", "HEAD", "refs/heads/main")

	var stdout, stderr bytes.Buffer
	initArgs := []string{"init", "--repo-root", repo, "--source", source, "--template-version", "v1", "--source-ref", initialRef}
	if exitCode := Run(initArgs, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("init failed with %d: %s", exitCode, stderr.String())
	}
	checkout := filepath.Join(repo, "memory-bank", ".repo")
	runCLIGit(t, filepath.Dir(checkout), "clone", "--quiet", remote, ".repo")
	if err := os.WriteFile(filepath.Join(source, "memory-bank", "dna", "rule.md"), []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	updatedRef := commitCLISource(t, source, "update source")
	runCLIGit(t, source, "push", "--quiet", "origin", "main")

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"update", "--repo-root", repo}, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("default update failed with %d: %s", exitCode, stderr.String())
	}
	if got, err := os.ReadFile(filepath.Join(repo, "memory-bank", "dna", "rule.md")); err != nil || string(got) != "v2\n" {
		t.Fatalf("default update did not apply main: %q, %v", got, err)
	}
	lock, err := os.ReadFile(filepath.Join(repo, "memory-bank", ".lock"))
	if err != nil || !strings.Contains(string(lock), updatedRef) || !strings.Contains(string(lock), "main@"+updatedRef[:12]) {
		t.Fatalf("lock does not record fetched main: %q, %v", lock, err)
	}
	if got := runCLIGit(t, checkout, "rev-parse", "HEAD"); got != initialRef {
		t.Fatalf("user upstream checkout was changed: got %s want %s", got, initialRef)
	}
}

func TestInitWithoutSourceUsesRepoUpstreamMain(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "memory-bank", "dna"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "memory-bank", "dna", "rule.md"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ref := commitCLISource(t, source, "initial source")
	remote := filepath.Join(t.TempDir(), "upstream.git")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	runCLIGit(t, remote, "init", "--bare", "--quiet")
	runCLIGit(t, source, "branch", "-M", "main")
	runCLIGit(t, source, "remote", "add", "origin", remote)
	runCLIGit(t, source, "push", "--quiet", "origin", "main")
	runCLIGit(t, remote, "symbolic-ref", "HEAD", "refs/heads/main")
	checkout := filepath.Join(repo, "memory-bank", ".repo")
	if err := os.MkdirAll(filepath.Dir(checkout), 0o755); err != nil {
		t.Fatal(err)
	}
	runCLIGit(t, filepath.Dir(checkout), "clone", "--quiet", remote, ".repo")

	var stdout, stderr bytes.Buffer
	if exitCode := Run([]string{"init", "--repo-root", repo}, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("default init failed with %d: %s", exitCode, stderr.String())
	}
	if got, err := os.ReadFile(filepath.Join(repo, "memory-bank", "dna", "rule.md")); err != nil || string(got) != "v1\n" {
		t.Fatalf("default init did not apply main: %q, %v", got, err)
	}
	lock, err := os.ReadFile(filepath.Join(repo, "memory-bank", ".lock"))
	if err != nil || !strings.Contains(string(lock), ref) || !strings.Contains(string(lock), "main@"+ref[:12]) {
		t.Fatalf("lock does not record fetched main: %q, %v", lock, err)
	}
}

func TestResolveUpdateUpstreamUsesFallbackWithoutRepoCheckout(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "memory-bank"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "memory-bank", "README.md"), []byte("template\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ref := commitCLISource(t, source, "source")
	remote := filepath.Join(t.TempDir(), "upstream.git")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	runCLIGit(t, remote, "init", "--bare", "--quiet")
	runCLIGit(t, source, "branch", "-M", "main")
	runCLIGit(t, source, "remote", "add", "origin", remote)
	runCLIGit(t, source, "push", "--quiet", "origin", "main")

	resolved, err := resolveUpdateUpstreamFrom(repo, remote)
	if err != nil {
		t.Fatal(err)
	}
	defer resolved.cleanup()
	if resolved.ref != ref || resolved.version != "main@"+ref[:12] {
		t.Fatalf("unexpected resolved upstream: %#v", resolved)
	}
	if _, err := os.Stat(filepath.Join(resolved.sourceRoot, "memory-bank", "README.md")); err != nil {
		t.Fatalf("temporary checkout lacks fetched payload: %v", err)
	}
}

func TestUpdateWithoutSourceRejectsDirtyRepoCheckoutWithoutMutation(t *testing.T) {
	repo := t.TempDir()
	checkout := filepath.Join(repo, "memory-bank", ".repo")
	if err := os.MkdirAll(checkout, 0o755); err != nil {
		t.Fatal(err)
	}
	runCLIGit(t, checkout, "init", "--quiet")
	if err := os.WriteFile(filepath.Join(checkout, "dirty"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := []byte("unchanged\n")
	if err := os.WriteFile(filepath.Join(repo, "sentinel"), before, 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if exitCode := Run([]string{"update", "--repo-root", repo}, "test", &stdout, &stderr); exitCode != 1 || !strings.Contains(stderr.String(), "dirty") {
		t.Fatalf("unexpected exit=%d stderr=%q", exitCode, stderr.String())
	}
	after, err := os.ReadFile(filepath.Join(repo, "sentinel"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("failed resolution changed downstream: %q, %v", after, err)
	}
}

func TestInitWithoutSourceRejectsInvalidRepoCheckoutWithoutMutation(t *testing.T) {
	repo := t.TempDir()
	checkout := filepath.Join(repo, "memory-bank", ".repo")
	if err := os.MkdirAll(checkout, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "sentinel"), []byte("unchanged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if exitCode := Run([]string{"init", "--repo-root", repo}, "test", &stdout, &stderr); exitCode != 1 || !strings.Contains(stderr.String(), "not a Git checkout") {
		t.Fatalf("unexpected exit=%d stderr=%q", exitCode, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(repo, "memory-bank", ".lock")); !os.IsNotExist(err) {
		t.Fatalf("failed resolution wrote lock: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(repo, "sentinel")); err != nil || string(got) != "unchanged\n" {
		t.Fatalf("failed resolution changed repository: %q, %v", got, err)
	}
}

func TestInitWithoutSourceRejectsRepoCheckoutWithoutOriginWithoutMutation(t *testing.T) {
	repo := t.TempDir()
	checkout := filepath.Join(repo, "memory-bank", ".repo")
	if err := os.MkdirAll(checkout, 0o755); err != nil {
		t.Fatal(err)
	}
	runCLIGit(t, checkout, "init", "--quiet")
	var stdout, stderr bytes.Buffer
	if exitCode := Run([]string{"init", "--repo-root", repo}, "test", &stdout, &stderr); exitCode != 1 || !strings.Contains(stderr.String(), "has no usable origin") {
		t.Fatalf("unexpected exit=%d stderr=%q", exitCode, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(repo, "memory-bank", ".lock")); !os.IsNotExist(err) {
		t.Fatalf("failed resolution wrote lock: %v", err)
	}
}

func TestDoctorAndAlternativeAgentTarget(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "memory-bank"), 0o755); err != nil {
		t.Fatal(err)
	}
	readme := "---\ndoc_function: index\npurpose: Test index for doctor.\nstatus: active\n---\n# Memory Bank\n"
	if err := os.WriteFile(filepath.Join(source, "memory-bank", "README.md"), []byte(readme), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "CLAUDE.md"), []byte("project rules\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sourceRef := commitCLISource(t, source, "source")
	args := []string{"init", "--repo-root", repo, "--source", source, "--template-version", "v1", "--source-ref", sourceRef, "--agent-file", "CLAUDE.md"}
	var stdout, stderr bytes.Buffer
	if exitCode := Run(args, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("init failed with %d: %s", exitCode, stderr.String())
	}
	claude, err := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if err != nil || !strings.HasPrefix(string(claude), "project rules\n\n<!-- MEMORY BANK START -->") {
		t.Fatalf("alternative target did not preserve content: %q, %v", claude, err)
	}
	if _, err := os.Stat(filepath.Join(repo, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("default target was also created: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"doctor", "--repo-root", repo, "--agent-file", "CLAUDE.md", "--json"}, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("doctor failed with %d: %s", exitCode, stderr.String())
	}
	var report struct {
		FormatVersion int `json:"format_version"`
		Summary       struct {
			Errors int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil || report.FormatVersion != doctor.ReportFormatVersion || report.Summary.Errors != 0 {
		t.Fatalf("unexpected doctor report: %s, %v", stdout.String(), err)
	}
	lockBefore, err := os.ReadFile(filepath.Join(repo, "memory-bank", ".lock"))
	if err != nil {
		t.Fatal(err)
	}
	outdated := []byte("project rules\n\n<!-- MEMORY BANK START -->\nold\n<!-- MEMORY BANK END -->\n")
	if err := os.WriteFile(filepath.Join(repo, "CLAUDE.md"), outdated, 0o644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	updateArgs := append([]string{"update"}, args[1:]...)
	if exitCode := Run(updateArgs, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("managed-block update failed with %d: %s", exitCode, stderr.String())
	}
	after, _ := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if !bytes.Equal(after, claude) {
		t.Fatalf("managed-block update did not restore current payload: %q", after)
	}
	lockAfter, err := os.ReadFile(filepath.Join(repo, "memory-bank", ".lock"))
	if err != nil || !bytes.Equal(lockBefore, lockAfter) {
		t.Fatalf("agent-only update changed template lock: %v", err)
	}
	before := append([]byte(nil), after...)
	stdout.Reset()
	stderr.Reset()
	if exitCode := Run(updateArgs, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("idempotent update failed with %d: %s", exitCode, stderr.String())
	}
	after, _ = os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if !bytes.Equal(before, after) {
		t.Fatalf("idempotent update changed target")
	}
}

func TestDoctorFixAdoptsMissingLockWithExplicitProvenance(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	readme := "---\ndoc_function: index\npurpose: Test index for doctor.\nstatus: active\n---\n# Memory Bank\n"
	if err := os.MkdirAll(filepath.Join(source, "memory-bank"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "memory-bank"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "memory-bank", "README.md"), []byte(readme), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "memory-bank", "README.md"), []byte(readme), 0o644); err != nil {
		t.Fatal(err)
	}
	sourceRef := commitCLISource(t, source, "source")
	base := []string{"doctor", "--fix", "--repo-root", repo, "--source", source, "--template-version", "v1", "--source-ref", sourceRef}

	var stdout, stderr bytes.Buffer
	if exitCode := Run(append(append([]string{}, base...), "--dry-run", "--json"), "test", &stdout, &stderr); exitCode != 1 {
		t.Fatalf("dry-run exit = %d, stderr=%s", exitCode, stderr.String())
	}
	var dryRun struct {
		Repair struct {
			Finding string `json:"finding"`
			Plan    struct {
				DryRun    bool `json:"dry_run"`
				Decisions []struct {
					Path string `json:"path"`
				} `json:"decisions"`
			} `json:"plan"`
		} `json:"repair"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &dryRun); err != nil {
		t.Fatalf("invalid dry-run report: %v\n%s", err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "memory-bank-cli doctor --fix") {
		t.Fatalf("missing-lock remediation does not guide the repair: %s", stdout.String())
	}
	if dryRun.Repair.Finding != "template.identity_missing" || !dryRun.Repair.Plan.DryRun || len(dryRun.Repair.Plan.Decisions) == 0 {
		t.Fatalf("unexpected dry-run repair: %#v", dryRun)
	}
	lockPlanned := false
	lockDecisionCount := 0
	for _, decision := range dryRun.Repair.Plan.Decisions {
		if decision.Path == "memory-bank/.lock" {
			lockPlanned = true
			lockDecisionCount++
		}
	}
	if !lockPlanned {
		t.Fatalf("dry-run omitted the lock creation: %#v", dryRun.Repair.Plan.Decisions)
	}
	if lockDecisionCount != 1 {
		t.Fatalf("dry-run should contain one lock creation, got %d decisions: %#v", lockDecisionCount, dryRun.Repair.Plan.Decisions)
	}
	if _, err := os.Stat(filepath.Join(repo, "memory-bank", ".lock")); !os.IsNotExist(err) {
		t.Fatalf("dry-run created a lock: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run(base, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("repair exit = %d, stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "commit memory-bank/.lock") {
		t.Fatalf("repair did not recommend committing the lock: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "create\t\tmemory-bank/.lock\trecord successful adoption") {
		t.Fatalf("repair report omitted the lock creation: %s", stdout.String())
	}
	lock, err := os.ReadFile(filepath.Join(repo, "memory-bank", ".lock"))
	if err != nil || !strings.Contains(string(lock), sourceRef) {
		t.Fatalf("repair did not create the pinned lock: %q, %v", lock, err)
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"doctor", "--fix", "--repo-root", repo, "--json"}, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("repair with an existing lock failed: exit=%d stderr=%s", exitCode, stderr.String())
	}
	var existing struct {
		Repair any `json:"repair"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &existing); err != nil || existing.Repair != nil {
		t.Fatalf("existing lock should not produce a repair: %#v, %v", existing, err)
	}
}

func TestDoctorFixRequiresProvenanceAndPreservesConflictingManagedContent(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	readme := "---\ndoc_function: index\npurpose: Test index for doctor.\nstatus: active\n---\n# Memory Bank\n"
	if err := os.MkdirAll(filepath.Join(source, "template", "memory-bank"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "memory-bank"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "template", "memory-bank", "README.md"), []byte(readme), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "memory-bank", "README.md"), []byte("local customization\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sourceRef := commitCLISource(t, source, "source")

	var stdout, stderr bytes.Buffer
	if exitCode := Run([]string{"doctor", "--fix", "--repo-root", repo}, "test", &stdout, &stderr); exitCode != exitUsage {
		t.Fatalf("missing provenance exit = %d, stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stderr.String(), "requires --source, --template-version, and --source-ref") {
		t.Fatalf("missing provenance error is not actionable: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	arguments := []string{"doctor", "--fix", "--repo-root", repo, "--source", source, "--template-version", "v1", "--source-ref", sourceRef, "--json"}
	if exitCode := Run(arguments, "test", &stdout, &stderr); exitCode != exitFailure {
		t.Fatalf("conflicting repair exit = %d, stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "existing managed file does not match initialization source") {
		t.Fatalf("conflicting managed content was not reported: %s", stdout.String())
	}
	if data, err := os.ReadFile(filepath.Join(repo, "memory-bank", "README.md")); err != nil || string(data) != "local customization\n" {
		t.Fatalf("conflicting managed content changed: %q, %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(repo, "memory-bank", ".lock")); !os.IsNotExist(err) {
		t.Fatalf("conflicting repair created a lock: %v", err)
	}
}

func TestDoctorRejectsUnknownProfile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if exitCode := Run([]string{"doctor", "--profile", "mystery"}, "test", &stdout, &stderr); exitCode != 2 {
		t.Fatalf("unexpected exit %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "auto, template, or downstream") {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}

func TestDoctorRejectsScopeRootOutsideRepository(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if exitCode := Run([]string{"doctor", "--scope-root", "../sibling"}, "test", &stdout, &stderr); exitCode != 1 {
		t.Fatalf("unexpected exit %d: %s", exitCode, stderr.String())
	}
	if !strings.Contains(stderr.String(), "must not contain parent-directory traversal") {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}

func TestInitRejectsAmbiguousMarkersWithoutMutation(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "memory-bank"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "memory-bank", "README.md"), []byte("template\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	original := []byte("rules\n<!-- MEMORY BANK START -->\n<!-- MEMORY BANK START -->\n<!-- MEMORY BANK END -->\n")
	if err := os.WriteFile(filepath.Join(repo, "AGENTS.md"), original, 0o644); err != nil {
		t.Fatal(err)
	}
	sourceRef := commitCLISource(t, source, "source")
	arguments := []string{"init", "--repo-root", repo, "--source", source, "--template-version", "v1", "--source-ref", sourceRef}
	var stdout, stderr bytes.Buffer
	if exitCode := Run(arguments, "test", &stdout, &stderr); exitCode != 1 {
		t.Fatalf("unexpected exit %d: %s", exitCode, stderr.String())
	}
	current, err := os.ReadFile(filepath.Join(repo, "AGENTS.md"))
	if err != nil || !bytes.Equal(current, original) {
		t.Fatalf("conflict changed user content: %q, %v", current, err)
	}
	if _, err := os.Stat(filepath.Join(repo, "memory-bank", "README.md")); !os.IsNotExist(err) {
		t.Fatalf("conflict partially installed template: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "memory-bank", ".lock")); !os.IsNotExist(err) {
		t.Fatalf("conflict created lock: %v", err)
	}
}

func TestOwnershipDryRunJSONReportsNewManagedPathCollision(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	dna := filepath.Join(source, "memory-bank", "dna")
	if err := os.MkdirAll(dna, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dna, "seed.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	initialRef := commitCLISource(t, source, "initial source")
	baseArguments := []string{"--repo-root", repo, "--source", source}
	initArguments := append([]string{"init"}, baseArguments...)
	initArguments = append(initArguments, "--template-version", "v1", "--source-ref", initialRef)
	var stdout, stderr bytes.Buffer
	if exitCode := Run(initArguments, "test", &stdout, &stderr); exitCode != 0 {
		t.Fatalf("unexpected init exit %d: %s", exitCode, stderr.String())
	}

	collision := filepath.Join("memory-bank", "dna", "collision.md")
	if err := os.WriteFile(filepath.Join(repo, collision), []byte("downstream\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, collision), []byte("upstream\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	updatedRef := commitCLISource(t, source, "add collision")
	stdout.Reset()
	stderr.Reset()
	updateArguments := append([]string{"update"}, baseArguments...)
	updateArguments = append(updateArguments, "--template-version", "v2", "--source-ref", updatedRef, "--dry-run", "--json")
	if exitCode := Run(updateArguments, "test", &stdout, &stderr); exitCode != 1 {
		t.Fatalf("unexpected update exit %d: %s", exitCode, stderr.String())
	}
	var report struct {
		Decisions []struct {
			Path      string `json:"path"`
			Ownership string `json:"ownership"`
			Action    string `json:"action"`
		} `json:"decisions"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("invalid report: %v\n%s", err, stdout.String())
	}
	for _, decision := range report.Decisions {
		if decision.Path == filepath.ToSlash(collision) {
			if decision.Action != "conflict" || decision.Ownership != "user-owned" {
				t.Fatalf("collision decision disagrees with safety contract: %#v", decision)
			}
			return
		}
	}
	t.Fatalf("collision decision missing from report: %#v", report.Decisions)
}
