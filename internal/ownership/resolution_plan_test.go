package ownership

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReviewedCanonicalMigrationMergeAppliesAtomicallyAndNextPullIsNoOp(t *testing.T) {
	source, repo := t.TempDir(), t.TempDir()
	adapted := "memory-bank/domain/model.md"
	managed := "memory-bank/dna/rule.md"
	write(t, source, adapted, "title\nbase body\nlocal anchor\n")
	write(t, source, managed, "managed v1\n")
	baseRef := commitTestSource(t, source)
	initOptions := Options{RepoRoot: repo, SourceRoot: source, TemplateVersion: "base", SourceRef: baseRef, Now: func() time.Time { return fixedTime }}
	report, err := Init(initOptions)
	if err != nil || !report.Applied {
		t.Fatalf("init report=%#v err=%v", report, err)
	}
	write(t, repo, adapted, "title\nbase body\nlocal project section\nlocal anchor\n")

	if err := os.RemoveAll(filepath.Join(source, "memory-bank")); err != nil {
		t.Fatal(err)
	}
	write(t, source, "template/"+adapted, "title\nupstream behavior section\nbase body\nlocal anchor\n")
	write(t, source, "template/"+managed, "managed v2\n")
	runGitTest(t, source, "add", "--all")
	runGitTest(t, source, "-c", "user.name=Memory Bank Tests", "-c", "user.email=tests@example.invalid", "commit", "--quiet", "-m", "canonical source")
	targetRef := runGitTest(t, source, "rev-parse", "HEAD")
	options := Options{RepoRoot: repo, SourceRoot: source, TemplateVersion: "target", SourceRef: targetRef, Now: func() time.Time { return fixedTime }}

	lockBefore := read(t, repo, LockFileName)
	plan, err := PlanPull(options)
	if err != nil {
		t.Fatal(err)
	}
	if got := read(t, repo, LockFileName); got != lockBefore {
		t.Fatal("planning changed the lock")
	}
	entry := resolutionEntryFor(t, plan, adapted)
	if !entry.RequiresHumanDecision || entry.Merge == nil || !containsResolutionAction(entry.AllowedActions, "apply-reviewed-merge") {
		t.Fatalf("adapted entry has no reviewed merge: %#v", entry)
	}
	merged, err := base64.StdEncoding.DecodeString(entry.Merge.ContentBase64)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(merged), "title\nupstream behavior section\nbase body\nlocal project section\nlocal anchor\n"; got != want {
		t.Fatalf("merge=%q, want %q", got, want)
	}
	for index := range plan.Entries {
		if plan.Entries[index].Path == adapted {
			plan.Entries[index].SelectedAction = "apply-reviewed-merge"
		}
	}
	report, err = ApplyResolutionPlan(options, plan)
	if err != nil || !report.Applied || report.ConflictCount != 0 {
		t.Fatalf("apply report=%#v err=%v", report, err)
	}
	if got := read(t, repo, adapted); got != string(merged) {
		t.Fatalf("applied merge=%q", got)
	}
	if got := read(t, repo, managed); got != "managed v2\n" {
		t.Fatalf("managed update=%q", got)
	}
	lock, exists, err := ReadLock(repo)
	if err != nil || !exists || lock.Template.SourceRef != targetRef || lock.Files[adapted].Ownership != Adapted || lock.Files[adapted].BaseDigest != digest([]byte("title\nupstream behavior section\nbase body\nlocal anchor\n")) {
		t.Fatalf("unexpected lock: %#v exists=%v err=%v", lock, exists, err)
	}
	report, err = Update(options)
	if err != nil || report.Applied || report.ConflictCount != 0 {
		t.Fatalf("second pull is not a no-op: report=%#v err=%v", report, err)
	}
}

func TestReviewedPlanRejectsUnresolvedStaleAndAlteredState(t *testing.T) {
	source, repo, options, path := resolutionConflictFixture(t)
	_ = source
	plan, err := PlanPull(options)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyResolutionPlan(options, plan); err == nil || !strings.Contains(err.Error(), "unresolved") {
		t.Fatalf("unresolved apply error=%v", err)
	}
	for index := range plan.Entries {
		if plan.Entries[index].Path == path {
			plan.Entries[index].SelectedAction = "keep-local"
		}
	}
	write(t, repo, path, "changed after review\n")
	if _, err := ApplyResolutionPlan(options, plan); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale apply error=%v", err)
	}
	write(t, repo, path, "local\nbase tail\n")
	plan, err = PlanPull(options)
	if err != nil {
		t.Fatal(err)
	}
	for index := range plan.Entries {
		if plan.Entries[index].Path == path {
			plan.Entries[index].SelectedAction = "keep-local"
			plan.Entries[index].Reason = "misleading reason"
		}
	}
	if _, err := ApplyResolutionPlan(options, plan); err == nil || !strings.Contains(err.Error(), "altered") {
		t.Fatalf("altered apply error=%v", err)
	}
}

func TestReviewedPlanTakeUpstreamAndKeepLocalAdvanceOwnership(t *testing.T) {
	for _, action := range []string{"keep-local", "take-upstream"} {
		t.Run(action, func(t *testing.T) {
			_, repo, options, path := resolutionConflictFixture(t)
			plan, err := PlanPull(options)
			if err != nil {
				t.Fatal(err)
			}
			for index := range plan.Entries {
				if plan.Entries[index].Path == path {
					plan.Entries[index].SelectedAction = action
				}
			}
			report, err := ApplyResolutionPlan(options, plan)
			if err != nil || !report.Applied || report.ConflictCount != 0 {
				t.Fatalf("apply report=%#v err=%v", report, err)
			}
			lock, _, err := ReadLock(repo)
			if err != nil {
				t.Fatal(err)
			}
			if action == "keep-local" {
				if got := read(t, repo, path); got != "local\nbase tail\n" || lock.Files[path].Ownership != Adapted {
					t.Fatalf("keep result=%q lock=%#v", got, lock.Files[path])
				}
			} else if got := read(t, repo, path); got != "upstream\nbase tail\n" || lock.Files[path].Ownership != Managed {
				t.Fatalf("take result=%q lock=%#v", got, lock.Files[path])
			}
			report, err = Update(options)
			if err != nil || report.ConflictCount != 0 {
				t.Fatalf("repeat pull report=%#v err=%v", report, err)
			}
		})
	}
}

func TestReviewedPlanKeepsAndDetachesUserOwnedRemoval(t *testing.T) {
	source, repo := t.TempDir(), t.TempDir()
	userPath := "memory-bank/features/FT-1/notes.md"
	seedPath := "memory-bank/dna/rule.md"
	write(t, source, userPath, "project notes\n")
	write(t, source, seedPath, "seed\n")
	baseRef := commitTestSource(t, source)
	report, err := Init(Options{RepoRoot: repo, SourceRoot: source, TemplateVersion: "base", SourceRef: baseRef, Now: func() time.Time { return fixedTime }})
	if err != nil || !report.Applied {
		t.Fatalf("init report=%#v err=%v", report, err)
	}
	if err := os.Remove(filepath.Join(source, filepath.FromSlash(userPath))); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, source, "add", "--all")
	runGitTest(t, source, "-c", "user.name=Memory Bank Tests", "-c", "user.email=tests@example.invalid", "commit", "--quiet", "-m", "remove user path")
	targetRef := runGitTest(t, source, "rev-parse", "HEAD")
	options := Options{RepoRoot: repo, SourceRoot: source, TemplateVersion: "target", SourceRef: targetRef, Now: func() time.Time { return fixedTime }}
	plan, err := PlanPull(options)
	if err != nil {
		t.Fatal(err)
	}
	entry := resolutionEntryFor(t, plan, userPath)
	if entry.RequiresHumanDecision || entry.ProposedAction != Preserve {
		t.Fatalf("user-owned removal is not deterministic: %#v", entry)
	}
	report, err = ApplyResolutionPlan(options, plan)
	if err != nil || !report.Applied || report.ConflictCount != 0 {
		t.Fatalf("apply report=%#v err=%v", report, err)
	}
	if got := read(t, repo, userPath); got != "project notes\n" {
		t.Fatalf("user-owned content changed: %q", got)
	}
	lock, _, err := ReadLock(repo)
	if err != nil {
		t.Fatal(err)
	}
	if _, tracked := lock.Files[userPath]; tracked {
		t.Fatalf("user-owned removal remained tracked: %#v", lock.Files[userPath])
	}
}

func TestReviewedPlanTransactionFailureRollsBackLock(t *testing.T) {
	_, repo, options, path := resolutionConflictFixture(t)
	plan, err := PlanPull(options)
	if err != nil {
		t.Fatal(err)
	}
	for index := range plan.Entries {
		if plan.Entries[index].Path == path {
			plan.Entries[index].SelectedAction = "keep-local"
		}
	}
	lockBefore := read(t, repo, LockFileName)
	localBefore := read(t, repo, path)
	options.BeforeMutation = func(decision Decision) error {
		if decision.Path == LockFileName {
			return errors.New("injected lock failure")
		}
		return nil
	}
	if _, err := ApplyResolutionPlan(options, plan); err == nil || !strings.Contains(err.Error(), "injected lock failure") {
		t.Fatalf("apply error=%v", err)
	}
	if got := read(t, repo, LockFileName); got != lockBefore {
		t.Fatal("failed apply changed lock")
	}
	if got := read(t, repo, path); got != localBefore {
		t.Fatalf("failed apply changed local content: %q", got)
	}
}

func TestReviewedPlanDoesNotOfferUnrepresentableLocalDeletion(t *testing.T) {
	_, repo, options, path := resolutionConflictFixture(t)
	if err := os.Remove(filepath.Join(repo, filepath.FromSlash(path))); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanPull(options)
	if err != nil {
		t.Fatal(err)
	}
	entry := resolutionEntryFor(t, plan, path)
	if got, want := strings.Join(entry.AllowedActions, ","), "take-upstream"; got != want {
		t.Fatalf("allowed actions=%q, want %q; entry=%#v", got, want, entry)
	}
	if !strings.Contains(entry.MergeUnavailableReason, "lock v1") {
		t.Fatalf("missing representation reason: %#v", entry)
	}
}

func TestOrdinaryPullLeavesOverlappingAdaptedChangesForReview(t *testing.T) {
	_, repo, options, path := resolutionConflictFixture(t)
	beforeLock := read(t, repo, LockFileName)
	beforeContent := read(t, repo, path)

	report, err := Update(options)
	if err != nil {
		t.Fatal(err)
	}
	if report.Applied || report.ConflictCount != 1 {
		t.Fatalf("overlapping ordinary pull report=%#v", report)
	}
	if got := read(t, repo, path); got != beforeContent {
		t.Fatalf("overlapping ordinary pull changed adapted file: %q", got)
	}
	if got := read(t, repo, LockFileName); got != beforeLock {
		t.Fatal("overlapping ordinary pull changed lock")
	}
}

func resolutionConflictFixture(t *testing.T) (string, string, Options, string) {
	t.Helper()
	source, repo := t.TempDir(), t.TempDir()
	path := "memory-bank/domain/model.md"
	write(t, source, path, "base\nbase tail\n")
	baseRef := commitTestSource(t, source)
	report, err := Init(Options{RepoRoot: repo, SourceRoot: source, TemplateVersion: "base", SourceRef: baseRef, Now: func() time.Time { return fixedTime }})
	if err != nil || !report.Applied {
		t.Fatalf("init report=%#v err=%v", report, err)
	}
	write(t, repo, path, "local\nbase tail\n")
	if err := os.RemoveAll(filepath.Join(source, "memory-bank")); err != nil {
		t.Fatal(err)
	}
	write(t, source, "template/"+path, "upstream\nbase tail\n")
	runGitTest(t, source, "add", "--all")
	runGitTest(t, source, "-c", "user.name=Memory Bank Tests", "-c", "user.email=tests@example.invalid", "commit", "--quiet", "-m", "target")
	targetRef := runGitTest(t, source, "rev-parse", "HEAD")
	return source, repo, Options{RepoRoot: repo, SourceRoot: source, TemplateVersion: "target", SourceRef: targetRef, Now: func() time.Time { return fixedTime }}, path
}

func resolutionEntryFor(t *testing.T, plan ResolutionPlan, path string) ResolutionPlanEntry {
	t.Helper()
	for _, entry := range plan.Entries {
		if entry.Path == path {
			return entry
		}
	}
	t.Fatalf("missing plan entry %s: %#v", path, plan.Entries)
	return ResolutionPlanEntry{}
}
