package contracts

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
)

type Identity struct {
	ID          string `json:"id"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	ContextRoot string `json:"context_root"`
}
type Finding struct {
	DocumentID string `json:"document_id"`
	Code       string `json:"code"`
	RuleID     string `json:"rule_id"`
	Subject    string `json:"subject"`
}

func finding(id Identity, p, code, rule string) Finding {
	subject := "@context"
	if p != "" {
		subject = strings.TrimPrefix(p, id.ContextRoot+"/")
	}
	return Finding{id.ID, code, rule, subject}
}
func SortFindings(f []Finding) []Finding {
	sort.SliceStable(f, func(i, j int) bool { a, _ := Canonical(f[i]); b, _ := Canonical(f[j]); return string(a) < string(b) })
	return f
}
func ValidateRules(d Document, r Rules, id Identity, context map[string]Document) []Finding {
	out := []Finding{}
	if !d.HasFrontmatter {
		out = append(out, finding(id, d.Path, "governance.frontmatter_missing", "frontmatter"))
	}
	for _, key := range Keys(r.Fields) {
		values := r.Fields[key]
		v, ok := d.Fields[key].(string)
		if !ok || strings.TrimSpace(v) == "" || (len(values) > 0 && !Contains(values, v)) {
			out = append(out, finding(id, d.Path, "contract.field_invalid", "field/"+key))
		}
	}
	headings := Headings(d)
	for _, h := range r.Sections {
		if !headings[h] {
			out = append(out, finding(id, d.Path, "contract.section_missing", "section/"+h))
		}
	}
	if r.ActiveRequiresUpstream && d.String("status") == "active" && d.Path != "memory-bank/dna/principles.md" && len(DerivedPaths(d)) == 0 {
		out = append(out, finding(id, d.Path, "contract.upstream_missing", "active-upstream"))
	}
	if r.FeatureLifecycle {
		out = append(out, featureFindings(d, id, context, false)...)
	}
	return SortFindings(out)
}
func ValidateBundle(d Document, b Bundle, id Identity, context map[string]Document) []Finding {
	if b.Legacy {
		out := legacyMetadata(d, id, id.Type == "feature", id.Type == "adr")
		validContext := map[string]Document{}
		for _, p := range Keys(context) {
			other := context[p]
			if !strings.HasPrefix(p, id.ContextRoot+"/") {
				continue
			}
			if p != d.Path && (id.Type == "feature" || id.Type == "research" || id.Type == "epic") {
				out = append(out, legacyMetadata(other, id, false, false)...)
			}
			fields, found, err := parseLegacyFrontmatter(other.Raw)
			if err == nil && found {
				other.Fields = fields
				validContext[p] = other
			}
		}
		fields, found, err := parseLegacyFrontmatter(d.Raw)
		if err == nil && found {
			d.Fields = fields
			validContext[d.Path] = d
		}
		if id.Type == "feature" {
			if _, valid := validContext[d.Path]; valid {
				out = append(out, featureFindings(d, id, validContext, true)...)
			} else if len(validContext) > 0 {
				out = append(out, finding(id, "", "lifecycle.feature_brief_missing", "lifecycle.feature_brief_missing"))
			}
		}
		return SortFindings(out)
	}
	r, _ := MergeRules(b.DNA, b.Base, b.Extension)
	return ValidateRules(d, r, id, context)
}
func LegacyBundle(id, typ string) (Bundle, error) {
	if !Contains([]string{"adr", "feature", "prd", "use_case", "research", "epic"}, typ) {
		return Bundle{}, fmt.Errorf("unsupported legacy type %s", typ)
	}
	b := Bundle{SchemaVersion: 1, ID: id, Type: typ, Engine: EngineRef{EngineID, EngineDigest()}, Legacy: true, DNA: Rules{Fields: map[string][]string{"status": {"active", "archived", "draft"}}, ActiveRequiresUpstream: true}}
	if typ == "adr" {
		b.Base = Rules{Fields: map[string][]string{"decision_status": {"accepted", "proposed", "rejected", "superseded"}}}
	}
	if typ == "feature" {
		b.Extension = Rules{FeatureLifecycle: true}
	}
	return b, nil
}
func legacyMetadata(d Document, id Identity, featureOwner, adr bool) []Finding {
	out := []Finding{}
	add := func(code string) { out = append(out, finding(id, d.Path, code, code)) }
	fields, found, err := parseLegacyFrontmatter(d.Raw)
	if err != nil {
		add("governance.frontmatter_invalid")
		return out
	}
	d.Fields = fields
	if !found {
		add("governance.frontmatter_missing")
		return out
	}
	if !Contains([]string{"draft", "active", "archived"}, d.String("status")) {
		add("governance.status_invalid")
	}
	if d.Has("delivery_status") {
		if !Contains([]string{"planned", "in_progress", "done", "cancelled"}, d.String("delivery_status")) {
			add("governance.delivery_status_invalid")
		}
		if !featureOwner {
			add("lifecycle.delivery_status_wrong_owner")
		}
	}
	if d.Has("decision_status") {
		if !Contains([]string{"proposed", "accepted", "superseded", "rejected"}, d.String("decision_status")) {
			add("governance.decision_status_invalid")
		}
		if d.String("doc_kind") != "adr" {
			add("lifecycle.decision_status_wrong_owner")
		}
	}
	if d.String("status") == "active" && !d.Has("derived_from") {
		add("governance.derived_from_missing")
	}
	if d.Has("derived_from") && len(DerivedPaths(d)) == 0 {
		add("governance.derived_from_invalid")
	}
	if adr && !d.Has("decision_status") {
		add("lifecycle.adr_decision_status_missing")
	}
	return out
}

var designDecisionPattern = regexp.MustCompile("(?i)^\\s*(?:(?:[-+*]|\\d+[.)])\\s+)?(?:\\|\\s*)?`?design\\s+required\\s*:\\s*`?(yes|no)`?(?:\\s*`)?(?:\\s*\\|.*|\\s*[.,;:]?\\s*)$")

func DesignDecision(d Document) (string, bool) {
	inSection := false
	depth := 0
	decision := ""
	for _, line := range VisibleLines(d.Body) {
		if level, title, ok := headingParts(line); ok {
			if strings.EqualFold(title, "Design Requirement Decision") {
				inSection = true
				depth = level
				continue
			}
			if inSection && level <= depth {
				break
			}
		}
		if inSection {
			if m := designDecisionPattern.FindStringSubmatch(line); len(m) > 0 {
				if decision != "" && decision != strings.ToLower(m[1]) {
					return "", false
				}
				decision = strings.ToLower(m[1])
			}
		}
	}
	return decision, decision != ""
}
func featureFindings(d Document, id Identity, context map[string]Document, legacy bool) []Finding {
	out := []Finding{}
	add := func(p, code string) { out = append(out, finding(id, p, code, code)) }
	design, hasDesign := context[path.Join(id.ContextRoot, "design.md")]
	plan, hasPlan := context[path.Join(id.ContextRoot, "implementation-plan.md")]
	delivery := d.String("delivery_status")
	if delivery == "" {
		add(d.Path, "lifecycle.delivery_status_missing")
	}
	if (hasDesign || hasPlan) && d.String("status") != "active" {
		add(d.Path, "lifecycle.plan_brief_not_active")
	}
	decision, valid := DesignDecision(d)
	if legacy {
		decision, valid = legacyDesignDecision(string(d.Raw))
	}
	if (hasDesign || hasPlan) && !valid {
		add(d.Path, "lifecycle.design_requirement_decision_invalid")
	}
	if hasDesign && valid && decision == "no" {
		add(design.Path, "lifecycle.design_present_when_not_required")
	}
	if hasPlan && valid && decision == "yes" && !hasDesign {
		add(plan.Path, "lifecycle.plan_without_design")
	}
	if hasPlan && hasDesign && valid && decision == "yes" && design.String("status") != "active" {
		add(design.Path, "lifecycle.plan_design_not_active")
	}
	if delivery == "in_progress" && (!hasPlan || plan.String("status") != "active") {
		add("", "lifecycle.execution_plan_not_active")
	}
	if delivery == "done" && (!hasPlan || plan.String("status") != "archived") {
		add("", "lifecycle.done_plan_not_archived")
	}
	if delivery == "cancelled" && hasPlan && plan.String("status") != "archived" {
		add("", "lifecycle.cancelled_plan_not_archived")
	}
	return out
}
