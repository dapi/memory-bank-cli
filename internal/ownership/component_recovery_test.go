package ownership

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComponentRecoveryRequiresCompleteRestoration(t *testing.T) {
	if !componentHost() {
		t.Skip("POSIX component writer")
	}
	root := t.TempDir()
	first := "memory-bank/a.md"
	second := "memory-bank/b.md"
	write(t, root, first, "before-a")
	write(t, root, second, "before-b")
	repo, err := pinRepoRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	before := map[string]observation{}
	for _, p := range []string{first, second} {
		o, _, e := observeComponent(repo, p)
		if e != nil {
			t.Fatal(e)
		}
		before[p] = o
	}
	mutations := []mutation{{decision: Decision{Path: first, Action: UpdateFile}, data: []byte("after-a"), expectedExists: true}, {decision: Decision{Path: second, Action: UpdateFile}, data: []byte("after-b"), expectedExists: true}}
	calls := 0
	options := Options{RepoRoot: root, componentTransaction: true, componentObservations: before, BeforeMutation: func(Decision) error {
		calls++
		if calls == 2 {
			return errors.New("interrupt")
		}
		return nil
	}}
	ops := transactionOps{writeFile: os.WriteFile, rename: os.Rename, link: func(from, to string) error {
		if filepath.Base(filepath.Dir(from)) == "old" {
			return errors.New("rollback blocked")
		}
		return os.Link(from, to)
	}}
	err = applyAtomicallyPinnedWithOps(options, mutations, repo, ops)
	if err == nil || !strings.Contains(err.Error(), "rollback incomplete") {
		t.Fatalf("expected retained failure: %v", err)
	}
	staging, _ := filepath.Glob(filepath.Join(root, ".memory-bank-update-*"))
	if len(staging) != 1 {
		t.Fatal(staging)
	}
	if err = checkComponentRecovery(repo, true); err == nil {
		t.Fatal("mixed state allowed")
	}
	// Restore the original using the journal's numbered backup, without replay.
	b, err := os.ReadFile(filepath.Join(staging[0], "old", "000000"))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, first), b, 0644); err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(filepath.Join(root, first), 0600); err != nil {
		t.Fatal(err)
	}
	if err = checkComponentRecovery(repo, true); err == nil {
		t.Fatal("permission-only partial restoration allowed")
	}
	if err = os.Chmod(filepath.Join(root, first), 0644); err != nil {
		t.Fatal(err)
	}
	if err = checkComponentRecovery(repo, true); err != nil {
		t.Fatal(err)
	}
	if err = checkComponentRecovery(repo, true); err != nil {
		t.Fatal(err)
	}
	assertNoTransactionStaging(t, root)
}
func TestComponentRejectsHardLinksAndPortableAliases(t *testing.T) {
	if !componentHost() {
		t.Skip("POSIX")
	}
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("safe"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "memory-bank"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(outside, filepath.Join(root, "memory-bank", "target.md")); err != nil {
		t.Fatal(err)
	}
	repo, err := pinRepoRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = observeComponent(repo, "memory-bank/target.md"); err == nil {
		t.Fatal("external hardlink accepted")
	}
	if err = os.Remove(filepath.Join(root, "memory-bank", "target.md")); err != nil {
		t.Fatal(err)
	}
	write(t, root, "memory-bank/Upper.md", "safe")
	if _, _, err = observeComponent(repo, "memory-bank/upper.md"); err == nil {
		t.Fatal("case alias accepted")
	}
}
