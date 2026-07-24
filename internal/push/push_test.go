package push

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestDryRunIncludesOnlyManagedPaths(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "memory-bank", "dna"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "memory-bank", "dna", "rule.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "memory-bank", "product"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "memory-bank", "dna", "rule.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "memory-bank", "product", "note.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "init", "--quiet")
	git(t, root, "add", ".")
	git(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "--quiet", "-m", "base")
	upstream := filepath.Join(root, "memory-bank", ".repo")
	if err := os.MkdirAll(upstream, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, upstream, "init", "--quiet")
	git(t, upstream, "remote", "add", "origin", "https://github.com/example/upstream.git")
	if err := os.WriteFile(filepath.Join(root, "memory-bank", "dna", "rule.md"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "memory-bank", "product", "note.md"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := Run(Options{RepoRoot: root, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !report.DryRun || len(report.Decisions) < 2 {
		t.Fatalf("unexpected report: %#v", report)
	}
	decisions := map[string]string{}
	for _, d := range report.Decisions {
		decisions[d.Path] = d.Action
	}
	if decisions["memory-bank/dna/rule.md"] != "include" {
		t.Fatalf("managed path was not included: %#v", report.Decisions)
	}
	if decisions["memory-bank/product/note.md"] != "exclude" {
		t.Fatalf("adapted path was not excluded: %#v", report.Decisions)
	}
	if _, err := os.Stat(filepath.Join(upstream, ".git")); err != nil {
		t.Fatalf("dry run changed checkout: %v", err)
	}
}

func TestRejectsDirtyUpstreamCheckout(t *testing.T) {
	root := t.TempDir()
	upstream := filepath.Join(root, "memory-bank", ".repo")
	if err := os.MkdirAll(filepath.Join(root, "memory-bank", "dna"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "memory-bank", "dna", "rule.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "init", "--quiet")
	git(t, root, "add", ".")
	git(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "--quiet", "-m", "base")
	if err := os.MkdirAll(upstream, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, upstream, "init", "--quiet")
	git(t, upstream, "remote", "add", "origin", "https://github.com/example/upstream.git")
	if err := os.WriteFile(filepath.Join(upstream, "dirty.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Run(Options{RepoRoot: root, DryRun: true})
	if err == nil || !strings.Contains(err.Error(), "dirty") {
		t.Fatalf("want dirty error, got %v", err)
	}
}
