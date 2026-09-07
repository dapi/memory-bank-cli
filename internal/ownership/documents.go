package ownership

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dapi/memory-bank-cli/internal/contracts"
)

type DocumentOptions struct {
	From           string
	RepoRoot       string
	Operation      string
	Type           string
	Path           string
	To             string
	ID             string
	Contract       string
	LegacyFlow     bool
	Evidence       []string
	DryRun         bool
	BeforeMutation func(Decision) error
}

func DocumentOperation(o DocumentOptions) (Report, error) {
	if !componentHost() {
		return Report{}, errors.New("component mutations require Linux or macOS")
	}
	if !contracts.Contains([]string{"create", "adopt", "transition", "move"}, o.Operation) {
		return Report{}, errors.New("unsupported document operation")
	}
	if !contracts.DocumentPath(o.Path) {
		return Report{}, errors.New("document path must be project Markdown under memory-bank")
	}
	if o.Operation == "move" && (!contracts.DocumentPath(o.To) || o.ID == "") {
		return Report{}, errors.New("move requires --id and a project Markdown --to path")
	}
	if o.From != "" && (o.Operation != "create" || !contracts.ValidPath(o.From) || !strings.HasSuffix(strings.ToLower(o.From), ".md") || o.From == o.Path) {
		return Report{}, errors.New("--from requires create and a different repository-relative Markdown input")
	}
	if o.Operation != "move" && (o.To != "" || o.ID != "") {
		return Report{}, errors.New("--to/--id are only supported by move")
	}
	if o.Operation == "move" && (o.Contract != "" || o.Type != "" || o.LegacyFlow || len(o.Evidence) > 0) {
		return Report{}, errors.New("move only accepts identity and paths")
	}
	if o.LegacyFlow && o.Contract != "" {
		return Report{}, errors.New("--legacy-flow and --contract are mutually exclusive")
	}
	repo, err := pinRepoRoot(o.RepoRoot)
	if err != nil {
		return Report{}, err
	}
	if err = checkComponentRecovery(repo, !o.DryRun); err != nil {
		return Report{}, err
	}
	lock, exists, _, err := readLockSnapshot(repo)
	if err != nil {
		return Report{}, err
	}
	if !exists || lock.SchemaVersion != 2 || lock.Installation == nil || !lock.Installation.Has("documents") {
		return Report{}, errors.New("document operations require a schema-2 Documents installation")
	}
	extra := []string{o.Path}
	if o.To != "" {
		extra = append(extra, o.To)
	}
	readPaths := append([]string{}, extra...)
	if o.From != "" {
		readPaths = append(readPaths, o.From)
	}
	tree, err := readComponentTree(repo, lock, readPaths)
	if err != nil {
		return Report{}, err
	}
	state, err := decodeComponentState(tree.files, lock, "AGENTS.md", true)
	if err != nil {
		return Report{}, err
	}
	if len(state.findings) > 0 {
		return Report{}, fmt.Errorf("installed document gates fail: %v", state.findings)
	}
	for name, f := range lock.Files {
		if f.Ownership == Managed && tree.observed[name].Mode != f.PayloadMode {
			return Report{}, fmt.Errorf("managed mode drift: %s", name)
		}
	}
	for _, p := range extra {
		if f, ok := lock.Files[p]; ok && (f.Ownership == Managed || f.Ownership == Generated) {
			return Report{}, fmt.Errorf("managed/generated document mutation forbidden: %s", p)
		}
	}
	p := componentPlan{tree: tree, next: lock, manifest: state.manifest, report: Report{FormatVersion: 1, DryRun: o.DryRun, Decisions: []Decision{}}}
	installation := *lock.Installation
	p.next.Installation = &installation
	p.next.Files = map[string]File{}
	for name, f := range lock.Files {
		p.next.Files[name] = f
	}
	binding := contracts.Binding{}
	for _, b := range state.bindings {
		if b.Identity.Path == o.Path {
			binding = b
		}
	}
	typ := o.Type
	if typ == "" {
		if d, ok := state.docs[o.Path]; ok {
			typ = d.String("document_type")
		}
	}
	if typ == "" && binding.Identity.ID != "" {
		typ = binding.Identity.Type
	}
	if o.Operation != "move" {
		if _, ok := state.catalog.Types[typ]; !ok {
			return Report{}, errors.New("known --type or existing document_type is required")
		}
	}
	if o.LegacyFlow {
		if installation.LegacySourceRef == "" {
			return Report{}, errors.New("installation has no pinned legacy creation contract")
		}
		o.Contract = state.manifest.LegacySources[installation.LegacySourceRef].Contracts[typ]
	}
	if existing, ok := state.docs[o.Path]; ok && o.Operation != "move" && existing.Has("document_type") && existing.String("document_type") != typ {
		return Report{}, errors.New("operation cannot change an existing document type")
	}
	adopting := o.Contract != "" || o.Operation != "create"
	if adopting && !installation.Has("flows") {
		return Report{}, errors.New("adoption requires installed Flows")
	}
	if o.Operation != "move" && adopting {
		b, ok := state.catalog.Bundles[o.Contract]
		if !ok || b.Type != typ {
			return Report{}, errors.New("installed matching --contract is required")
		}
	}
	evidenceSet := map[string]bool{}
	for _, ref := range o.Evidence {
		evidenceSet[ref] = true
	}
	evidence := contracts.Keys(evidenceSet)
	if !contracts.SortedSet(evidence) {
		return Report{}, errors.New("evidence must contain unique nonempty references")
	}
	for _, ref := range evidence {
		if strings.TrimSpace(ref) != ref || strings.ContainsAny(ref, "\r\n\x00") {
			return Report{}, errors.New("invalid evidence reference")
		}
	}
	registry := state.registry
	// A retry is accepted only with the current complete identity and last move.
	if o.Operation == "move" && !tree.observed[o.Path].Exists {
		b, ok := state.bindings[o.ID]
		if ok && b.Identity.Path == o.To {
			for i := len(registry.History) - 1; i >= 0; i-- {
				last := registry.History[i]
				if last.DocumentID != o.ID {
					continue
				}
				if last.Operation == "move" && last.FromPath == o.Path && last.ToPath == o.To {
					return p.report, nil
				}
				break
			}
		}

		return Report{}, errors.New("move source missing or retry identity/history mismatch")
	}
	raw := tree.files[o.Path]
	switch o.Operation {
	case "create":
		if tree.observed[o.Path].Exists {
			return Report{}, errors.New("document already exists")
		}
		templatePath := state.catalog.Types[typ].Template
		raw, err = contracts.RelocateBaseDocument(state.catalog.Files[templatePath], templatePath, o.Path)
		if err != nil {
			return Report{}, err
		}
		if o.From != "" {
			if !tree.observed[o.From].Exists {
				return Report{}, errors.New("draft input missing")
			}
			draft, e := contracts.ParseDocument(o.From, tree.files[o.From])
			if e != nil {
				return Report{}, e
			}
			if draft.Has("document_id") || draft.Has("flow_contract") || (draft.Has("document_type") && draft.String("document_type") != typ) || (draft.Has("doc_kind") && draft.String("doc_kind") != typ) {
				return Report{}, errors.New("draft has adoption or incompatible type metadata")
			}
			raw, e = contracts.RelocateBaseDocument(draft.Raw, draft.Path, o.Path)
			if e != nil {
				return Report{}, e
			}
		}

	case "adopt":
		if !tree.observed[o.Path].Exists {
			return Report{}, errors.New("document missing")
		}
		if binding.Identity.ID != "" {
			if binding.Identity.Type == typ && binding.ContractID == o.Contract {
				return p.report, nil
			}
			return Report{}, errors.New("already adopted; use transition")
		}
	case "transition":
		if binding.Identity.ID == "" {
			return Report{}, errors.New("transition requires current adoption")
		}
		if typ != binding.Identity.Type {
			return Report{}, errors.New("transition cannot change document type")
		}
		if err = state.catalog.ValidateTransition(binding.ContractID, o.Contract, typ, evidence); err != nil {
			return Report{}, err
		}
		if findings := contracts.ValidateBundle(state.docs[o.Path], state.catalog.Bundles[binding.ContractID], binding.Identity, state.docs); len(findings) > 0 {
			return Report{}, fmt.Errorf("current contract gates fail: %v", findings)
		}
	case "move":
		if binding.Identity.ID != o.ID {
			return Report{}, errors.New("move identity mismatch")
		}
		if tree.observed[o.To].Exists || o.To == o.Path {
			return Report{}, errors.New("move destination exists")
		}
		relocated, e := contracts.RelocateBaseDocument(raw, o.Path, o.To)
		if e != nil {
			return Report{}, e
		}
		if string(relocated) != string(raw) {
			return Report{}, errors.New("move would change relative reference meaning; use location-independent references first")
		}
		if !strings.HasPrefix(o.To, binding.Identity.ContextRoot+"/") {
			return Report{}, errors.New("move cannot change adoption context")
		}
	}
	if o.Operation == "move" {
		// The original document stays byte-identical. Owned navigation references
		// are updated separately; references inside the moved document are audited.
		p.add(o.To, raw, UserOwned, "move adopted document")
		last := len(p.mutations) - 1
		p.mutations[last].mode = pMode(tree.observed[o.Path].Permissions)
		p.mutations = append(p.mutations, mutation{decision: Decision{Path: o.Path, Action: Delete, Ownership: UserOwned, Reason: "move adopted document"}, expectedExists: true, expectedDigest: tree.observed[o.Path].Digest, expectedMode: tree.observed[o.Path].Mode})
		p.report.Decisions = append(p.report.Decisions, p.mutations[len(p.mutations)-1].decision)
		for i := range registry.Records {
			if registry.Records[i].ID == o.ID {
				registry.Records[i].Path = o.To
			}
		}
		for i := range registry.Selectors {
			for k := range registry.Selectors[i].Snapshot {
				if registry.Selectors[i].Snapshot[k].ID == o.ID && !contracts.Contains(registry.Selectors[i].Exclusions, o.ID) {
					registry.Selectors[i].Snapshot[k].Path = o.To
				}
			}
		}
		registry.History = append(registry.History, contracts.Event{Operation: "move", DocumentID: o.ID, FromPath: o.Path, ToPath: o.To, FromContract: binding.ContractID, ToContract: binding.ContractID, Evidence: []string{}})
		if err = p.moveNavigation(o.Path, o.To); err != nil {
			return Report{}, err
		}
		delete(p.next.Files, o.Path)
	} else {
		projection := map[string]string{"document_type": typ}
		if adopting {
			id := binding.Identity.ID
			contextRoot := binding.Identity.ContextRoot
			if id == "" {
				b := make([]byte, 32)
				if _, err = rand.Read(b); err != nil {
					return Report{}, err
				}
				id = "doc-" + hex.EncodeToString(b)
				contextRoot, err = contracts.ContextRoot(o.Path, typ)
				if err != nil {
					return Report{}, err
				}
			}
			projection["document_id"] = id
			projection["flow_contract"] = o.Contract
			rec := contracts.Record{ID: id, Path: o.Path, Type: typ, ContextRoot: contextRoot, ContractID: o.Contract, BundleDigest: state.manifest.Contracts[o.Contract].Digest}
			found := false
			for i := range registry.Records {
				if registry.Records[i].ID == id {
					registry.Records[i] = rec
					found = true
				}
			}
			if !found {
				registry.Records = append(registry.Records, rec)
			}
			if binding.SelectorID != "" {
				for i := range registry.Selectors {
					if registry.Selectors[i].ID == binding.SelectorID {
						registry.Selectors[i].Exclusions = append(registry.Selectors[i].Exclusions, id)
						sort.Strings(registry.Selectors[i].Exclusions)
					}
				}
			}
			fromPath := o.Path
			if o.Operation == "create" {
				fromPath = ""
			}
			registry.History = append(registry.History, contracts.Event{Operation: o.Operation, DocumentID: id, FromPath: fromPath, ToPath: o.Path, FromContract: binding.ContractID, ToContract: o.Contract, Evidence: evidence})
		}
		raw, err = contracts.Project(raw, projection)
		if err != nil {
			return Report{}, err
		}
		p.add(o.Path, raw, UserOwned, "write explicit document projection")
		if o.Operation == "create" {
			if err = p.addNavigation(o.Path); err != nil {
				return Report{}, err
			}
		}
	}
	if _, tracked := p.next.Files[o.Path]; tracked && o.Operation != "move" {
		p.next.Files[o.Path] = File{Ownership: UserOwned}
	}
	if installation.Has("flows") {
		sort.Slice(registry.Records, func(i, j int) bool { return registry.Records[i].ID < registry.Records[j].ID })
		b, e := contracts.RegistryBytes(registry)
		if e != nil {
			return Report{}, e
		}
		installation.AdoptionDigest = digest(b)
		p.add(contracts.RegistryPath, b, Generated, "record adoption history")
	}
	nextState, err := decodeComponentState(p.future(), p.next, "AGENTS.md", true)
	if err != nil {
		return Report{}, err
	}
	if len(nextState.findings) > 0 {
		return Report{}, fmt.Errorf("document gates fail: %v", nextState.findings)
	}
	nav, err := componentNavigation(p.future(), state.manifest, installation)
	if err != nil {
		return Report{}, err
	}
	if nav.ExitCode != 0 {
		return Report{}, fmt.Errorf("document navigation invalid: %+v", nav.Errors)
	}
	p.next.LastUpdate = UpdateRecord{lock.Template.Version, time.Now().UTC()}
	b, err := marshalLock(p.next)
	if err != nil {
		return Report{}, err
	}
	p.add(LockFileName, b, Generated, "record document transaction")
	return applyComponentPlan(Options{RepoRoot: repo.root, DryRun: o.DryRun, BeforeMutation: o.BeforeMutation, componentDraftInput: o.From}, repo, p)
}

func (p *componentPlan) addNavigation(document string) error {
	parent := path.Dir(document)
	index := parent + "/README.md"
	if document == index {
		parent = path.Dir(parent)
		index = parent + "/README.md"
	}
	if parent == "." {
		return errors.New("cannot index outside memory-bank")
	}
	ref, err := relativeDocumentPath(parent, document)
	if err != nil {
		return err
	}
	if _, ok := p.tree.files[index]; !ok {
		data := []byte("---\nstatus: draft\ndoc_function: index\npurpose: Navigate project documents.\n---\n\n# " + path.Base(parent) + "\n\n- [" + path.Base(document) + "](" + ref + ") — project document.\n")
		if err := p.observeNew(index); err != nil {
			return err
		}
		p.add(index, data, UserOwned, "create project document index")
		return p.addNavigation(index)
	}
	f, tracked := p.next.Files[index]
	if tracked && (f.Ownership == Managed || f.Ownership == Generated) {
		return fmt.Errorf("managed index requires an explicit project override: %s", index)
	}
	data := append([]byte{}, p.tree.files[index]...)
	if len(data) > 0 && data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	data = append(data, []byte("\n- ["+path.Base(document)+"]("+ref+") — project document.\n")...)
	p.add(index, data, UserOwned, "index created document")
	if tracked {
		p.next.Files[index] = File{Ownership: UserOwned}
	}
	return nil
}
func (p *componentPlan) observeNew(name string) error {
	// readComponentTree already scanned memory-bank; a missing key is an observed
	// absence only after a no-follow destination check using the pinned caller.
	if _, ok := p.tree.observed[name]; !ok {
		p.tree.observed[name] = observation{}
	}
	return nil
}
func (p *componentPlan) moveNavigation(from, to string) error {
	for _, name := range contracts.Keys(p.tree.files) {
		if !strings.HasSuffix(name, ".md") || name == from {
			continue
		}
		if !strings.HasPrefix(name, "memory-bank/") {
			continue
		}
		oldRef, err := relativeDocumentPath(path.Dir(name), from)
		if err != nil {
			return err
		}
		newRef, err := relativeDocumentPath(path.Dir(name), to)
		if err != nil {
			return err
		}
		before := string(p.tree.files[name])
		after := strings.ReplaceAll(before, "]("+oldRef+")", "]("+newRef+")")
		if after == before {
			continue
		}
		if f, ok := p.next.Files[name]; ok && (f.Ownership == Managed || f.Ownership == Generated) {
			return fmt.Errorf("managed reference blocks move: %s", name)
		}
		p.add(name, []byte(after), UserOwned, "update project navigation after move")
	}
	return nil
}

func pMode(s string) os.FileMode { var m os.FileMode; fmt.Sscanf(s, "%o", &m); return m }
func relativeDocumentPath(from, to string) (string, error) {
	p, e := filepath.Rel(filepath.FromSlash(from), filepath.FromSlash(to))
	return filepath.ToSlash(p), e
}
