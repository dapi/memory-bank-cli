package ownership

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/dapi/memory-bank-cli/internal/contracts"
)

func TestComponentLegacyMigration(t *testing.T) {
	source := os.Getenv("MEMORY_BANK_COMPONENT_SOURCE")
	legacySource := os.Getenv("MEMORY_BANK_LEGACY_SOURCE")
	if source == "" || legacySource == "" {
		t.Skip("set MEMORY_BANK_COMPONENT_SOURCE and MEMORY_BANK_LEGACY_SOURCE")
	}
	for _, invalid := range []bool{false, true} {
		name := "valid"
		if invalid {
			name = "invalid"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			oldRef := SupportedLegacySourceRefs()[0]
			init := Options{RepoRoot: root, SourceRoot: legacySource, TemplateVersion: "legacy", SourceRef: oldRef}
			if r, err := Init(init); err != nil || !r.Applied || r.ConflictCount != 0 {
				t.Fatalf("legacy init: %+v %v", r, err)
			}
			p := "memory-bank/features/FT-321/brief.md"
			meta := "---\nstatus: draft\ndoc_kind: feature\ndelivery_status: planned\n---\n\n# Feature\n"
			if invalid {
				meta = strings.Replace(meta, "delivery_status: planned\n", "", 1)
			}
			write(t, root, p, meta)
			write(t, root, "memory-bank/features/FT-321/README.md", "---\nstatus: draft\ndoc_function: index\npurpose: Navigate feature.\n---\n\n# Feature\n\n- [Brief](brief.md) — project brief.\n")

			second := "memory-bank/features/FT-322/brief.md"
			write(t, root, second, "---\nstatus: draft\ndoc_kind: feature\ndelivery_status: planned\n---\n\n# Second feature\n")
			write(t, root, "memory-bank/features/FT-322/README.md", "---\nstatus: draft\ndoc_function: index\npurpose: Navigate feature.\n---\n\n# Feature\n\n- [Brief](brief.md) — project brief.\n")
			index := filepath.Join(root, "memory-bank/features/README.md")
			b, err := os.ReadFile(index)
			if err != nil {
				t.Fatal(err)
			}
			b = append(b, []byte("\n- [FT-321](FT-321/README.md) — feature package.\n- [FT-322](FT-322/README.md) — second feature package.\n")...)
			if err = os.WriteFile(index, b, 0644); err != nil {
				t.Fatal(err)
			}
			originalLock, _ := os.ReadFile(filepath.Join(root, LockFileName))
			o := Options{RepoRoot: root, SourceRoot: source, TemplateVersion: "components", SourceRef: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", verifySource: func(string, string) error { return nil }, DryRun: true}
			if _, err = Update(o); err == nil {
				t.Fatal("flagless legacy migration accepted")
			}
			o.MigrateComponents = true
			if _, err = Update(o); err == nil {
				t.Fatal("managed index drift silently adopted")
			}
			o.MigrationResolution = []byte(`{"schema_version":1,"documents":[],"ownership":{"memory-bank/features/README.md":"keep-local"}}`)
			preview, err := Update(o)
			if err != nil || preview.ConflictCount > 0 || !contracts.ValidDigest(preview.MigrationPlanDigest) {
				t.Fatalf("preview: %+v %v", preview, err)
			}
			repeated, err := Update(o)
			if err != nil || preview.MigrationPlanDigest != repeated.MigrationPlanDigest {
				t.Fatalf("unstable preview: %v %v", repeated, err)
			}
			afterPreview, _ := os.ReadFile(filepath.Join(root, LockFileName))
			if string(originalLock) != string(afterPreview) {
				t.Fatal("preview mutated lock")
			}
			o.DryRun = false
			o.MigrationPlanDigest = preview.MigrationPlanDigest
			if err = os.Chmod(filepath.Join(root, p), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err = Update(o); err == nil {
				t.Fatal("permission change did not stale digest")
			}
			if err = os.Chmod(filepath.Join(root, p), 0644); err != nil {
				t.Fatal(err)
			}
			r, err := Update(o)
			if err != nil || !r.Applied {
				t.Fatalf("apply: %+v %v", r, err)
			}
			lock, _, err := ReadLock(root)
			if err != nil || lock.SchemaVersion != 2 {
				t.Fatalf("new lock: %+v %v", lock, err)
			}
			data, _ := os.ReadFile(filepath.Join(root, contracts.RegistryPath))
			registry, err := contracts.ReadRegistry(data)
			if err != nil || len(registry.Selectors) != 1 || len(registry.Records) != 0 {
				t.Fatalf("registry: %+v %v", registry, err)
			}
			if len(registry.Selectors[0].Snapshot) != 2 {
				t.Fatal("migration snapshot does not cover both documents")
			}
			_, findings, nav, err := ValidateComponents(root, "")
			if err != nil || nav.ExitCode != 0 || (len(findings) > 0) != invalid {
				t.Fatalf("preserved verdict: %+v %+v %v", findings, nav.Errors, err)
			}
			if !invalid {
				original, err := os.ReadFile(filepath.Join(root, p))
				if err != nil {
					t.Fatal(err)
				}
				ready := strings.Replace(string(original), "status: draft\n", "status: draft\ntitle: Feature\npurpose: Feature scope.\n", 1) + "\n## Acceptance\n\n## Outcome\n\n## Problem\n\n## Scope\n\n## Design Requirement Decision\n\nDesign required: no\n\n## Validation Profile Decision\n\nDocs checks.\n\n## Verify\n\nEvidence.\n"
				write(t, root, p, ready)
				transition := DocumentOptions{RepoRoot: root, Operation: "transition", Path: p, Contract: "feature/v1"}
				if _, err = DocumentOperation(transition); err == nil {
					t.Fatal("transition omitted required evidence")
				}
				transition.Evidence = []string{"review:test"}
				before := treeSnapshot(t, root)
				transition.BeforeMutation = func(d Decision) error {
					if d.Path == contracts.RegistryPath {
						return errors.New("interrupt registry")
					}
					return nil
				}
				if _, err = DocumentOperation(transition); err == nil {
					t.Fatal("interrupted transition succeeded")
				}
				if !reflect.DeepEqual(before, treeSnapshot(t, root)) {
					t.Fatal("transition rollback changed tree")
				}
				transition.BeforeMutation = nil
				if r, err := DocumentOperation(transition); err != nil || !r.Applied {
					t.Fatalf("selector transition: %+v %v", r, err)
				}
				data, _ = os.ReadFile(filepath.Join(root, contracts.RegistryPath))
				registry, err = contracts.ReadRegistry(data)
				if err != nil || len(registry.Records) != 1 || len(registry.Selectors[0].Exclusions) != 1 {
					t.Fatalf("selector exclusion: %+v %v", registry, err)
				}
				if r, err := DocumentOperation(DocumentOptions{RepoRoot: root, Operation: "create", Type: "feature", Path: "memory-bank/features/FT-323/brief.md"}); err != nil || !r.Applied {
					t.Fatalf("post-migration base creation: %+v %v", r, err)
				}
				data, _ = os.ReadFile(filepath.Join(root, contracts.RegistryPath))
				registry, err = contracts.ReadRegistry(data)
				if err != nil || len(registry.Records) != 1 || len(registry.Selectors[0].Snapshot) != 2 {
					t.Fatal("new base joined legacy snapshot")
				}
			}

		})
	}
}

func TestComponentAmbiguousMigrationRequiresExactResolution(t *testing.T) {
	source, legacy := os.Getenv("MEMORY_BANK_COMPONENT_SOURCE"), os.Getenv("MEMORY_BANK_LEGACY_SOURCE")
	if source == "" || legacy == "" {
		t.Skip("set both source fixtures")
	}
	root := t.TempDir()
	if _, err := Init(Options{RepoRoot: root, SourceRoot: legacy, TemplateVersion: "legacy", SourceRef: SupportedLegacySourceRefs()[0]}); err != nil {
		t.Fatal(err)
	}
	p := "memory-bank/features/FT-444/proposal.md"
	write(t, root, p, "---\nstatus: draft\ndoc_kind: feature\ndelivery_status: planned\n---\n# Feature\n")
	write(t, root, "memory-bank/features/FT-444/README.md", "---\nstatus: draft\ndoc_function: index\npurpose: Navigate package.\n---\n# Feature\n\n- [Proposal](proposal.md) — primary document.\n")
	index := filepath.Join(root, "memory-bank/features/README.md")
	b, _ := os.ReadFile(index)
	write(t, root, "memory-bank/features/README.md", string(b)+"\n- [Feature](FT-444/README.md) — feature package.\n")
	o := Options{RepoRoot: root, SourceRoot: source, SourceRef: strings.Repeat("b", 40), TemplateVersion: "components", MigrateComponents: true, DryRun: true, verifySource: func(string, string) error { return nil }}
	before := treeSnapshot(t, root)
	for _, resolution := range []string{
		`{"schema_version":1,"documents":[],"ownership":{"memory-bank/features/README.md":"keep-local"}}`,
		`{"schema_version":1,"documents":[{"path":"memory-bank/features/FT-444/proposal.md","type":"feature","contract_id":"legacy/f1f04de/feature/v1","evidence":[]}],"ownership":{"memory-bank/features/README.md":"keep-local"}}`,
		`{"schema_version":1,"documents":[{"path":"memory-bank/features/FT-444/proposal.md","type":"feature","contract_id":"feature/v1","evidence":["review:classification"]}],"ownership":{"memory-bank/features/README.md":"keep-local"}}`,
	} {
		o.MigrationResolution = []byte(resolution)
		if _, err := Update(o); err == nil {
			t.Fatal("incomplete or weakening resolution accepted")
		}
		if !reflect.DeepEqual(before, treeSnapshot(t, root)) {
			t.Fatal("failed preview mutated tree")
		}
	}
	o.MigrationResolution = []byte(`{"schema_version":1,"documents":[{"path":"memory-bank/features/FT-444/proposal.md","type":"feature","contract_id":"legacy/f1f04de/feature/v1","evidence":["review:classification"]}],"ownership":{"memory-bank/features/README.md":"keep-local"}}`)
	preview, err := Update(o)
	if err != nil {
		t.Fatal(err)
	}
	o.DryRun = false
	o.MigrationPlanDigest = preview.MigrationPlanDigest
	if r, err := Update(o); err != nil || !r.Applied {
		t.Fatalf("resolved migration: %+v %v", r, err)
	}
	b, _ = os.ReadFile(filepath.Join(root, contracts.RegistryPath))
	registry, err := contracts.ReadRegistry(b)
	if err != nil || len(registry.Selectors) != 1 || registry.Selectors[0].Snapshot[0].Path != p {
		t.Fatalf("selector: %+v %v", registry, err)
	}
}
