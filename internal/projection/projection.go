// Package projection recognises an intentional payload projection: a symlink
// inside a template source repository that points at the repository's own
// payload instead of a copy of it.
//
// A repository that owns template/ is not a downstream installation of itself.
// Duplicating the payload into the installed tree buys nothing there, so such a
// repository may represent generic documents as symlinks into template/. The
// symlink then equals the payload by construction and cannot drift.
//
// This is deliberately narrow. The symlink guards elsewhere in this CLI exist to
// stop a write from escaping the repository root through a redirected path, and
// that protection is unchanged: a link is only a projection when it resolves
// inside the same repository AND lands exactly on the payload file that backs
// the destination path. Everything else stays an unsafe path.
package projection

import (
	"path/filepath"
	"strings"
)

// PayloadRoot is the tracked payload tree of a template source repository.
// A destination path X is backed by PayloadRoot/X.
const PayloadRoot = "template"

// Resolve reports whether the repository-relative destination path resolves,
// through one or more symlinks, to the payload file that backs it, and returns
// that resolved payload path.
//
// The returned path is what the destination actually reads. A caller must not
// assume it matches an incoming source payload: the local payload can be older
// than the source being installed, and only comparing content can tell.
//
// It returns false rather than an error for the ordinary negative cases — a
// regular file, a broken link, a missing payload counterpart — because callers
// use this to decide between two legitimate behaviours, not to detect faults.
func Resolve(repoRoot, relative string) (string, bool) {
	if repoRoot == "" || relative == "" {
		return "", false
	}
	osRelative := filepath.FromSlash(relative)
	if !filepath.IsLocal(osRelative) {
		return "", false
	}

	resolvedRoot, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		return "", false
	}
	destination, err := filepath.EvalSymlinks(filepath.Join(repoRoot, osRelative))
	if err != nil {
		return "", false
	}
	payload, err := filepath.EvalSymlinks(filepath.Join(repoRoot, PayloadRoot, osRelative))
	if err != nil {
		return "", false
	}
	if destination != payload || !within(resolvedRoot, destination) {
		return "", false
	}
	return destination, true
}

// IsPayloadProjection reports whether the destination path is a projection of
// the payload file backing it.
func IsPayloadProjection(repoRoot, relative string) bool {
	_, ok := Resolve(repoRoot, relative)
	return ok
}

// within reports whether target is the root itself or lives below it. Both
// paths must already be resolved, so no symlink can move the target afterwards.
func within(root, target string) bool {
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}
