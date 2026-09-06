package ownership

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const legacyDeclaration = `{"schema_version":1,"payload_format":"legacy/v1","capabilities":["legacy/v1"]}`

func rawSourceCommit(t *testing.T, source string) string {
	t.Helper()
	runGitTest(t, source, "init", "--quiet")
	runGitTest(t, source, "add", "--all")
	runGitTest(t, source, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "--quiet", "-m", "source-format fixture")
	return runGitTest(t, source, "rev-parse", "HEAD")
}

func treeSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		if d.IsDir() {
			result[rel] = "dir"
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		result[rel] = info.Mode().String() + ":" + string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestSourceFormatRejectionIsNonMutating(t *testing.T) {
	cases := []struct{ name, declaration, marker, want string }{
		{"unknown manifestless", "", "", "unsupported manifestless"},
		{"unknown schema", `{"schema_version":2,"payload_format":"legacy/v1","capabilities":["legacy/v1"]}`, "", "unsupported source format"},
		{"component format", `{"schema_version":1,"payload_format":"components/v1","capabilities":["components/v1"]}`, "", "unsupported source format"},
		{"unknown field", `{"schema_version":1,"payload_format":"legacy/v1","capabilities":["legacy/v1"],"extra":true}`, "", "unknown field"},
		{"case-aliased field", `{"Schema_Version":1,"payload_format":"legacy/v1","capabilities":["legacy/v1"]}`, "", "unknown field"},
		{"wrong field type", `{"schema_version":"1","payload_format":"legacy/v1","capabilities":["legacy/v1"]}`, "", "cannot unmarshal"},
		{"duplicate field", `{"schema_version":1,"schema_version":1,"payload_format":"legacy/v1","capabilities":["legacy/v1"]}`, "", "duplicate JSON field"},
		{"trailing JSON", legacyDeclaration + `{}`, "", "trailing JSON"},
		{"missing capability", `{"schema_version":1,"payload_format":"legacy/v1"}`, "", "requires capability"},
		{"unsupported capability", `{"schema_version":1,"payload_format":"legacy/v1","capabilities":["legacy/v1","components/v1"]}`, "", "unsupported source capability"},
		{"duplicate capability", `{"schema_version":1,"payload_format":"legacy/v1","capabilities":["legacy/v1","legacy/v1"]}`, "", "duplicate source capability"},
		{"legacy with component marker", legacyDeclaration, "template/memory-bank/components.json", "unsupported component source"},
		{"manifestless with marker", "", "template/memory-bank/components.json", "unsupported component source"},
		{"null", `null`, "", "unsupported source format"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source, repo := t.TempDir(), t.TempDir()
			write(t, source, "template/memory-bank/dna/rule.md", "incoming\n")
			if tc.declaration != "" {
				write(t, source, SourceDeclarationFile, tc.declaration)
			}
			if tc.marker != "" {
				write(t, source, tc.marker, "{}")
			}
			ref := rawSourceCommit(t, source)
			options := Options{RepoRoot: repo, SourceRoot: source, TemplateVersion: "fixture", SourceRef: ref}
			write(t, repo, "user.md", "preserve me")
			check := func(err error, before map[string]string) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("want %q, got %v", tc.want, err)
				}
				if !reflect.DeepEqual(before, treeSnapshot(t, repo)) {
					t.Fatal("rejected operation changed downstream")
				}
			}
			for _, dry := range []bool{false, true} {
				options.DryRun = dry
				before := treeSnapshot(t, repo)
				_, err := Init(options)
				check(err, before)
			}
			validSource := t.TempDir()
			write(t, validSource, "template/memory-bank/dna/rule.md", "base\n")
			validRef := commitTestSource(t, validSource)
			if _, err := Init(Options{RepoRoot: repo, SourceRoot: validSource, TemplateVersion: "base", SourceRef: validRef}); err != nil {
				t.Fatal(err)
			}
			for _, dry := range []bool{false, true} {
				options.DryRun = dry
				before := treeSnapshot(t, repo)
				_, err := Update(options)
				check(err, before)
			}
			before := treeSnapshot(t, repo)
			_, err := PlanPull(options)
			check(err, before)
			// A saved pre-bridge plan cannot exempt its source from revalidation.
			_, err = ApplyResolutionPlan(options, ResolutionPlan{FormatVersion: 1, Template: Template{Version: "fixture", SourceRef: ref}})
			check(err, before)
		})
	}
}

func TestSourceDeclarationMustBeRegularTrackedBlob(t *testing.T) {
	source := t.TempDir()
	write(t, source, "template/memory-bank/README.md", "payload")
	write(t, source, "declaration.json", legacyDeclaration)
	symlinkForTest(t, "declaration.json", filepath.Join(source, SourceDeclarationFile))
	ref := rawSourceCommit(t, source)
	if err := verifySourceCheckout(source, ref); err == nil || !strings.Contains(err.Error(), "regular Git blob") {
		t.Fatalf("got %v", err)
	}
}

func TestComponentMarkerRejectedInLegacyPayloadRoots(t *testing.T) {
	for _, root := range []string{"memory-bank", "memory-bank-template"} {
		t.Run(root, func(t *testing.T) {
			source := t.TempDir()
			write(t, source, root+"/components.json", "{}")
			write(t, source, SourceDeclarationFile, legacyDeclaration)
			ref := rawSourceCommit(t, source)
			if err := verifySourceCheckout(source, ref); err == nil || !strings.Contains(err.Error(), "unsupported component source") {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestPinnedSourceIgnoresGitReplacementObjects(t *testing.T) {
	source, repo := t.TempDir(), t.TempDir()
	write(t, source, "template/memory-bank/README.md", "original pinned content\n")
	write(t, source, SourceDeclarationFile, legacyDeclaration)
	original := rawSourceCommit(t, source)
	write(t, source, "template/memory-bank/README.md", "replacement content\n")
	replacement := rawSourceCommit(t, source)
	runGitTest(t, source, "replace", original, replacement)
	runGitTest(t, source, "checkout", "--quiet", "--detach", original)
	before := treeSnapshot(t, repo)
	_, err := Init(Options{RepoRoot: repo, SourceRoot: source, TemplateVersion: "pinned", SourceRef: original})
	if err == nil {
		t.Fatal("installed replacement objects under an unchanged source commit identity")
	}
	if !reflect.DeepEqual(before, treeSnapshot(t, repo)) {
		t.Fatal("replaced source changed downstream")
	}
}

func TestUnchangedPullPersistsManagedAdaptation(t *testing.T) {
	source, repo := t.TempDir(), t.TempDir()
	target := "memory-bank/domain/model.md"
	write(t, source, "template/"+target, "template model\n")
	ref := commitTestSource(t, source)
	options := Options{RepoRoot: repo, SourceRoot: source, TemplateVersion: "fixture", SourceRef: ref}
	if _, err := Init(options); err != nil {
		t.Fatal(err)
	}
	write(t, repo, target, "template model\nproject adaptation\n")
	report, err := Update(options)
	if err != nil {
		t.Fatal(err)
	}
	lock, _, err := ReadLock(repo)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Applied || lock.Files[target].Ownership != Adapted {
		t.Fatalf("ownership change was not persisted: report=%#v file=%#v", report, lock.Files[target])
	}
	before := treeSnapshot(t, repo)
	report, err = Update(options)
	if err != nil || report.Applied || !reflect.DeepEqual(before, treeSnapshot(t, repo)) {
		t.Fatalf("repeat pull is not a no-op: report=%#v err=%v", report, err)
	}
}
