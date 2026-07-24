package ownership

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

var fixedTime = time.Date(2026, 7, 21, 10, 0, 0, 0, time.UTC)

func write(t *testing.T, root, path, contents string) {
	t.Helper()
	target := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root, path string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func opts(repo, source, version string) Options {
	return Options{
		RepoRoot: repo, SourceRoot: source, TemplateVersion: version,
		SourceRef: strings.Repeat(version, 40/len(version)+1)[:40], Now: func() time.Time { return fixedTime },
		verifySource: func(string, string) error { return nil },
	}
}

func initialize(t *testing.T, repo, source string) {
	t.Helper()
	report, err := Init(opts(repo, source, "a"))
	if err != nil || !report.Applied || report.ConflictCount != 0 {
		t.Fatalf("init failed: report=%#v err=%v", report, err)
	}
}

func decisionFor(t *testing.T, report Report, path string) Decision {
	t.Helper()
	for _, decision := range report.Decisions {
		if decision.Path == path {
			return decision
		}
	}
	t.Fatalf("no decision for %s: %#v", path, report.Decisions)
	return Decision{}
}

func TestAgentFileRejectsCaseAliasesOfMemoryBank(t *testing.T) {
	repo := t.TempDir()
	for _, target := range []string{"memory-bank", "Memory-Bank/README.md", "MEMORY-BANK/dna/README.md"} {
		t.Run(target, func(t *testing.T) {
			if _, err := Doctor(repo, target); err == nil || !strings.Contains(err.Error(), "outside memory-bank") {
				t.Fatalf("target %q was not rejected: %v", target, err)
			}
		})
	}
}

func TestReadLockRejectsAbsoluteManagedPath(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	write(t, source, "memory-bank/dna/rule.md", "managed\n")
	initialize(t, repo, source)
	lockPath := filepath.Join(repo, filepath.FromSlash(LockFileName))
	lock := read(t, repo, LockFileName)
	lock = strings.Replace(lock, "memory-bank/dna/rule.md", "/outside.md", 1)
	if err := os.WriteFile(lockPath, []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadLock(repo); err == nil || !strings.Contains(err.Error(), "invalid path") {
		t.Fatalf("absolute lock path was accepted: %v", err)
	}
}

func TestReadLockRejectsGitMetadataPath(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	write(t, source, "memory-bank/dna/rule.md", "managed\n")
	initialize(t, repo, source)
	lockPath := filepath.Join(repo, filepath.FromSlash(LockFileName))
	lock := read(t, repo, LockFileName)
	lock = strings.Replace(lock, "memory-bank/dna/rule.md", ".GiT/config", 1)
	if err := os.WriteFile(lockPath, []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadLock(repo); err == nil || !strings.Contains(err.Error(), "invalid path") {
		t.Fatalf("Git metadata lock path was accepted: %v", err)
	}
}

func TestAgentPlanPreservesExplicitZeroMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose Unix permission bits")
	}
	repoRoot := t.TempDir()
	target := "AGENTS.md"
	write(t, repoRoot, target, "project rules\n")
	targetPath := filepath.Join(repoRoot, target)
	if err := os.Chmod(targetPath, 0); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	repo, err := pinRepoRoot(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	plan, _, err := buildAgentPlanWithReader(repo, target, func(pinnedRepo, string) (os.FileInfo, []byte, error) {
		return info, []byte("project rules\n"), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan == nil || !plan.modeSet || plan.mode.Perm() != 0 {
		t.Fatalf("zero mode was not captured explicitly: %#v", plan)
	}
}

func TestCleanUpdateAndRepeatedUpdateAreIdempotent(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	path := "memory-bank/dna/rule.md"
	write(t, source, path, "one\n")
	initialize(t, repo, source)
	write(t, source, path, "two\n")

	report, err := Update(opts(repo, source, "b"))
	if err != nil || !report.Applied || decisionFor(t, report, path).Action != UpdateFile {
		t.Fatalf("clean update failed: report=%#v err=%v", report, err)
	}
	if got := read(t, repo, path); got != "two\n" {
		t.Fatalf("unexpected payload: %q", got)
	}
	lockBefore := read(t, repo, LockFileName)
	report, err = Update(opts(repo, source, "b"))
	if err != nil || report.Applied || decisionFor(t, report, path).Action != Preserve {
		t.Fatalf("idempotent update failed: report=%#v err=%v", report, err)
	}
	if lockAfter := read(t, repo, LockFileName); lockAfter != lockBefore {
		t.Fatal("no-op update rewrote the lock")
	}
}

func TestFilesystemSourceReaderTranslatesTargetRoot(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	write(t, source, "template/memory-bank/dna/rule.md", "template\n")

	initialize(t, repo, source)
	if got := read(t, repo, "memory-bank/dna/rule.md"); got != "template\n" {
		t.Fatalf("target-root filesystem reader did not translate downstream path: %q", got)
	}
	if _, err := os.Stat(filepath.Join(repo, "template")); !os.IsNotExist(err) {
		t.Fatalf("source root leaked into downstream: %v", err)
	}
}

func TestUpdateFromTargetRootPreservesDownstreamPathAndIsIdempotent(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	path := "memory-bank/dna/rule.md"
	write(t, source, "template/memory-bank/dna/rule.md", "one\n")
	initialize(t, repo, source)

	write(t, source, "template/memory-bank/dna/rule.md", "two\n")
	report, err := Update(opts(repo, source, "b"))
	if err != nil || !report.Applied || decisionFor(t, report, path).Action != UpdateFile {
		t.Fatalf("target-root update failed: report=%#v err=%v", report, err)
	}
	if got := read(t, repo, path); got != "two\n" {
		t.Fatalf("target-root update wrote unexpected downstream payload: %q", got)
	}
	if _, err := os.Stat(filepath.Join(repo, "template")); !os.IsNotExist(err) {
		t.Fatalf("source root leaked into downstream after update: %v", err)
	}

	lockBefore := read(t, repo, LockFileName)
	report, err = Update(opts(repo, source, "b"))
	if err != nil || report.Applied || decisionFor(t, report, path).Action != Preserve {
		t.Fatalf("target-root repeated update was not idempotent: report=%#v err=%v", report, err)
	}
	if lockAfter := read(t, repo, LockFileName); lockAfter != lockBefore {
		t.Fatal("target-root no-op update rewrote the lock")
	}
}

func TestLegacyLockMigratesToCanonicalTemplateWithoutRemovingManagedFiles(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	write(t, source, "memory-bank/dna/rule.md", "v1\n")
	initialize(t, repo, source)

	if err := os.RemoveAll(filepath.Join(source, "memory-bank")); err != nil {
		t.Fatal(err)
	}
	write(t, source, "template/memory-bank/dna/rule.md", "v1\n")
	write(t, source, "template/.config/new", "new\n")
	report, err := Update(opts(repo, source, "b"))
	if err != nil || !report.Applied {
		t.Fatalf("canonical migration failed: report=%#v err=%v", report, err)
	}
	if got := read(t, repo, "memory-bank/dna/rule.md"); got != "v1\n" {
		t.Fatalf("legacy managed file changed during migration: %q", got)
	}
	if got := read(t, repo, ".config/new"); got != "new\n" {
		t.Fatalf("canonical addition was not installed: %q", got)
	}
	lock, exists, err := ReadLock(repo)
	if err != nil || !exists || lock.Files["memory-bank/dna/rule.md"].Ownership != Managed || lock.Files[".config/new"].Ownership != Managed {
		t.Fatalf("migration lock is incomplete: %#v, exists=%v, err=%v", lock, exists, err)
	}
}

func TestLegacyLockMigrationPromotesOnlySafeCanonicalFiles(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	adaptedClean := "memory-bank/domain/model.md"
	userClean := "memory-bank/features/FT-001/brief.md"
	userDrift := "memory-bank/features/FT-001/local.md"
	for _, path := range []string{adaptedClean, userClean, userDrift} {
		write(t, source, path, "v1\n")
	}
	initialize(t, repo, source)
	write(t, repo, userDrift, "local customization\n")

	if err := os.RemoveAll(filepath.Join(source, "memory-bank")); err != nil {
		t.Fatal(err)
	}
	write(t, source, "template/"+adaptedClean, "v2\n")
	write(t, source, "template/"+userClean, "v1\n")
	write(t, source, "template/"+userDrift, "v1\n")

	report, err := Update(opts(repo, source, "b"))
	if err != nil || !report.Applied || report.ConflictCount != 0 {
		t.Fatalf("canonical migration failed: report=%#v err=%v", report, err)
	}
	if got := read(t, repo, adaptedClean); got != "v2\n" {
		t.Fatalf("clean adapted file was not updated: %q", got)
	}
	if got := read(t, repo, userDrift); got != "local customization\n" {
		t.Fatalf("legacy user customization was overwritten: %q", got)
	}
	lock, exists, err := ReadLock(repo)
	if err != nil || !exists {
		t.Fatalf("read migrated lock: exists=%v err=%v", exists, err)
	}
	if lock.Files[adaptedClean].Ownership != Managed || lock.Files[userClean].Ownership != Managed {
		t.Fatalf("safe legacy files were not promoted: %#v", lock.Files)
	}
	if lock.Files[userDrift].Ownership != UserOwned {
		t.Fatalf("drifted legacy file lost its ownership protection: %#v", lock.Files[userDrift])
	}
}

func TestInitRejectsReservedLockPathInTemplate(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	write(t, source, "memory-bank/dna/rule.md", "template\n")
	write(t, source, LockFileName, "not runtime metadata\n")

	report, err := Init(opts(repo, source, "a"))
	if err == nil || !strings.Contains(err.Error(), "reserved metadata path") {
		t.Fatalf("expected reserved-path error, got report=%#v err=%v", report, err)
	}
	if _, statErr := os.Lstat(filepath.Join(repo, LockFileName)); !os.IsNotExist(statErr) {
		t.Fatalf("failed init created a lock: %v", statErr)
	}
}

func TestExecutableModeIsInstalledAndUpdated(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose Unix executable permission bits")
	}
	repo, source := t.TempDir(), t.TempDir()
	path := "memory-bank/flows/tool.md"
	write(t, source, path, "tool\n")
	if err := os.Chmod(filepath.Join(source, filepath.FromSlash(path)), 0o755); err != nil {
		t.Fatal(err)
	}
	initialize(t, repo, source)
	assertMode := func(want os.FileMode) {
		t.Helper()
		info, err := os.Stat(filepath.Join(repo, filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != want {
			t.Fatalf("unexpected installed mode: got %04o want %04o", got, want)
		}
	}
	assertMode(0o755)

	if err := os.Chmod(filepath.Join(source, filepath.FromSlash(path)), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := Update(opts(repo, source, "b"))
	if err != nil || !report.Applied || decisionFor(t, report, path).Action != UpdateFile {
		t.Fatalf("mode-only update failed: report=%#v err=%v", report, err)
	}
	assertMode(0o644)
	lock, _, err := ReadLock(repo)
	if err != nil {
		t.Fatal(err)
	}
	if got := lock.Files[path]; got.BaseMode != "100644" || got.PayloadMode != "100644" {
		t.Fatalf("mode contract was not recorded: %#v", got)
	}
}

func TestAdaptedCustomizationIsPreserved(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	path := "memory-bank/domain/model.md"
	write(t, source, path, "template\n")
	initialize(t, repo, source)
	write(t, repo, path, "project model\n")

	report, err := Update(opts(repo, source, "a"))
	if err != nil || report.Applied || decisionFor(t, report, path).Action != Preserve {
		t.Fatalf("customization was not preserved: report=%#v err=%v", report, err)
	}
	if got := read(t, repo, path); got != "project model\n" {
		t.Fatalf("adapted file was overwritten: %q", got)
	}
}

func TestAdaptedUpstreamAndDownstreamChangesConflictWithoutPartialApply(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	adapted := "memory-bank/domain/model.md"
	managed := "memory-bank/dna/rule.md"
	write(t, source, adapted, "base\n")
	write(t, source, managed, "managed v1\n")
	initialize(t, repo, source)
	write(t, repo, adapted, "downstream\n")
	write(t, source, adapted, "upstream\n")
	write(t, source, managed, "managed v2\n")
	lockBefore := read(t, repo, LockFileName)

	report, err := Update(opts(repo, source, "b"))
	if err != nil || report.Applied || report.ConflictCount != 1 || decisionFor(t, report, adapted).Action != Conflict {
		t.Fatalf("expected conflict: report=%#v err=%v", report, err)
	}
	if got := read(t, repo, managed); got != "managed v1\n" {
		t.Fatalf("managed file was partially updated: %q", got)
	}
	if got := read(t, repo, LockFileName); got != lockBefore {
		t.Fatal("conflicting update changed lock")
	}
}

func TestManagedContentDriftIsPreservedWhenTemplateIsUnchanged(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	path := "memory-bank/flows/feature.md"
	write(t, source, path, "base\n")
	initialize(t, repo, source)
	write(t, repo, path, "local drift\n")

	for index := 0; index < 2; index++ {
		report, err := Update(opts(repo, source, "a"))
		if err != nil || report.ConflictCount != 0 || decisionFor(t, report, path).Reason != "preserve local managed content while template is unchanged" {
			t.Fatalf("drift run %d: report=%#v err=%v", index, report, err)
		}
	}
}

func TestInterruptedUpdateRollsBackTreeAndLock(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	first, second := "memory-bank/dna/a.md", "memory-bank/dna/b.md"
	write(t, source, first, "a1\n")
	write(t, source, second, "b1\n")
	initialize(t, repo, source)
	write(t, source, first, "a2\n")
	write(t, source, second, "b2\n")
	lockBefore := read(t, repo, LockFileName)
	options := opts(repo, source, "b")
	count := 0
	options.BeforeMutation = func(Decision) error {
		count++
		if count == 2 {
			return errors.New("simulated interruption")
		}
		return nil
	}

	if _, err := Update(options); err == nil {
		t.Fatal("expected interrupted update error")
	}
	if read(t, repo, first) != "a1\n" || read(t, repo, second) != "b1\n" {
		t.Fatal("interrupted update left partial template changes")
	}
	if got := read(t, repo, LockFileName); got != lockBefore {
		t.Fatal("interrupted update changed lock")
	}
}

func TestSchemaZeroIsUpgradedAfterSuccessfulUpdate(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	path := "memory-bank/dna/rule.md"
	write(t, source, path, "base\n")
	initialize(t, repo, source)
	legacy := strings.Replace(read(t, repo, LockFileName), `"schema_version": 1`, `"schema_version": 0`, 1)
	write(t, repo, LockFileName, legacy)

	report, err := Update(opts(repo, source, "a"))
	if err != nil || !report.Applied {
		t.Fatalf("schema upgrade failed: report=%#v err=%v", report, err)
	}
	lock, exists, err := ReadLock(repo)
	if err != nil || !exists || lock.SchemaVersion != CurrentSchemaVersion {
		t.Fatalf("unexpected upgraded lock: %#v exists=%v err=%v", lock, exists, err)
	}
}

func TestUserOwnedFileIsNeverOverwrittenOrDeleted(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	path := "memory-bank/features/FT-001/brief.md"
	write(t, source, path, "seed\n")
	initialize(t, repo, source)
	write(t, repo, path, "user document\n")
	if err := os.Remove(filepath.Join(source, filepath.FromSlash(path))); err != nil {
		t.Fatal(err)
	}
	report, err := Update(opts(repo, source, "b"))
	if err != nil || decisionFor(t, report, path).Action != Preserve || read(t, repo, path) != "user document\n" {
		t.Fatalf("user-owned file was not preserved: report=%#v err=%v", report, err)
	}
}

func TestRemovedCleanManagedFileIsDeleted(t *testing.T) {
	repo, source := t.TempDir(), t.TempDir()
	path := "memory-bank/dna/obsolete.md"
	write(t, source, path, "obsolete\n")
	initialize(t, repo, source)
	if err := os.Remove(filepath.Join(source, filepath.FromSlash(path))); err != nil {
		t.Fatal(err)
	}

	report, err := Update(opts(repo, source, "b"))
	if err != nil || !report.Applied || decisionFor(t, report, path).Action != Delete {
		t.Fatalf("managed delete failed: report=%#v err=%v", report, err)
	}
	if _, err := os.Stat(filepath.Join(repo, filepath.FromSlash(path))); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("managed file was not deleted: %v", err)
	}
}
