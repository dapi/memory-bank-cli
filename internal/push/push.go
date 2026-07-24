// Package push implements safe, opt-in publication of managed Memory Bank changes.
package push

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dapi/memory-bank-cli/internal/ownership"
)

type Decision struct {
	Path   string `json:"path"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

type Report struct {
	FormatVersion int        `json:"format_version"`
	DryRun        bool       `json:"dry_run"`
	Branch        string     `json:"branch,omitempty"`
	PRURL         string     `json:"pr_url,omitempty"`
	Decisions     []Decision `json:"decisions"`
}

type Options struct {
	RepoRoot string
	DryRun   bool
	Now      func() time.Time
	Run      func(dir, name string, args ...string) (string, error)
}

func Run(options Options) (Report, error) {
	if options.RepoRoot == "" {
		return Report{}, errors.New("repository root is required")
	}
	run := options.Run
	if run == nil {
		run = command
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	if _, err := run(options.RepoRoot, "git", "rev-parse", "--show-toplevel"); err != nil {
		return Report{}, fmt.Errorf("push must run inside a Git repository: %w", err)
	}
	checkout, err := safeCheckout(options.RepoRoot)
	if err != nil {
		return Report{}, err
	}
	if err := clean(run, checkout); err != nil {
		return Report{}, err
	}
	remote, err := run(checkout, "git", "remote", "get-url", "origin")
	if err != nil || strings.TrimSpace(remote) == "" {
		return Report{}, fmt.Errorf("upstream checkout has no valid origin remote")
	}
	paths, err := changedPaths(run, options.RepoRoot)
	if err != nil {
		return Report{}, err
	}
	report := Report{FormatVersion: 1, DryRun: options.DryRun}
	for _, path := range paths {
		class := ownership.Classify(path)
		if class != ownership.Managed {
			report.Decisions = append(report.Decisions, Decision{Path: path, Action: "exclude", Reason: string(class) + " paths are not published"})
			continue
		}
		report.Decisions = append(report.Decisions, Decision{Path: path, Action: "include", Reason: "managed template path"})
	}
	sort.Slice(report.Decisions, func(i, j int) bool { return report.Decisions[i].Path < report.Decisions[j].Path })
	if options.DryRun {
		return report, nil
	}
	included := make([]string, 0)
	for _, decision := range report.Decisions {
		if decision.Action == "include" {
			included = append(included, decision.Path)
		}
	}
	if len(included) == 0 {
		return report, errors.New("no managed Memory Bank changes to publish")
	}
	branch := fmt.Sprintf("memory-bank-cli/push-%s", now().UTC().Format("20060102-150405"))
	if _, err := run(checkout, "git", "checkout", "-b", branch); err != nil {
		return report, fmt.Errorf("create upstream branch: %w", err)
	}
	report.Branch = branch
	failed := func(cause error) (Report, error) {
		_, _ = run(checkout, "git", "checkout", "-")
		return report, cause
	}
	for _, path := range included {
		if err := copyFile(filepath.Join(options.RepoRoot, filepath.FromSlash(path)), filepath.Join(checkout, filepath.FromSlash(path))); err != nil {
			return failed(fmt.Errorf("stage %s: %w", path, err))
		}
	}
	addArgs := append([]string{"add", "--"}, included...)
	if _, err := run(checkout, "git", addArgs...); err != nil {
		return failed(fmt.Errorf("stage upstream changes: %w", err))
	}
	if _, err := run(checkout, "git", "commit", "-m", "Publish managed Memory Bank changes"); err != nil {
		return failed(fmt.Errorf("commit upstream changes: %w", err))
	}
	if _, err := run(checkout, "git", "push", "-u", "origin", branch); err != nil {
		return failed(fmt.Errorf("push upstream branch: %w", err))
	}
	pr, err := run(checkout, "gh", "pr", "create", "--head", branch, "--fill")
	if err != nil {
		if _, deleteErr := run(checkout, "git", "push", "origin", "--delete", branch); deleteErr != nil {
			return failed(fmt.Errorf("create PR: %w; remote branch %q remains and must be deleted manually", err, branch))
		}
		return failed(fmt.Errorf("create PR: %w; remote branch was deleted", err))
	}
	report.PRURL = strings.TrimSpace(pr)
	if report.PRURL == "" {
		return failed(errors.New("GitHub did not return a PR URL"))
	}
	return report, nil
}

func safeCheckout(root string) (string, error) {
	path := filepath.Join(root, "memory-bank", ".repo")
	info, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("inspect upstream checkout: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", errors.New("memory-bank/.repo must be a real directory, not a symlink")
	}
	return path, nil
}

func clean(run func(string, string, ...string) (string, error), dir string) error {
	if _, err := run(dir, "git", "rev-parse", "--is-inside-work-tree"); err != nil {
		return fmt.Errorf("memory-bank/.repo is not a Git checkout: %w", err)
	}
	status, err := run(dir, "git", "status", "--porcelain")
	if err != nil {
		return err
	}
	if strings.TrimSpace(status) != "" {
		return errors.New("upstream checkout is dirty; commit, stash, or discard its changes first")
	}
	conflicts, err := run(dir, "git", "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return err
	}
	if strings.TrimSpace(conflicts) != "" {
		return errors.New("upstream checkout has unresolved conflicts")
	}
	return nil
}

func changedPaths(run func(string, string, ...string) (string, error), root string) ([]string, error) {
	out, err := run(root, "git", "status", "--porcelain", "--untracked-files=all", "--", "memory-bank")
	if err != nil {
		return nil, fmt.Errorf("inspect downstream changes: %w", err)
	}
	set := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSuffix(out, "\n"), "\n") {
		if len(line) < 3 {
			continue
		}
		pathStart := 2
		if line[2] == ' ' {
			pathStart = 3
		}
		path := strings.TrimSpace(line[pathStart:])
		if before, _, ok := strings.Cut(path, " -> "); ok {
			path = strings.TrimSpace(before)
		}
		path = filepath.ToSlash(path)
		if !strings.HasPrefix(path, "memory-bank/") || strings.Contains(path, "../") {
			return nil, fmt.Errorf("ambiguous changed path %q", path)
		}
		set[path] = true
	}
	paths := make([]string, 0, len(set))
	for path := range set {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

func copyFile(source, destination string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destination, data, 0o644)
}

func command(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		if text != "" {
			return "", fmt.Errorf("%s: %w", text, err)
		}
		return "", err
	}
	return text, nil
}
