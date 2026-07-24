// Package push implements safe, opt-in publication of managed Memory Bank changes.
package push

import (
	"crypto/rand"
	"encoding/hex"
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
	RepoRoot   string
	DryRun     bool
	Now        func() time.Time
	BranchName func(time.Time) (string, error)
	Run        func(dir, name string, args ...string) (string, error)
}

type change struct {
	path   string
	delete bool
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
	githubRepo, err := githubRepository(strings.TrimSpace(remote))
	if err != nil {
		return Report{}, err
	}
	if !options.DryRun {
		if identity, err := run(checkout, "gh", "repo", "view", githubRepo, "--json", "id"); err != nil || strings.TrimSpace(identity) == "" {
			return Report{}, errors.New("upstream origin is not an accessible GitHub repository for PR creation")
		}
	}
	defaultBranch, err := defaultBranch(run, checkout)
	if err != nil {
		return Report{}, err
	}
	if !options.DryRun {
		if _, err := run(checkout, "git", "fetch", "origin", defaultBranch+":refs/remotes/origin/"+defaultBranch); err != nil {
			return Report{}, fmt.Errorf("refresh upstream default branch: %w", err)
		}
	}
	payloadRoot, err := selectPayloadRootAt(run, checkout, "origin/"+defaultBranch)
	if err != nil {
		return Report{}, err
	}
	originalBranch, err := run(checkout, "git", "branch", "--show-current")
	if err != nil || strings.TrimSpace(originalBranch) == "" {
		return Report{}, errors.New("upstream checkout must be on a named branch")
	}
	originalHead, err := run(checkout, "git", "rev-parse", "HEAD")
	if err != nil {
		return Report{}, fmt.Errorf("resolve upstream HEAD: %w", err)
	}
	changes, err := changedPaths(run, options.RepoRoot)
	if err != nil {
		return Report{}, err
	}
	report := Report{FormatVersion: 1, DryRun: options.DryRun}
	for _, item := range changes {
		class := ownership.Classify(item.path)
		if class != ownership.Managed {
			report.Decisions = append(report.Decisions, Decision{Path: item.path, Action: "exclude", Reason: string(class) + " paths are not published"})
			continue
		}
		action := "include"
		if item.delete {
			action = "delete"
		}
		report.Decisions = append(report.Decisions, Decision{Path: item.path, Action: action, Reason: "managed template path"})
	}
	sort.Slice(report.Decisions, func(i, j int) bool { return report.Decisions[i].Path < report.Decisions[j].Path })
	if options.DryRun {
		return report, nil
	}
	included := make([]change, 0)
	for _, item := range changes {
		if ownership.Classify(item.path) == ownership.Managed {
			included = append(included, item)
		}
	}
	if len(included) == 0 {
		return report, errors.New("no managed Memory Bank changes to publish")
	}
	makeBranch := options.BranchName
	if makeBranch == nil {
		makeBranch = branchName
	}
	branch, err := makeBranch(now())
	if err != nil {
		return report, err
	}
	if _, err := run(checkout, "git", "ls-remote", "--exit-code", "--heads", "origin", branch); err == nil {
		return report, fmt.Errorf("upstream branch %q already exists; retry the command", branch)
	}
	branchCreated, remoteCreated := false, false
	stagePaths := make([]string, 0, len(included))
	failed := func(cause error) (Report, error) {
		var cleanup []string
		if remoteCreated {
			if _, err := run(checkout, "git", "push", "origin", "--delete", branch); err != nil {
				cleanup = append(cleanup, "remote branch remains: "+err.Error())
			}
		}
		if _, err := run(checkout, "git", "reset", "--hard"); err != nil {
			cleanup = append(cleanup, "reset failed: "+err.Error())
		}
		if _, err := run(checkout, "git", "checkout", strings.TrimSpace(originalBranch)); err != nil {
			cleanup = append(cleanup, "restore branch failed: "+err.Error())
		}
		if branchCreated {
			if _, err := run(checkout, "git", "branch", "-D", branch); err != nil {
				cleanup = append(cleanup, "remove local branch failed: "+err.Error())
			}
		}
		if _, err := run(checkout, "git", "reset", "--hard", strings.TrimSpace(originalHead)); err != nil {
			cleanup = append(cleanup, "restore HEAD failed: "+err.Error())
		}
		if len(stagePaths) > 0 {
			args := append([]string{"clean", "-fd", "--"}, stagePaths...)
			if _, err := run(checkout, "git", args...); err != nil {
				cleanup = append(cleanup, "remove untracked staged paths failed: "+err.Error())
			}
		}
		if status, err := run(checkout, "git", "status", "--porcelain"); err != nil || strings.TrimSpace(status) != "" {
			cleanup = append(cleanup, "checkout restoration is not clean")
		}
		if len(cleanup) > 0 {
			return report, fmt.Errorf("%w; cleanup: %s", cause, strings.Join(cleanup, "; "))
		}
		return report, cause
	}
	if _, err := run(checkout, "git", "checkout", "-b", branch, "origin/"+defaultBranch); err != nil {
		return failed(fmt.Errorf("create upstream branch: %w", err))
	}
	branchCreated = true
	report.Branch = branch
	actualPayloadRoot, err := selectPayloadRoot(checkout)
	if err != nil {
		return failed(err)
	}
	if actualPayloadRoot != payloadRoot {
		return failed(fmt.Errorf("upstream payload root changed from %q to %q while switching to default branch", payloadRoot, actualPayloadRoot))
	}
	for _, item := range included {
		relative := strings.TrimPrefix(item.path, "memory-bank/")
		destinationRoot := payloadDestinationRoot(checkout, payloadRoot)
		destination := filepath.Join(destinationRoot, filepath.FromSlash(relative))
		stagePaths = append(stagePaths, filepath.ToSlash(filepath.Join(payloadRootForPath(payloadRoot), relative)))
		if item.delete {
			if err := removeRegular(destination, destinationRoot); err != nil && !errors.Is(err, os.ErrNotExist) {
				return failed(fmt.Errorf("delete %s: %w", item.path, err))
			}
		} else if err := copyRegular(filepath.Join(options.RepoRoot, filepath.FromSlash(item.path)), destination, filepath.Join(options.RepoRoot, "memory-bank"), destinationRoot); err != nil {
			return failed(fmt.Errorf("stage %s: %w", item.path, err))
		}
	}
	addArgs := append([]string{"add", "--"}, stagePaths...)
	if _, err := run(checkout, "git", addArgs...); err != nil {
		return failed(fmt.Errorf("stage upstream changes: %w", err))
	}
	if _, err := run(checkout, "git", "commit", "-m", "Publish managed Memory Bank changes"); err != nil {
		return failed(fmt.Errorf("commit upstream changes: %w", err))
	}
	if _, err := run(checkout, "git", "push", "-u", "origin", branch); err != nil {
		return failed(fmt.Errorf("push upstream branch: %w; remote branch ownership is unproven and will not be deleted", err))
	}
	remoteCreated = true
	pr, err := run(checkout, "gh", "pr", "create", "--repo", githubRepo, "--head", branch, "--base", defaultBranch, "--fill")
	if err != nil {
		if existing, queryErr := run(checkout, "gh", "pr", "list", "--repo", githubRepo, "--head", branch, "--json", "url", "--jq", ".[0].url"); queryErr == nil && strings.TrimSpace(existing) != "" {
			remoteCreated = false
			return failed(fmt.Errorf("create PR returned an error, but PR %s may have been created; remote branch was retained", strings.TrimSpace(existing)))
		}
		return failed(fmt.Errorf("create PR: %w", err))
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

func githubRepository(remote string) (string, error) {
	value := strings.TrimSuffix(remote, ".git")
	if strings.HasPrefix(value, "git@github.com:") {
		value = strings.TrimPrefix(value, "git@github.com:")
	} else if strings.HasPrefix(value, "https://github.com/") {
		value = strings.TrimPrefix(value, "https://github.com/")
	} else {
		return "", fmt.Errorf("upstream origin must be a GitHub repository: %q", remote)
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid GitHub origin: %q", remote)
	}
	return value, nil
}

func clean(run func(string, string, ...string) (string, error), dir string) error {
	if _, err := run(dir, "git", "rev-parse", "--is-inside-work-tree"); err != nil {
		return fmt.Errorf("memory-bank/.repo is not a Git checkout: %w", err)
	}
	top, err := run(dir, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("resolve upstream Git root: %w", err)
	}
	absDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return err
	}
	absTop, err := filepath.EvalSymlinks(strings.TrimSpace(top))
	if err != nil {
		return err
	}
	if filepath.Clean(absDir) != filepath.Clean(absTop) {
		return fmt.Errorf("memory-bank/.repo must be its own Git worktree (got %q, want %q)", absTop, absDir)
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

func defaultBranch(run func(string, string, ...string) (string, error), checkout string) (string, error) {
	ref, err := run(checkout, "git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", fmt.Errorf("resolve upstream default branch: %w", err)
	}
	branch := strings.TrimPrefix(strings.TrimSpace(ref), "origin/")
	if branch == "" || branch == ref {
		return "", errors.New("upstream origin/HEAD does not name a default branch")
	}
	if _, err := run(checkout, "git", "rev-parse", "--verify", "origin/"+branch); err != nil {
		return "", fmt.Errorf("default branch %q is not available locally: %w", branch, err)
	}
	return branch, nil
}

func selectPayloadRoot(checkout string) (string, error) {
	var roots []string
	for _, candidate := range []string{"template", "memory-bank-template", "memory-bank"} {
		info, err := os.Lstat(filepath.Join(checkout, candidate))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("upstream payload root %q must be a real directory", candidate)
		}
		roots = append(roots, candidate)
	}
	if len(roots) != 1 {
		return "", fmt.Errorf("upstream checkout must contain exactly one payload root (template, memory-bank-template, or memory-bank), found %v", roots)
	}
	return roots[0], nil
}

func selectPayloadRootAt(run func(string, string, ...string) (string, error), checkout, ref string) (string, error) {
	var roots []string
	for _, candidate := range []string{"template", "memory-bank-template", "memory-bank"} {
		out, err := run(checkout, "git", "ls-tree", "-d", "--name-only", ref, "--", candidate)
		if err != nil {
			return "", fmt.Errorf("inspect upstream payload root %q: %w", candidate, err)
		}
		if strings.TrimSpace(out) == candidate {
			roots = append(roots, candidate)
		}
	}
	if len(roots) != 1 {
		return "", fmt.Errorf("default branch must contain exactly one payload root (template, memory-bank-template, or memory-bank), found %v", roots)
	}
	return roots[0], nil
}

// payloadDestinationRoot is the inverse of the canonical source projection
// for downstream memory-bank paths. Legacy roots retain their old layout.
func payloadDestinationRoot(checkout, payloadRoot string) string {
	if payloadRoot == "template" {
		return filepath.Join(checkout, payloadRoot, "memory-bank")
	}
	return filepath.Join(checkout, payloadRoot)
}

func payloadRootForPath(payloadRoot string) string {
	if payloadRoot == "template" {
		return filepath.Join(payloadRoot, "memory-bank")
	}
	return payloadRoot
}

func changedPaths(run func(string, string, ...string) (string, error), root string) ([]change, error) {
	out, err := run(root, "git", "status", "--porcelain=v1", "-z", "--untracked-files=all", "--", "memory-bank")
	if err != nil {
		return nil, fmt.Errorf("inspect downstream changes: %w", err)
	}
	set := map[string]change{}
	records := strings.Split(out, "\x00")
	for index := 0; index < len(records); index++ {
		line := records[index]
		if len(line) < 4 {
			continue
		}
		status, paths := line[:2], line[3:]
		if strings.Contains(status, "U") {
			return nil, fmt.Errorf("downstream path has unresolved Git conflict: %q", paths)
		}
		if strings.Contains(status, "R") {
			if index+1 >= len(records) || records[index+1] == "" {
				return nil, fmt.Errorf("ambiguous rename %q", line)
			}
			old, next := records[index+1], paths
			index++
			for _, item := range []change{{path: old, delete: true}, {path: next}} {
				if err := addChange(set, item); err != nil {
					return nil, err
				}
			}
			continue
		}
		if err := addChange(set, change{path: paths, delete: strings.Contains(status, "D")}); err != nil {
			return nil, err
		}
	}
	paths := make([]change, 0, len(set))
	for _, item := range set {
		paths = append(paths, item)
	}
	sort.Slice(paths, func(i, j int) bool {
		if paths[i].path == paths[j].path {
			return paths[i].delete
		}
		return paths[i].path < paths[j].path
	})
	return paths, nil
}

func branchName(now time.Time) (string, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate unique upstream branch name: %w", err)
	}
	return fmt.Sprintf("memory-bank-cli/push-%s-%s", now.UTC().Format("20060102-150405.000000000"), hex.EncodeToString(bytes)), nil
}

func addChange(set map[string]change, item change) error {
	item.path = filepath.ToSlash(item.path)
	if !strings.HasPrefix(item.path, "memory-bank/") || strings.Contains(item.path, "../") {
		return fmt.Errorf("ambiguous changed path %q", item.path)
	}
	set[item.path] = item
	return nil
}

func copyRegular(source, destination, sourceRoot, destinationRoot string) error {
	if err := validateRegular(source, sourceRoot); err != nil {
		return err
	}
	if err := ensureSafeParents(filepath.Dir(destination), destinationRoot); err != nil {
		return err
	}
	if info, err := os.Lstat(destination); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("destination is a symlink: %s", destination)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, data, 0o644)
}

func removeRegular(path, root string) error {
	if err := ensureSafeParents(filepath.Dir(path), root); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("destination is not a regular file: %s", path)
	}
	return os.Remove(path)
}

func validateRegular(path, root string) error {
	if err := safeExistingParents(filepath.Dir(path), root); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("source is not a regular file: %s", path)
	}
	return nil
}

func safeExistingParents(directory, root string) error {
	relative, err := filepath.Rel(root, directory)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("path escapes payload root")
	}
	current := root
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		if part == "." || part == "" {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("unsafe path component: %s", current)
		}
	}
	return nil
}

func ensureSafeParents(directory, root string) error {
	relative, err := filepath.Rel(root, directory)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("path escapes payload root")
	}
	current := root
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		if part == "." || part == "" {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(current, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("unsafe path component: %s", current)
		}
	}
	return nil
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
	return string(out), nil
}
