package push

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/dapi/memory-bank-cli/internal/ownership"
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

func TestRunCreatesBranchCopiesManagedFileAndReturnsPR(t *testing.T) {
	root := pushFixture(t)
	tool := filepath.Join(root, ".config", "tool")
	if err := os.MkdirAll(filepath.Dir(tool), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tool, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeManagedLock(t, root, "memory-bank/dna/rule.md", ".config/tool")
	checkout := filepath.Join(root, "memory-bank", ".repo")
	var calls []string
	run := func(dir, name string, args ...string) (string, error) {
		call := name + " " + strings.Join(args, " ")
		calls = append(calls, call)
		if call == "git rev-parse --show-toplevel" {
			return dir, nil
		}
		switch call {
		case "git rev-parse --is-inside-work-tree", "git status --porcelain", "git diff --name-only --diff-filter=U", "git rev-parse --verify origin/main", "git add -- template/.config/tool template/memory-bank/dna/rule.md", "git commit -m Publish managed Memory Bank changes", "git push -u origin memory-bank-cli/push-20260724-120000":
			return "", nil
		case "git remote get-url origin":
			return "https://github.com/example/upstream.git", nil
		case "gh repo view example/upstream --json id":
			return "{\"id\":\"R_1\"}", nil
		case "git ls-remote --exit-code --heads origin memory-bank-cli/push-20260724-120000":
			return "", errors.New("not found")
		case "git symbolic-ref --short refs/remotes/origin/HEAD":
			return "origin/main", nil
		case "git fetch origin main:refs/remotes/origin/main":
			return "", nil
		case "git branch --show-current":
			return "main", nil
		case "git rev-parse HEAD":
			return "abc123", nil
		case "git ls-tree -d --name-only origin/main -- template":
			return "template", nil
		case "git ls-tree -d --name-only origin/main -- memory-bank-template":
			return "", nil
		case "git ls-tree -d --name-only origin/main -- memory-bank":
			return "", nil
		case "git status --porcelain=v1 -z --untracked-files=all":
			return " M .config/tool\x00 M memory-bank/dna/rule.md\x00 M memory-bank/product/note.md\x00", nil
		case "git checkout -b memory-bank-cli/push-20260724-120000 origin/main":
			if dir != checkout {
				t.Fatalf("branch created outside checkout: %q", dir)
			}
			return "", nil
		case "gh pr create --repo example/upstream --head memory-bank-cli/push-20260724-120000 --base main --fill":
			return "https://github.com/example/upstream/pull/1", nil
		default:
			return "", errors.New("unexpected command: " + call)
		}
	}
	report, err := Run(Options{RepoRoot: root, Now: func() time.Time { return time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC) }, BranchName: func(time.Time) (string, error) { return "memory-bank-cli/push-20260724-120000", nil }, Run: run})
	if err != nil {
		t.Fatal(err)
	}
	if report.Branch != "memory-bank-cli/push-20260724-120000" || report.PRURL != "https://github.com/example/upstream/pull/1" {
		t.Fatalf("unexpected report: %#v", report)
	}
	data, err := os.ReadFile(filepath.Join(checkout, "template", "memory-bank", "dna", "rule.md"))
	if err != nil || string(data) != "changed\n" {
		t.Fatalf("managed file was not copied: %q, %v", data, err)
	}
	data, err = os.ReadFile(filepath.Join(checkout, "template", ".config", "tool"))
	if err != nil || string(data) != "#!/bin/sh\n" {
		t.Fatalf("canonical root file was not copied: %q, %v", data, err)
	}
	if runtime.GOOS != "windows" {
		info, statErr := os.Stat(filepath.Join(checkout, "template", ".config", "tool"))
		if statErr != nil {
			t.Fatal(statErr)
		}
		if info.Mode().Perm() != 0o755 {
			t.Fatalf("canonical executable mode = %v", info.Mode())
		}
	}
	if _, err := os.Stat(filepath.Join(checkout, "template", "memory-bank", "product", "note.md")); !os.IsNotExist(err) {
		t.Fatalf("excluded path was copied: %v", err)
	}
	for _, call := range calls {
		if call == "git checkout -" {
			t.Fatalf("successful run restored checkout: %v", calls)
		}
	}
}

func TestRunCompensatesRemoteBranchWhenPRCreationFails(t *testing.T) {
	root := pushFixture(t)
	var calls []string
	run := func(dir string, name string, args ...string) (string, error) {
		call := name + " " + strings.Join(args, " ")
		calls = append(calls, call)
		if call == "git rev-parse --show-toplevel" {
			return dir, nil
		}
		switch call {
		case "git rev-parse --is-inside-work-tree", "git status --porcelain", "git diff --name-only --diff-filter=U", "git rev-parse --verify origin/main", "git add -- template/memory-bank/dna/rule.md", "git commit -m Publish managed Memory Bank changes", "git push -u origin memory-bank-cli/push-20260724-120000", "git push origin --delete memory-bank-cli/push-20260724-120000", "git reset --hard", "git checkout main", "git branch -D memory-bank-cli/push-20260724-120000", "git reset --hard abc123", "git clean -fd -- template/memory-bank/dna/rule.md":
			return "", nil
		case "git remote get-url origin":
			return "https://github.com/example/upstream.git", nil
		case "gh repo view example/upstream --json id":
			return "{\"id\":\"R_1\"}", nil
		case "git symbolic-ref --short refs/remotes/origin/HEAD":
			return "origin/main", nil
		case "git fetch origin main:refs/remotes/origin/main":
			return "", nil
		case "git branch --show-current":
			return "main", nil
		case "git rev-parse HEAD":
			return "abc123", nil
		case "git ls-tree -d --name-only origin/main -- template":
			return "template", nil
		case "git ls-tree -d --name-only origin/main -- memory-bank-template":
			return "", nil
		case "git ls-tree -d --name-only origin/main -- memory-bank":
			return "", nil
		case "git status --porcelain=v1 -z --untracked-files=all":
			return " M memory-bank/dna/rule.md\x00", nil
		case "git ls-remote --exit-code --heads origin memory-bank-cli/push-20260724-120000":
			return "", errors.New("not found")
		case "git checkout -b memory-bank-cli/push-20260724-120000 origin/main":
			return "", nil
		case "gh pr create --repo example/upstream --head memory-bank-cli/push-20260724-120000 --base main --fill":
			return "", errors.New("forbidden")
		default:
			return "", errors.New("unexpected command: " + call)
		}
	}
	_, err := Run(Options{RepoRoot: root, Now: func() time.Time { return time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC) }, BranchName: func(time.Time) (string, error) { return "memory-bank-cli/push-20260724-120000", nil }, Run: run})
	if err == nil || !strings.Contains(err.Error(), "create PR: forbidden") {
		t.Fatalf("want compensated failure, got %v", err)
	}
	joined := strings.Join(calls, "\n")
	for _, want := range []string{"git push origin --delete memory-bank-cli/push-20260724-120000", "git reset --hard", "git checkout main"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing compensation %q in %v", want, calls)
		}
	}
}

func pushFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "memory-bank", "dna"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "memory-bank", ".repo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "memory-bank", ".repo", "template", "memory-bank"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "memory-bank", "dna", "rule.md"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func writeManagedLock(t *testing.T, root string, paths ...string) {
	t.Helper()
	files := make(map[string]ownership.File, len(paths))
	for _, relative := range paths {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		value := fmt.Sprintf("sha256:%x", sha256.Sum256(data))
		mode := "100644"
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0o111 != 0 {
			mode = "100755"
		}
		files[relative] = ownership.File{Ownership: ownership.Managed, BaseDigest: value, PayloadDigest: value, BaseMode: mode, PayloadMode: mode}
	}
	lock := ownership.Lock{
		SchemaVersion: ownership.CurrentSchemaVersion,
		Template:      ownership.Template{Version: "v1", SourceRef: strings.Repeat("a", 40)},
		LastUpdate:    ownership.UpdateRecord{Version: "v1", At: time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)},
		Files:         files,
	}
	data, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, filepath.FromSlash(ownership.LockFileName))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		t.Fatal(err)
	}
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
	if err := os.MkdirAll(filepath.Join(upstream, "memory-bank-template"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(upstream, "memory-bank-template", ".keep"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, upstream, "add", ".")
	git(t, upstream, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "--quiet", "-m", "base")
	git(t, upstream, "update-ref", "refs/remotes/origin/master", "HEAD")
	git(t, upstream, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/master")
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

func TestDryRunIncludesCanonicalTemplatePathsOutsideMemoryBank(t *testing.T) {
	source, root := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "template", "memory-bank", "dna"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(source, "template", ".config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "template", "memory-bank", "dna", "rule.md"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "template", ".config", "hidden"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, source, "init", "--quiet")
	git(t, source, "add", ".")
	git(t, source, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "--quiet", "-m", "template")
	ref := git(t, source, "rev-parse", "HEAD")
	if _, err := ownership.Init(ownership.Options{RepoRoot: root, SourceRoot: source, TemplateVersion: "v1", SourceRef: ref}); err != nil {
		t.Fatal(err)
	}
	git(t, root, "init", "--quiet")
	git(t, root, "add", ".")
	git(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "--quiet", "-m", "downstream")
	upstream := filepath.Join(root, "memory-bank", ".repo")
	if err := os.MkdirAll(filepath.Join(upstream, "template", "memory-bank"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(upstream, "template", "memory-bank", ".keep"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, upstream, "init", "--quiet")
	git(t, upstream, "remote", "add", "origin", "https://github.com/example/upstream.git")
	git(t, upstream, "add", ".")
	git(t, upstream, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "--quiet", "-m", "base")
	git(t, upstream, "update-ref", "refs/remotes/origin/master", "HEAD")
	git(t, upstream, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/master")
	if err := os.WriteFile(filepath.Join(root, ".config", "hidden"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Run(Options{RepoRoot: root, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, decision := range report.Decisions {
		if decision.Path == ".config/hidden" && decision.Action == "include" {
			return
		}
	}
	t.Fatalf("canonical path outside memory-bank was not included: %#v", report.Decisions)
}

func TestCanonicalTemplateRootTakesPrecedenceOverLegacyRoots(t *testing.T) {
	checkout := t.TempDir()
	for _, directory := range []string{"template", "memory-bank-template", "memory-bank"} {
		if err := os.Mkdir(filepath.Join(checkout, directory), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if got, err := selectPayloadRoot(checkout); err != nil || got != "template" {
		t.Fatalf("worktree payload root = %q, %v; want template", got, err)
	}
	run := func(_ string, name string, args ...string) (string, error) {
		if name != "git" || len(args) != 6 || args[0] != "ls-tree" {
			return "", errors.New("unexpected command")
		}
		return args[len(args)-1], nil
	}
	if got, err := selectPayloadRootAt(run, checkout, "origin/main"); err != nil || got != "template" {
		t.Fatalf("Git payload root = %q, %v; want template", got, err)
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

func TestChangedPathsRepresentsDeletionAndRename(t *testing.T) {
	changes, err := changedPaths(func(_ string, _ string, _ ...string) (string, error) {
		return " D memory-bank/dna/removed.md\x00R  memory-bank/dna/new.md\x00memory-bank/dna/old.md\x00", nil
	}, "unused")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"memory-bank/dna/removed.md": true, "memory-bank/dna/old.md": true}
	for _, item := range changes {
		if want[item.path] && item.delete {
			delete(want, item.path)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing deletion changes: %#v", changes)
	}
	foundNew := false
	for _, item := range changes {
		if item.path == "memory-bank/dna/new.md" && !item.delete {
			foundNew = true
		}
	}
	if !foundNew {
		t.Fatalf("missing rename destination: %#v", changes)
	}
}

func TestCopyRegularRejectsSymlinkSource(t *testing.T) {
	root := t.TempDir()
	sourceRoot, destinationRoot := filepath.Join(root, "source"), filepath.Join(root, "destination")
	if err := os.MkdirAll(filepath.Join(sourceRoot, "dna"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destinationRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "outside"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "outside"), filepath.Join(sourceRoot, "dna", "link.md")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	err := copyRegular(filepath.Join(sourceRoot, "dna", "link.md"), filepath.Join(destinationRoot, "dna", "link.md"), sourceRoot, destinationRoot)
	if err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("want symlink rejection, got %v", err)
	}
}
