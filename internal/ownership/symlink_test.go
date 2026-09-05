package ownership

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func symlinkForTest(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
}

func TestInitRejectsDestinationSymlinkAncestor(t *testing.T) {
	repo, source, outside := t.TempDir(), t.TempDir(), t.TempDir()
	path := "memory-bank/dna/rule.md"
	write(t, source, path, "template\n")
	if err := os.MkdirAll(filepath.Join(repo, "memory-bank"), 0o755); err != nil {
		t.Fatal(err)
	}
	symlinkForTest(t, outside, filepath.Join(repo, "memory-bank", "dna"))

	report, err := Init(opts(repo, source, "a"))
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected destination symlink error, got report=%#v err=%v", report, err)
	}
	if _, err := os.Lstat(filepath.Join(outside, "rule.md")); !os.IsNotExist(err) {
		t.Fatalf("init wrote through destination symlink: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(repo, LockFileName)); !os.IsNotExist(err) {
		t.Fatalf("failed init created a lock: %v", err)
	}
}

func TestReadLockRejectsSymlinkLeaf(t *testing.T) {
	source, lockOwner, repo := t.TempDir(), t.TempDir(), t.TempDir()
	write(t, source, "memory-bank/dna/rule.md", "template\n")
	initialize(t, lockOwner, source)
	outsideLock := filepath.Join(lockOwner, LockFileName)
	if err := os.MkdirAll(filepath.Join(repo, "memory-bank"), 0o755); err != nil {
		t.Fatal(err)
	}
	symlinkForTest(t, outsideLock, filepath.Join(repo, LockFileName))

	_, exists, err := ReadLock(repo)
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected lock symlink error, got exists=%v err=%v", exists, err)
	}
	if exists {
		t.Fatal("symlinked lock was reported as an owned repository lock")
	}
}

func TestUpdateRejectsSymlinkAncestorInjectedBeforeMutation(t *testing.T) {
	repo, source, outside := t.TempDir(), t.TempDir(), t.TempDir()
	path := "memory-bank/dna/rule.md"
	write(t, source, path, "one\n")
	initialize(t, repo, source)
	write(t, source, path, "two\n")
	write(t, outside, "rule.md", "outside sentinel\n")
	lockBefore := read(t, repo, LockFileName)

	options := opts(repo, source, "b")
	injected := false
	options.BeforeMutation = func(decision Decision) error {
		if decision.Path != path {
			return nil
		}
		parent := filepath.Join(repo, "memory-bank", "dna")
		if err := os.Rename(parent, parent+".original"); err != nil {
			return err
		}
		if err := os.Symlink(outside, parent); err != nil {
			if restoreErr := os.Rename(parent+".original", parent); restoreErr != nil {
				t.Fatalf("symlink unavailable (%v) and parent restore failed: %v", err, restoreErr)
			}
			t.Skipf("symlinks are unavailable: %v", err)
		}
		injected = true
		return nil
	}

	report, err := Update(options)
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected apply-time symlink error, got report=%#v err=%v", report, err)
	}
	if !injected {
		t.Fatal("test did not inject the destination symlink")
	}
	if got := read(t, outside, "rule.md"); got != "outside sentinel\n" {
		t.Fatalf("update wrote outside the repository: %q", got)
	}
	if got := read(t, repo, LockFileName); got != lockBefore {
		t.Fatal("failed update changed the ownership lock")
	}
}

func TestInitRejectsRepoRootReboundAfterPlanning(t *testing.T) {
	repo, source, outside := t.TempDir(), t.TempDir(), t.TempDir()
	write(t, source, "memory-bank/dna/rule.md", "template\n")
	movedRepo := repo + ".original"
	rebound := false
	defer func() {
		if !rebound {
			return
		}
		_ = os.Remove(repo)
		_ = os.Rename(movedRepo, repo)
	}()

	options := opts(repo, source, "a")
	options.Now = func() time.Time {
		if err := os.Rename(repo, movedRepo); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, repo); err != nil {
			if restoreErr := os.Rename(movedRepo, repo); restoreErr != nil {
				t.Fatalf("symlink unavailable (%v) and repo restore failed: %v", err, restoreErr)
			}
			t.Skipf("symlinks are unavailable: %v", err)
		}
		rebound = true
		return fixedTime
	}
	report, err := Init(options)
	if err == nil || !strings.Contains(err.Error(), "repo root") {
		t.Fatalf("expected rebound repo-root error, got report=%#v err=%v", report, err)
	}
	if _, err := os.Lstat(filepath.Join(outside, "memory-bank")); !os.IsNotExist(err) {
		t.Fatalf("init wrote through rebound repo root: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(outside, LockFileName)); !os.IsNotExist(err) {
		t.Fatalf("init wrote lock through rebound repo root: %v", err)
	}
}

func TestInitRejectsRepoRootReplacedByAnotherDirectory(t *testing.T) {
	repo, source, replacement := t.TempDir(), t.TempDir(), t.TempDir()
	write(t, source, "memory-bank/dna/rule.md", "template\n")
	movedRepo := repo + ".original"
	rebound := false
	defer func() {
		if !rebound {
			return
		}
		_ = os.Rename(repo, replacement)
		_ = os.Rename(movedRepo, repo)
	}()

	options := opts(repo, source, "a")
	options.Now = func() time.Time {
		if err := os.Rename(repo, movedRepo); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(replacement, repo); err != nil {
			if restoreErr := os.Rename(movedRepo, repo); restoreErr != nil {
				t.Fatalf("repo rebind failed (%v) and repo restore failed: %v", err, restoreErr)
			}
			t.Fatal(err)
		}
		rebound = true
		return fixedTime
	}
	report, err := Init(options)
	if err == nil || !strings.Contains(err.Error(), "changed during update") {
		t.Fatalf("expected pinned-root identity error, got report=%#v err=%v", report, err)
	}
	if _, err := os.Lstat(filepath.Join(repo, "memory-bank")); !os.IsNotExist(err) {
		t.Fatalf("init wrote into replacement repo root: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(repo, LockFileName)); !os.IsNotExist(err) {
		t.Fatalf("init wrote lock into replacement repo root: %v", err)
	}
}

// A template source repository may project its payload instead of copying it:
// the destination is a symlink onto the very file this path installs from. Its
// content equals the payload by construction, so pull has nothing to write and
// must not abort the run the way it does for a link leaving the repository.
func TestPullPreservesPayloadProjection(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	write(t, source, "memory-bank/dna/rule.md", "payload\n")
	initialize(t, repo, source)

	// Re-shape the installed copy into a projection of the repository's own
	// payload, mirroring a dual-role repository that owns template/.
	write(t, repo, "template/memory-bank/dna/rule.md", "payload\n")
	installed := filepath.Join(repo, "memory-bank", "dna", "rule.md")
	if err := os.Remove(installed); err != nil {
		t.Fatal(err)
	}
	symlinkForTest(t, filepath.Join("..", "..", "template", "memory-bank", "dna", "rule.md"), installed)

	report, err := Update(opts(repo, source, "b"))
	if err != nil {
		t.Fatalf("update rejected a payload projection: %v", err)
	}
	decision := decisionFor(t, report, "memory-bank/dna/rule.md")
	if decision.Action != Preserve || !strings.Contains(decision.Reason, "projection") {
		t.Fatalf("expected the projection to be preserved, got %#v", decision)
	}
	if info, statErr := os.Lstat(installed); statErr != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("pull replaced the projection with a regular file: info=%v err=%v", info, statErr)
	}
}

// A projection equals the payload it resolves to, not the payload being
// installed. When the repository's own template/ is older than the incoming
// source, treating the destination as current would record a digest the file
// does not have — a lock that lies about installed content.
func TestPullRejectsStalePayloadProjection(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	write(t, source, "memory-bank/dna/rule.md", "payload v1\n")
	initialize(t, repo, source)

	write(t, repo, "template/memory-bank/dna/rule.md", "payload v1\n")
	installed := filepath.Join(repo, "memory-bank", "dna", "rule.md")
	if err := os.Remove(installed); err != nil {
		t.Fatal(err)
	}
	symlinkForTest(t, filepath.Join("..", "..", "template", "memory-bank", "dna", "rule.md"), installed)

	// The source moves ahead while the repository's own payload stays behind.
	write(t, source, "memory-bank/dna/rule.md", "payload v2\n")

	report, err := Update(opts(repo, source, "b"))
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	decision := decisionFor(t, report, "memory-bank/dna/rule.md")
	if decision.Action != Conflict || !strings.Contains(decision.Reason, "stale") {
		t.Fatalf("expected a stale-projection conflict, got %#v", decision)
	}
	if report.Applied {
		t.Fatal("a conflicting run must not be applied")
	}

	lock, _, err := ReadLock(repo)
	if err != nil {
		t.Fatal(err)
	}
	onDisk, err := os.ReadFile(installed)
	if err != nil {
		t.Fatal(err)
	}
	if recorded := lock.Files["memory-bank/dna/rule.md"].PayloadDigest; recorded != digest(onDisk) {
		t.Fatalf("lock records %s but the projection reads %s", recorded, digest(onDisk))
	}
}
