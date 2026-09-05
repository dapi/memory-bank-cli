package lint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestRunMatchesPythonGoldenReport(t *testing.T) {
	repositoryRoot, err := filepath.Abs(filepath.Join("testdata", "repository"))
	if err != nil {
		t.Fatal(err)
	}
	report, err := Run(Options{RepoRoot: repositoryRoot, ScopeRoot: "memory-bank", MaxDepth: 1})
	if err != nil {
		t.Fatal(err)
	}
	report.RepoRoot = "repo-root"

	actual, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	actual = append(actual, '\n')
	expected, err := os.ReadFile(filepath.Join("testdata", "expected-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != string(expected) {
		t.Fatalf("report differs from the Python golden contract\n--- expected\n%s\n--- actual\n%s", expected, actual)
	}
}

func TestConfiguredEntrypointPrefersScope(t *testing.T) {
	repositoryRoot, err := filepath.Abs(filepath.Join("testdata", "repository"))
	if err != nil {
		t.Fatal(err)
	}
	report, err := Run(Options{
		RepoRoot: repositoryRoot, ScopeRoot: "memory-bank", Entrypoints: []string{"README.md", "README.md"}, MaxDepth: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(report.Entrypoints, []string{"memory-bank/README.md"}) {
		t.Fatalf("unexpected entrypoints: %#v", report.Entrypoints)
	}
}

func TestExtractInternalMarkdownLinks(t *testing.T) {
	text := strings.Join([]string{
		`[Titled](guide.md "Guide")[Adjacent](adjacent.md)`,
		`[Angle](<folder/a b.md>)`,
		`![Image](image.md)`,
		"```markdown\n[Ignored](ignored.md)\n```",
		`[External](https://example.com/page.md)`,
	}, "\n")
	want := []string{"memory-bank/guide.md", "memory-bank/adjacent.md", "memory-bank/folder/a b.md"}
	if got := extractInternalMarkdownLinks("memory-bank/README.md", text); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected links: got %#v, want %#v", got, want)
	}
}

func TestIndexAnnotationSkipsImagesBeforeLinks(t *testing.T) {
	text := "- ![Diagram](diagram.md) [Child](child.md) — Detailed child documentation.\n"
	got := annotationTextForChildLinks("memory-bank/README.md", text)
	want := []childAnnotation{{target: "memory-bank/child.md", annotation: "![Diagram](diagram.md)  — Detailed child documentation."}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected annotations: got %#v, want %#v", got, want)
	}
}

func TestParseFrontmatterDerivedFromForms(t *testing.T) {
	frontmatter := parseFrontmatter("---\nderived_from:\n  - path: one.md\n  - {path: two.md, role: source}\n  - three.md\n---\n")
	want := []string{"one.md", "two.md", "three.md"}
	if got := extractDerivedFromPaths(frontmatter); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected dependencies: got %#v, want %#v", got, want)
	}
}

func TestNormalizeScopeRootRejectsCurrentDirectory(t *testing.T) {
	if _, err := NormalizeScopeRoot("."); err == nil {
		t.Fatal("expected an error for current-directory scope")
	}
}

func TestNormalizeScopeRootRejectsEscapingRepository(t *testing.T) {
	for _, scopeRoot := range []string{"../sibling", "memory-bank/../sibling", "/tmp/sibling", `..\\sibling`} {
		t.Run(scopeRoot, func(t *testing.T) {
			if _, err := NormalizeScopeRoot(scopeRoot); err == nil {
				t.Fatalf("expected an error for scope root %q", scopeRoot)
			}
		})
	}
}

func writeDocument(t *testing.T, root, relative, contents string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A repository may project part of its tree through a directory symlink.
// Skipping it would drop every document below and report the referring links
// as broken, which is indistinguishable from a genuinely missing document.
func TestLoadDocumentsFollowsInRepoDirectorySymlink(t *testing.T) {
	repo := t.TempDir()
	writeDocument(t, repo, "template/memory-bank/flows/routing.md", "# Routing\n")
	writeDocument(t, repo, "memory-bank/README.md", "[routing](flows/routing.md)\n")
	if err := os.Symlink(filepath.Join("..", "template", "memory-bank", "flows"), filepath.Join(repo, "memory-bank", "flows")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	documents, err := loadDocuments(repo)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := documents["memory-bank/flows/routing.md"]; !ok {
		t.Fatalf("document behind a directory symlink is missing; loaded: %v", documents)
	}
}

func TestLoadDocumentsRecordsEveryLinkToTheSameTarget(t *testing.T) {
	repo := t.TempDir()
	writeDocument(t, repo, "template/memory-bank/flows/routing.md", "# Routing\n")
	for _, link := range []string{"memory-bank/flows", "docs/flows"} {
		full := filepath.Join(repo, filepath.FromSlash(link))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join("..", "template", "memory-bank", "flows"), full); err != nil {
			t.Skipf("symlinks are unavailable: %v", err)
		}
	}

	documents, err := loadDocuments(repo)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"memory-bank/flows/routing.md", "docs/flows/routing.md"} {
		if _, ok := documents[expected]; !ok {
			t.Fatalf("%s is missing; loaded: %v", expected, keysOf(documents))
		}
	}
}

// A link to an ancestor must not duplicate the tree under a phantom prefix.
func TestLoadDocumentsStopsAtAnAncestorCycle(t *testing.T) {
	repo := t.TempDir()
	writeDocument(t, repo, "a/doc.md", "# Doc\n")
	if err := os.Symlink("..", filepath.Join(repo, "a", "up")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	documents, err := loadDocuments(repo)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := documents["a/up/a/doc.md"]; ok {
		t.Fatalf("phantom duplicate recorded; loaded: %v", keysOf(documents))
	}
}

// Ignoring must follow what the link points at, not how it is named.
func TestLoadDocumentsIgnoresLinkIntoAnIgnoredDirectory(t *testing.T) {
	repo := t.TempDir()
	writeDocument(t, repo, "vendor/pkg/readme.md", "# Vendored\n")
	if err := os.Symlink(filepath.Join("..", "vendor"), filepath.Join(repo, "docs")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	documents, err := loadDocuments(repo)
	if err != nil {
		t.Fatal(err)
	}
	for documentPath := range documents {
		if strings.HasPrefix(documentPath, "docs/") {
			t.Fatalf("a link into vendor/ contributed %s", documentPath)
		}
	}
}

// A link inside a projected subtree that points back into the repository is
// still inside the repository and must be followed.
func TestLoadDocumentsFollowsLinkWithinProjectedSubtree(t *testing.T) {
	repo := t.TempDir()
	writeDocument(t, repo, "template/memory-bank/flows/routing.md", "# Routing\n")
	writeDocument(t, repo, "shared/note.md", "# Note\n")
	if err := os.Symlink(filepath.Join("template", "memory-bank", "flows"), filepath.Join(repo, "flows")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join("..", "..", "..", "shared", "note.md"), filepath.Join(repo, "template", "memory-bank", "flows", "note.md")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	documents, err := loadDocuments(repo)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := documents["flows/note.md"]; !ok {
		t.Fatalf("in-repository link inside a projected subtree was dropped; loaded: %v", keysOf(documents))
	}
}

func keysOf(documents map[string]document) []string {
	paths := make([]string, 0, len(documents))
	for documentPath := range documents {
		paths = append(paths, documentPath)
	}
	sort.Strings(paths)
	return paths
}

// A downstream repository may symlink a shared document into memory-bank/.
// It was audited under its in-repository path before projections existed and
// must stay audited: dropping it turns every reference into a broken link.
func TestLoadDocumentsKeepsDocumentLinkedFromOutside(t *testing.T) {
	repo, outside := t.TempDir(), t.TempDir()
	writeDocument(t, outside, "shared.md", "# Shared\n")
	if err := os.MkdirAll(filepath.Join(repo, "memory-bank"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "shared.md"), filepath.Join(repo, "memory-bank", "shared.md")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	documents, err := loadDocuments(repo)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := documents["memory-bank/shared.md"]; !ok {
		t.Fatalf("document linked from outside was dropped; loaded: %v", keysOf(documents))
	}
}
