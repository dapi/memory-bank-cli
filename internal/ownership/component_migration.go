package ownership

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/dapi/memory-bank-cli/internal/contracts"
)

type documentResolution struct {
	Path       string   `json:"path"`
	Type       string   `json:"type"`
	ContractID string   `json:"contract_id"`
	Evidence   []string `json:"evidence"`
}
type migrationResolution struct {
	SchemaVersion int                  `json:"schema_version"`
	Documents     []documentResolution `json:"documents"`
	Ownership     map[string]string    `json:"ownership"`
}
type proposedState struct {
	Exists      bool   `json:"exists"`
	Digest      string `json:"digest"`
	Mode        string `json:"mode"`
	Permissions string `json:"permissions"`
	DigestKind  string `json:"digest_kind"`
}
type writeIntent struct {
	Path      string        `json:"path"`
	Action    Action        `json:"action"`
	Ownership Class         `json:"ownership"`
	Reason    string        `json:"reason"`
	Before    observation   `json:"before"`
	After     proposedState `json:"after"`
}
type componentMigration struct {
	SourceRef        string                    `json:"source_ref"`
	OldLockDigest    string                    `json:"old_lock_digest"`
	ResolutionDigest string                    `json:"resolution_digest"`
	Observed         map[string]observation    `json:"observed"`
	Directories      map[string]directoryState `json:"directories"`
	Changes          []writeIntent             `json:"changes"`
	Installation     contracts.Installation    `json:"installation"`
	NewTemplate      Template                  `json:"new_template"`
	Semantics        string                    `json:"semantics"`
}
type legacyMigration struct {
	registry         contracts.Registry
	documents        map[string][]byte
	findings         []contracts.Finding
	resolutionDigest string
	resolution       migrationResolution
}

func readMigrationResolution(data []byte) (migrationResolution, string, error) {
	r := migrationResolution{SchemaVersion: 1, Documents: []documentResolution{}, Ownership: map[string]string{}}
	if len(data) == 0 {
		return r, digest([]byte("null")), nil
	}
	if err := contracts.Decode(data, &r); err != nil {
		return r, "", err
	}
	if r.SchemaVersion != 1 {
		return r, "", errors.New("unknown migration resolution schema")
	}
	sort.Slice(r.Documents, func(i, j int) bool { return r.Documents[i].Path < r.Documents[j].Path })
	for i := range r.Documents {
		d := &r.Documents[i]
		if !contracts.DocumentPath(d.Path) || !contracts.ValidContractID(d.ContractID) || d.Type == "" || (i > 0 && r.Documents[i-1].Path == d.Path) {
			return r, "", errors.New("invalid or duplicate document resolution")
		}
		set := map[string]bool{}
		for _, e := range d.Evidence {
			if strings.TrimSpace(e) == "" || strings.ContainsAny(e, "\r\n\x00") {
				return r, "", errors.New("invalid resolution evidence")
			}
			set[e] = true
		}
		d.Evidence = contracts.Keys(set)
		if len(d.Evidence) == 0 {
			return r, "", errors.New("classification resolution requires evidence")
		}
	}
	for p, a := range r.Ownership {
		if !contracts.ValidPath(p) || (a != "keep-local" && a != "take-upstream") {
			return r, "", errors.New("invalid ownership resolution")
		}
	}
	b, err := contracts.Canonical(r)
	return r, digest(b), err
}

var canonicalLegacyName = map[string]*regexp.Regexp{
	"adr": regexp.MustCompile(`^ADR-[^/]+\.md$`), "prd": regexp.MustCompile(`^PRD-[^/]+\.md$`), "use_case": regexp.MustCompile(`^UC-[^/]+\.md$`),
}

func legacyCandidate(p string, d contracts.Document, docs map[string]contracts.Document) (typ string, ambiguous bool, err error) {
	parts := strings.Split(p, "/")
	if len(parts) < 3 {
		return "", false, nil
	}
	section := map[string]string{"features": "feature", "research": "research", "epics": "epic", "adr": "adr", "prd": "prd", "use-cases": "use_case"}[parts[1]]
	base := path.Base(p)
	kind := d.String("doc_kind")
	knownKind := contracts.Contains([]string{"feature", "research", "epic", "adr", "prd", "use_case"}, kind)
	if d.Has("document_id") || d.Has("flow_contract") {
		return "", false, fmt.Errorf("legacy adoption projection requires explicit repair: %s", p)
	}
	companions := map[string][]string{"feature": {"design.md", "implementation-plan.md"}, "research": {"plan.md", "evidence.md", "synthesis.md", "decision.md"}, "epic": {"roadmap.md", "decision-log.md", "risks.md", "subissues.md"}}
	packageDocument := len(parts) >= 4 && (section == "feature" || section == "research" || section == "epic")
	if packageDocument && contracts.Contains(companions[section], base) {
		if d.Has("delivery_status") || d.Has("flow_contract") {
			return "", false, fmt.Errorf("wrong-owner flow fields: %s", p)
		}
		return "", false, nil
	}
	if base == "README.md" {
		if section == "epic" && packageDocument {
			if _, hasCharter := docs[path.Dir(p)+"/charter.md"]; !hasCharter {
				if d.String("doc_function") != "canonical" {
					return "", false, fmt.Errorf("epic package has no canonical primary: %s", p)
				}
				return "epic", kind != "epic", nil
			}
		}
		return "", false, nil
	}
	if section == "" {
		if knownKind {
			return kind, true, nil
		}
		if d.Has("delivery_status") || d.Has("research_status") {
			return "", false, fmt.Errorf("unresolved flow-bearing document %s", p)
		}
		return "", false, nil
	}
	canonical := false
	if pattern := canonicalLegacyName[section]; pattern != nil {
		canonical = len(parts) == 3 && pattern.MatchString(base)
	}
	if packageDocument {
		canonical = (section == "feature" && strings.HasPrefix(parts[2], "FT-") && base == "brief.md") || (section == "research" && base == "brief.md") || (section == "epic" && base == "charter.md")
	}
	if knownKind && kind != section {
		return "", false, fmt.Errorf("document kind contradicts legacy section: %s", p)
	}
	if d.Has("document_type") && d.String("document_type") != section {
		return "", false, fmt.Errorf("document type contradicts legacy section: %s", p)
	}
	if d.Has("delivery_status") && section != "feature" {
		return "", false, fmt.Errorf("wrong-owner lifecycle fields: %s", p)
	}
	return section, !canonical || kind != section, nil
}
func prepareLegacyMigration(options Options, old Lock, tree componentTree, m contracts.Manifest, source map[string]payload, lockDigest string) (legacyMigration, error) {
	migration := legacyMigration{registry: contracts.EmptyRegistry(), documents: map[string][]byte{}}
	if !options.MigrateComponents {
		return migration, errors.New("legacy installation requires --migrate-components and a reviewed preview digest")
	}
	if options.Preset != "" && options.Preset != "legacy" {
		return migration, errors.New("legacy migration only supports preset legacy")
	}
	compat, ok := m.LegacySources[old.Template.SourceRef]
	if !ok {
		return migration, errors.New("unsupported legacy source; keep using its pinned template")
	}
	r, rd, err := readMigrationResolution(options.MigrationResolution)
	if err != nil {
		return migration, err
	}
	migration.resolution = r
	migration.resolutionDigest = rd
	resolutions := map[string]documentResolution{}
	for _, d := range r.Documents {
		resolutions[d.Path] = d
	}
	inventory := map[string][]byte{}
	for p, f := range source {
		inventory[p] = f.data
	}
	catalog, err := contracts.LoadCatalog(m, inventory, nil)
	if err != nil {
		return migration, err
	}
	docs := map[string]contracts.Document{}
	for _, p := range contracts.Keys(tree.files) {
		if !contracts.DocumentPath(p) {
			continue
		}
		if f, ok := old.Files[p]; ok && (f.Ownership == Managed || f.Ownership == Generated) {
			continue
		}
		d, e := contracts.ParseDocument(p, tree.files[p])
		if e != nil {
			return migration, fmt.Errorf("repair unsupported legacy YAML in %s: %w", p, e)
		}
		docs[p] = d
	}
	selectors := map[string]contracts.Selector{}
	primaries := map[string]bool{}
	packages := map[string]bool{}
	for _, p := range contracts.Keys(docs) {
		d := docs[p]
		parts := strings.Split(p, "/")
		if len(parts) >= 4 && contracts.Contains([]string{"features", "research", "epics"}, parts[1]) {
			packages[strings.Join(parts[:3], "/")] = true
		}
		typ, ambiguous, e := legacyCandidate(p, d, docs)
		if e != nil {
			return migration, e
		}
		resolution, resolved := resolutions[p]
		if typ == "" {
			if resolved {
				return migration, fmt.Errorf("resolution targets a non-candidate: %s", p)
			}
			continue
		}
		contract := compat.Contracts[typ]
		if contract == "" {
			return migration, fmt.Errorf("legacy type has no compatibility contract: %s", typ)
		}
		if ambiguous && !resolved {
			return migration, fmt.Errorf("ambiguous legacy document requires resolution: %s", p)
		}
		if resolved {
			if resolution.Type != typ || resolution.ContractID != contract {
				return migration, fmt.Errorf("incompatible classification resolution: %s", p)
			}
			delete(resolutions, p)
		}
		id := contracts.MigrationID(lockDigest, p, digest(d.Raw))
		root, e := contracts.ContextRoot(p, typ)
		if e != nil {
			return migration, e
		}
		primaries[root] = true
		identity := contracts.Identity{ID: id, Path: p, Type: typ, ContextRoot: root}
		bundle := m.Contracts[contract]
		sid := contracts.SelectorID(old.Template.SourceRef, typ, contract, bundle.Digest)
		sel, ok := selectors[sid]
		if !ok {
			sel = contracts.Selector{ID: sid, SourceRef: old.Template.SourceRef, Type: typ, ContractID: contract, BundleDigest: bundle.Digest, Snapshot: []contracts.Identity{}, Exclusions: []string{}}
		}
		sel.Snapshot = append(sel.Snapshot, identity)
		selectors[sid] = sel
		evidence := []string{}
		if resolved {
			evidence = resolution.Evidence
		}
		migration.registry.History = append(migration.registry.History, contracts.Event{Operation: "migrate", DocumentID: id, FromPath: p, ToPath: p, ToContract: contract, Evidence: evidence})
		projected, e := contracts.Project(d.Raw, map[string]string{"document_type": typ, "document_id": id})
		if e != nil {
			return migration, e
		}
		migration.documents[p] = projected
		migration.findings = append(migration.findings, contracts.ValidateBundle(d, catalog.Bundles[contract], identity, docs)...)
	}
	if len(resolutions) > 0 {
		return migration, errors.New("resolution refers to unknown document")
	}
	for p := range packages {
		if !primaries[p] {
			return migration, fmt.Errorf("legacy package has no primary identity: %s", p)
		}
	}
	for _, id := range contracts.Keys(selectors) {
		sel := selectors[id]
		sort.Slice(sel.Snapshot, func(i, j int) bool { return sel.Snapshot[i].ID < sel.Snapshot[j].ID })
		migration.registry.Selectors = append(migration.registry.Selectors, sel)
	}
	sort.Slice(migration.registry.History, func(i, j int) bool {
		return migration.registry.History[i].DocumentID < migration.registry.History[j].DocumentID
	})
	contracts.SortFindings(migration.findings)
	return migration, nil
}

func (p *componentPlan) describeMigration(repo pinnedRepo, old Lock, lockDigest string, legacy legacyMigration) error {
	changes := map[string]writeIntent{}
	for _, d := range p.report.Decisions {
		before := p.tree.observed[d.Path]
		after := proposedState{before.Exists, before.Digest, before.Mode, before.Permissions, "bytes/v1"}
		changes[d.Path] = writeIntent{d.Path, d.Action, d.Ownership, d.Reason, before, after}
	}
	for _, item := range p.mutations {
		before := p.tree.observed[item.decision.Path]
		after := proposedState{DigestKind: "bytes/v1"}
		if item.decision.Action != Delete {
			mode := item.mode
			if mode == 0 && !item.modeSet {
				mode = 0644
			}
			after = proposedState{true, digest(item.data), gitMode(mode), fmt.Sprintf("%04o", mode.Perm()), "bytes/v1"}
		}
		if item.decision.Path == LockFileName && after.Exists {
			var value map[string]any
			if err := json.Unmarshal(item.data, &value); err != nil {
				return err
			}
			value["last_update"].(map[string]any)["at"] = "<execution-time>"
			b, err := contracts.Canonical(value)
			if err != nil {
				return err
			}
			after.Digest = digest(b)
			after.DigestKind = "lock-projection/v1"
		}
		changes[item.decision.Path] = writeIntent{item.decision.Path, item.decision.Action, item.decision.Ownership, item.decision.Reason, before, after}
	}
	directories, err := componentPlanDirectories(repo, p.tree.observed, p.mutations)
	if err != nil {
		return err
	}
	p.directories = directories
	preview := componentMigration{SourceRef: old.Template.SourceRef, OldLockDigest: lockDigest, ResolutionDigest: legacy.resolutionDigest, Observed: p.tree.observed, Directories: directories, Changes: []writeIntent{}, Installation: *p.next.Installation, NewTemplate: p.next.Template, Semantics: "blanket-to-explicit-adoption/v1"}
	for _, name := range contracts.Keys(changes) {
		intent := changes[name]
		if err = validateWriteIntent(intent); err != nil {
			return err
		}
		preview.Changes = append(preview.Changes, intent)
	}
	b, err := contracts.Canonical(preview)
	if err != nil {
		return err
	}
	registry := p.future()[contracts.RegistryPath]
	p.report.Migration = &preview
	p.report.MigrationPlanDigest = "sha256:" + contracts.FramedID("memory-bank/migration-plan/v1", string(b), string(registry))
	return nil
}
func validateWriteIntent(i writeIntent) error {
	switch i.Action {
	case Create:
		if i.Before.Exists || !i.After.Exists {
			return errors.New("invalid create intent")
		}
	case UpdateFile:
		if !i.Before.Exists || !i.After.Exists {
			return errors.New("invalid update intent")
		}
	case Delete:
		if !i.Before.Exists || i.After.Exists {
			return errors.New("invalid delete intent")
		}
	case Preserve:
		if i.After.DigestKind != "bytes/v1" || !reflect.DeepEqual(i.Before, observation{i.After.Exists, i.After.Digest, i.After.Mode, i.After.Permissions}) {
			return errors.New("invalid preserve intent")
		}
	default:
		return errors.New("unresolved migration intent")
	}
	return nil
}
func componentPlanDirectories(repo pinnedRepo, observed map[string]observation, mutations []mutation) (map[string]directoryState, error) {
	directories := map[string]directoryState{}
	for p := range observed {
		for parent := path.Dir(p); parent != "."; parent = path.Dir(parent) {
			if _, ok := directories[parent]; ok {
				continue
			}
			if err := checkComponentAncestors(repo, parent+"/plan-check"); err != nil {
				return nil, err
			}
			info, err := os.Lstat(filepath.Join(repo.root, filepath.FromSlash(parent)))
			d := directoryState{}
			if err == nil {
				if !info.IsDir() || info.Mode()&(os.ModeSymlink|os.ModeSetuid|os.ModeSetgid|os.ModeSticky) != 0 {
					return nil, errors.New("unsafe planned directory")
				}
				mode := fmt.Sprintf("%04o", info.Mode().Perm())
				d = directoryState{true, mode, true, mode}
			} else if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
			directories[parent] = d
		}
	}
	for _, item := range mutations {
		if item.topology != nil {
			return nil, errors.New("component topology migration unsupported")
		}
		if item.decision.Action == Delete {
			continue
		}
		for parent := path.Dir(item.decision.Path); parent != "."; parent = path.Dir(parent) {
			d := directories[parent]
			if !d.BeforeExists {
				d.AfterExists = true
				d.AfterMode = "0755"
				directories[parent] = d
			}
		}
	}
	return directories, nil
}

// A source ownership change must not hide existing managed drift. Resolutions
// are accepted only for concrete reported conflicts and cannot adapt rule assets.
func resolveLegacyOwnership(repo pinnedRepo, source map[string]payload, old Lock, tree componentTree, r migrationResolution, options Options) (Lock, Options, error) {
	planning := old
	planning.Files = map[string]File{}
	for p, f := range old.Files {
		planning.Files[p] = f
	}
	_, decisions, _, err := buildPlan(repo, source, old, true, nil, nil, false)
	if err != nil {
		return old, options, err
	}
	conflicts := map[string]Decision{}
	for _, d := range decisions {
		if d.Action == Conflict {
			conflicts[d.Path] = d
		}
	}
	drift := map[string]bool{}
	for p, f := range old.Files {
		if f.Ownership != Managed && f.Ownership != Generated {
			continue
		}
		o := tree.observed[p]
		if !o.Exists || o.Digest != f.PayloadDigest || (f.PayloadMode != "" && o.Mode != f.PayloadMode) {
			drift[p] = true
			conflicts[p] = Decision{Path: p, Action: Conflict, Ownership: f.Ownership}
		}
	}
	options.UserOwnedResolutions = map[string]bool{}
	options.AdaptedResolutions = map[string]AdaptedResolution{}
	for p := range drift {
		if _, resolved := r.Ownership[p]; !resolved {
			return old, options, fmt.Errorf("managed drift requires ownership resolution: %s", p)
		}
	}
	for p, action := range r.Ownership {
		d, ok := conflicts[p]
		if !ok {
			return old, options, fmt.Errorf("ownership resolution is not a reported conflict: %s", p)
		}
		incoming, present := source[p]
		if drift[p] {
			if !tree.observed[p].Exists {
				return old, options, fmt.Errorf("restore missing managed input before migration: %s", p)
			}
			if action == "keep-local" && (!present || incoming.class != UserOwned && p != "memory-bank/README.md") {
				return old, options, fmt.Errorf("cannot keep local bytes at managed component path: %s", p)
			}
			if action == "take-upstream" && !present {
				return old, options, fmt.Errorf("cannot discard removed managed input without explicit deletion plan: %s", p)
			}
			planning.Files[p] = File{Ownership: UserOwned}
			options.UserOwnedResolutions[p] = action == "take-upstream"
		} else if d.Ownership == UserOwned {
			options.UserOwnedResolutions[p] = action == "take-upstream"
		} else if d.Ownership == Adapted {
			options.AdaptedResolutions[p] = AdaptedResolution{Action: action}
		} else {
			return old, options, fmt.Errorf("unsupported ownership resolution: %s", p)
		}
	}
	return planning, options, nil
}
