package contracts

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

const ManifestPath = "memory-bank/components.json"
const RegistryPath = "memory-bank/.adoption.json"

var contractIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*(/[a-z0-9_-]+)*/v[1-9][0-9]*$`)

func ValidContractID(s string) bool { return contractIDPattern.MatchString(s) }

var sourceRefPattern = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

type Component struct {
	Dependencies []string `json:"dependencies"`
	Adapter      bool     `json:"adapter"`
	Legacy       bool     `json:"legacy"`
}
type File struct {
	Component string `json:"component"`
	Ownership string `json:"ownership"`
}
type BundleRef struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}
type Compatibility struct {
	Classifier string            `json:"classifier"`
	Contracts  map[string]string `json:"contracts"`
}
type PathMigration struct {
	To     string `json:"to"`
	Policy string `json:"policy"`
}
type Manifest struct {
	SchemaVersion          int                      `json:"schema_version"`
	Capabilities           []string                 `json:"capabilities"`
	DNAContract            string                   `json:"dna_contract"`
	Components             map[string]Component     `json:"components"`
	Presets                map[string][]string      `json:"presets"`
	Files                  map[string]File          `json:"files"`
	DocumentTypes          map[string]string        `json:"document_types"`
	Contracts              map[string]BundleRef     `json:"contracts"`
	LegacySources          map[string]Compatibility `json:"legacy_sources"`
	LegacyDefaultSourceRef string                   `json:"legacy_default_source_ref"`
	MigrationPaths         map[string]PathMigration `json:"migration_paths,omitempty"`
}
type Installation struct {
	RendererVersion *int     `json:"renderer_version,omitempty"`
	Preset          string   `json:"preset"`
	Components      []string `json:"components"`
	Adapters        []string `json:"adapters"`
	ManifestDigest  string   `json:"manifest_digest"`
	AdoptionDigest  string   `json:"adoption_digest,omitempty"`
	LegacySourceRef string   `json:"legacy_source_ref,omitempty"`
}

func (s Installation) Has(id string) bool {
	for _, v := range append(append([]string{}, s.Components...), s.Adapters...) {
		if v == id {
			return true
		}
	}
	return false
}
func ReadManifest(data []byte, inventory map[string][]byte) (Manifest, error) {
	var m Manifest
	if err := Decode(data, &m); err != nil {
		return m, err
	}
	if m.SchemaVersion != 1 {
		return m, errors.New("unsupported component schema")
	}
	if !SortedSet(m.Capabilities) {
		return m, errors.New("capabilities must be a sorted set")
	}
	caps := map[string]bool{}
	for _, c := range m.Capabilities {
		switch c {
		case "components/v1", "adoption/v1", "source-format/v1", "legacy/v1":
			caps[c] = true
		default:
			return m, fmt.Errorf("unsupported capability %q", c)
		}
	}
	if !caps["components/v1"] || !caps["adoption/v1"] {
		return m, errors.New("component capabilities missing")
	}
	required := map[string][]string{"dna": {}, "documents": {"dna"}, "flows": {"dna", "documents"}}
	for name, deps := range required {
		c, ok := m.Components[name]
		if !ok || c.Adapter || c.Legacy || !reflect.DeepEqual(c.Dependencies, deps) {
			return m, fmt.Errorf("invalid %s component dependencies", name)
		}
	}
	for id, c := range m.Components {
		if id == "" || !SortedSet(c.Dependencies) {
			return m, errors.New("invalid component or dependencies")
		}
		if _, ok := required[id]; !ok && !c.Adapter {
			return m, fmt.Errorf("extra non-adapter component %s", id)
		}
		if _, err := m.closure([]string{id}); err != nil {
			return m, err
		}
	}
	if len(m.Presets) != 4 {
		return m, errors.New("exactly four presets are required")
	}
	for p, ids := range map[string][]string{"core": {"dna"}, "docs": {"dna", "documents"}, "full": {"dna", "documents", "flows"}} {
		if !reflect.DeepEqual(m.Presets[p], ids) {
			return m, fmt.Errorf("invalid preset %s", p)
		}
	}
	if !SortedSet(m.Presets["legacy"]) {
		return m, errors.New("invalid legacy preset")
	}
	legacy, err := m.closure(m.Presets["legacy"])
	if err != nil {
		return m, err
	}
	expected := map[string]bool{"dna": true, "documents": true, "flows": true}
	for id, c := range m.Components {
		if c.Legacy {
			expected[id] = true
		}
	}
	expectedClosure, err := m.closure(Keys(expected))
	if err != nil || !reflect.DeepEqual(legacy, expectedClosure) {
		return m, errors.New("legacy preset does not preserve declared legacy adapters")
	}
	if inventory != nil && len(m.Files) != len(inventory) {
		return m, errors.New("inventory does not match payload")
	}
	if err := CheckPortable(Keys(m.Files)); err != nil {
		return m, err
	}
	for p, f := range m.Files {
		if !ValidPath(p) || p == "AGENTS.md" || p == RegistryPath || p == "memory-bank/.lock" || strings.HasPrefix(p, ".memory-bank-update-") || strings.HasPrefix(p, "memory-bank/.repo/") {
			return m, fmt.Errorf("reserved or unsafe payload path %s", p)
		}
		if _, ok := inventory[p]; inventory != nil && !ok {
			return m, fmt.Errorf("missing payload %s", p)
		}
		if _, ok := m.Components[f.Component]; !ok {
			return m, fmt.Errorf("unknown file component %s", f.Component)
		}
		if f.Ownership != "managed" && f.Ownership != "user-owned" {
			return m, fmt.Errorf("invalid ownership %s", p)
		}
	}
	for _, p := range []string{ManifestPath, "memory-bank/README.md"} {
		if m.Files[p] != (File{"dna", "managed"}) {
			return m, fmt.Errorf("reserved composed path %s must be managed DNA", p)
		}
	}
	if err := m.requireFile(m.DNAContract, "dna"); err != nil {
		return m, err
	}
	if len(m.DocumentTypes) == 0 || len(m.Contracts) == 0 {
		return m, errors.New("document types and contracts are required")
	}
	for typ, p := range m.DocumentTypes {
		if typ == "" {
			return m, errors.New("empty type")
		}
		if err := m.requireFile(p, "documents"); err != nil {
			return m, err
		}
	}
	for id, b := range m.Contracts {
		if !ValidContractID(id) || !ValidDigest(b.Digest) {
			return m, errors.New("invalid bundle reference")
		}
		if err := m.requireFile(b.Path, "flows"); err != nil {
			return m, err
		}
		if inventory != nil && Digest(inventory[b.Path]) != b.Digest {
			return m, fmt.Errorf("bundle digest mismatch: %s", id)
		}
	}
	if _, ok := m.LegacySources[m.LegacyDefaultSourceRef]; !ok {
		return m, errors.New("legacy creation baseline missing")
	}
	for ref, c := range m.LegacySources {
		if !sourceRefPattern.MatchString(ref) || ref != "f1f04de843aef45a2425d4a7351d577bbf89e940" || c.Classifier != "legacy-f1f04de/v1" || len(c.Contracts) == 0 {
			return m, errors.New("unsupported legacy classifier")
		}
		for typ, id := range c.Contracts {
			if _, ok := m.DocumentTypes[typ]; !ok {
				return m, errors.New("unknown compatibility type")
			}
			if _, ok := m.Contracts[id]; !ok {
				return m, errors.New("missing compatibility bundle")
			}
		}
	}
	for from, migration := range m.MigrationPaths {
		if migration.Policy != "retain-wrapper" || !strings.HasPrefix(from, "memory-bank/flows/templates/") {
			return m, errors.New("unsupported path migration")
		}
		if err := m.requireFile(from, "flows"); err != nil {
			return m, err
		}
		if err := m.requireFile(migration.To, "documents"); err != nil {
			return m, err
		}
	}
	return m, nil
}
func (m Manifest) requireFile(p, component string) error {
	f, ok := m.Files[p]
	if !ok || !ValidPath(p) || f.Component != component || f.Ownership != "managed" {
		return fmt.Errorf("%s must be a managed %s file", p, component)
	}
	return nil
}
func (m Manifest) closure(ids []string) (map[string]bool, error) {
	state := map[string]int{}
	result := map[string]bool{}
	var visit func(string) error
	visit = func(id string) error {
		c, ok := m.Components[id]
		if !ok {
			return fmt.Errorf("unknown component %q", id)
		}
		if state[id] == 1 {
			return errors.New("component dependency cycle")
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		for _, dep := range c.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}
		state[id] = 2
		result[id] = true
		return nil
	}
	for _, id := range ids {
		if err := visit(id); err != nil {
			return nil, err
		}
	}
	return result, nil
}
func (m Manifest) Select(preset string, adapters []string, old *Installation) (Installation, error) {
	if old != nil && preset == "" && len(adapters) == 0 {
		ids := append(append([]string{}, old.Components...), old.Adapters...)
		closure, err := m.closure(ids)
		if err != nil {
			return Installation{}, err
		}
		if !reflect.DeepEqual(Keys(closure), sorted(ids)) {
			return Installation{}, errors.New("source changed locked dependency closure")
		}
		return *old, nil
	}
	if preset == "" {
		preset = "legacy"
		if old != nil {
			preset = old.Preset
		}
	}
	ids, ok := m.Presets[preset]
	if !ok {
		return Installation{}, fmt.Errorf("unknown preset %q", preset)
	}
	ids = append([]string{}, ids...)
	if old != nil {
		adapters = append(append([]string{}, adapters...), old.Adapters...)
	}
	for _, id := range adapters {
		c, ok := m.Components[id]
		if !ok || !c.Adapter {
			return Installation{}, fmt.Errorf("unknown adapter %q", id)
		}
	}
	ids = append(ids, adapters...)
	resolved, err := m.closure(ids)
	if err != nil {
		return Installation{}, err
	}
	if old != nil {
		for _, id := range append(append([]string{}, old.Components...), old.Adapters...) {
			if !resolved[id] {
				return Installation{}, errors.New("component removal is unsupported")
			}
		}
	}
	s := Installation{Preset: preset, Components: []string{}, Adapters: []string{}}
	for _, id := range Keys(resolved) {
		if m.Components[id].Adapter {
			s.Adapters = append(s.Adapters, id)
		} else {
			s.Components = append(s.Components, id)
		}
	}
	if old != nil {
		s.LegacySourceRef = old.LegacySourceRef
		s.AdoptionDigest = old.AdoptionDigest
	}
	if preset == "legacy" && s.LegacySourceRef == "" {
		s.LegacySourceRef = m.LegacyDefaultSourceRef
	}
	return s, nil
}
func sorted(values []string) []string {
	seen := map[string]bool{}
	for _, v := range values {
		seen[v] = true
	}
	return Keys(seen)
}

// ValidateInstallation checks a persisted closure rather than treating it as new opt-in.
func (m Manifest) ValidateInstallation(s Installation) error {
	if s.RendererVersion != nil && *s.RendererVersion != 1 && *s.RendererVersion != 2 {
		return errors.New("unsupported component renderer")
	}
	if !ValidDigest(s.ManifestDigest) || !SortedSet(s.Components) || !SortedSet(s.Adapters) {
		return errors.New("invalid installation metadata")
	}
	if s.Has("flows") != ValidDigest(s.AdoptionDigest) || (!s.Has("flows") && s.AdoptionDigest != "") {
		return errors.New("invalid adoption digest for selection")
	}
	for _, id := range s.Components {
		c, ok := m.Components[id]
		if !ok || c.Adapter {
			return errors.New("invalid installed component")
		}
	}
	for _, id := range s.Adapters {
		c, ok := m.Components[id]
		if !ok || !c.Adapter {
			return errors.New("invalid installed adapter")
		}
	}
	if s.Preset == "legacy" {
		// The legacy preset is an installation default, not permission to adopt
		// adapters newly marked legacy in a subsequent source.
		if !reflect.DeepEqual(s.Components, []string{"dna", "documents", "flows"}) {
			return errors.New("invalid legacy component closure")
		}
		ids := append(append([]string{}, s.Components...), s.Adapters...)
		closure, e := m.closure(ids)
		if e != nil || !reflect.DeepEqual(Keys(closure), sorted(ids)) {
			return errors.New("invalid locked legacy adapter closure")
		}
	} else {
		wanted, err := m.Select(s.Preset, s.Adapters, nil)
		if err != nil || !reflect.DeepEqual(wanted.Components, s.Components) || !reflect.DeepEqual(wanted.Adapters, s.Adapters) {
			return errors.New("installed closure does not match preset and adapters")
		}
	}
	if s.LegacySourceRef != "" {
		if _, ok := m.LegacySources[s.LegacySourceRef]; !ok || !s.Has("flows") {
			return errors.New("unsupported legacy creation reference")
		}
	}
	return nil
}
