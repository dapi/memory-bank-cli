package contracts

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var documentIDPattern = regexp.MustCompile(`^doc-[0-9a-f]{64}$`)

type Record struct {
	ID           string `json:"id"`
	Path         string `json:"path"`
	Type         string `json:"type"`
	ContextRoot  string `json:"context_root"`
	ContractID   string `json:"contract_id"`
	BundleDigest string `json:"bundle_digest"`
}

func (r Record) Identity() Identity { return Identity{r.ID, r.Path, r.Type, r.ContextRoot} }

type Selector struct {
	ID           string     `json:"id"`
	SourceRef    string     `json:"source_ref"`
	Type         string     `json:"type"`
	ContractID   string     `json:"contract_id"`
	BundleDigest string     `json:"bundle_digest"`
	Snapshot     []Identity `json:"snapshot"`
	Exclusions   []string   `json:"exclusions"`
}
type Event struct {
	Operation    string   `json:"operation"`
	DocumentID   string   `json:"document_id"`
	FromPath     string   `json:"from_path"`
	ToPath       string   `json:"to_path"`
	FromContract string   `json:"from_contract"`
	ToContract   string   `json:"to_contract"`
	Evidence     []string `json:"evidence"`
}
type Registry struct {
	SchemaVersion int        `json:"schema_version"`
	Records       []Record   `json:"records"`
	Selectors     []Selector `json:"selectors"`
	History       []Event    `json:"history"`
}
type Binding struct {
	Identity                             Identity
	ContractID, BundleDigest, SelectorID string
}

func EmptyRegistry() Registry { return Registry{1, []Record{}, []Selector{}, []Event{}} }
func RegistryBytes(r Registry) ([]byte, error) {
	b, e := Canonical(r)
	if e != nil {
		return nil, e
	}
	return append(b, '\n'), nil
}
func SelectorID(source, typ, contract, bundle string) string {
	return "sel-" + FramedID("memory-bank/selector-id/v1", source, typ, contract, bundle)
}
func MigrationID(lock, p, content string) string {
	return "doc-" + FramedID("memory-bank/document-id/v1", lock, p, content)
}
func ReadRegistry(data []byte) (Registry, error) {
	var r Registry
	e := Decode(data, &r)
	if e != nil {
		return r, e
	}
	canonical, e := RegistryBytes(r)
	if e != nil || string(canonical) != string(data) {
		return r, errors.New("registry must use canonical JSON plus LF")
	}
	return r, nil
}

// ValidateRegistry checks a registry whose exact bytes have already been authenticated
// against installation.adoption_digest. CTR-01 deliberately does not require retaining
// inactive historical bundles; the trusted lock protects previously checked history.
// New writes must additionally call Catalog.ValidateTransition before appending an event.
func ValidateRegistry(r Registry, c Catalog, docs map[string]Document) (map[string]Binding, error) {
	active := map[string]Binding{}
	all := map[string]Identity{}
	records := map[string]Record{}
	snapshots := map[string]Selector{}
	if r.SchemaVersion != 1 {
		return nil, errors.New("unsupported adoption registry schema")
	}
	previous := ""
	checkIdentity := func(id Identity) error {
		if !documentIDPattern.MatchString(id.ID) || !DocumentPath(id.Path) || !ValidPath(id.ContextRoot) || !strings.HasPrefix(id.Path, id.ContextRoot+"/") {
			return errors.New("invalid document identity or context")
		}
		if _, ok := c.Types[id.Type]; !ok {
			return errors.New("identity type is not installed")
		}
		return nil
	}
	checkBundle := func(typ, id, digest string) error {
		b, ok := c.Bundles[id]
		ref := c.Manifest.Contracts[id]
		if !ok || b.Type != typ || ref.Digest != digest {
			return fmt.Errorf("missing or changed pinned bundle %s", id)
		}
		return nil
	}
	for _, rec := range r.Records {
		if rec.ID <= previous {
			return nil, errors.New("records must be uniquely ID-sorted")
		}
		previous = rec.ID
		if err := checkIdentity(rec.Identity()); err != nil {
			return nil, err
		}
		if err := checkBundle(rec.Type, rec.ContractID, rec.BundleDigest); err != nil {
			return nil, err
		}
		records[rec.ID] = rec
		all[rec.ID] = rec.Identity()
		active[rec.ID] = Binding{rec.Identity(), rec.ContractID, rec.BundleDigest, ""}
	}
	previous = ""
	for _, sel := range r.Selectors {
		if len(sel.Snapshot) == 0 || sel.ID <= previous || sel.ID != SelectorID(sel.SourceRef, sel.Type, sel.ContractID, sel.BundleDigest) || !SortedSet(sel.Exclusions) {
			return nil, errors.New("invalid selector identity or ordering")
		}
		previous = sel.ID
		compat, ok := c.Manifest.LegacySources[sel.SourceRef]
		if !ok || compat.Contracts[sel.Type] != sel.ContractID {
			return nil, errors.New("unsupported selector provenance")
		}
		if err := checkBundle(sel.Type, sel.ContractID, sel.BundleDigest); err != nil {
			return nil, err
		}
		if !c.Bundles[sel.ContractID].Legacy {
			return nil, errors.New("selector requires compatibility bundle")
		}
		members := map[string]bool{}
		priorID := ""
		for _, id := range sel.Snapshot {
			if id.ID <= priorID || id.Type != sel.Type {
				return nil, errors.New("invalid selector snapshot ordering/type")
			}
			priorID = id.ID
			if err := checkIdentity(id); err != nil {
				return nil, err
			}
			if _, duplicate := snapshots[id.ID]; duplicate {
				return nil, errors.New("identity occurs in multiple selectors")
			}
			members[id.ID] = true
			snapshots[id.ID] = sel
			if rec, ok := records[id.ID]; ok {
				if rec.Type != id.Type || rec.ContextRoot != id.ContextRoot || !Contains(sel.Exclusions, id.ID) {
					return nil, errors.New("multiple applicable adoption records")
				}
			} else {
				all[id.ID] = id
			}
			if !Contains(sel.Exclusions, id.ID) {
				active[id.ID] = Binding{id, sel.ContractID, sel.BundleDigest, sel.ID}
			}
		}
		for _, excluded := range sel.Exclusions {
			if !members[excluded] {
				return nil, errors.New("exclusion is not a snapshot member")
			}
			if _, ok := records[excluded]; !ok {
				return nil, errors.New("exclusion lacks resulting record")
			}
		}
	}
	type state struct {
		path, contract, initialPath, initialOperation string
		selectorTransition                            bool
	}
	states := map[string]state{}
	lastMigrated := ""
	migrationEnded := false
	for _, event := range r.History {
		if event.Operation == "migrate" {
			if migrationEnded || event.DocumentID <= lastMigrated {
				return nil, errors.New("migration history must be an initial ID-sorted block")
			}
			lastMigrated = event.DocumentID
		} else {
			migrationEnded = true
		}
		if !documentIDPattern.MatchString(event.DocumentID) || !SortedSet(event.Evidence) || !DocumentPath(event.ToPath) || !ValidContractID(event.ToContract) {
			return nil, errors.New("invalid history event")
		}
		if _, ok := all[event.DocumentID]; !ok {
			return nil, errors.New("history refers to unknown identity")
		}
		prior, exists := states[event.DocumentID]
		if !exists {
			if event.FromContract != "" {
				return nil, errors.New("initial event has a previous contract")
			}
			switch event.Operation {
			case "create":
				if event.FromPath != "" {
					return nil, errors.New("create has a previous path")
				}
			case "adopt", "migrate":
				if event.FromPath != event.ToPath {
					return nil, errors.New("initial binding path mismatch")
				}
			default:
				return nil, errors.New("history has no initial binding event")
			}
			prior.initialPath = event.ToPath
			prior.initialOperation = event.Operation
			sel, inSelector := snapshots[event.DocumentID]
			if (event.Operation == "migrate") != inSelector || (inSelector && event.ToContract != sel.ContractID) {
				return nil, errors.New("migration history does not match snapshot")
			}
		} else {
			if event.FromPath != prior.path || event.FromContract != prior.contract {
				return nil, errors.New("history continuity broken")
			}
			switch event.Operation {
			case "move":
				if event.ToPath == prior.path || event.ToContract != prior.contract {
					return nil, errors.New("invalid move event")
				}
			case "transition":
				oldBundle, oldAvailable := c.Bundles[event.FromContract]
				newBundle, newAvailable := c.Bundles[event.ToContract]
				if ((oldAvailable && oldBundle.TransitionEvidence) || (newAvailable && newBundle.TransitionEvidence)) && len(event.Evidence) == 0 {
					return nil, errors.New("transition history lacks required evidence")
				}
				if event.ToPath != prior.path || event.ToContract == prior.contract {
					return nil, errors.New("invalid transition event")
				}
				if sel, ok := snapshots[event.DocumentID]; ok && event.FromContract == sel.ContractID {
					prior.selectorTransition = true
				}
			default:
				return nil, errors.New("duplicate initial or unsupported event")
			}
		}
		id := all[event.DocumentID]
		if !strings.HasPrefix(event.ToPath, id.ContextRoot+"/") {
			return nil, errors.New("event escaped identity context")
		}
		prior.path = event.ToPath
		prior.contract = event.ToContract
		states[event.DocumentID] = prior
	}
	paths := map[string]string{}
	for id, binding := range active {
		current, ok := states[id]
		if !ok || current.path != binding.Identity.Path || current.contract != binding.ContractID {
			return nil, errors.New("history final binding disagrees with registry")
		}
		expectedRoot, err := ContextRoot(current.initialPath, binding.Identity.Type)
		if err != nil || expectedRoot != binding.Identity.ContextRoot {
			return nil, errors.New("identity context does not match initial binding")
		}
		if sel, ok := snapshots[id]; ok && Contains(sel.Exclusions, id) && !current.selectorTransition {
			return nil, errors.New("selector exclusion lacks explicit transition")
		}
		if other, duplicate := paths[binding.Identity.Path]; duplicate && other != id {
			return nil, errors.New("multiple identities share a path")
		}
		paths[binding.Identity.Path] = id
		d, present := docs[binding.Identity.Path]
		if !present || d.String("document_id") != id || d.String("document_type") != binding.Identity.Type {
			return nil, errors.New("missing target or identity/type projection")
		}
		if binding.SelectorID == "" {
			if d.String("flow_contract") != binding.ContractID {
				return nil, errors.New("missing or changed contract projection")
			}
		} else if d.Has("flow_contract") && d.String("flow_contract") != binding.ContractID {
			return nil, errors.New("selector contract projection mismatch")
		}
	}
	for p, d := range docs {
		if d.Has("document_type") {
			typ := d.String("document_type")
			if _, ok := c.Types[typ]; !ok {
				return nil, fmt.Errorf("%s: unknown document type", p)
			}
			if d.Has("doc_kind") && d.String("doc_kind") != typ {
				return nil, fmt.Errorf("%s: contradictory document kind", p)
			}
		}
		if d.Has("document_id") || d.Has("flow_contract") {
			id := d.String("document_id")
			binding, ok := active[id]
			if !ok || binding.Identity.Path != p {
				return nil, fmt.Errorf("%s: orphan adoption projection", p)
			}
		}
	}
	return active, nil
}
