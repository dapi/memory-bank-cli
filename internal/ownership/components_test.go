package ownership

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dapi/memory-bank-cli/internal/contracts"
)

// Producer/consumer integration uses the actual candidate payload, avoiding a
// miniature manifest that cannot expose navigation/selection incompatibilities.
func TestComponentPayloadMatrix(t *testing.T) {
	source := os.Getenv("MEMORY_BANK_COMPONENT_SOURCE")
	if source == "" {
		t.Skip("set MEMORY_BANK_COMPONENT_SOURCE to the template checkout")
	}
	for _, preset := range []string{"core", "docs", "full", "legacy"} {
		t.Run(preset, func(t *testing.T) {
			root := t.TempDir()
			o := Options{RepoRoot: root, SourceRoot: source, TemplateVersion: "candidate", SourceRef: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Preset: preset, verifySource: func(string, string) error { return nil }}
			r, err := Init(o)
			if err != nil || !r.Applied || r.ConflictCount != 0 {
				t.Fatalf("init: %+v %v", r, err)
			}
			lock, _, err := ReadLock(root)
			if err != nil {
				t.Fatal(err)
			}
			if lock.SchemaVersion != 2 || lock.Installation.Preset != preset {
				t.Fatal(lock)
			}
			before, err := os.ReadFile(filepath.Join(root, LockFileName))
			if err != nil {
				t.Fatal(err)
			}
			o.Preset = ""
			r, err = Update(o)
			if err != nil || r.Applied || r.ConflictCount != 0 {
				t.Fatalf("repeat: %+v %v", r, err)
			}
			after, _ := os.ReadFile(filepath.Join(root, LockFileName))
			if string(before) != string(after) {
				t.Fatal("no-op changed lock")
			}
			handled, findings, nav, err := ValidateComponents(root, "")
			if !handled || err != nil || len(findings) != 0 || nav.ExitCode != 0 {
				t.Fatalf("audit: %v %+v %+v %v", handled, findings, nav.Errors, err)
			}
			if preset == "docs" {
				o.Preset = "full"
				r, err = Update(o)
				if err != nil || !r.Applied {
					t.Fatalf("upgrade: %+v %v", r, err)
				}
				if _, err = os.Stat(filepath.Join(root, contracts.RegistryPath)); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestComponentDocumentsLifecycle(t *testing.T) {
	source := os.Getenv("MEMORY_BANK_COMPONENT_SOURCE")
	if source == "" {
		t.Skip("set MEMORY_BANK_COMPONENT_SOURCE")
	}
	root := t.TempDir()
	o := Options{RepoRoot: root, SourceRoot: source, TemplateVersion: "candidate", SourceRef: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Preset: "legacy", verifySource: func(string, string) error { return nil }}
	if r, err := Init(o); err != nil || !r.Applied {
		t.Fatalf("init: %+v %v", r, err)
	}
	p := "memory-bank/features/FT-123/brief.md"
	d := DocumentOptions{RepoRoot: root, Operation: "create", Type: "feature", Path: p}
	if r, err := DocumentOperation(d); err != nil || !r.Applied {
		t.Fatalf("base create: %+v %v", r, err)
	}
	b, _ := os.ReadFile(filepath.Join(root, p))
	parsed, err := contracts.ParseDocument(p, b)
	if err != nil || parsed.Has("document_id") || parsed.Has("flow_contract") {
		t.Fatalf("implicit adoption: %+v %v", parsed, err)
	}
	b = []byte(strings.Replace(string(b), "status: draft\n", "status: draft\ndelivery_status: planned\n", 1))
	if err = os.WriteFile(filepath.Join(root, p), b, 0644); err != nil {
		t.Fatal(err)
	}
	d.Operation = "adopt"
	d.LegacyFlow = true
	if r, err := DocumentOperation(d); err != nil || !r.Applied {
		t.Fatalf("adopt: %+v %v", r, err)
	}
	if r, err := DocumentOperation(d); err != nil || r.Applied {
		t.Fatalf("repeat adopt: %+v %v", r, err)
	}
	b, _ = os.ReadFile(filepath.Join(root, p))
	parsed, _ = contracts.ParseDocument(p, b)
	move := DocumentOptions{RepoRoot: root, Operation: "move", Path: p, To: "memory-bank/features/FT-123/proposal.md", ID: parsed.String("document_id")}
	if r, err := DocumentOperation(move); err != nil || !r.Applied {
		t.Fatalf("move: %+v %v", r, err)
	}
	moved, _ := os.ReadFile(filepath.Join(root, move.To))
	if string(moved) != string(b) {
		t.Fatal("move rewrote document")
	}
	if r, err := DocumentOperation(move); err != nil || r.Applied {
		t.Fatalf("repeat move: %+v %v", r, err)
	}
	move.ID = "doc-" + strings.Repeat("0", 64)
	if _, err := DocumentOperation(move); err == nil {
		t.Fatal("wrong retry identity accepted")
	}
	if handled, findings, nav, err := ValidateComponents(root, ""); !handled || err != nil || len(findings) > 0 || nav.ExitCode != 0 {
		t.Fatalf("audit: %v %+v %+v %v", handled, findings, nav.Errors, err)
	}
}
