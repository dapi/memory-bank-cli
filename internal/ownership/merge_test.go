package ownership

import (
	"strings"
	"testing"
)

func TestMechanicalMergeCombinesNonOverlappingLineChanges(t *testing.T) {
	merged, mode, err := mechanicalMerge(
		[]byte("one\ntwo\nthree\n"),
		[]byte("ONE\ntwo\nthree\n"),
		[]byte("one\ntwo\nTHREE\n"),
		"100644", "100755", "100644",
	)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(merged), "ONE\ntwo\nTHREE\n"; got != want {
		t.Fatalf("merged=%q, want %q", got, want)
	}
	if mode != "100755" {
		t.Fatalf("mode=%q, want 100755", mode)
	}
}

func TestMechanicalMergeAcceptsIdenticalOverlappingChange(t *testing.T) {
	merged, _, err := mechanicalMerge([]byte("one\n"), []byte("same\n"), []byte("same\n"), "100644", "100644", "100644")
	if err != nil || string(merged) != "same\n" {
		t.Fatalf("merged=%q err=%v", merged, err)
	}
}

func TestMechanicalMergeRejectsContentAndModeConflicts(t *testing.T) {
	if _, _, err := mechanicalMerge([]byte("one\n"), []byte("local\n"), []byte("upstream\n"), "100644", "100644", "100644"); err == nil || !strings.Contains(err.Error(), "overlap") {
		t.Fatalf("content conflict error=%v", err)
	}
	if _, _, err := mechanicalMerge([]byte("one\n"), []byte("one\n"), []byte("one\n"), "100644", "100755", "100600"); err == nil || !strings.Contains(err.Error(), "mode") {
		t.Fatalf("mode conflict error=%v", err)
	}
}

func TestMechanicalMergeRejectsBinaryAndExcessiveWork(t *testing.T) {
	if _, _, err := mechanicalMerge([]byte("a\x00b"), []byte("a\x00b"), []byte("a\x00b"), "100644", "100644", "100644"); err == nil {
		t.Fatal("binary merge was accepted")
	}
	lines := strings.Repeat("line\n", 2100)
	if _, _, err := mechanicalMerge([]byte(lines), []byte(lines), []byte(lines), "100644", "100644", "100644"); err == nil || !strings.Contains(err.Error(), "work limit") {
		t.Fatalf("work limit error=%v", err)
	}
}
