package projection

import (
	"os"
	"path/filepath"
	"testing"
)

func symlinkForTest(t *testing.T, target, link string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
}

func writeForTest(t *testing.T, root, relative, contents string) string {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return full
}

func TestRecognisesFileProjection(t *testing.T) {
	repo := t.TempDir()
	writeForTest(t, repo, "template/memory-bank/dna/rule.md", "payload\n")
	symlinkForTest(t, "../../template/memory-bank/dna/rule.md", filepath.Join(repo, "memory-bank/dna/rule.md"))

	if !IsPayloadProjection(repo, "memory-bank/dna/rule.md") {
		t.Fatal("symlink into the repository's own payload must be recognised as a projection")
	}
}

func TestRecognisesProjectionThroughDirectorySymlink(t *testing.T) {
	repo := t.TempDir()
	writeForTest(t, repo, "template/memory-bank/flows/routing.md", "payload\n")
	symlinkForTest(t, "../template/memory-bank/flows", filepath.Join(repo, "memory-bank/flows"))

	if !IsPayloadProjection(repo, "memory-bank/flows/routing.md") {
		t.Fatal("a directory symlink into the payload projects the files below it")
	}
}

func TestRejectsSymlinkEscapingTheRepository(t *testing.T) {
	repo, outside := t.TempDir(), t.TempDir()
	writeForTest(t, repo, "template/memory-bank/dna/rule.md", "payload\n")
	writeForTest(t, outside, "rule.md", "payload\n")
	symlinkForTest(t, filepath.Join(outside, "rule.md"), filepath.Join(repo, "memory-bank/dna/rule.md"))

	if IsPayloadProjection(repo, "memory-bank/dna/rule.md") {
		t.Fatal("a link leaving the repository root must never count as a projection")
	}
}

func TestRejectsSymlinkToADifferentPayloadFile(t *testing.T) {
	repo := t.TempDir()
	writeForTest(t, repo, "template/memory-bank/dna/rule.md", "payload\n")
	writeForTest(t, repo, "template/memory-bank/dna/other.md", "payload\n")
	symlinkForTest(t, "../../template/memory-bank/dna/other.md", filepath.Join(repo, "memory-bank/dna/rule.md"))

	if IsPayloadProjection(repo, "memory-bank/dna/rule.md") {
		t.Fatal("only the payload file backing this exact destination is a projection")
	}
}

func TestRejectsRegularFileAndBrokenLink(t *testing.T) {
	repo := t.TempDir()
	writeForTest(t, repo, "template/memory-bank/dna/rule.md", "payload\n")
	writeForTest(t, repo, "memory-bank/dna/rule.md", "local override\n")
	symlinkForTest(t, "../../template/memory-bank/dna/missing.md", filepath.Join(repo, "memory-bank/dna/broken.md"))

	if IsPayloadProjection(repo, "memory-bank/dna/rule.md") {
		t.Fatal("a regular file is an override, not a projection")
	}
	if IsPayloadProjection(repo, "memory-bank/dna/broken.md") {
		t.Fatal("a broken link is not a projection")
	}
}

func TestRejectsPathsOutsideTheRepositoryRoot(t *testing.T) {
	repo := t.TempDir()
	writeForTest(t, repo, "template/memory-bank/dna/rule.md", "payload\n")

	for _, relative := range []string{"", "../escape.md", "/absolute.md"} {
		if IsPayloadProjection(repo, relative) {
			t.Fatalf("path %q must not be treated as a projection", relative)
		}
	}
}
