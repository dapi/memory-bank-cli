package ownership

import (
	"errors"
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

func TestComponentResolutionPlanBindsPermissions(t *testing.T) {
	source := os.Getenv("MEMORY_BANK_COMPONENT_SOURCE")
	if source == "" {
		t.Skip("set MEMORY_BANK_COMPONENT_SOURCE")
	}
	root := t.TempDir()
	o := Options{RepoRoot: root, SourceRoot: source, TemplateVersion: "candidate", SourceRef: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Preset: "docs", verifySource: func(string, string) error { return nil }}
	if r, err := Init(o); err != nil || !r.Applied {
		t.Fatalf("init: %+v %v", r, err)
	}
	p := "memory-bank/features/FT-124/brief.md"
	if r, err := DocumentOperation(DocumentOptions{RepoRoot: root, Operation: "create", Type: "feature", Path: p}); err != nil || !r.Applied {
		t.Fatalf("create: %+v %v", r, err)
	}
	o.Preset = "full"
	plan, err := PlanPull(o)
	if err != nil || plan.FormatVersion != 2 || !contracts.ValidDigest(plan.PreconditionDigest) {
		t.Fatalf("plan: %+v %v", plan, err)
	}
	if err = os.Chmod(filepath.Join(root, p), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = ApplyResolutionPlan(o, plan); err == nil {
		t.Fatal("untracked document permission drift accepted")
	}
	if err = os.Chmod(filepath.Join(root, p), 0644); err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(filepath.Join(root, "memory-bank/features"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = ApplyResolutionPlan(o, plan); err == nil {
		t.Fatal("directory permission drift accepted")
	}
	if err = os.Chmod(filepath.Join(root, "memory-bank/features"), 0755); err != nil {
		t.Fatal(err)
	}
	if r, err := ApplyResolutionPlan(o, plan); err != nil || !r.Applied {
		t.Fatalf("apply: %+v %v", r, err)
	}
	b, _ := os.ReadFile(filepath.Join(root, p))
	d, _ := contracts.ParseDocument(p, b)
	if d.Has("document_id") {
		t.Fatal("upgrade implicitly adopted existing document")
	}
	plan.FormatVersion = 1
	if _, err = ApplyResolutionPlan(o, plan); err == nil {
		t.Fatal("legacy format applied component source")
	}
}

func TestComponentAtomicLegacyFlowCreation(t *testing.T) {
	source := os.Getenv("MEMORY_BANK_COMPONENT_SOURCE")
	if source == "" {
		t.Skip("set MEMORY_BANK_COMPONENT_SOURCE")
	}
	root := t.TempDir()
	o := Options{RepoRoot: root, SourceRoot: source, TemplateVersion: "candidate", SourceRef: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Preset: "legacy", verifySource: func(string, string) error { return nil }}
	if r, err := Init(o); err != nil || !r.Applied {
		t.Fatalf("init: %+v %v", r, err)
	}
	draft := "---\nstatus: draft\ndoc_kind: feature\ndelivery_status: planned\n---\n\n# Feature\n"
	write(t, root, "drafts/feature.md", draft)
	d := DocumentOptions{RepoRoot: root, Operation: "create", Type: "feature", Path: "memory-bank/features/FT-125/brief.md", From: "drafts/feature.md", LegacyFlow: true}
	if r, err := DocumentOperation(d); err != nil || !r.Applied {
		t.Fatalf("flow create: %+v %v", r, err)
	}
	input, _ := os.ReadFile(filepath.Join(root, d.From))
	if string(input) != draft {
		t.Fatal("draft changed")
	}
	b, _ := os.ReadFile(filepath.Join(root, contracts.RegistryPath))
	registry, err := contracts.ReadRegistry(b)
	if err != nil || len(registry.Records) != 1 || len(registry.Selectors) != 0 || registry.History[0].Operation != "create" {
		t.Fatalf("atomic binding: %+v %v", registry, err)
	}
}

func TestComponentAdapterMatrix(t *testing.T) {
	source := os.Getenv("MEMORY_BANK_COMPONENT_SOURCE")
	if source == "" {
		t.Skip("set MEMORY_BANK_COMPONENT_SOURCE")
	}
	for _, adapter := range []string{"bootstrap", "codex", "start-issue", "symphony"} {
		t.Run(adapter, func(t *testing.T) {
			root := t.TempDir()
			o := Options{RepoRoot: root, SourceRoot: source, TemplateVersion: "candidate", SourceRef: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Preset: "core", Adapters: []string{adapter}, verifySource: func(string, string) error { return nil }}
			if r, err := Init(o); err != nil || !r.Applied {
				t.Fatalf("adapter init: %+v %v", r, err)
			}
			lock, _, err := ReadLock(root)
			if err != nil || !lock.Installation.Has(adapter) || !lock.Installation.Has("flows") || len(lock.Installation.Adapters) != 1 {
				t.Fatalf("selection: %+v %v", lock, err)
			}
			o.Preset = ""
			o.Adapters = nil
			if r, err := Update(o); err != nil || r.Applied {
				t.Fatalf("adapter repeat: %+v %v", r, err)
			}
		})
	}
}

func TestComponentCommittedRecoveryRetriesCleanup(t *testing.T) {
	source := os.Getenv("MEMORY_BANK_COMPONENT_SOURCE")
	if source == "" {
		t.Skip("set MEMORY_BANK_COMPONENT_SOURCE")
	}
	root := t.TempDir()
	o := Options{RepoRoot: root, SourceRoot: source, TemplateVersion: "candidate", SourceRef: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Preset: "core", verifySource: func(string, string) error { return nil }}
	if r, err := Init(o); err != nil || !r.Applied {
		t.Fatalf("init: %+v %v", r, err)
	}
	repo, err := pinRepoRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	old, _, lockDigest, err := readLockSnapshot(repo)
	if err != nil {
		t.Fatal(err)
	}
	pinned, err := pinSourceRoot(source)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := readSource(pinned)
	if err != nil {
		t.Fatal(err)
	}
	o.Preset = "docs"
	p, err := prepareComponents(o, old, true, repo, lockDigest, payload)
	if err != nil {
		t.Fatal(err)
	}
	draftPath := "drafts/input.md"
	write(t, root, draftPath, "prepared draft")
	observation, _, err := observeComponent(repo, draftPath)
	if err != nil {
		t.Fatal(err)
	}
	p.tree.observed[draftPath] = observation
	if _, err = p.capturePreconditions(repo); err != nil {
		t.Fatal(err)
	}
	o.componentDraftInput = draftPath
	o.componentTransaction = true
	o.componentObservations = p.tree.observed
	o.componentDirectories = p.directories
	ops := osTransactionOps
	ops.removeAll = func(string) error { return errors.New("cleanup blocked") }
	err = applyAtomicallyPinnedWithOps(o, p.mutations, repo, ops)
	var committed *committedError
	if !errors.As(err, &committed) {
		t.Fatalf("expected committed cleanup error: %v", err)
	}
	staging, _ := filepath.Glob(filepath.Join(root, ".memory-bank-update-*"))
	if len(staging) != 1 {
		t.Fatal(staging)
	}
	snapshot, err := os.ReadFile(filepath.Join(staging[0], "inputs/000000"))
	if err != nil || string(snapshot) != "prepared draft" {
		t.Fatalf("draft snapshot: %s %v", snapshot, err)
	}
	if err = os.Remove(filepath.Join(root, draftPath)); err != nil {
		t.Fatal(err)
	}
	if err = checkComponentRecovery(repo, true); err == nil {
		t.Fatal("deleted read input allowed cleanup")
	}
	write(t, root, draftPath, string(snapshot))
	if err = checkComponentRecovery(repo, true); err != nil {
		t.Fatal(err)
	}
	assertNoTransactionStaging(t, root)
	o.Preset = ""
	if r, err := Update(o); err != nil || r.Applied {
		t.Fatalf("reentry: %+v %v", r, err)
	}
}
