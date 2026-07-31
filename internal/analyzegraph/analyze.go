package analyzegraph

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

const fanOutThreshold = 2

type node struct {
	NodeRef
	Path     string
	Declared bool
	Task     bool
	Weak     bool
}

type relation struct {
	Evidence
	MissingSource bool
}

type handoff struct {
	Kind          string
	SchemaVersion int
	Task          Task
	Nodes         []*node
	Relations     []*relation
}

// AnalyzeFile reads and analyses one JSON execution handoff. It never writes
// to the handoff or any other project file.
func AnalyzeFile(path string) (Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Report{}, fmt.Errorf("read handoff %q: %w", path, err)
	}
	return Analyze(path, data)
}

// Analyze analyses JSON data as an execution handoff. The graph is accepted
// from explicit nodes/relations. Known handoff sections are normalized into
// nodes, but no relation is inferred from a shared category or path.
func Analyze(path string, data []byte) (Report, error) {
	h, err := decodeHandoff(data)
	if err != nil {
		return Report{}, fmt.Errorf("parse handoff %q: %w", path, err)
	}
	return analyze(path, h), nil
}

func decodeHandoff(data []byte) (handoff, error) {
	var object map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&object); err != nil {
		return handoff{}, err
	}
	if object == nil {
		return handoff{}, errors.New("handoff must be a JSON object")
	}

	h := handoff{Kind: stringValue(object["kind"]), SchemaVersion: intValue(object["schema_version"])}
	if h.SchemaVersion == 0 {
		h.SchemaVersion = intValue(object["format_version"])
	}
	if h.Kind != "" && h.Kind != "execution_handoff" {
		return handoff{}, fmt.Errorf("unsupported kind %q", h.Kind)
	}
	h.Task = parseTask(object["task"])
	if h.Task.ID == "" && h.Task.Source == "" {
		h.Task = parseTask(object["starting_task"])
	}

	seenNodes := make(map[string]*node)
	addNode := func(candidate *node) {
		if candidate == nil {
			return
		}
		if existing, ok := seenNodes[candidate.ID]; ok {
			mergeNode(existing, candidate)
			return
		}
		seenNodes[candidate.ID] = candidate
		h.Nodes = append(h.Nodes, candidate)
	}

	if h.Task.ID != "" || h.Task.Source != "" {
		id := h.Task.ID
		if id == "" {
			id = "task"
		}
		addNode(&node{NodeRef: NodeRef{ID: "task:" + id, Type: "task", Label: id, Source: h.Task.Source}, Task: true})
	}

	knownNodeKeys := []string{
		"nodes", "declared_context", "declared_priming_context", "observed_context",
		"observed_execution_context", "documents", "decisions", "files", "commits",
		"verification", "verification_artifacts", "artifacts", "artefacts", "changed_files",
		"dependencies", "context", "graph", "execution_context_graph",
	}
	for _, key := range knownNodeKeys {
		parseNodeContainer(object[key], key, key == "declared_context" || key == "declared_priming_context", addNode)
	}

	var relations []*relation
	knownRelationKeys := []string{"relations", "edges"}
	for _, key := range knownRelationKeys {
		relations = append(relations, parseRelationContainer(object[key])...)
	}
	for _, key := range knownNodeKeys {
		collectNestedRelations(object[key], &relations)
		collectNestedEvidence(object[key], addNode, &relations)
	}
	for _, key := range []string{"evidence", "observed_evidence"} {
		parseEvidenceContainer(object[key], addNode, &relations)
	}
	h.Relations = relations

	if len(h.Nodes) == 0 && len(h.Relations) == 0 && h.Task.ID == "" && h.Task.Source == "" {
		return handoff{}, errors.New("handoff contains no explicit nodes, relations, or task")
	}
	return h, nil
}

func analyze(path string, h handoff) Report {
	report := Report{
		FormatVersion:        ReportFormatVersion,
		HandoffPath:          path,
		HandoffKind:          h.Kind,
		HandoffSchemaVersion: h.SchemaVersion,
		Task:                 h.Task,
		Findings:             make([]Finding, 0),
		Recommendations:      make([]Recommendation, 0),
		Evidence:             make([]Evidence, 0, len(h.Nodes)+len(h.Relations)),
	}

	nodes := make(map[string]*node, len(h.Nodes))
	aliases := make(map[string]string, len(h.Nodes)*3)
	for _, item := range h.Nodes {
		nodes[item.ID] = item
		for _, alias := range []string{item.ID, item.Label, item.Path} {
			if alias != "" {
				aliases[alias] = item.ID
			}
		}
		report.Evidence = append(report.Evidence, Evidence{Kind: "node", Node: item.ID, Type: item.Type, Source: item.Source, Sources: item.Sources, Weak: item.Weak || item.Source == ""})
		if item.Source == "" {
			report.Findings = append(report.Findings, Finding{
				Kind:          "weak_evidence",
				Severity:      "warning",
				Subject:       item.ID,
				Message:       "node has no direct source reference",
				Nodes:         []string{item.ID},
				RelationTypes: []string{"node"},
				Sources:       []string{path},
				Evidence:      []Evidence{{Kind: "node", Node: item.ID, Type: item.Type, Source: "", Weak: true}},
			})
		}
	}

	resolvedRelations := make([]*relation, 0, len(h.Relations))
	outgoing := make(map[string]map[string][]int)
	coAccess := make(map[string][]int)
	for index, item := range h.Relations {
		from := resolveAlias(item.From, aliases)
		to := resolveAlias(item.To, aliases)
		evidence := item.Evidence
		evidence.From, evidence.To = from, to
		if evidence.Count < 1 {
			evidence.Count = 1
		}
		item.Evidence = evidence
		evidenceIndex := len(report.Evidence)
		report.Evidence = append(report.Evidence, evidence)

		if item.Type == "" || from == "" || to == "" || nodes[from] == nil || nodes[to] == nil {
			nodesInFinding := compactStrings([]string{from, to})
			report.Findings = append(report.Findings, Finding{
				Kind:          "unresolved_link",
				Severity:      "warning",
				Subject:       relationSubject(index, item),
				Message:       unresolvedMessage(item, from, to),
				Nodes:         nodesInFinding,
				RelationTypes: []string{relationType(item.Type)},
				Sources:       findingSources(item.Source, item.Sources, path),
				Evidence:      []Evidence{evidence},
			})
			continue
		}
		item.Evidence.From, item.Evidence.To = from, to
		resolvedRelations = append(resolvedRelations, item)
		if outgoing[from] == nil {
			outgoing[from] = make(map[string][]int)
		}
		outgoing[from][to] = append(outgoing[from][to], evidenceIndex)
		if isCoAccessType(item.Type) {
			coAccess[pairKey(from, to)] = append(coAccess[pairKey(from, to)], evidenceIndex)
		}
		if item.Source == "" && len(item.Sources) == 0 || item.Weak || item.MissingSource {
			report.Findings = append(report.Findings, Finding{
				Kind:          "weak_evidence",
				Severity:      "warning",
				Subject:       relationSubject(index, item),
				Message:       "relation is explicit but has weak or missing provenance",
				Nodes:         []string{from, to},
				RelationTypes: []string{item.Type},
				Sources:       findingSources(item.Source, item.Sources, path),
				Evidence:      []Evidence{evidence},
			})
		}
	}

	for _, item := range h.Nodes {
		if targets := outgoing[item.ID]; len(targets) >= fanOutThreshold {
			targetIDs := make([]string, 0, len(targets))
			evidence := make([]Evidence, 0)
			for target, indexes := range targets {
				targetIDs = append(targetIDs, target)
				for _, index := range indexes {
					evidence = append(evidence, report.Evidence[index])
				}
			}
			sort.Strings(targetIDs)
			report.Findings = append(report.Findings, Finding{
				Kind:          "high_fan_out",
				Severity:      "info",
				Subject:       item.ID,
				Message:       fmt.Sprintf("node points to %d distinct nodes; review it as a context hub", len(targets)),
				Nodes:         append([]string{item.ID}, targetIDs...),
				RelationTypes: relationTypes(evidence),
				Sources:       sources(evidence, path),
				Evidence:      evidence,
			})
		}
	}

	keys := make([]string, 0, len(coAccess))
	for key := range coAccess {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		indexes := coAccess[key]
		count := 0
		evidence := make([]Evidence, 0, len(indexes))
		for _, index := range indexes {
			count += report.Evidence[index].Count
			evidence = append(evidence, report.Evidence[index])
		}
		if count < 2 {
			continue
		}
		parts := strings.Split(key, "\x00")
		report.Findings = append(report.Findings, Finding{
			Kind:          "repeated_co_access",
			Severity:      "info",
			Subject:       key,
			Message:       fmt.Sprintf("nodes are explicitly co-accessed %d times", count),
			Nodes:         parts,
			RelationTypes: relationTypes(evidence),
			Sources:       sources(evidence, path),
			Evidence:      evidence,
		})
	}

	if recommendation := buildRecommendation(path, h, nodes, resolvedRelations); recommendation != nil {
		report.Recommendations = append(report.Recommendations, *recommendation)
	}
	return report
}

func buildRecommendation(path string, h handoff, nodes map[string]*node, relations []*relation) *Recommendation {
	seed := make(map[string]bool)
	for _, item := range h.Nodes {
		if item.Declared || item.Task {
			seed[item.ID] = true
		}
	}
	if len(seed) == 0 {
		for _, item := range h.Nodes {
			seed[item.ID] = true
			break
		}
	}
	if len(seed) == 0 {
		return nil
	}

	selected := make(map[string]bool, len(seed))
	for id := range seed {
		selected[id] = true
	}
	for _, item := range relations {
		if selected[item.From] || selected[item.To] {
			selected[item.From] = true
			selected[item.To] = true
		}
	}

	selectedIDs := make([]string, 0, len(selected))
	for id := range selected {
		if nodes[id] != nil {
			selectedIDs = append(selectedIDs, id)
		}
	}
	sort.Strings(selectedIDs)
	if len(selectedIDs) == 0 {
		return nil
	}

	contributingNodes := make([]NodeRef, 0, len(selectedIDs))
	for _, id := range selectedIDs {
		contributingNodes = append(contributingNodes, nodes[id].NodeRef)
	}
	evidence := make([]Evidence, 0)
	for _, item := range h.Nodes {
		if !selected[item.ID] {
			continue
		}
		if item.Declared {
			evidence = append(evidence, Evidence{Kind: "declaration", Node: item.ID, Type: "declared_context", Source: item.Source, Sources: item.Sources, Weak: item.Source == ""})
		}
		if item.Task {
			evidence = append(evidence, Evidence{Kind: "declaration", Node: item.ID, Type: "task_owner", Source: item.Source, Sources: item.Sources, Weak: item.Source == ""})
		}
	}
	for _, item := range relations {
		if selected[item.From] && selected[item.To] {
			evidence = append(evidence, item.Evidence)
		}
	}
	if len(evidence) == 0 {
		evidence = append(evidence, Evidence{Kind: "declaration", Type: "context_selection", Source: path, Count: 1})
	}
	return &Recommendation{
		Kind:              "context_set",
		Summary:           "Review this advisory context set before continuing the task; selection is based only on the listed typed evidence.",
		ContributingNodes: contributingNodes,
		RelationTypes:     relationTypes(evidence),
		Sources:           sources(evidence, path),
		Evidence:          evidence,
	}
}

func parseTask(raw json.RawMessage) Task {
	var object map[string]json.RawMessage
	if !unmarshalObject(raw, &object) {
		return Task{ID: stringValue(raw), Source: stringValue(raw)}
	}
	return Task{ID: firstString(object, "id", "key", "name", "task"), Source: firstString(object, "source", "source_ref", "source_path", "path", "owner", "primary_source")}
}

func parseNodeContainer(raw json.RawMessage, category string, declared bool, add func(*node)) {
	if len(raw) == 0 || string(raw) == "null" {
		return
	}
	var list []json.RawMessage
	if unmarshalArray(raw, &list) {
		for index, item := range list {
			candidate := parseNode(item, category, index, declared)
			add(candidate)
		}
		return
	}
	var object map[string]json.RawMessage
	if !unmarshalObject(raw, &object) {
		add(parseNode(raw, category, 0, declared))
		return
	}
	for _, key := range []string{"nodes", "documents", "decisions", "files", "commits", "verification", "verification_artifacts", "artifacts", "artefacts", "changed_files", "dependencies", "context", "graph", "execution_context_graph"} {
		if object[key] != nil {
			parseNodeContainer(object[key], key, declared || key == "declared_context", add)
		}
	}
	if object["id"] != nil || object["path"] != nil || object["label"] != nil || object["name"] != nil || object["source"] != nil {
		add(parseNode(raw, category, 0, declared))
	}
}

func collectNestedRelations(raw json.RawMessage, result *[]*relation) {
	if len(raw) == 0 || string(raw) == "null" {
		return
	}
	var object map[string]json.RawMessage
	if unmarshalObject(raw, &object) {
		for _, key := range []string{"relations", "edges"} {
			*result = append(*result, parseRelationContainer(object[key])...)
		}
		for _, key := range []string{"nodes", "declared_context", "declared_priming_context", "observed_context", "observed_execution_context", "documents", "decisions", "files", "commits", "verification", "verification_artifacts", "artifacts", "artefacts", "changed_files", "dependencies", "context", "graph", "execution_context_graph"} {
			collectNestedRelations(object[key], result)
		}
		return
	}
	var list []json.RawMessage
	if unmarshalArray(raw, &list) {
		for _, item := range list {
			collectNestedRelations(item, result)
		}
	}
}

func parseEvidenceContainer(raw json.RawMessage, addNode func(*node), relations *[]*relation) {
	if len(raw) == 0 || string(raw) == "null" {
		return
	}
	var list []json.RawMessage
	if unmarshalArray(raw, &list) {
		for index, item := range list {
			var object map[string]json.RawMessage
			if unmarshalObject(item, &object) && isRelationObject(object) {
				*relations = append(*relations, parseRelation(item))
				continue
			}
			category := "evidence"
			if unmarshalObject(item, &object) {
				if value := firstString(object, "type", "kind", "category"); value != "" {
					category = value
				}
			}
			addNode(parseNode(item, category, index, false))
		}
		return
	}
	var object map[string]json.RawMessage
	if !unmarshalObject(raw, &object) {
		return
	}
	for _, key := range []string{"evidence", "observed_evidence"} {
		parseEvidenceContainer(object[key], addNode, relations)
	}
	for _, key := range []string{"relations", "edges"} {
		*relations = append(*relations, parseRelationContainer(object[key])...)
	}
}

func collectNestedEvidence(raw json.RawMessage, addNode func(*node), relations *[]*relation) {
	if len(raw) == 0 || string(raw) == "null" {
		return
	}
	var object map[string]json.RawMessage
	if unmarshalObject(raw, &object) {
		for _, key := range []string{"evidence", "observed_evidence"} {
			parseEvidenceContainer(object[key], addNode, relations)
		}
		for _, key := range []string{"nodes", "declared_context", "declared_priming_context", "observed_context", "observed_execution_context", "documents", "decisions", "files", "commits", "verification", "verification_artifacts", "artifacts", "artefacts", "changed_files", "dependencies", "context", "graph", "execution_context_graph"} {
			collectNestedEvidence(object[key], addNode, relations)
		}
		return
	}
	var list []json.RawMessage
	if unmarshalArray(raw, &list) {
		for _, item := range list {
			collectNestedEvidence(item, addNode, relations)
		}
	}
}

func isRelationObject(object map[string]json.RawMessage) bool {
	hasFrom := firstString(object, "from", "from_id", "source_node", "source_node_id", "subject") != ""
	hasTo := firstString(object, "to", "to_id", "target", "target_id", "target_node", "destination") != ""
	hasGraphEndpoints := stringValue(object["source"]) != "" && stringValue(object["target"]) != ""
	return (hasFrom && hasTo) || hasGraphEndpoints
}

func parseRelationContainer(raw json.RawMessage) []*relation {
	var list []json.RawMessage
	if !unmarshalArray(raw, &list) {
		return nil
	}
	result := make([]*relation, 0, len(list))
	for _, item := range list {
		result = append(result, parseRelation(item))
	}
	return result
}

func parseNode(raw json.RawMessage, category string, index int, declared bool) *node {
	if text := stringValue(raw); text != "" {
		return &node{NodeRef: NodeRef{ID: category + ":" + text, Type: category, Label: text, Source: text}, Declared: declared}
	}
	var object map[string]json.RawMessage
	if !unmarshalObject(raw, &object) {
		return &node{NodeRef: NodeRef{ID: fmt.Sprintf("%s:%d", category, index), Type: category}, Declared: declared, Weak: true}
	}
	id := firstString(object, "id", "key", "node_id")
	label := firstString(object, "label", "name", "title", "path", "file", "sha", "id")
	path := firstString(object, "path", "file")
	source := firstString(object, "source", "source_ref", "source_path", "ref")
	if id == "" {
		id = label
	}
	if id == "" {
		id = fmt.Sprintf("%s:%d", category, index)
	}
	if label == "" {
		label = id
	}
	nodeType := firstString(object, "type", "kind", "category")
	if nodeType == "" {
		nodeType = category
	}
	sources := stringList(object["sources"])
	if source == "" && len(sources) > 0 {
		source = sources[0]
	}
	return &node{NodeRef: NodeRef{ID: id, Type: nodeType, Label: label, Source: source, Sources: sources}, Path: path, Declared: declared, Weak: boolValue(object["weak"]) || strings.EqualFold(firstString(object, "strength", "confidence"), "weak")}
}

func parseRelation(raw json.RawMessage) *relation {
	var object map[string]json.RawMessage
	if !unmarshalObject(raw, &object) {
		return &relation{Evidence: Evidence{Kind: "relation", Type: ""}, MissingSource: true}
	}
	from := firstString(object, "from", "from_id", "source_node", "source_node_id", "subject")
	to := firstString(object, "to", "to_id", "target", "target_id", "target_node", "destination")
	if from == "" && stringValue(object["source"]) != "" && stringValue(object["target"]) != "" {
		from = stringValue(object["source"])
	}
	typeName := firstString(object, "type", "relation_type", "relation", "kind")
	source := firstString(object, "source_ref", "evidence_source", "source_path", "provenance", "evidence")
	if source == "" && from != stringValue(object["source"]) {
		source = stringValue(object["source"])
	}
	sources := stringList(object["sources"])
	if source == "" && len(sources) > 0 {
		source = sources[0]
	}
	count := intValue(object["count"])
	if count == 0 {
		count = intValue(object["occurrences"])
	}
	if count == 0 {
		count = intValue(object["frequency"])
	}
	weak := boolValue(object["weak"]) || strings.EqualFold(firstString(object, "strength", "confidence"), "weak")
	return &relation{Evidence: Evidence{Kind: "relation", From: from, To: to, Type: typeName, Source: source, Sources: sources, Count: count, Weak: weak}, MissingSource: source == "" && len(sources) == 0}
}

func mergeNode(target, source *node) {
	if target.Type == "" {
		target.Type = source.Type
	}
	if target.Label == "" {
		target.Label = source.Label
	}
	if target.Path == "" {
		target.Path = source.Path
	}
	if target.Source == "" {
		target.Source = source.Source
	}
	target.Sources = uniqueStrings(append(target.Sources, source.Sources...))
	target.Declared = target.Declared || source.Declared
	target.Task = target.Task || source.Task
	target.Weak = target.Weak || source.Weak
}

func resolveAlias(value string, aliases map[string]string) string {
	if value == "" {
		return ""
	}
	if resolved, ok := aliases[value]; ok {
		return resolved
	}
	return value
}

func relationSubject(index int, item *relation) string {
	if item.From != "" || item.To != "" {
		return strings.Trim(strings.Join([]string{item.From, item.To}, " -> "), " ->")
	}
	return "relation:" + strconv.Itoa(index)
}

func unresolvedMessage(item *relation, from, to string) string {
	if item.Type == "" {
		return "relation has no explicit relation type"
	}
	if from == "" || to == "" {
		return "relation has an empty endpoint"
	}
	return fmt.Sprintf("relation target is not present in the handoff nodes: %s -> %s", from, to)
}

func relationType(value string) string {
	if value == "" {
		return "untyped"
	}
	return value
}

func isCoAccessType(value string) bool {
	switch strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(value, "-", "_"), " ", "_")) {
	case "co_access", "coaccess", "observed_together", "shared_execution":
		return true
	default:
		return false
	}
}

func pairKey(left, right string) string {
	if left > right {
		left, right = right, left
	}
	return left + "\x00" + right
}

func relationTypes(evidence []Evidence) []string {
	result := make([]string, 0, len(evidence))
	for _, item := range evidence {
		if item.Type != "" {
			result = append(result, item.Type)
		}
	}
	return uniqueStrings(result)
}

func sources(evidence []Evidence, fallback string) []string {
	result := make([]string, 0)
	for _, item := range evidence {
		result = append(result, item.Source)
		result = append(result, item.Sources...)
	}
	result = uniqueStrings(result)
	if len(result) == 0 && fallback != "" {
		return []string{fallback}
	}
	return result
}

func findingSources(primary string, additional []string, fallback string) []string {
	return sources([]Evidence{{Source: primary, Sources: additional}}, fallback)
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			result = append(result, value)
		}
	}
	return uniqueStrings(result)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func firstString(object map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		if value := stringValue(object[key]); value != "" {
			return value
		}
	}
	return ""
}

func stringValue(raw json.RawMessage) string {
	var value string
	if json.Unmarshal(raw, &value) == nil {
		return strings.TrimSpace(value)
	}
	return ""
}

func stringList(raw json.RawMessage) []string {
	var values []string
	if json.Unmarshal(raw, &values) != nil {
		return nil
	}
	return compactStrings(values)
}

func intValue(raw json.RawMessage) int {
	var value int
	if json.Unmarshal(raw, &value) == nil {
		return value
	}
	return 0
}

func boolValue(raw json.RawMessage) bool {
	var value bool
	return json.Unmarshal(raw, &value) == nil && value
}

func unmarshalObject(raw json.RawMessage, target *map[string]json.RawMessage) bool {
	if len(raw) == 0 || string(raw) == "null" {
		return false
	}
	return json.Unmarshal(raw, target) == nil && *target != nil
}

func unmarshalArray(raw json.RawMessage, target *[]json.RawMessage) bool {
	if len(raw) == 0 || string(raw) == "null" {
		return false
	}
	return json.Unmarshal(raw, target) == nil
}

// PrintMarkdown writes the advisory report in a stable, source-oriented
// format. It intentionally includes relation types and sources next to every
// finding and recommendation.
func PrintMarkdown(writer io.Writer, report Report) {
	fmt.Fprintln(writer, "# Execution Context Graph Analysis")
	fmt.Fprintf(writer, "\nHandoff: `%s`\n", report.HandoffPath)
	if report.HandoffKind != "" {
		fmt.Fprintf(writer, "Kind: `%s`\n", report.HandoffKind)
	}
	if report.HandoffSchemaVersion > 0 {
		fmt.Fprintf(writer, "Schema version: `%d`\n", report.HandoffSchemaVersion)
	}

	fmt.Fprintln(writer, "\n## Recommendations")
	if len(report.Recommendations) == 0 {
		fmt.Fprintln(writer, "\nNo context recommendation could be formed from the handoff.")
	} else {
		for _, item := range report.Recommendations {
			fmt.Fprintf(writer, "\n### %s\n\n%s\n", item.Kind, item.Summary)
			fmt.Fprintf(writer, "- Contributing nodes: %s\n", formatNodes(item.ContributingNodes))
			fmt.Fprintf(writer, "- Relation types: %s\n", formatStrings(item.RelationTypes))
			fmt.Fprintf(writer, "- Sources: %s\n", formatStrings(item.Sources))
			printEvidence(writer, item.Evidence)
		}
	}

	fmt.Fprintln(writer, "\n## Findings")
	if len(report.Findings) == 0 {
		fmt.Fprintln(writer, "\nNo findings.")
	} else {
		for _, item := range report.Findings {
			fmt.Fprintf(writer, "\n- **%s** `%s`: %s\n", item.Severity, item.Kind, item.Message)
			fmt.Fprintf(writer, "  - Subject: `%s`\n  - Nodes: %s\n  - Relation types: %s\n  - Sources: %s\n", item.Subject, formatStrings(item.Nodes), formatStrings(item.RelationTypes), formatStrings(item.Sources))
		}
	}

	fmt.Fprintln(writer, "\n## Evidence")
	printEvidence(writer, report.Evidence)
}

func printEvidence(writer io.Writer, evidence []Evidence) {
	if len(evidence) == 0 {
		fmt.Fprintln(writer, "\nNo evidence.")
		return
	}
	for _, item := range evidence {
		description := item.Node
		if item.From != "" || item.To != "" {
			description = item.From + " -> " + item.To
		}
		if description == "" {
			description = "(selection)"
		}
		fmt.Fprintf(writer, "\n- `%s` `%s` (%s); source: %s", item.Kind, item.Type, description, formatSource(item.Source, item.Sources))
		if item.Count > 1 {
			fmt.Fprintf(writer, "; count: %d", item.Count)
		}
		if item.Weak {
			fmt.Fprint(writer, "; weak")
		}
		fmt.Fprintln(writer)
	}
}

func formatNodes(nodes []NodeRef) string {
	values := make([]string, 0, len(nodes))
	for _, item := range nodes {
		label := item.ID
		if item.Label != "" && item.Label != item.ID {
			label += " (" + item.Label + ")"
		}
		values = append(values, "`"+label+"`")
	}
	return formatStrings(values)
}

func formatStrings(values []string) string {
	if len(values) == 0 {
		return "(none)"
	}
	return strings.Join(values, ", ")
}

func formatSource(source string, additional []string) string {
	values := compactStrings(append([]string{source}, additional...))
	if len(values) == 0 {
		return "(missing)"
	}
	return strings.Join(values, ", ")
}
