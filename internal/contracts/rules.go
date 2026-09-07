package contracts

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"path"
	"strings"
)

// EngineArtifact is deliberately frozen with its implementation and corpus. New
// semantics require a new engine identifier and artifact, never edits in place.
//
//go:embed engines/governance-v1.json
var engineFiles embed.FS

const EngineID = "governance/v1"

func EngineArtifact() []byte {
	b, err := engineFiles.ReadFile("engines/governance-v1.json")
	if err != nil {
		panic(err)
	}
	return b
}
func EngineDigest() string { return Digest(EngineArtifact()) }

type Rules struct {
	Fields                 map[string][]string `json:"fields,omitempty"`
	Sections               []string            `json:"sections,omitempty"`
	ActiveRequiresUpstream bool                `json:"active_requires_upstream,omitempty"`
	FeatureLifecycle       bool                `json:"feature_lifecycle,omitempty"`
}
type DNA struct {
	SchemaVersion int   `json:"schema_version"`
	Rules         Rules `json:"rules"`
}
type DocumentType struct {
	SchemaVersion int    `json:"schema_version"`
	Type          string `json:"type"`
	Template      string `json:"template"`
	Rules         Rules  `json:"rules"`
}
type EngineRef struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}
type Bundle struct {
	SchemaVersion      int       `json:"schema_version"`
	ID                 string    `json:"id"`
	Type               string    `json:"type"`
	Engine             EngineRef `json:"engine"`
	DNA                Rules     `json:"dna"`
	Base               Rules     `json:"base"`
	Extension          Rules     `json:"extension"`
	Legacy             bool      `json:"legacy"`
	TransitionEvidence bool      `json:"transition_evidence"`
}
type Catalog struct {
	DNA      DNA
	Types    map[string]DocumentType
	Bundles  map[string]Bundle
	Manifest Manifest
	Files    map[string][]byte
}

func MergeRules(sets ...Rules) (Rules, error) {
	out := Rules{Fields: map[string][]string{}, Sections: []string{}}
	sections := map[string]bool{}
	for _, r := range sets {
		if !SortedSet(r.Sections) {
			return out, errors.New("sections must be a sorted set")
		}
		for k, values := range r.Fields {
			if strings.TrimSpace(k) == "" || !SortedSet(values) {
				return out, errors.New("invalid field rule")
			}
			if prior, exists := out.Fields[k]; exists && len(prior) > 0 {
				if len(values) == 0 {
					return out, fmt.Errorf("field %s weakens its parent enum", k)
				}
				for _, v := range values {
					if !Contains(prior, v) {
						return out, fmt.Errorf("field %s widens its parent enum", k)
					}
				}
			}
			out.Fields[k] = append([]string{}, values...)
		}
		for _, section := range r.Sections {
			sections[section] = true
		}
		out.ActiveRequiresUpstream = out.ActiveRequiresUpstream || r.ActiveRequiresUpstream
		out.FeatureLifecycle = out.FeatureLifecycle || r.FeatureLifecycle
	}
	out.Sections = Keys(sections)
	return out, nil
}
func Contains(values []string, s string) bool {
	for _, v := range values {
		if v == s {
			return true
		}
	}
	return false
}

// LoadCatalog reads only installed definitions when selection is non-nil. A full
// source calls it with nil after validating the exhaustive inventory.
func LoadCatalog(m Manifest, files map[string][]byte, selection *Installation) (Catalog, error) {
	c := Catalog{Manifest: m, Files: files, Types: map[string]DocumentType{}, Bundles: map[string]Bundle{}}
	if err := Decode(files[m.DNAContract], &c.DNA); err != nil {
		return c, fmt.Errorf("DNA contract: %w", err)
	}
	if c.DNA.SchemaVersion != 1 {
		return c, errors.New("unsupported DNA schema")
	}
	if _, err := MergeRules(c.DNA.Rules); err != nil {
		return c, err
	}
	if c.DNA.Rules.FeatureLifecycle {
		return c, errors.New("DNA cannot contain feature lifecycle")
	}
	if selection == nil {
		for _, p := range Keys(files) {
			if !strings.HasPrefix(p, "memory-bank/") || !strings.HasSuffix(p, ".md") {
				continue
			}
			doc, err := ParseDocument(p, files[p])
			if err != nil {
				return c, fmt.Errorf("source metadata %s: %w", p, err)
			}
			if findings := ValidateRules(doc, c.DNA.Rules, Identity{Path: p, ContextRoot: path.Dir(p)}, nil); len(findings) != 0 {
				return c, fmt.Errorf("source metadata %s: %v", p, findings)
			}
		}
	}
	if selection != nil && !selection.Has("documents") {
		return c, nil
	}
	for typ, p := range m.DocumentTypes {
		var d DocumentType
		if err := Decode(files[p], &d); err != nil {
			return c, fmt.Errorf("base type %s: %w", typ, err)
		}
		if d.SchemaVersion != 1 || d.Type != typ {
			return c, fmt.Errorf("incompatible base type %s", typ)
		}
		if err := m.requireFile(d.Template, "documents"); err != nil {
			return c, err
		}
		if _, err := MergeRules(c.DNA.Rules, d.Rules); err != nil {
			return c, err
		}
		if d.Rules.FeatureLifecycle {
			return c, errors.New("base type cannot contain flow lifecycle")
		}
		template, exists := files[d.Template]
		if !exists {
			return c, fmt.Errorf("missing base template %s", d.Template)
		}
		doc, err := ParseDocument(d.Template, template)
		if err != nil || !doc.HasFrontmatter {
			return c, fmt.Errorf("invalid base template %s: %v", d.Template, err)
		}
		if doc.String("status") != "draft" || doc.String("document_type") != typ || doc.String("doc_kind") != typ || doc.Has("document_id") || doc.Has("flow_contract") {
			return c, fmt.Errorf("invalid base template metadata %s", d.Template)
		}
		if hasEmbeddedFrontmatter(doc.Body) {
			return c, fmt.Errorf("embedded frontmatter is unsupported in %s", d.Template)
		}
		if _, err := RelocateBaseDocument(template, d.Template, "memory-bank/features/FT-probe/brief.md"); err != nil {
			return c, fmt.Errorf("base references %s: %w", d.Template, err)
		}
		c.Types[typ] = d
	}
	if selection != nil && !selection.Has("flows") {
		return c, nil
	}
	for id, ref := range m.Contracts {
		data, exists := files[ref.Path]
		if !exists || Digest(data) != ref.Digest {
			return c, fmt.Errorf("missing or changed bundle %s", id)
		}
		var b Bundle
		if err := Decode(data, &b); err != nil {
			return c, fmt.Errorf("bundle %s: %w", id, err)
		}
		if b.SchemaVersion != 1 || b.ID != id || !ValidContractID(id) || b.Engine.ID != EngineID || b.Engine.Digest != EngineDigest() {
			return c, fmt.Errorf("unsupported bundle or engine %s", id)
		}
		if _, ok := c.Types[b.Type]; !ok {
			return c, fmt.Errorf("unknown bundle type %s", b.Type)
		}
		r, err := MergeRules(b.DNA, b.Base, b.Extension)
		if err != nil {
			return c, err
		}
		if r.FeatureLifecycle && b.Type != "feature" {
			return c, errors.New("feature lifecycle requires feature type")
		}
		if b.Legacy {
			expected, err := LegacyBundle(id, b.Type)
			actualBytes, _ := Canonical(b)
			expectedBytes, _ := Canonical(expected)
			if err != nil || !bytes.Equal(actualBytes, expectedBytes) {
				return c, fmt.Errorf("legacy bundle %s differs from frozen classifier rules", id)
			}
		}
		c.Bundles[id] = b
	}
	for _, compat := range m.LegacySources {
		for typ, id := range compat.Contracts {
			b, ok := c.Bundles[id]
			if !ok || !b.Legacy || b.Type != typ {
				return c, errors.New("incompatible legacy bundle mapping")
			}
		}
	}
	return c, nil
}

func hasEmbeddedFrontmatter(body []byte) bool {
	lines := strings.Split(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n")
	for i, line := range lines {
		if line != "---" {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if lines[j] != "---" {
				continue
			}
			embedded, err := ParseDocument("", []byte(strings.Join(lines[i:j+1], "\n")+"\n"))
			if err == nil {
				for _, key := range []string{"status", "title", "purpose", "doc_kind", "doc_function", "derived_from", "audience", "delivery_status", "research_status", "decision_status", "document_type", "document_id", "flow_contract", "template_for", "template_target_path", "canonical_for", "must_not_define"} {
					if embedded.Has(key) {
						return true
					}
				}
			}
			break
		}
	}
	return false
}

// ValidateTransition applies to a NEW operation. Both endpoints must be installed;
// unlike lock-authenticated history replay, this may not rely on retired contracts.
func (c Catalog) ValidateTransition(from, to, typ string, evidence []string) error {
	old, oldOK := c.Bundles[from]
	next, nextOK := c.Bundles[to]
	if !oldOK || !nextOK || old.Type != typ || next.Type != typ || from == to {
		return errors.New("transition requires two installed compatible contracts")
	}
	if !SortedSet(evidence) || ((old.TransitionEvidence || next.TransitionEvidence) && len(evidence) == 0) {
		return errors.New("transition requires valid evidence references")
	}
	return nil
}
