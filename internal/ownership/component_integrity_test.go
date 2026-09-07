package ownership

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/dapi/memory-bank-cli/internal/agentinstructions"
	"github.com/dapi/memory-bank-cli/internal/contracts"
)

func TestComponentIntegrityRejectsTamperingWithoutMutation(t *testing.T) {
	source := os.Getenv("MEMORY_BANK_COMPONENT_SOURCE")
	if source == "" {
		t.Skip("set MEMORY_BANK_COMPONENT_SOURCE")
	}
	root := t.TempDir()
	options := Options{RepoRoot: root, SourceRoot: source, TemplateVersion: "candidate", SourceRef: strings.Repeat("a", 40), Preset: "legacy", verifySource: func(string, string) error { return nil }}
	if _, err := Init(options); err != nil {
		t.Fatal(err)
	}
	draft := "---\nstatus: draft\ndoc_kind: feature\ndelivery_status: planned\n---\n# Feature\n"
	write(t, root, "drafts/input.md", draft)
	document := "memory-bank/features/FT-901/brief.md"
	if _, err := DocumentOperation(DocumentOptions{RepoRoot: root, Operation: "create", Type: "feature", Path: document, From: "drafts/input.md", LegacyFlow: true}); err != nil {
		t.Fatal(err)
	}
	lock, _, err := ReadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := ValidateComponents(root, "CLAUDE.md"); err == nil {
		t.Fatal("alternate component agent target accepted")
	}
	manifestBytes, _ := os.ReadFile(filepath.Join(root, contracts.ManifestPath))
	manifest, err := contracts.ReadManifest(manifestBytes, nil)
	if err != nil {
		t.Fatal(err)
	}
	bundle := manifest.Contracts["feature/v1"].Path
	cases := []struct {
		name, path string
		change     func([]byte) []byte
	}{
		{"missing registry", contracts.RegistryPath, func([]byte) []byte { return nil }},
		{"manifest", contracts.ManifestPath, func(b []byte) []byte { return append(b, ' ') }},
		{"bundle", bundle, func(b []byte) []byte { return append(b, ' ') }},
		{"orphan identity", document, func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `document_id: "doc-`, `document_id: "doc-a`, 1))
		}},
		{"type contradiction", document, func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `document_type: "feature"`, `document_type: "adr"`, 1))
		}},
		{"README marker", "memory-bank/README.md", func(b []byte) []byte {
			return []byte(strings.Replace(string(b), "MEMORY BANK START", "BROKEN START", 1))
		}},
		{"AGENTS marker", "AGENTS.md", func(b []byte) []byte {
			return []byte(strings.Replace(string(b), "BLOCK VERSION: 4", "BLOCK VERSION: 9", 1))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			full := filepath.Join(root, tc.path)
			original, err := os.ReadFile(full)
			if err != nil {
				t.Fatal(err)
			}
			changed := tc.change(original)
			if changed != nil && string(changed) == string(original) {
				t.Fatal("fixture did not change bytes")
			}
			if changed == nil {
				err = os.Remove(full)
			} else {
				err = os.WriteFile(full, changed, 0644)
			}
			if err != nil {
				t.Fatal(err)
			}
			before := treeSnapshot(t, root)
			options.Preset = ""
			if _, err = Update(options); err == nil {
				t.Fatal("pull accepted tampering")
			}
			if _, err = DocumentOperation(DocumentOptions{RepoRoot: root, Operation: "create", Type: "feature", Path: "memory-bank/features/FT-902/brief.md"}); err == nil {
				t.Fatal("document command accepted tampering")
			}
			if !reflect.DeepEqual(before, treeSnapshot(t, root)) {
				t.Fatal("rejection mutated tree")
			}
			if err = os.WriteFile(full, original, 0644); err != nil {
				t.Fatal(err)
			}
		})
	}
	// A historical renderer-1 installation remains readable and upgrades explicitly
	// through pull; an unknown discriminator fails closed.
	for _, historicalVersion := range []int{1, 2} {
		readmePath := filepath.Join(root, "memory-bank/README.md")
		readme, _ := os.ReadFile(readmePath)
		lock, _, _ = ReadLock(root)
		version := historicalVersion
		lock.Installation.RendererVersion = &version
		plan := agentinstructions.BuildPlanWithBlock(readme, contracts.ReadmeBlock(manifest, *lock.Installation))
		if err = os.WriteFile(readmePath, plan.Data, 0644); err != nil {
			t.Fatal(err)
		}
		f := lock.Files["memory-bank/README.md"]
		f.PayloadDigest = digest(plan.Data)
		lock.Files["memory-bank/README.md"] = f
		b, err := marshalLock(lock)
		if err != nil {
			t.Fatal(err)
		}
		write(t, root, LockFileName, string(b))
		beforeAudit := treeSnapshot(t, root)
		if handled, findings, nav, err := ValidateComponents(root, ""); !handled || err != nil || len(findings) > 0 || nav.ExitCode != 0 {
			t.Fatalf("v1 audit: %v %v %v", err, findings, nav.Errors)
		}
		if !reflect.DeepEqual(beforeAudit, treeSnapshot(t, root)) {
			t.Fatal("historical audit mutated installation")
		}
		if r, err := Update(options); err != nil || !r.Applied {
			t.Fatalf("v1 upgrade: %+v %v", r, err)
		}
		lock, _, _ = ReadLock(root)
		if *lock.Installation.RendererVersion != 3 {
			t.Fatal("renderer not upgraded")
		}
	}
	version := 99
	lock.Installation.RendererVersion = &version
	b, _ := marshalLock(lock)
	write(t, root, LockFileName, string(b))
	if _, err = Update(options); err == nil {
		t.Fatal("unknown renderer accepted")
	}
}

func TestComponentBaseTypeMatrixRemainsUnadopted(t *testing.T) {
	source := os.Getenv("MEMORY_BANK_COMPONENT_SOURCE")
	if source == "" {
		t.Skip("set MEMORY_BANK_COMPONENT_SOURCE")
	}
	root := t.TempDir()
	if _, err := Init(Options{RepoRoot: root, SourceRoot: source, SourceRef: strings.Repeat("a", 40), TemplateVersion: "candidate", Preset: "full", verifySource: func(string, string) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ typ, path string }{
		{"feature", "memory-bank/features/FT-811/brief.md"},
		{"adr", "memory-bank/adr/ADR-811-decision.md"},
		{"prd", "memory-bank/prd/PRD-811-product.md"},
		{"use_case", "memory-bank/use-cases/UC-811-scenario.md"},
		{"research", "memory-bank/research/RS-811/brief.md"},
		{"epic", "memory-bank/epics/EP-811/charter.md"},
	} {
		t.Run(tc.typ, func(t *testing.T) {
			if r, err := DocumentOperation(DocumentOptions{RepoRoot: root, Operation: "create", Type: tc.typ, Path: tc.path}); err != nil || !r.Applied {
				t.Fatalf("create: %+v %v", r, err)
			}
			b, _ := os.ReadFile(filepath.Join(root, tc.path))
			d, err := contracts.ParseDocument(tc.path, b)
			if err != nil || d.Has("flow_contract") || d.Has("document_id") || d.String("document_type") != tc.typ {
				t.Fatal("base unexpectedly adopted", err)
			}
		})
	}
	b, _ := os.ReadFile(filepath.Join(root, contracts.RegistryPath))
	r, err := contracts.ReadRegistry(b)
	if err != nil || len(r.Records) > 0 || len(r.Selectors) > 0 || len(r.History) > 0 {
		t.Fatal("base creation changed adoption", err)
	}
}
