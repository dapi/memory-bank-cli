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
	"os"
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

// Covers reports whether the destination path is a projection or lies below a
// projected directory, even when the file itself does not exist yet.
//
// A caller needs this to tell "nothing to install here, the payload backs it"
// from "unsafe path". When upstream adds a file below a projected directory,
// the destination is absent until the local payload catches up, yet the path is
// still governed by the projection and must not be written through.
func Covers(repoRoot, relative string) bool {
	if repoRoot == "" || relative == "" {
		return false
	}
	osRelative := filepath.FromSlash(relative)
	if !filepath.IsLocal(osRelative) {
		return false
	}
	if IsPayloadProjection(repoRoot, relative) || Declares(repoRoot, relative) {
		return true
	}
	// Walk towards the root. A plain directory ancestor says nothing: it may
	// itself be reached through a projected symlink higher up, so keep looking
	// instead of concluding the path is unprojected.
	prefix := osRelative
	for {
		parent := filepath.Dir(prefix)
		if parent == "." || parent == string(filepath.Separator) || parent == prefix {
			return false
		}
		prefix = parent
		info, err := os.Lstat(filepath.Join(repoRoot, prefix))
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return IsPayloadProjection(repoRoot, filepath.ToSlash(prefix))
		}
	}
}

// Declares reports whether the destination is a symlink written to point at the
// payload backing it, even when that payload no longer exists.
//
// Removing a file from template/ leaves its projection dangling. The path is
// still governed by the projection and must reach a decision rather than a
// hard failure, so a caller needs to recognise it without resolving it. This
// never enables a read or a write: a dangling link has nothing to read.
func Declares(repoRoot, relative string) bool {
	if repoRoot == "" || relative == "" {
		return false
	}
	osRelative := filepath.FromSlash(relative)
	if !filepath.IsLocal(osRelative) {
		return false
	}
	link := filepath.Join(repoRoot, osRelative)
	info, err := os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return false
	}
	target, err := os.Readlink(link)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(link), target)
	}
	return filepath.Clean(target) == filepath.Join(repoRoot, PayloadRoot, osRelative)
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
