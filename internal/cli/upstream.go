package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const defaultTemplateRemote = "https://github.com/dapi/memory-bank.git"

type resolvedUpstream struct {
	sourceRoot string
	version    string
	ref        string
	cleanup    func()
}

func resolveUpdateUpstream(repoRoot string) (resolvedUpstream, error) {
	return resolveUpdateUpstreamFrom(repoRoot, defaultTemplateRemote)
}

// resolveUpdateUpstreamFrom creates a disposable, detached checkout of main.
// It deliberately never fetches or checks out in memory-bank/.repo.
func resolveUpdateUpstreamFrom(repoRoot, fallbackRemote string) (resolvedUpstream, error) {
	remote, err := upstreamRemote(repoRoot, fallbackRemote)
	if err != nil {
		return resolvedUpstream{}, err
	}
	checkout, err := os.MkdirTemp("", "memory-bank-cli-update-")
	if err != nil {
		return resolvedUpstream{}, fmt.Errorf("create temporary upstream checkout: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(checkout) }
	fail := func(action string, err error) (resolvedUpstream, error) {
		cleanup()
		return resolvedUpstream{}, fmt.Errorf("%s %q: %w; check network access and that the upstream repository exposes branch main", action, remote, err)
	}
	if err := runUpdateGit(checkout, "init", "--quiet"); err != nil {
		return fail("initialize temporary checkout for", err)
	}
	if err := runUpdateGit(checkout, "remote", "add", "origin", remote); err != nil {
		return fail("configure upstream", err)
	}
	if err := runUpdateGit(checkout, "fetch", "--no-tags", "origin", "refs/heads/main"); err != nil {
		return fail("fetch main from", err)
	}
	if err := runUpdateGit(checkout, "checkout", "--quiet", "--detach", "FETCH_HEAD"); err != nil {
		return fail("checkout fetched main from", err)
	}
	ref, err := updateGitOutput(checkout, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return fail("resolve fetched main from", err)
	}
	return resolvedUpstream{sourceRoot: checkout, version: "main@" + ref[:12], ref: ref, cleanup: cleanup}, nil
}

func upstreamRemote(repoRoot, fallbackRemote string) (string, error) {
	checkout := filepath.Join(repoRoot, "memory-bank", ".repo")
	info, err := os.Lstat(checkout)
	if errors.Is(err, fs.ErrNotExist) {
		return fallbackRemote, nil
	}
	if err != nil {
		return "", fmt.Errorf("inspect memory-bank/.repo: %w; repair it or remove it to use the default upstream", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", errors.New("memory-bank/.repo must be a real, clean Git checkout; repair or remove it to use the default upstream")
	}
	top, err := updateGitOutput(checkout, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("memory-bank/.repo is not a Git checkout: %w; repair it or remove it to use the default upstream", err)
	}
	absCheckout, err := filepath.EvalSymlinks(checkout)
	if err != nil {
		return "", fmt.Errorf("resolve memory-bank/.repo: %w", err)
	}
	absTop, err := filepath.EvalSymlinks(top)
	if err != nil || filepath.Clean(absCheckout) != filepath.Clean(absTop) {
		return "", errors.New("memory-bank/.repo must be its own Git worktree; repair it or remove it to use the default upstream")
	}
	status, err := updateGitOutput(checkout, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return "", fmt.Errorf("inspect memory-bank/.repo status: %w", err)
	}
	if status != "" {
		return "", errors.New("memory-bank/.repo is dirty; commit, stash, or discard its changes before update")
	}
	remote, err := updateGitOutput(checkout, "remote", "get-url", "origin")
	if err != nil {
		return "", fmt.Errorf("memory-bank/.repo has no usable origin remote: %w; set origin or remove .repo to use the default upstream", err)
	}
	if remote == "" {
		return "", errors.New("memory-bank/.repo has no usable origin remote; set origin or remove .repo to use the default upstream")
	}
	return remote, nil
}

func runUpdateGit(dir string, arguments ...string) error {
	_, err := updateGitBytes(dir, arguments...)
	return err
}

func updateGitOutput(dir string, arguments ...string) (string, error) {
	output, err := updateGitBytes(dir, arguments...)
	return strings.TrimSpace(string(output)), err
}

func updateGitBytes(dir string, arguments ...string) ([]byte, error) {
	command := exec.Command("git", append([]string{"-C", dir}, arguments...)...)
	command.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	output, err := command.CombinedOutput()
	if err == nil {
		return output, nil
	}
	if message := strings.TrimSpace(string(output)); message != "" {
		return nil, errors.New(message)
	}
	return nil, err
}
