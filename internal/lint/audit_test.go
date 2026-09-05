package lint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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

func TestLoadDocumentsIgnoresSymlinkLeavingTheRepository(t *testing.T) {
	repo, outside := t.TempDir(), t.TempDir()
	writeDocument(t, outside, "secret.md", "# Secret\n")
	if err := os.Symlink(filepath.Join(outside, "secret.md"), filepath.Join(repo, "leak.md")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	documents, err := loadDocuments(repo)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := documents["leak.md"]; ok {
		t.Fatal("a link pointing outside the repository must not contribute a document")
	}
}
